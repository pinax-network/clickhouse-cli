package clickhouse

import (
	"clickhouse-cli/pkg/log"
	"context"
	"fmt"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.uber.org/zap"
)

// Client represents a Clickhouse client
type Client struct {
	conn driver.Conn
}

// NewClient creates a new Clickhouse client
func NewClient(ctx context.Context, node, user, password string) (*Client, error) {

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{node},
		Auth: clickhouse.Auth{
			Username: user,
			Password: password,
		},
		Debug: os.Getenv("DEBUG") == "true",
		Debugf: func(format string, v ...any) {
			log.Debugf(format, v...)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Clickhouse: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Clickhouse: %w", err)
	}

	log.Info("successfully connected to Clickhouse", zap.String("node", node))

	return &Client{conn: conn}, nil
}

// Close closes the connection to the Clickhouse server
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) execute(ctx context.Context, query string, parameters clickhouse.Parameters) error {
	log.Debug("executing query", zap.String("query", query), zap.Any("parameters", parameters))
	return c.conn.Exec(getContext(ctx, parameters), query)
}

func (c *Client) queryRows(ctx context.Context, query string, parameters clickhouse.Parameters) (driver.Rows, error) {
	log.Debug("executing query", zap.String("query", query), zap.Any("parameters", parameters))
	return c.conn.Query(getContext(ctx, parameters), query)
}

func (c *Client) queryStruct(ctx context.Context, query string, parameters clickhouse.Parameters, result any) error {
	log.Debug("executing query", zap.String("query", query), zap.Any("parameters", parameters))
	err := c.conn.QueryRow(getContext(ctx, parameters), query, result).ScanStruct(result)
	if err != nil {
		return fmt.Errorf("failed to execute Clickhouse query: %w", err)
	}
	return nil
}

func getContext(ctx context.Context, parameters clickhouse.Parameters) context.Context {
	chCtx := ctx
	if parameters != nil {
		chCtx = clickhouse.Context(ctx, clickhouse.WithParameters(parameters))
	}
	return chCtx
}
