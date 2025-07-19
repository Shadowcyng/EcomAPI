// api/handlers/auth_handlers.go
package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"mabletask/api/models"
	"mabletask/api/store" // NEW: Import your store package
	"mabletask/api/utils" // Your utilities for session management
)

type AuthHandlers struct {
	UserStore *store.UserStore
}

func NewAuthHandlers(userStore *store.UserStore) *AuthHandlers {
	return &AuthHandlers{UserStore: userStore}
}

func (h *AuthHandlers) Signup(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 1. Check if the user's email already exists in the database.
	_, err := h.UserStore.GetUserByEmail(c.Request.Context(), req.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}
	// Check if the error is actually "not found". If it's a different DB error, return 500.
	if err.Error() != fmt.Sprintf("user with email '%s' not found", req.Email) { // Using fmt.Sprintf to match exact error string from store
		log.Printf("ERROR: Database error during signup email check: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
		return
	}

	// 2. Hash the password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ERROR: Failed to hash password for %s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// 3. Store the user's email and hashed password in your persistent user database.
	user, err := h.UserStore.CreateUser(c.Request.Context(), req.Email, hashedPassword)
	if err != nil {
		log.Printf("ERROR: Failed to create user in DB for email %s: %v", req.Email, err)
		if err.Error() == fmt.Sprintf("user with email '%s' already exists", req.Email) {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		}
		return
	}

	log.Printf("User registered successfully via DB: ID=%d, Email=%s", user.ID, user.Email)
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user_email": user.Email})
}

// Login handles user authentication and JWT token creation.
func (h *AuthHandlers) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 1. Retrieve the user from the database by email.
	user, err := h.UserStore.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		log.Printf("Login failed for email %s: %v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// 2. Compare the provided password with the stored hashed password.
	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(req.Password))
	if err != nil {
		log.Printf("Login failed for email %s: password mismatch", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// 3. Generate JWT token.
	tokenString, err := utils.GenerateJWT(user)
	if err != nil {
		log.Printf("ERROR: Failed to generate JWT for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	c.SetCookie(
		"jwt_token",
		tokenString,
		int(24*time.Hour/time.Second),
		"/",
		"",
		false,
		true,
	)

	log.Printf("User logged in: ID=%d, Email=%s. JWT issued.", user.ID, user.Email)
	c.JSON(http.StatusOK, gin.H{
		"message":    "Login successful",
		"user_email": user.Email,
	})
}

func (h *AuthHandlers) Logout(c *gin.Context) {
	// Clear the JWT cookie by setting its MaxAge to -1 (immediately expire).
	c.SetCookie(
		"jwt_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	log.Println("User logged out (JWT cookie cleared).")
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
