package main

import (
	"clickhouse-cli/pkg/clickhouse"
	"context"

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
			Usage:   "Create the schema migration table and database if it does not exist. ",
			Value:   false,
			Sources: cli.EnvVars("CLICKHOUSE_CREATE_MIGRATIONS_TABLE"),
		},
	},
}

func runMigrate(ctx context.Context, c *cli.Command) error {

	clickhouseClient, err := clickhouse.NewClient(ctx, c.String("node"), c.String("user"), c.String("password"))
	if err != nil {
		return err
	}

	migration, err := clickhouse.NewMigration(clickhouseClient, c.Args().First(), c.String("schema-table"), c.Bool("create-migrations-table"))
	if err != nil {
		return err
	}

	err = migration.Run(ctx)
	if err != nil {
		return err
	}

	return clickhouseClient.Close()
}
