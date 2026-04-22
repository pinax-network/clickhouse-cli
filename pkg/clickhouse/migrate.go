package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pinax-network/clickhouse-cli/pkg/log"

	"go.uber.org/zap"
)

const (
	schemaMigrationsDatabaseDDLCluster = "CREATE DATABASE `%s` ON CLUSTER '{cluster}' ENGINE=Replicated('/clickhouse/databases/{database}', '{shard}', '{replica}')"
	schemaMigrationsTableDDLCluster    = "CREATE TABLE `%s`.`%s` (version UInt32, dirty Bool, sequence DateTime64) ENGINE = ReplicatedMergeTree ORDER BY sequence"

	schemaMigrationsDatabaseDDLSingle = "CREATE DATABASE `%s`"
	schemaMigrationsTableDDLSingle    = "CREATE TABLE `%s`.`%s` (version UInt32, dirty Bool, sequence DateTime64) ENGINE = MergeTree ORDER BY sequence"
)

// MigrationFilePattern matches schema migration files, which need to be named after <sequential_number>_*.sql.
//
// Note: migration files are split on ';' before execution, which means semicolons inside string literals,
// comments, or compound statements (e.g. CREATE FUNCTION) are not supported. Write one statement per
// delimiter or keep literals semicolon-free.
var MigrationFilePattern = regexp.MustCompile(`^(\d+)_.*\.sql$`)

// identifierPattern restricts database and table names to a safe character set. Required because
// Clickhouse parameter placeholders cannot be used for identifiers, so names are interpolated
// directly into DDL.
var identifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

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
	schemaFS    fs.FS
	database    string
	table       string
	createTable bool
	clusterMode bool
}

func NewMigration(chClient *Client, schemaFS fs.FS, table string, createTable, clusterMode bool) (*Migration, error) {

	tableSplit := strings.Split(table, ".")
	if len(tableSplit) != 2 {
		return nil, fmt.Errorf("invalid table name expecting format <database>.<table>, got: %s", table)
	}

	database, tableName := tableSplit[0], tableSplit[1]
	if !identifierPattern.MatchString(database) {
		return nil, fmt.Errorf("invalid database identifier: %q", database)
	}
	if !identifierPattern.MatchString(tableName) {
		return nil, fmt.Errorf("invalid table identifier: %q", tableName)
	}

	return &Migration{
		chClient:    chClient,
		schemaFS:    schemaFS,
		database:    database,
		table:       tableName,
		createTable: createTable,
		clusterMode: clusterMode,
	}, nil
}

func (m *Migration) Run(ctx context.Context) error {

	if err := m.checkMigrationTable(ctx); err != nil {
		return err
	}

	migrationFiles, err := m.parseMigrationDirectory()
	if err != nil {
		return err
	}

	latestVersion, isDirty, err := m.fetchLatestVersion(ctx)
	if err != nil {
		return err
	}

	if isDirty {
		return errors.New("schema migrations are in a dirty state, won't execute any migrations")
	}

	for _, mf := range migrationFiles {
		migration, err := fs.ReadFile(m.schemaFS, mf.File)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", mf.File, err)
		}

		if mf.Seq <= latestVersion {
			log.Info("skipping migration file as it's already applied", zap.String("file", mf.File))
			continue
		}

		for query := range strings.SplitSeq(string(migration), ";") {
			query = strings.TrimSpace(query)
			if query == "" {
				continue
			}

			if err := m.chClient.Execute(ctx, query, nil); err != nil {
				log.Error("failed to execute migration", zap.String("query", query), zap.Error(err))
				if updateErr := m.updateTable(ctx, mf.Seq, true); updateErr != nil {
					return fmt.Errorf("failed to update migration table: %w", updateErr)
				}
				return fmt.Errorf("failed to execute migration: %w", err)
			}
		}
		log.Info("successfully applied migration", zap.String("file", mf.File))

		if err := m.updateTable(ctx, mf.Seq, false); err != nil {
			return fmt.Errorf("failed to update migration table: %w", err)
		}
	}

	return nil
}

