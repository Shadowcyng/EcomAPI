package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type DBClient struct {
	DB *sql.DB
}

func NewPostgresDB() (*DBClient, error) {
	fmt.Println("DATABASE_URL:", os.Getenv("DATABASE_URL"))
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL environment variable not set. Using default for local development.")
		// IMPORTANT: Replace this with your actual local PostgreSQL connection string
		dbURL = "postgres://postgres:password@localhost:5432/mabledb?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to the database (ping failed): %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database!")
	return &DBClient{DB: db}, nil
}

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
