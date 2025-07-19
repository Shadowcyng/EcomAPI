// api/middleware/cors.go
package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware provides a Gin middleware function for handling Cross-Origin Resource Sharing.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set 'Access-Control-Allow-Origin'. During development, this should be your Remix app's URL.
		// For deployment, use an environment variable (e.g., os.Getenv("FE_ORIGIN"))
		// or list specific allowed domains. Avoid "*" in production for security.
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Default for Remix dev server
		if os.Getenv("FE_ORIGIN") != "" { // Allow overriding with environment variable for deployment
			c.Writer.Header().Set("Access-Control-Allow-Origin", os.Getenv("FE_ORIGIN"))
		}

		// Allow credentials (like cookies/sessions) to be sent with cross-origin requests.
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Specify which headers can be used in the actual request.
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")

		// Specify which HTTP methods are allowed for cross-origin requests.
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// Handle preflight requests (OPTIONS method). Browsers send these before complex cross-origin requests.
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent) // Respond with 204 No Content for preflight
			return
		}
		c.Next() // Continue processing the request chain
	}
}