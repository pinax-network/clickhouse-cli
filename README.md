# Clickhouse CLI

A command-line interface for ClickHouse operations.

```bash
$ clickhouse-cli 
NAME:
   clickhouse-cli - CLI tool for ClickHouse operations

USAGE:
   clickhouse-cli [global options] [command [command options]]

COMMANDS:
   migrate  Run Clickhouse migrations.
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --node string, -n string      ClickHouse node to execute commands against. Should use the port for the native protocol. (default: "localhost:9000") [$CLICKHOUSE_NODE]
   --user string, -u string      ClickHouse user to execute commands as. (default: "default") [$CLICKHOUSE_USER]
   --password string, -p string  Password for the ClickHouse user. [$CLICKHOUSE_PASSWORD]
   --help, -h                    show help
```