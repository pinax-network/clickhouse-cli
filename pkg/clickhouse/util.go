package clickhouse

import (
	"context"
	"fmt"

	"github.com/pinax-network/clickhouse-cli/pkg/log"
	"go.uber.org/zap"
)

func (c *Client) databaseExists(ctx context.Context, databaseName string) (bool, error) {

	rows, err := c.QueryRows(ctx,
		`SELECT 1 FROM system.databases WHERE name = {name:String} LIMIT 1`,
		map[string]string{"name": databaseName},
	)
	if err != nil {
		return false, fmt.Errorf("failed to check if database exists: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error("failed to call close on rows", zap.Error(err))
		}
	}()

	exists := rows.Next()
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("error iterating rows: %w", err)
	}

	return exists, nil
}

func (c *Client) tableExists(ctx context.Context, databaseName, tableName string) (bool, error) {

	rows, err := c.QueryRows(ctx,
		`SELECT 1 FROM system.tables WHERE database = {database:String} AND name = {name:String} LIMIT 1`,
		map[string]string{"database": databaseName, "name": tableName},
	)
	if err != nil {
		return false, fmt.Errorf("failed to check if table exists: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error("failed to call close on rows", zap.Error(err))
		}
	}()

	exists := rows.Next()
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("error iterating rows: %w", err)
	}

	return exists, nil
}
