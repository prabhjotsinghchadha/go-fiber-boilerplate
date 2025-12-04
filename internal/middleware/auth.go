package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var (
	cachedKeys   = make(map[string]*rsa.PublicKey)
	cacheExpiries = make(map[string]time.Time)
	cachedKeysMu sync.RWMutex
	cacheTTL     = 1 * time.Hour // Default cache TTL
	gracePeriod  = 5 * time.Minute // Grace period to use stale cache on fetch failure
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

		// Parse token header to extract kid for RS256 key matching
		parser := jwt.NewParser()
		token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			// Log the full error server-side with context
			requestID := c.Locals("requestid")
			if requestID == nil {
				requestID = c.IP()
			}
			log.Printf("ERROR: Token parsing failed: %v | RequestID: %v | IP: %s", err, requestID, c.IP())
			
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Extract kid from token header for RS256 key matching
		var kid string
		if kidValue, ok := token.Header["kid"].(string); ok {
			kid = kidValue
		}

		// Parse token with validation
		token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
				// HS256
				if jwtSecret == "" {
					return nil, fmt.Errorf("JWT_SECRET not configured")
				}
				return []byte(jwtSecret), nil
			}

			if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
				// RS256 - fetch public key from Supabase using kid
				if supabaseURL == "" {
					return nil, fmt.Errorf("SUPABASE_URL not configured for RS256")
				}
				if kid == "" {
					return nil, fmt.Errorf("kid header missing from token")
				}
				return getSupabasePublicKey(supabaseURL, kid)
			}

			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		})

		if err != nil {
			// Log the full error server-side with context
			requestID := c.Locals("requestid")
			if requestID == nil {
				requestID = c.IP()
			}
			// Try to extract user info from token if available (even if invalid)
			userInfo := "unknown"
			if token != nil {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if sub, ok := claims["sub"].(string); ok {
						userInfo = sub
					}
				}
			}
			log.Printf("ERROR: Token validation failed: %v | RequestID: %v | IP: %s | User: %s", err, requestID, c.IP(), userInfo)
			
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication failed",
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

// getSupabasePublicKey fetches the public key from Supabase JWKS endpoint matching the provided kid
// Uses a package-level cache to avoid fetching JWKS on every call
func getSupabasePublicKey(supabaseURL, kid string) (*rsa.PublicKey, error) {
	now := time.Now()
	
	// First, try to read from cache with read lock
	cachedKeysMu.RLock()
	if key, ok := cachedKeys[kid]; ok {
		if expiry, ok := cacheExpiries[kid]; ok && now.Before(expiry) {
			cachedKeysMu.RUnlock()
			return key, nil
		}
	}
	cachedKeysMu.RUnlock()

	// Cache miss or expired, acquire write lock
	cachedKeysMu.Lock()
	defer cachedKeysMu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have updated it)
	if key, ok := cachedKeys[kid]; ok {
		if expiry, ok := cacheExpiries[kid]; ok && now.Before(expiry) {
			return key, nil
		}
	}

	// Check if we have a stale cache within grace period for the same kid
	var staleKey *rsa.PublicKey
	if key, ok := cachedKeys[kid]; ok {
		if expiry, ok := cacheExpiries[kid]; ok && now.Before(expiry.Add(gracePeriod)) {
			staleKey = key
		}
	}

	// Fetch JWKS from Supabase
	jwksURL := strings.TrimSuffix(supabaseURL, "/") + "/.well-known/jwks.json"
	resp, err := http.Get(jwksURL)
	if err != nil {
		// If fetch fails but we have a stale key within grace period for the same kid, use it
		if staleKey != nil {
			log.Printf("WARN: JWKS fetch failed, using stale cached key for kid '%s': %v", kid, err)
			return staleKey, nil
		}
		// Don't update cache on error - return error and keep existing cache (if any)
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If fetch fails but we have a stale key within grace period for the same kid, use it
		if staleKey != nil {
			log.Printf("WARN: JWKS fetch returned status %d, using stale cached key for kid '%s'", resp.StatusCode, kid)
			return staleKey, nil
		}
		// Don't update cache on error - return error and keep existing cache (if any)
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
		// If decode fails but we have a stale key within grace period for the same kid, use it
		if staleKey != nil {
			log.Printf("WARN: JWKS decode failed, using stale cached key for kid '%s': %v", kid, err)
			return staleKey, nil
		}
		// Don't update cache on error - return error and keep existing cache (if any)
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	if len(jwks.Keys) == 0 {
		// If no keys but we have a stale key within grace period for the same kid, use it
		if staleKey != nil {
			log.Printf("WARN: No keys found in JWKS, using stale cached key for kid '%s'", kid)
			return staleKey, nil
		}
		// Don't update cache on error - return error and keep existing cache (if any)
		return nil, fmt.Errorf("no keys found in JWKS")
	}

	// Find the key matching the kid from the token header
	var key *struct {
		Kty string `json:"kty"`
		Use string `json:"use"`
		Kid string `json:"kid"`
		N   string `json:"n"`
		E   string `json:"e"`
	}
	for i := range jwks.Keys {
		if jwks.Keys[i].Kid == kid {
			key = &jwks.Keys[i]
			break
		}
	}

	if key == nil {
		// Kid not found in JWKS - this is not a transient error, so never use stale fallback
		// The requested kid is missing, which means the key rotation may have occurred
		// and we should not use a stale key for a different kid
		log.Printf("ERROR: Key with kid '%s' not found in JWKS", kid)
		return nil, fmt.Errorf("key with kid '%s' not found in JWKS", kid)
	}

	// Decode base64url encoded modulus and exponent
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		// If decode fails but we have a stale key within grace period for the same kid, use it
		if staleKey != nil {
			log.Printf("WARN: Failed to decode modulus for kid '%s', using stale cached key: %v", kid, err)
			return staleKey, nil
		}
		// Don't update cache on error - return error and keep existing cache (if any)
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		// If decode fails but we have a stale key within grace period for the same kid, use it
		if staleKey != nil {
			log.Printf("WARN: Failed to decode exponent for kid '%s', using stale cached key: %v", kid, err)
			return staleKey, nil
		}
		// Don't update cache on error - return error and keep existing cache (if any)
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

	// Successfully fetched and built the key - update cache
	cachedKeys[kid] = publicKey
	cacheExpiries[kid] = now.Add(cacheTTL)

	return publicKey, nil
}

