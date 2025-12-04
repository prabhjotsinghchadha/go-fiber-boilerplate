package app

import (
	"boilerplate/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// NewApp creates and configures a new Fiber application
func NewApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ReadBufferSize:  65536, // 64KB read buffer (increased for Docker/browser compatibility)
		WriteBufferSize: 65536, // 64KB write buffer
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Setup routes
	setupRoutes(app)

	return app
}

// setupRoutes registers all application routes
func setupRoutes(app *fiber.App) {
	// Public routes (no auth required)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Protected API routes (auth + rate limiting required)
	// Auth runs first to identify user, then rate limiting uses user ID if available
	api := app.Group("/api", middleware.Auth(), middleware.RateLimit())
	
	// Test route to verify auth middleware
	api.Get("/profile", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user": c.Locals("user"),
		})
	})
}

