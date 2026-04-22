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
   clickhouse-cli migrate [command options] <migration directory>

OPTIONS:
   --schema-table string         Table containing schema migrations in the format of <database>.<table>. (default: "default.schema_migrations") [$CLICKHOUSE_SCHEMA_TABLE]
   --create-migrations-table     Create the schema migration table and database if it does not exist. (default: false) [$CLICKHOUSE_CREATE_MIGRATIONS_TABLE]
   --cluster-mode                Use Replicated database and ReplicatedMergeTree engines with ON CLUSTER. Disable for single-node setups. (default: true) [$CLICKHOUSE_CLUSTER_MODE]
   --help, -h                    show help
```

## Migration File Format

Migration files should be named with a sequence number prefix followed by a descriptive name:

```
001_create_initial_tables.sql
002_add_indexes.sql
003_create_materialized_views.sql
```

Each migration file can contain multiple SQL statements separated by semicolons. Note that the splitter is naive — semicolons inside string literals, comments, or compound statements (e.g. `CREATE FUNCTION`) are not supported.

## Environment Variables

The following environment variables can be used instead of command-line flags:

- `CLICKHOUSE_NODE`: ClickHouse server address (default: "localhost:9000")
- `CLICKHOUSE_USER`: ClickHouse username (default: "default")
- `CLICKHOUSE_PASSWORD`: ClickHouse password
- `CLICKHOUSE_SCHEMA_TABLE`: Schema migrations table in the form `<database>.<table>`
- `CLICKHOUSE_CREATE_MIGRATIONS_TABLE`: Whether to create the migrations table if missing
- `CLICKHOUSE_CLUSTER_MODE`: Whether to emit cluster-aware DDL (Replicated engines + ON CLUSTER)
- `DEBUG`: Set to any non-empty value to enable debug logging

## License

[License Information]

## Contributing

[Contributing Guidelines]