func (m *Migration) updateTable(ctx context.Context, seq int, dirty bool) error {
	query := fmt.Sprintf("INSERT INTO `%s`.`%s` VALUES ({seq:UInt32}, {dirty:Bool}, now())", m.database, m.table)
	return m.chClient.Execute(ctx, query, map[string]string{
		"seq":   strconv.Itoa(seq),
		"dirty": strconv.FormatBool(dirty),
	})
}

func (m *Migration) fetchLatestVersion(ctx context.Context) (version int, dirty bool, err error) {
	var res MigrationRow
	query := fmt.Sprintf("SELECT * FROM `%s`.`%s` ORDER BY version DESC, sequence DESC LIMIT 1", m.database, m.table)
	if err := m.chClient.QueryStruct(ctx, query, nil, &res); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("failed to fetch latest migration sequence: %w", err)
	}

	return int(res.Version), res.Dirty, nil
}

// parseMigrationDirectory reads the schema directory and returns the migrations in sequence order.
// An empty directory (no matching files) returns an empty slice, not an error.
func (m *Migration) parseMigrationDirectory() ([]MigrationFile, error) {
	return parseMigrationDirectory(m.schemaFS)
}

func parseMigrationDirectory(schemaFS fs.FS) ([]MigrationFile, error) {

	files, err := fs.ReadDir(schemaFS, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrationFiles []MigrationFile

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		matches := MigrationFilePattern.FindStringSubmatch(fileName)
		if matches == nil {
			continue
		}

		seqNum, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse sequence number from file %s: %w", fileName, err)
		}

		migrationFiles = append(migrationFiles, MigrationFile{File: fileName, Seq: seqNum})
	}

	sort.Slice(migrationFiles, func(i, j int) bool {
		return migrationFiles[i].Seq < migrationFiles[j].Seq
	})

	for i, mf := range migrationFiles {
		expected := i + 1
		if mf.Seq == expected {
			continue
		}
		if i > 0 && mf.Seq == migrationFiles[i-1].Seq {
			return nil, fmt.Errorf("invalid migration files, duplicate sequence number: %d", mf.Seq)
		}
		return nil, fmt.Errorf("missing sequence number: expected %d, but found %d", expected, mf.Seq)
	}

	return migrationFiles, nil
}

func (m *Migration) checkMigrationTable(ctx context.Context) error {

	dbExists, err := m.chClient.databaseExists(ctx, m.database)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}
	if !dbExists {
		if !m.createTable {
			return fmt.Errorf("database %q for the schema migrations doesn't exist and should not be created", m.database)
		}
		log.Info("creating database for schema_migrations", zap.String("database", m.database))
		if err := m.chClient.Execute(ctx, fmt.Sprintf(m.databaseDDL(), m.database), nil); err != nil {
			return fmt.Errorf("failed to create database for schema migrations: %w", err)
		}
	}

	tableExists, err := m.chClient.tableExists(ctx, m.database, m.table)
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	}
	if !tableExists {
		if !m.createTable {
			return fmt.Errorf("table '%s.%s' for the schema migrations doesn't exist and should not be created", m.database, m.table)
		}
		log.Info("creating table for schema_migrations", zap.String("database", m.database), zap.String("table", m.table))
		if err := m.chClient.Execute(ctx, fmt.Sprintf(m.tableDDL(), m.database, m.table), nil); err != nil {
			return fmt.Errorf("failed to create table for schema migrations: %w", err)
		}
	}

	return nil
}

func (m *Migration) databaseDDL() string {
	if m.clusterMode {
		return schemaMigrationsDatabaseDDLCluster
	}
	return schemaMigrationsDatabaseDDLSingle
}

func (m *Migration) tableDDL() string {
	if m.clusterMode {
		return schemaMigrationsTableDDLCluster
	}
	return schemaMigrationsTableDDLSingle
}
