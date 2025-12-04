package middleware

import (
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimit middleware applies rate limiting based on user ID (if authenticated) or IP address
//
// IP-based rate limiting relies on correct proxy configuration in the Fiber app.
// When deployed behind reverse proxies/load balancers, ensure the app is configured with:
// - ENABLE_TRUSTED_PROXY_CHECK=true (or set TRUSTED_PROXIES)
// - TRUSTED_PROXIES set to your proxy IP ranges (e.g., "172.16.0.0/12,10.0.0.0/8")
// This ensures c.IP() returns the real client IP from X-Forwarded-For header.
// Without proper proxy configuration, all users behind the same proxy will share a rate limit.
func RateLimit() fiber.Handler {
	// Get rate limit from environment, default to 100 requests per minute
	maxRequests := 100
	if maxStr := os.Getenv("RATE_LIMIT_MAX"); maxStr != "" {
		if parsed, err := strconv.Atoi(maxStr); err == nil && parsed > 0 {
			maxRequests = parsed
		}
	}

	return limiter.New(limiter.Config{
		Max:        maxRequests,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// If user is authenticated, use user ID as key (most accurate)
			if userID := c.Locals("user"); userID != nil {
				if userIDStr, ok := userID.(string); ok && userIDStr != "" {
					return "user:" + userIDStr
				}
			}
			// Otherwise, use IP address (requires proper proxy configuration for accuracy)
			// c.IP() will return the real client IP if proxy support is configured in app.go
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		},
	})
}

