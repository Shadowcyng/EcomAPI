// api/database/postgres.go
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver, imported for its side effects (registering itself)
)

// DBClient represents our PostgreSQL database connection.
type DBClient struct {
	DB *sql.DB
}

func NewPostgresDB() (*DBClient, error) {
	fmt.Println("DATABASE_URL:", os.Getenv("DATABASE_URL")) // Debugging line to check if DATABASE_URL is set
	// DATABASE_URL example: "postgres://user:password@host:port/dbname?sslmode=disable"
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Provide a sensible default or error out if not found for development
		log.Println("DATABASE_URL environment variable not set. Using default for local development.")
		// IMPORTANT: Replace this with your actual local PostgreSQL connection string
		dbURL = "postgres://postgres:password@localhost:5432/mabledb?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	// Set connection pool settings for better performance in production
	db.SetMaxOpenConns(25)                 // Max number of open connections
	db.SetMaxIdleConns(5)                  // Max number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Max time a connection can be reused

	// Ping the database to verify the connection is alive
	if err = db.Ping(); err != nil {
		db.Close() // Close connection if ping fails
		return nil, fmt.Errorf("error connecting to the database (ping failed): %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database!")
	return &DBClient{DB: db}, nil
}

// Close closes the database connection. Call this when the application shuts down.
func (c *DBClient) Close() {
	if c.DB != nil {
		err := c.DB.Close()
		if err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			log.Println("PostgreSQL database connection closed.")
		}
	}
}
