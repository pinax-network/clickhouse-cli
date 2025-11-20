package clickhouse

import (
	"clickhouse-cli/pkg/log"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	SchemaMigrationsDatabaseDDL = "CREATE DATABASE `%s` ON CLUSTER '{cluster}' ENGINE=Replicated('/clickhouse/databases/{database}', '{shard}', '{replica}')"
	SchemaMigrationsTableDDL    = "CREATE TABLE `%s`.`%s` (version UInt32, dirty Bool, sequence DateTime64) ENGINE = ReplicatedMergeTree ORDER BY sequence"
)

// MigrationFilePattern matches schema migration files, which need to be named after <sequential_number>_*.sql
var MigrationFilePattern = regexp.MustCompile(`^(\d+)_.*\.sql$`)

type MigrationFile struct {
	File string
	Seq  int
}

type MigrationRow struct {
	Version  uint32    `ch:"version"`
	Dirty    bool      `ch:"dirty"`
	Sequence time.Time `ch:"sequence"`
}

type Migration struct {
	chClient    *Client
	schemaDir   string
	database    string
	table       string
	createTable bool
}

func NewMigration(chClient *Client, schemaDir, table string, createTable bool) (*Migration, error) {

	tableSplit := strings.Split(table, ".")
	if len(tableSplit) != 2 {
		return nil, fmt.Errorf("invalid table name expecting format <database>.<table>, got: %s", table)
	}

	return &Migration{
		chClient:    chClient,
		schemaDir:   schemaDir,
		database:    tableSplit[0],
		table:       tableSplit[1],
		createTable: createTable,
	}, nil
}

func (c *Migration) Run(ctx context.Context) error {

	if err := c.checkMigrationTable(ctx); err != nil {
		return err
	}

	migrationFiles, err := c.parseMigrationDirectory()
	if err != nil {
		return err
	}

	latestVersion, isDirty, err := c.fetchLatestVersion(ctx)
	if err != nil {
		return err
	}

	if isDirty {
		return errors.New("schema migrations are in a dirty state, won't execute any migrations")
	}

	for _, m := range migrationFiles {
		migration, err := os.ReadFile(path.Join(c.schemaDir, m.File))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", m.File, err)
		}

		if m.Seq <= latestVersion {
			log.Info("skipping migration file as it's already applied", zap.String("file", m.File))
			continue
		}

		queries := strings.Split(string(migration), ";")
		for _, query := range queries {
			query = strings.TrimSpace(query)
			if query == "" {
				continue
			}

			if err := c.chClient.execute(ctx, query, nil); err != nil {
				log.Error("failed to execute migration", zap.String("query", query), zap.Error(err))
				if err := c.updateTable(ctx, m.Seq, true); err != nil {
					return fmt.Errorf("failed to update migration table: %w", err)
				}
				return fmt.Errorf("failed to execute migration: %w", err)
			}
		}
		log.Info("successfully applied migration", zap.String("file", m.File))

		if err := c.updateTable(ctx, m.Seq, false); err != nil {
			return fmt.Errorf("failed to update migration table: %w", err)
		}
	}

	return nil
}

func (c *Migration) updateTable(ctx context.Context, seq int, dirty bool) error {
	query := fmt.Sprintf("INSERT INTO `%s`.`%s` VALUES ({seq:UInt32}, {dirty:Bool}, now())", c.database, c.table)
	return c.chClient.execute(ctx, query, map[string]string{"seq": fmt.Sprintf("%d", seq), "dirty": fmt.Sprintf("%t", dirty)})
}

func (c *Migration) fetchLatestVersion(ctx context.Context) (version int, dirty bool, err error) {
	var res MigrationRow
	err = c.chClient.queryStruct(ctx, fmt.Sprintf("SELECT * FROM `%s`.`%s` ORDER BY version DESC LIMIT 1", c.database, c.table), nil, &res)
	if err != nil {
		return 0, false, fmt.Errorf("failed to fetch latest migration sequence: %w", err)
	}

	return int(res.Version), res.Dirty, nil
}

func (c *Migration) parseMigrationDirectory() ([]MigrationFile, error) {

	files, err := os.ReadDir(c.schemaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory %s: %w", c.schemaDir, err)
	}

	var migrationFiles []MigrationFile

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		matches := MigrationFilePattern.FindStringSubmatch(fileName)
		if matches == nil {
			// Skip non-migration files (could be other files in the directory)
			continue
		}

		seqNum, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse sequence number from file %s: %w", fileName, err)
		}

		migrationFiles = append(migrationFiles, MigrationFile{File: fileName, Seq: seqNum})
	}

	if len(migrationFiles) == 0 {
		return nil, fmt.Errorf("no migration files found in directory %s", c.schemaDir)
	}

	// Sort migration files by sequence number
	sort.Slice(migrationFiles, func(i, j int) bool {
		return migrationFiles[i].Seq < migrationFiles[j].Seq
	})

	// Check for duplicates, gaps and missing sequence numbers in one loop
	for i, mf := range migrationFiles {
		expected := i + 1
		if mf.Seq != expected {
			if i > 0 && mf.Seq == migrationFiles[i-1].Seq {
				return nil, fmt.Errorf("invalid migration files, duplicate sequence number: %d", mf.Seq)
			} else {
				return nil, fmt.Errorf("missing sequence number: expected %d, but found %d", expected, mf.Seq)
			}
		}
	}

	return migrationFiles, nil
}

func (c *Migration) checkMigrationTable(ctx context.Context) error {

	if dbExists, err := c.chClient.databaseExists(ctx, c.database); err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	} else if !dbExists {
		if !c.createTable {
			return fmt.Errorf("database %q for the schema migrations doesn't exist and should not be created", c.database)
		}

		log.Info("creating database for schema_migrations", zap.String("database", c.database))
		query := fmt.Sprintf(SchemaMigrationsDatabaseDDL, c.database)
		err = c.chClient.execute(ctx, query, nil)
		if err != nil {
			return fmt.Errorf("failed to create database for schema migrations: %w", err)
		}
	}

	if tableExists, err := c.chClient.tableExists(ctx, c.database, c.table); err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	} else if !tableExists {
		if !c.createTable {
			return fmt.Errorf("table '%s.%s' for the schema migrations doesn't exist and should not be created", c.database, c.table)
		}

		log.Info("creating table for schema_migrations", zap.String("database", c.database), zap.String("table", c.table))
		query := fmt.Sprintf(SchemaMigrationsTableDDL, c.database, c.table)
		err = c.chClient.execute(ctx, query, nil)
		if err != nil {
			return fmt.Errorf("failed to create table for schema migrations: %w", err)
		}
	}

	return nil
}
