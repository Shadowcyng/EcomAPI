// api/middleware/auth_middleware.go
package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"mabletask/api/utils" // Import your utils package for JWT functions
)

// AuthRequired is a Gin middleware to check for a valid JWT token.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Try to get the token from the HttpOnly cookie.
		tokenString, err := c.Cookie("jwt_token")
		if err != nil {
			log.Printf("AuthRequired: No JWT token cookie found or error: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No token provided"})
			return
		}

		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			log.Printf("AuthRequired: Invalid JWT token: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid or expired token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		log.Printf("AuthRequired: User authenticated - ID: %d, Email: %s", claims.UserID, claims.Email)

		c.Next()
	}
}