package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	handlers "github.com/24tylerdurden/levo-api/internal/Handlers"
	"github.com/24tylerdurden/levo-api/internal/database"
	"github.com/24tylerdurden/levo-api/internal/services"
	"github.com/24tylerdurden/levo-api/pkg/config"
	"github.com/gin-gonic/gin"

	"log"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database with migrations
	db, err := database.InitializeDatabase(cfg.DBPath, cfg.MigrationsPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		// Test database connection
		err := db.Ping()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "unhealthy",
				"database": "disconnected",
				"error":    err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"database": "connected",
		})
	})

	// Initialize services

	schemaService := services.NewSchemaService()

	// Initialize handlers

	schemaHandler := handlers.NewSchemaHandler()

	// API routes
	api := router.Group("/api/v1")
	{
		apps := api.Group("/application/:aplication")
		{
			apps.POST("/schemas", schemaHandler.UploadApplicationSchema)

			apps.GET("/schemas/latest", schemaHandler.GetLatestApplicationSchema)

			apps.GET("/schemas/:version", schemaHandler.GetApplicationSchemVersion)
		}

		services := apps.Group("/services/:services")
		{
			services.POST("/schemas", schemaHandler.UploadServiceSchema)

			services.GET("/schemas/:latest", schemaHandler.GetLatestServiceSchema)

			services.GET("/schemas/:version", schemaHandler.GetServiceSchemaVersion)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Close database connection
	db.Close()
	log.Println("Server exited")
}
