package models

import "time"

type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	// Username  string `json:"username" binding:"required,min=3,max=20"`
	// FirstName string `json:"first_name" binding:"required,min=2,max=30"`
	// LastName  string `json:"last_name" binding:"required,min=2,max=30"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type User struct {
	ID             int       `json:"id"`
	Email          string    `json:"email"`
	HashedPassword []byte    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
