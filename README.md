# Clickhouse CLI

A command-line interface for ClickHouse operations with both TUI (Terminal User Interface) and CLI modes.

## Features

- Drop databases from a ClickHouse cluster
- Interactive TUI mode with bubbletea
- CLI mode for scripting and automation
- Consul service discovery integration
- Support for regex patterns to filter databases
- ON CLUSTER operations support

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd clickhouse-cli

# Build the application
./build.sh
```

## Usage

### TUI Mode

Run the application without any arguments to enter TUI mode:

```bash
./clickhouse-cli
```

In TUI mode, you'll first see a main menu where you can select the action you want to perform:

1. Select an action (e.g., "Drop Databases")

For the "Drop Databases" action, you'll be guided through the following steps:

1. Select a cluster from the available Consul services
2. Select a node to connect to
3. Choose how to select databases to drop (all, regex, or manual selection)
4. If using regex, enter a pattern; if manual, select databases from the list
5. Choose whether to drop locally or ON CLUSTER
6. Confirm the operation
7. Execute the statements

### CLI Mode

For automation and scripting, you can use the CLI mode with flags:

```bash
# Drop all databases on a specific host
./clickhouse-cli clickhouse drop databases --host localhost --confirm

# Drop databases matching a regex pattern
./clickhouse-cli clickhouse drop databases --host localhost --regex 'test.*' --confirm

# Drop specific databases
./clickhouse-cli clickhouse drop databases --host localhost --list 'db1,db2,db3' --confirm

# Connect to a cluster via Consul and drop databases ON CLUSTER
./clickhouse-cli clickhouse drop databases --cluster clickhouse-production --on-cluster --confirm
```

## Available Flags

- `--cluster`: Cluster to connect to (uses Consul service discovery)
- `--host`: Hostname to connect to
- `--port`: Port to connect to (default: 8123)
- `--user`: Username for authentication (default: default)
- `--password`: Password for authentication
- `--database`: Database to use (default: default)
- `--regex`: Regex pattern to match databases
- `--list`: Comma-separated list of databases to drop
- `--on-cluster`: Drop databases on the entire cluster
- `--confirm`: Confirm dropping databases without prompting

## Environment Variables

- `DEBUG`: Set to any value to enable debug logging

## Development

The application is structured as follows:

- `cmd/`: Command-line interface code
- `pkg/clickhouse/`: ClickHouse client and operations
- `pkg/consul/`: Consul service discovery
- `pkg/tui/`: Terminal User Interface components
  - `main_menu.go`: Main menu for selecting actions
  - `common.go`: Common types and utilities
  - `cluster_selection.go`: Cluster selection component
  - `node_selection.go`: Node selection component
  - `database_selection.go`: Database selection components
  - `drop_mode.go`: Drop mode selection component
  - `confirmation.go`: Confirmation and execution components
  - `drop_databases.go`: Drop databases flow

The TUI is designed with a modular architecture where each component is a separate model that implements the `tea.Model` interface. This allows for easy reuse of components across different flows. To add new features or commands:

1. Create a new action in `main_menu.go`
2. Create a new flow file (e.g., `your_feature.go`) that uses the existing components
3. Implement any new components needed for your feature
4. Update the documentation and build script