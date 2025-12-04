package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Auth middleware validates JWT tokens and attaches user ID to context
func Auth() fiber.Handler {
	jwtSecret := os.Getenv("JWT_SECRET")
	supabaseURL := os.Getenv("SUPABASE_URL")

	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing Authorization header",
			})
		}

		// Check for Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Authorization header format",
			})
		}

		tokenString := parts[1]

		// Parse token without validation first to check algorithm
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
				// HS256
				if jwtSecret == "" {
					return nil, fmt.Errorf("JWT_SECRET not configured")
				}
				return []byte(jwtSecret), nil
			}

			if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
				// RS256 - fetch public key from Supabase
				if supabaseURL == "" {
					return nil, fmt.Errorf("SUPABASE_URL not configured for RS256")
				}
				return getSupabasePublicKey(supabaseURL)
			}

			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
				"details": err.Error(),
			})
		}

		if !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Extract user ID from claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		// Try to get user ID from common claim fields
		var userID string
		if sub, ok := claims["sub"].(string); ok {
			userID = sub
		} else if userIDClaim, ok := claims["user_id"].(string); ok {
			userID = userIDClaim
		} else if id, ok := claims["id"].(string); ok {
			userID = id
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User ID not found in token",
			})
		}

		// Attach user ID to context
		c.Locals("user", userID)

		return c.Next()
	}
}

// getSupabasePublicKey fetches the public key from Supabase JWKS endpoint
func getSupabasePublicKey(supabaseURL string) (*rsa.PublicKey, error) {
	// Fetch JWKS from Supabase
	jwksURL := strings.TrimSuffix(supabaseURL, "/") + "/.well-known/jwks.json"
	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks struct {
		Keys []struct {
			Kty string `json:"kty"`
			Use string `json:"use"`
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	if len(jwks.Keys) == 0 {
		return nil, fmt.Errorf("no keys found in JWKS")
	}

	// Use the first key (typically Supabase has one key)
	key := jwks.Keys[0]

	// Decode base64url encoded modulus and exponent
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert exponent bytes to int
	var eInt int
	for _, b := range eBytes {
		eInt = eInt<<8 | int(b)
	}

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}

	return publicKey, nil
}

