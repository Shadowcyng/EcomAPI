package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"mabletask/api/models"
	"mabletask/api/store"
	"mabletask/api/utils"
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

	_, err := h.UserStore.GetUserByEmail(c.Request.Context(), req.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}
	if err.Error() != fmt.Sprintf("user with email '%s' not found", req.Email) {
		log.Printf("ERROR: Database error during signup email check: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ERROR: Failed to hash password for %s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

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
	tokenString, err := utils.GenerateJWT(user)
	if err != nil {
		log.Printf("ERROR: Failed to generate JWT for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}
	c.SetCookie(
		"jwt_token",
		tokenString,
		int(24*time.Hour),
		"/",
		"",
		true,
		true,
	)
	log.Printf("User registered and JWT issued: ID=%d, Email=%s", user.ID, user.Email)

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user_email": user.Email, "token": tokenString})
}

func (h *AuthHandlers) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	user, err := h.UserStore.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		log.Printf("Login failed for email %s: %v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(req.Password))
	if err != nil {
		log.Printf("Login failed for email %s: password mismatch", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	tokenString, err := utils.GenerateJWT(user)
	if err != nil {
		log.Printf("ERROR: Failed to generate JWT for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	c.SetCookie(
		"jwt_token",
		tokenString,
		int(24*time.Hour),
		"/",
		"",
		true,
		true,
	)

	log.Printf("User logged in: ID=%d, Email=%s. JWT issued.", user.ID, user.Email)
	c.JSON(http.StatusOK, gin.H{
		"message":    "Login successful",
		"user_email": user.Email,
		"token":      tokenString,
	})
}

func (h *AuthHandlers) Logout(c *gin.Context) {
	c.SetCookie(
		"jwt_token",
		"",
		-1,
		"/",
		"",
		true,
		true,
	)

	log.Println("User logged out (JWT cookie cleared).")
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandlers) GetUserByToken(c *gin.Context) {
	email := c.GetString("user_email")
	user, err := h.UserStore.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    user.ID,
		"user_email": user.Email,
	})
}
