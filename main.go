// cmd/server/main.go
package main

import (
	"auth-server/internal/api"
	"auth-server/internal/config"
	"auth-server/internal/database"
	"log"
	"os"

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
	if port == "" {
		// Try to get port from config
		serverAddr := cfg.Server.GetServerAddr()
		if serverAddr == "" {
			// Fall back to default port
			port = "8080"
			log.Printf("No port configuration found, using default port %s", port)
		} else {
			// Extract port from serverAddr if it exists
			log.Printf("Using configured server address: %s", serverAddr)
			port = serverAddr
			if serverAddr[0] == ':' {
				port = serverAddr[1:] // Remove leading colon if present
			}
		}
	} else {
		log.Printf("Using PORT from environment: %s", port)
	}

	// Ensure port has leading colon for proper address format
	serverAddr := ":" + port

	server := api.NewServer(db, serverAddr, cfg.Auth)
	server.SetupRoutes()

	// Start server
	log.Printf("Server starting on %s", serverAddr)
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
		os.Exit(1)
	}
}
