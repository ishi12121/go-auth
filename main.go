// cmd/server/main.go
package main

import (
	"auth-server/internal/api"
	"auth-server/internal/config"
	"auth-server/internal/database"
	"log"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
)

func main() {
	// Set up logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Starting auth server...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	sqlDB, err := cfg.Database.GetDatabaseWithLogging()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer sqlDB.Close()

	// Convert sql.DB to sqlx.DB
	sqlxDB := sqlx.NewDb(sqlDB, "postgres")

	// Create database wrapper
	db := &database.Database{DB: sqlxDB}

	// Get port from environment for Render deployment
	port := os.Getenv("PORT")
	var serverAddr string

	if port != "" {
		// Use environment variable port (for Render)
		log.Printf("Using PORT from environment: %s", port)
		serverAddr = ":" + port
	} else {
		// Try to get address from config
		serverAddr = cfg.Server.GetServerAddr()
		if serverAddr == "" {
			// Fall back to default port
			log.Printf("No server address configured, using default port 8080")
			serverAddr = ":8080"
		} else {
			log.Printf("Using configured server address: %s", serverAddr)
			// Make sure it has a leading colon for binding
			if !strings.Contains(serverAddr, ":") {
				serverAddr = ":" + serverAddr
			}
		}
	}

	// For Render deployment, we need to listen on 0.0.0.0 instead of localhost
	serverAddr = strings.Replace(serverAddr, "localhost", "0.0.0.0", 1)

	server := api.NewServer(db, serverAddr, cfg.Auth)
	server.SetupRoutes()

	// Start server
	log.Printf("Server starting on %s", serverAddr)
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
		os.Exit(1)
	}
}
