package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"boilerplate/internal/app"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
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

