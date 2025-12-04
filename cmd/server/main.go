package main

import (
	"log"
	"os"

	"boilerplate/internal/app"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("using environment variables; godotenv.Load() returned: %v", err)
	}

	// Get PORT from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Initialize app
	fiberApp := app.NewApp()

	// Start server
	log.Printf("Server starting on port %s", port)
	log.Fatal(fiberApp.Listen(":" + port))
}

