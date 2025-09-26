package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.uber.org/zap"
)

type ClickHouseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	Timeout  time.Duration
}

type ClickHouseClient struct {
	conn   driver.Conn
	db     *sql.DB
	config ClickHouseConfig
	logger *zap.Logger
}

func NewClickHouseClient(config ClickHouseConfig, logger *zap.Logger) (*ClickHouseClient, error) {
	// Create native ClickHouse connection
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: config.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Create SQL connection for compatibility
	db, err := sql.Open("clickhouse", fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s?dial_timeout=10s&read_timeout=20s",
		config.Username, config.Password, config.Host, config.Port, config.Database))
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL connection: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	logger.Info("ClickHouse connection established successfully",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Database))

	return &ClickHouseClient{
		conn:   conn,
		db:     db,
		config: config,
		logger: logger,
	}, nil
}

func (c *ClickHouseClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.logger.Error("Failed to close ClickHouse connection", zap.Error(err))
			return err
		}
	}
	if c.db != nil {
		if err := c.db.Close(); err != nil {
			c.logger.Error("Failed to close ClickHouse SQL connection", zap.Error(err))
			return err
		}
	}
	return nil
}

func (c *ClickHouseClient) GetConnection() driver.Conn {
	return c.conn
}

func (c *ClickHouseClient) GetSQLDB() *sql.DB {
	return c.db
}

func (c *ClickHouseClient) Execute(ctx context.Context, query string, args ...interface{}) error {
	return c.conn.Exec(ctx, query, args...)
}

func (c *ClickHouseClient) Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	return c.conn.Query(ctx, query, args...)
}

func (c *ClickHouseClient) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	return c.conn.QueryRow(ctx, query, args...)
}

func (c *ClickHouseClient) Insert(ctx context.Context, table string, columns []string, rows [][]interface{}) error {
	batch, err := c.conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s (%s)", table, fmt.Sprintf("%s", columns)))
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, row := range rows {
		if err := batch.Append(row...); err != nil {
			return fmt.Errorf("failed to append row: %w", err)
		}
	}

	return batch.Send()
}

func (c *ClickHouseClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.conn.Ping(ctx); err != nil {
		return fmt.Errorf("ClickHouse health check failed: %w", err)
	}

	return nil
}

func (c *ClickHouseClient) GetStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			name,
			value
		FROM system.metrics
		WHERE name IN ('Query', 'SelectQuery', 'InsertQuery', 'Merge', 'ReplicatedFetch')
	`

	rows, err := c.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]interface{})
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			return nil, fmt.Errorf("failed to scan stats row: %w", err)
		}
		stats[name] = value
	}

	return stats, nil
}
