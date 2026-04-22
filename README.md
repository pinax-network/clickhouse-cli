# ClickHouse CLI

A small Go command-line tool for running SQL schema migrations against a ClickHouse server. It connects over the native protocol, applies numbered `.sql` files in order, and records what has been applied in a tracking table so subsequent runs only execute the new ones.

Out of the box the CLI assumes a replicated ClickHouse cluster and emits cluster-aware DDL (`Replicated` database engine, `ReplicatedMergeTree`, `ON CLUSTER`). Single-node setups are supported via `--cluster-mode=false`.

## Installation

```bash
go install github.com/pinax-network/clickhouse-cli/cmd/clickhouse-cli@latest
```

## Running migrations

Migration files live in a directory passed as a positional argument. Each filename must start with a contiguous sequence number starting at 1, followed by an underscore and a descriptive name:

```
1_create_initial_tables.sql
2_add_indexes.sql
3_create_materialized_views.sql
```

Files that don't match the pattern are ignored, so a `README.md` sitting alongside the migrations is fine. Each file may contain multiple SQL statements separated by `;` — note that the splitter is naive, so avoid semicolons inside string literals or compound statements.

A typical invocation:

```bash
clickhouse-cli \
  --node clickhouse.internal:9000 \
  --user migrator \
  --password "$CLICKHOUSE_PASSWORD" \
  migrate --create-migrations-table ./migrations
```

All flags can also be supplied via the `CLICKHOUSE_*` environment variables shown in the help output below. Set `DEBUG` to any non-empty value to log every query sent to ClickHouse.

## `migrate --help`

```
NAME:
   clickhouse-cli migrate - Run Clickhouse migrations.

USAGE:
   clickhouse-cli migrate [options] <migration directory>

OPTIONS:
   --schema-table string      Table containing schema migrations in the format of <database>.<table>. (default: "default.schema_migrations") [$CLICKHOUSE_SCHEMA_TABLE]
   --create-migrations-table  Create the schema migration table and database if it does not exist. [$CLICKHOUSE_CREATE_MIGRATIONS_TABLE]
   --cluster-mode             Use Replicated database and ReplicatedMergeTree engines with ON CLUSTER. Disable for single-node setups. [$CLICKHOUSE_CLUSTER_MODE]
   --help, -h                 show help

GLOBAL OPTIONS:
   --node string, -n string      ClickHouse node to execute commands against. Should use the port for the native protocol. (default: "localhost:9000") [$CLICKHOUSE_NODE]
   --user string, -u string      ClickHouse user to execute commands as. (default: "default") [$CLICKHOUSE_USER]
   --password string, -p string  Password for the ClickHouse user. [$CLICKHOUSE_PASSWORD]
```

## Behavior notes

If a migration fails partway through, the tracking table is marked dirty and the CLI refuses to run again until the database is reconciled manually — either by fixing the partial state and inserting a clean row, or by rolling back the offending changes and the dirty row. Already-applied migrations are skipped based on the latest version recorded in the tracking table, so rerunning the command against the same directory is safe.

## License

[License Information]
