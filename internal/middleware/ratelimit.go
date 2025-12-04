package middleware

import (
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimit middleware applies rate limiting based on user ID (if authenticated) or IP address
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
			// If user is authenticated, use user ID as key
			if userID := c.Locals("user"); userID != nil {
				if userIDStr, ok := userID.(string); ok && userIDStr != "" {
					return "user:" + userIDStr
				}
			}
			// Otherwise, use IP address
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		},
	})
}

