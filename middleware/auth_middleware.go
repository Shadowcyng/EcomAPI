package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"mabletask/api/utils"

	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		defaultToken := c.GetHeader("X-API-KEY")
		if defaultToken == os.Getenv("AUTH_DEFAULT") {
			c.Next()
			return
		}
		tokenString, err := c.Cookie("jwt_token")
		if err != nil {
			tokenString = c.GetHeader("Authorization")
			if tokenString == "" {
				log.Println("AuthRequired: No JWT token found in cookie or header")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No token provided"})
				return
			}
			if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
				tokenString = tokenString[7:]
			}

		}
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			log.Printf("AuthRequired: Invalid JWT token: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid or expired token"})
			return
		}

		fmt.Println("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		log.Printf("AuthRequired: User authenticated - ID: %d, Email: %s", claims.UserID, claims.Email)
		c.Next()
	}
}
