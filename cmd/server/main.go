package main

import (
	"log"
	"os"

	"boilerplate/internal/app"
	"boilerplate/internal/cache"
	"boilerplate/internal/handlers"
	"boilerplate/internal/realtime"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("using environment variables; godotenv.Load() returned: %v", err)
	}

	// Initialize Redis cache
	if err := cache.Init(); err != nil {
		log.Printf("WARNING: Failed to initialize Redis cache: %v", err)
		log.Println("Continuing without cache...")
	}

	// Initialize WebSocket hub
	handlers.InitHub()

	// Initialize Supabase Realtime client
	if err := realtime.Init(); err != nil {
		log.Printf("WARNING: Failed to initialize Supabase Realtime client: %v", err)
		log.Println("Continuing without Realtime subscription...")
	}

	// Get PORT from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Initialize app
	fiberApp := app.NewApp()

	// Start Realtime subscriber in background
	go realtime.SubscribeToPrices()

	// Start server
	log.Printf("Server starting on port %s", port)
	log.Fatal(fiberApp.Listen(":" + port))
}

