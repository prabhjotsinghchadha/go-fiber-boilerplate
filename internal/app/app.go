package app

import (
	"boilerplate/internal/middleware"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// getTrustedProxies parses TRUSTED_PROXIES environment variable (comma-separated)
// Returns empty slice if not set, which is safe (Fiber will skip proxy headers)
func getTrustedProxies() []string {
	trustedProxiesStr := os.Getenv("TRUSTED_PROXIES")
	if trustedProxiesStr == "" {
		return []string{}
	}
	
	// Split by comma and trim whitespace
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

// NewApp creates and configures a new Fiber application
func NewApp() *fiber.App {
	// Configure proxy support for correct IP detection behind load balancers/reverse proxies
	// When EnableTrustedProxyCheck is true and TrustedProxies is set, c.IP() will read
	// from X-Forwarded-For header for requests from trusted proxy IPs
	trustedProxies := getTrustedProxies()
	enableProxyCheck := os.Getenv("ENABLE_TRUSTED_PROXY_CHECK")
	
	appConfig := fiber.Config{
		ReadBufferSize:  65536, // 64KB read buffer (increased for Docker/browser compatibility)
		WriteBufferSize: 65536, // 64KB write buffer
	}
	
	// Security: Only enable trusted proxy check when explicitly enabled AND trusted proxies are configured
	// This prevents accidentally trusting all proxies (IP spoofing vulnerability)
	if enableProxyCheck == "true" {
		if len(trustedProxies) == 0 {
			log.Fatalf("SECURITY ERROR: ENABLE_TRUSTED_PROXY_CHECK=true but TRUSTED_PROXIES is empty. " +
				"This would trust all proxies and allow IP spoofing. " +
				"Either set TRUSTED_PROXIES to a comma-separated list of trusted proxy IPs/CIDRs, " +
				"or set ENABLE_TRUSTED_PROXY_CHECK=false")
		}
		appConfig.EnableTrustedProxyCheck = true
		appConfig.TrustedProxies = trustedProxies
		// Use X-Forwarded-For header to get real client IP
		appConfig.ProxyHeader = fiber.HeaderXForwardedFor
		log.Printf("Trusted proxy check enabled with %d trusted proxy(ies): %v", len(trustedProxies), trustedProxies)
	} else if len(trustedProxies) > 0 {
		// If trusted proxies are configured but not explicitly enabled, log a warning
		log.Printf("WARNING: TRUSTED_PROXIES is set but ENABLE_TRUSTED_PROXY_CHECK is not 'true'. " +
			"Proxy checking is disabled. Set ENABLE_TRUSTED_PROXY_CHECK=true to enable.")
	}
	
	app := fiber.New(appConfig)

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	
	// CORS configuration
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		// Check environment indicator (GO_ENV or ENV)
		env := os.Getenv("GO_ENV")
		if env == "" {
			env = os.Getenv("ENV")
		}
		
		// Production safety check: fail fast if in production without ALLOWED_ORIGINS
		if env == "production" {
			log.Fatalf("SECURITY ERROR: ALLOWED_ORIGINS is empty in production environment. " +
				"This would allow requests from any origin, creating a security vulnerability. " +
				"Please set ALLOWED_ORIGINS to a comma-separated list of allowed origins.")
		}
		
		// Safe default for local development
		allowedOrigins = "http://localhost:3000,http://localhost:8080,http://127.0.0.1:3000,http://127.0.0.1:8080"
		log.Printf("INFO: ALLOWED_ORIGINS not set, using development fallback: %s", allowedOrigins)
	}
	
	corsConfig := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization,X-Requested-With",
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour
	}
	app.Use(cors.New(corsConfig))

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

