package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/atqamz/kogase-backend/docs" // Import for swagger docs
	"github.com/atqamz/kogase-backend/server"
	"github.com/joho/godotenv"
)

// @title Kogase Telemetry API
// @version 1.0
// @description Backend API for Kogase game telemetry system
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.kogase.io/support
// @contact.email support@kogase.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Create new server
	s, err := server.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing server: %v\n", err)
		os.Exit(1)
	}

	// Start server
	log.Println("Starting Kogase server...")
	if err := s.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running server: %v\n", err)
		os.Exit(1)
	}
}
