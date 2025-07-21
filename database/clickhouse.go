package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseClient struct {
	Conn clickhouse.Conn
}

func NewClickHouseDB() (*ClickHouseClient, error) {
	host := os.Getenv("CLICKHOUSE_HOST")
	nativePortStr := os.Getenv("CLICKHOUSE_NATIVE_PORT")
	dbName := os.Getenv("CLICKHOUSE_DB_NAME")
	username := os.Getenv("CLICKHOUSE_USERNAME")
	password := os.Getenv("CLICKHOUSE_PASSWORD")

	if host == "" || nativePortStr == "" || dbName == "" {
		return nil, fmt.Errorf("CLICKHOUSE_HOST, CLICKHOUSE_NATIVE_PORT, or CLICKHOUSE_DB_NAME environment variables are not set")
	}

	nativePort, err := strconv.Atoi(nativePortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CLICKHOUSE_NATIVE_PORT: %w", err)
	}

	options := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", host, nativePort)},
		Auth: clickhouse.Auth{
			Database: dbName,
			Username: username,
			Password: password,
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{{Name: "mable-api", Version: "1.0.0"}},
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout: time.Second * 5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := clickhouse.Open(options)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse via Native TCP: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	log.Println("Successfully connected to ClickHouse database via Native TCP (direct options)!")
	return &ClickHouseClient{Conn: conn}, nil
}

func (c *ClickHouseClient) Close() {
	if c.Conn != nil {
		c.Conn.Close()
		log.Println("ClickHouse connection closed.")
	}
}
