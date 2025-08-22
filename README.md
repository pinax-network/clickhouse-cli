# ClickHouse CLI

A Go-based command-line interface for managing ClickHouse databases, with a focus on schema migrations.

## Features

- **Connection Management**: Connect to ClickHouse servers using the native protocol
- **Schema Migrations**: Run SQL migrations in a controlled, sequential manner
- **Migration Tracking**: Automatically track applied migrations in a dedicated table
- **Error Handling**: Comprehensive error reporting and dirty state tracking for failed migrations

## Installation

```bash
# Install using Go
go install github.com/pinax-network/clickhouse-cli@latest
```

## Usage

```bash
$ clickhouse-cli 
NAME:
   clickhouse-cli - CLI tool for ClickHouse operations

USAGE:
   clickhouse-cli [global options] [command [command options]]

COMMANDS:
   migrate  Run Clickhouse migrations
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --node string, -n string      ClickHouse node to execute commands against. Should use the port for the native protocol. (default: "localhost:9000") [$CLICKHOUSE_NODE]
   --user string, -u string      ClickHouse user to execute commands as. (default: "default") [$CLICKHOUSE_USER]
   --password string, -p string  Password for the ClickHouse user. [$CLICKHOUSE_PASSWORD]
   --help, -h                    show help
```

## Migration Command

The primary feature of this CLI is running schema migrations against your ClickHouse database:

```bash
$ clickhouse-cli migrate --help
NAME:
   clickhouse-cli migrate - Run Clickhouse migrations

USAGE:
   clickhouse-cli migrate [command options] [arguments...]

OPTIONS:
   --dir string      Directory containing migration files (required)
   --table string    Table to track migrations in format <database>.<table> (default: "_migrations.schema_migrations")
   --create-table    Create the migrations tracking table if it doesn't exist (default: false)
   --help, -h        show help
```

## Migration File Format

Migration files should be named with a sequence number prefix followed by a descriptive name:

```
001_create_initial_tables.sql
002_add_indexes.sql
003_create_materialized_views.sql
```

Each migration file can contain multiple SQL statements separated by semicolons.

## Environment Variables

The following environment variables can be used instead of command-line flags:

- `CLICKHOUSE_NODE`: ClickHouse server address (default: "localhost:9000")
- `CLICKHOUSE_USER`: ClickHouse username (default: "default")
- `CLICKHOUSE_PASSWORD`: ClickHouse password
- `DEBUG`: Set to "true" to enable debug logging

## License

[License Information]

## Contributing

[Contributing Guidelines]