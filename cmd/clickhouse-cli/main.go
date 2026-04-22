package main

import (
	"context"
	"os"

	"github.com/pinax-network/clickhouse-cli/pkg/log"

	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
)

func init() {
	err := log.InitializeGlobalLogger(os.Getenv("DEBUG") != "")
	if err != nil {
		log.Fatal("failed to initialize logger", zap.Error(err))
	}
}

func main() {

	cmd := &cli.Command{
		Name:  "clickhouse-cli",
		Usage: "CLI tool for ClickHouse operations",
		Commands: []*cli.Command{
			migrateCmd,
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "node",
				Aliases: []string{"n"},
				Usage:   "ClickHouse node to execute commands against. Should use the port for the native protocol.",
				Value:   "localhost:9000",
				Sources: cli.EnvVars("CLICKHOUSE_NODE"),
			}, &cli.StringFlag{
				Name:    "user",
				Aliases: []string{"u"},
				Usage:   "ClickHouse user to execute commands as.",
				Value:   "default",
				Sources: cli.EnvVars("CLICKHOUSE_USER"),
			}, &cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				Usage:   "Password for the ClickHouse user.",
				Value:   "",
				Sources: cli.EnvVars("CLICKHOUSE_PASSWORD"),
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal("application error", zap.Error(err))
	}
}
