package middleware

import (
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimit applies rate limiting based on user ID (if authenticated) or IP address.
//
// Rate limiting behavior:
// - Authenticated users: Rate limit is per user ID (most accurate)
// - Unauthenticated users: Rate limit is per IP address
//
// Note: IP-based rate limiting requires proper proxy configuration when deployed
// behind reverse proxies/load balancers. Ensure the app is configured with:
// - ENABLE_TRUSTED_PROXY_CHECK=true
// - TRUSTED_PROXIES set to your proxy IP ranges
// This ensures c.IP() returns the real client IP from X-Forwarded-For header.
// Without proper configuration, all users behind the same proxy will share a rate limit.
func RateLimit() fiber.Handler {
	maxRequests := getMaxRequests()

	return limiter.New(limiter.Config{
		Max:        maxRequests,
		Expiration: 1 * time.Minute,
		KeyGenerator: generateRateLimitKey,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		},
	})
}

// getMaxRequests returns the maximum number of requests allowed per minute.
// Defaults to 100 if RATE_LIMIT_MAX is not set or invalid.
func getMaxRequests() int {
	maxStr := os.Getenv("RATE_LIMIT_MAX")
	if maxStr == "" {
		return 100
	}

	parsed, err := strconv.Atoi(maxStr)
	if err != nil || parsed <= 0 {
		return 100
	}

	return parsed
}

// generateRateLimitKey generates a unique key for rate limiting.
// Uses user ID if authenticated, otherwise falls back to IP address.
func generateRateLimitKey(c *fiber.Ctx) string {
	// Prefer user ID if available (more accurate for authenticated users)
	if userID := c.Locals("user"); userID != nil {
		if userIDStr, ok := userID.(string); ok && userIDStr != "" {
			return "user:" + userIDStr
		}
	}

	// Fall back to IP address for unauthenticated requests
	// Note: c.IP() returns the real client IP if proxy support is configured
	return c.IP()
}
