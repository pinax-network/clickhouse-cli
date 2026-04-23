package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.uber.org/zap"
)

// Client represents a Clickhouse client
type Client struct {
	conn   driver.Conn
	logger *zap.Logger
}

// NewClient creates a new Clickhouse client. The provided logger is used for
// all logging emitted by the client. When debug is true the underlying driver
// emits verbose logs through the same logger.
func NewClient(ctx context.Context, logger *zap.Logger, node, user, password string, debug bool) (*Client, error) {

	sugar := logger.Sugar()
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{node},
		Auth: clickhouse.Auth{
			Username: user,
			Password: password,
		},
		Debug: debug,
		Debugf: func(format string, v ...any) {
			sugar.Debugf(format, v...)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Clickhouse: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Clickhouse: %w", err)
	}

	logger.Info("successfully connected to Clickhouse", zap.String("node", node))

	return &Client{conn: conn, logger: logger}, nil
}

// Close closes the connection to the Clickhouse server
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Execute runs a statement that returns no rows.
func (c *Client) Execute(ctx context.Context, query string, parameters clickhouse.Parameters) error {
	c.logger.Debug("executing query", zap.String("query", query), zap.Any("parameters", parameters))
	return c.conn.Exec(queryContext(ctx, parameters), query)
}

// QueryRows runs a query and returns its rows. The caller owns the returned
// driver.Rows and must close it.
func (c *Client) QueryRows(ctx context.Context, query string, parameters clickhouse.Parameters) (driver.Rows, error) {
	c.logger.Debug("executing query", zap.String("query", query), zap.Any("parameters", parameters))
	return c.conn.Query(queryContext(ctx, parameters), query)
}

// QueryStruct runs a query expected to return a single row and scans it into result.
func (c *Client) QueryStruct(ctx context.Context, query string, parameters clickhouse.Parameters, result any) error {
	c.logger.Debug("executing query", zap.String("query", query), zap.Any("parameters", parameters))
	if err := c.conn.QueryRow(queryContext(ctx, parameters), query).ScanStruct(result); err != nil {
		return fmt.Errorf("failed to execute Clickhouse query: %w", err)
	}
	return nil
}

func queryContext(ctx context.Context, parameters clickhouse.Parameters) context.Context {
	if parameters == nil {
		return ctx
	}
	return clickhouse.Context(ctx, clickhouse.WithParameters(parameters))
}
