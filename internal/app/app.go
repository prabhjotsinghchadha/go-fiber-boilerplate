package app

import (
	"boilerplate/internal/handlers"
	"boilerplate/internal/middleware"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
)

// NewApp creates and configures a new Fiber application with middleware and routes.
func NewApp() *fiber.App {
	app := fiber.New(createAppConfig())

	// Apply global middleware
	setupMiddleware(app)

	// Register routes
	setupRoutes(app)

	return app
}

// createAppConfig creates the Fiber app configuration.
func createAppConfig() fiber.Config {
	config := fiber.Config{
		ReadBufferSize:  65536, // 64KB read buffer
		WriteBufferSize: 65536, // 64KB write buffer
	}

	// Configure proxy support if enabled
	configureProxy(&config)

	return config
}

// configureProxy configures trusted proxy settings for correct IP detection.
// When enabled, c.IP() will read from X-Forwarded-For header for requests from trusted proxies.
func configureProxy(config *fiber.Config) {
	enableProxyCheck := os.Getenv("ENABLE_TRUSTED_PROXY_CHECK")
	if enableProxyCheck != "true" {
		return
	}

	trustedProxies := getTrustedProxies()
	if len(trustedProxies) == 0 {
		log.Fatalf("SECURITY ERROR: ENABLE_TRUSTED_PROXY_CHECK=true but TRUSTED_PROXIES is empty. " +
			"This would trust all proxies and allow IP spoofing. " +
			"Either set TRUSTED_PROXIES to a comma-separated list of trusted proxy IPs/CIDRs, " +
			"or set ENABLE_TRUSTED_PROXY_CHECK=false")
	}

	config.EnableTrustedProxyCheck = true
	config.TrustedProxies = trustedProxies
	config.ProxyHeader = fiber.HeaderXForwardedFor

	log.Printf("Trusted proxy check enabled with %d trusted proxy(ies): %v", len(trustedProxies), trustedProxies)
}

// getTrustedProxies parses TRUSTED_PROXIES environment variable (comma-separated).
// Returns empty slice if not set, which is safe (Fiber will skip proxy headers).
func getTrustedProxies() []string {
	trustedProxiesStr := os.Getenv("TRUSTED_PROXIES")
	if trustedProxiesStr == "" {
		return []string{}
	}

	proxies := strings.Split(trustedProxiesStr, ",")
	result := make([]string, 0, len(proxies))
	for _, proxy := range proxies {
		trimmed := strings.TrimSpace(proxy)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// setupMiddleware applies global middleware to the application.
func setupMiddleware(app *fiber.App) {
	// Panic recovery middleware
	app.Use(recover.New())

	// Request logging middleware
	app.Use(logger.New())

	// CORS middleware
	app.Use(cors.New(createCORSConfig()))
}

// createCORSConfig creates the CORS configuration based on environment variables.
// Uses development defaults if ALLOWED_ORIGINS is not set and not in production.
func createCORSConfig() cors.Config {
	allowedOrigins := getAllowedOrigins()

	return cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization,X-Requested-With",
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour
	}
}

// getAllowedOrigins returns the allowed CORS origins from environment or defaults.
// Fails fast in production if ALLOWED_ORIGINS is not set.
func getAllowedOrigins() string {
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins != "" {
		return allowedOrigins
	}

	// Check if we're in production
	if isProduction() {
		log.Fatalf("SECURITY ERROR: ALLOWED_ORIGINS is empty in production environment. " +
			"This would allow requests from any origin, creating a security vulnerability. " +
			"Please set ALLOWED_ORIGINS to a comma-separated list of allowed origins.")
	}

	// Safe default for local development
	allowedOrigins = "http://localhost:3000,http://localhost:8080,http://127.0.0.1:3000,http://127.0.0.1:8080"
	log.Printf("INFO: ALLOWED_ORIGINS not set, using development fallback: %s", allowedOrigins)
	return allowedOrigins
}

// isProduction checks if the application is running in production environment.
func isProduction() bool {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = os.Getenv("ENV")
	}
	return env == "production"
}

// setupRoutes registers all application routes.
func setupRoutes(app *fiber.App) {
	// Public routes (no authentication required)
	setupPublicRoutes(app)

	// Protected API routes (authentication and rate limiting required)
	setupProtectedRoutes(app)
}

// setupPublicRoutes registers public routes that don't require authentication.
func setupPublicRoutes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// GraphQL proxy to Supabase (public for now; wrap in auth group later for mutations)
	app.All("/graphql", handlers.GraphQLProxy)

	// WebSocket endpoint for Realtime updates
	app.Use("/ws", handlers.UpgradeWebSocket)
	app.Get("/ws", websocket.New(handlers.WebSocketHandler))
}

// setupProtectedRoutes registers protected routes that require authentication and rate limiting.
// Auth middleware runs first to identify the user, then rate limiting uses the user ID if available.
func setupProtectedRoutes(app *fiber.App) {
	api := app.Group("/api", middleware.Auth(), middleware.RateLimit())

	// Example protected route
	api.Get("/profile", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user": c.Locals("user"),
		})
	})
}
