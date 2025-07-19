package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	// Added for `time.Time`
	"mabletask/api/models" // Import your models package to use the User struct
)

type UserStore struct {
	db *sql.DB 
}

// NewUserStore creates a new UserStore instance.
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// CreateUser inserts a new user into the database.
func (s *UserStore) CreateUser(ctx context.Context, email string, hashedPassword []byte) (*models.User, error) {
	user := &models.User{}
	query := `
		INSERT INTO users (email, hashed_password)
		VALUES ($1, $2)
		RETURNING id, email, created_at, updated_at;
	`
	// Use QueryRowContext for single row returns
	err := s.db.QueryRowContext(ctx, query, email, hashedPassword).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "idx_users_email"` ||
			err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` { // Depending on PG version/setup
			return nil, fmt.Errorf("user with email '%s' already exists", email)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("User created in DB: ID=%d, Email=%s", user.ID, user.Email)
	return user, nil
}

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, hashed_password, created_at, updated_at
		FROM users
		WHERE email = $1;
	`
	// Use QueryRowContext for single row returns
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.HashedPassword,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with email '%s' not found", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}
