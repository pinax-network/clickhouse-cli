package main

import (
	"context"

	"github.com/pinax-network/clickhouse-cli/pkg/clickhouse"

	"github.com/urfave/cli/v3"
)

var migrateCmd = &cli.Command{
	Name:      "migrate",
	Usage:     "Run Clickhouse migrations.",
	ArgsUsage: "<migration directory>",
	Action:    runMigrate,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "schema-table",
			Usage:   "Table containing schema migrations in the format of <database>.<table>.",
			Value:   "default.schema_migrations",
			Sources: cli.EnvVars("CLICKHOUSE_SCHEMA_TABLE"),
		},
		&cli.BoolFlag{
			Name:    "create-migrations-table",
			Usage:   "Create the schema migration table and database if it does not exist.",
			Value:   false,
			Sources: cli.EnvVars("CLICKHOUSE_CREATE_MIGRATIONS_TABLE"),
		},
		&cli.BoolFlag{
			Name:    "cluster-mode",
			Usage:   "Use Replicated database and ReplicatedMergeTree engines with ON CLUSTER. Disable for single-node setups.",
			Value:   true,
			Sources: cli.EnvVars("CLICKHOUSE_CLUSTER_MODE"),
		},
	},
}

func runMigrate(ctx context.Context, c *cli.Command) error {

	schemaDir := c.Args().First()
	if schemaDir == "" {
		return cli.Exit("migration directory argument is required (see `clickhouse-cli migrate --help`)", 1)
	}

	clickhouseClient, err := clickhouse.NewClient(ctx, c.String("node"), c.String("user"), c.String("password"), debugEnabled())
	if err != nil {
		return err
	}

	migration, err := clickhouse.NewMigration(
		clickhouseClient,
		schemaDir,
		c.String("schema-table"),
		c.Bool("create-migrations-table"),
		c.Bool("cluster-mode"),
	)
	if err != nil {
		return err
	}

	if err := migration.Run(ctx); err != nil {
		return err
	}

	return clickhouseClient.Close()
}
