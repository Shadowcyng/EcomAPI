// api/database/clickhouse.go
package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv" // Needed for parsing port string to int
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// ClickHouseClient holds the ClickHouse connection.
type ClickHouseClient struct {
	Conn clickhouse.Conn // This is still expected by AnalyticsStore
}

// NewClickHouseDB initializes and returns a new ClickHouseClient.
// It connects to ClickHouse using native TCP protocol by constructing options directly.
func NewClickHouseDB() (*ClickHouseClient, error) {
	// Retrieve connection details from environment variables
	host := os.Getenv("CLICKHOUSE_HOST")
	nativePortStr := os.Getenv("CLICKHOUSE_NATIVE_PORT") // Using NATIVE_PORT for TCP
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

	// Define ClickHouse connection options directly
	options := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", host, nativePort)}, // Use host and native TCP port
		Auth: clickhouse.Auth{
			Database: dbName,
			Username: username,
			Password: password,
		},
		ClientInfo: clickhouse.ClientInfo{ // Optional: Identify your application to ClickHouse
			Products: []struct {
				Name    string
				Version string
			}{{Name: "mable-api", Version: "1.0.0"}},
		},
		Compression: &clickhouse.Compression{ // Enable compression for native protocol
			Method: clickhouse.CompressionLZ4, // Or other methods like CompressionZSTD
		},
		DialTimeout: time.Second * 5, // Default dial timeout
		// No Protocol: clickhouse.HTTP here, as we are using native TCP
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Open the ClickHouse connection using options directly (no WithDSN)
	conn, err := clickhouse.Open(options) // Using clickhouse.Open with direct options
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse via Native TCP: %w", err)
	}

	// Ping to verify the connection
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	log.Println("Successfully connected to ClickHouse database via Native TCP (direct options)!")
	return &ClickHouseClient{Conn: conn}, nil
}

// Close closes the ClickHouse connection.
func (c *ClickHouseClient) Close() {
	if c.Conn != nil {
		c.Conn.Close()
		log.Println("ClickHouse connection closed.")
	}
}
