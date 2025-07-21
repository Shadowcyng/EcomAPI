package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"mabletask/api/database"
	"mabletask/api/handlers"
	"mabletask/api/middleware"
	"mabletask/api/store"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading .env: %v", err)
	}

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	dbClient, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL database: %v", err)
	}
	defer dbClient.Close()

	chClient, err := database.NewClickHouseDB()
	if err != nil {
		log.Fatalf("Failed to initialize ClickHouse database: %v", err)
	}
	defer chClient.Close()

	userStore := store.NewUserStore(dbClient.DB)
	analyticsStore := store.NewAnalyticsStore(chClient)

	authHandlers := handlers.NewAuthHandlers(userStore)
	analyticsHandlers := handlers.NewAnalyticsHandlers(analyticsStore)

	r := gin.Default()

	r.Use(middleware.CORSMiddleware())
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "Welcome to the Mable Analytics API!"})
	})
	api := r.Group("/api")
	{
		// Authentication Endpoints (no authentication required)
		api.POST("/signup", authHandlers.Signup)
		api.POST("/login", authHandlers.Login)
		api.POST("/logout", authHandlers.Logout)
		api.GET("/health", handlers.HealthCheck)
		api.POST("/track", analyticsHandlers.TrackEvent)
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": "Welcome to the Mable Analytics API!"})
		})
		// Protected Routes (require a valid JWT token)
		protected := api.Group("/")
		protected.Use(middleware.AuthRequired())
		{
			protected.POST("/validate-user", authHandlers.GetUserByToken)
			// Example protected endpoint (e.g., get user profile)
			protected.GET("/profile", func(c *gin.Context) {
				userID := c.MustGet("user_id").(int)
				userEmail := c.MustGet("user_email").(string)
				ipAddress := c.ClientIP()

				c.JSON(http.StatusOK, gin.H{
					"message":    "Welcome to your profile!",
					"user_id":    userID,
					"user_email": userEmail,
					"ip_address": ipAddress,
				})
			})

			analyticsGroup := protected.Group("/stats")
			{
				analyticsGroup.GET("/event-counts", analyticsHandlers.GetEventCountsOverTime)
				analyticsGroup.GET("/average-event-duration", analyticsHandlers.GetAverageEventDuration)
				analyticsGroup.GET("/average-custom-param", analyticsHandlers.GetAverageCustomEventParameter)
				analyticsGroup.GET("/unique-users", analyticsHandlers.GetUniqueUsersOverTime)
				analyticsGroup.GET("/top-paths", analyticsHandlers.GetTopNPagePaths)

			}
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Go API server starting on http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Go API server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting.")
}
