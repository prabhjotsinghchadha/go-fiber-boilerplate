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
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Cache for storing RSA public keys from Supabase JWKS
var (
	cachedKeys    = make(map[string]*rsa.PublicKey)
	cacheExpiries = make(map[string]time.Time)
	cacheMu       sync.RWMutex
	cacheTTL      = 1 * time.Hour
)

// Auth validates JWT tokens and attaches the user ID to the request context.
// Supports both HS256 (symmetric) and RS256 (asymmetric) signing methods.
func Auth() fiber.Handler {
	jwtSecret := os.Getenv("JWT_SECRET")
	supabaseURL := os.Getenv("SUPABASE_URL")

	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		tokenString, err := extractTokenFromHeader(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Validate token and get claims
		claims, err := validateToken(tokenString, jwtSecret, supabaseURL)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication failed",
			})
		}

		// Extract user ID from claims
		userID, err := extractUserIDFromClaims(claims)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Attach user ID to context for use in handlers
		c.Locals("user", userID)
		return c.Next()
	}
}

// extractTokenFromHeader extracts the JWT token from the Authorization header.
// Returns an error if the header is missing or malformed.
func extractTokenFromHeader(c *fiber.Ctx) (string, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	return parts[1], nil
}

// validateToken parses and validates a JWT token.
// Returns the token claims if valid, or an error if validation fails.
func validateToken(tokenString, jwtSecret, supabaseURL string) (jwt.MapClaims, error) {
	// Parse token header to extract kid for RS256 key matching
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}

	// Extract kid from token header for RS256
	kid := extractKidFromToken(token)

	// Parse and validate token with appropriate key
	token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return getSigningKey(token, jwtSecret, supabaseURL, kid)
	})
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// extractKidFromToken extracts the key ID (kid) from the JWT token header.
// Returns empty string if kid is not present.
func extractKidFromToken(token *jwt.Token) string {
	if kidValue, ok := token.Header["kid"].(string); ok {
		return kidValue
	}
	return ""
}

// getSigningKey returns the appropriate signing key based on the token's algorithm.
// For HS256, returns the JWT secret. For RS256, fetches the public key from Supabase.
func getSigningKey(token *jwt.Token, jwtSecret, supabaseURL, kid string) (interface{}, error) {
	// Handle HS256 (symmetric) tokens
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
		if jwtSecret == "" {
			return nil, fmt.Errorf("JWT_SECRET not configured")
		}
		return []byte(jwtSecret), nil
	}

	// Handle RS256 (asymmetric) tokens
	if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
		if supabaseURL == "" {
			return nil, fmt.Errorf("SUPABASE_URL not configured for RS256")
		}
		if kid == "" {
			return nil, fmt.Errorf("kid header missing from token")
		}
		return getSupabasePublicKey(supabaseURL, kid)
	}

	return nil, fmt.Errorf("unsupported signing method: %v", token.Header["alg"])
}

// extractUserIDFromClaims extracts the user ID from JWT claims.
// Checks common claim fields: sub, user_id, and id.
func extractUserIDFromClaims(claims jwt.MapClaims) (string, error) {
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		return sub, nil
	}
	if userID, ok := claims["user_id"].(string); ok && userID != "" {
		return userID, nil
	}
	if id, ok := claims["id"].(string); ok && id != "" {
		return id, nil
	}
	return "", fmt.Errorf("user ID not found in token")
}

// getSupabasePublicKey fetches and caches RSA public keys from Supabase JWKS endpoint.
// Uses a simple cache with TTL to avoid fetching keys on every request.
func getSupabasePublicKey(supabaseURL, kid string) (*rsa.PublicKey, error) {
	// Check cache first
	if key := getCachedKey(kid); key != nil {
		return key, nil
	}

	// Cache miss - fetch from Supabase
	key, err := fetchAndCacheKey(supabaseURL, kid)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// getCachedKey retrieves a cached public key if it exists and hasn't expired.
func getCachedKey(kid string) *rsa.PublicKey {
	cacheMu.RLock()
	defer cacheMu.RUnlock()

	key, exists := cachedKeys[kid]
	if !exists {
		return nil
	}

	expiry, exists := cacheExpiries[kid]
	if !exists || time.Now().After(expiry) {
		return nil
	}

	return key
}

// fetchAndCacheKey fetches a public key from Supabase JWKS and caches it.
func fetchAndCacheKey(supabaseURL, kid string) (*rsa.PublicKey, error) {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	// Double-check cache after acquiring write lock
	if key := getCachedKeyUnsafe(kid); key != nil {
		return key, nil
	}

	// Fetch JWKS from Supabase
	jwks, err := fetchJWKS(supabaseURL)
	if err != nil {
		return nil, err
	}

	// Find the key matching the kid
	keyData := findKeyByKid(jwks, kid)
	if keyData == nil {
		return nil, fmt.Errorf("key with kid '%s' not found in JWKS", kid)
	}

	// Build RSA public key from JWKS data
	publicKey, err := buildRSAPublicKey(keyData)
	if err != nil {
		return nil, err
	}

	// Cache the key
	cachedKeys[kid] = publicKey
	cacheExpiries[kid] = time.Now().Add(cacheTTL)

	return publicKey, nil
}

// getCachedKeyUnsafe retrieves a cached key without locking (caller must hold lock).
func getCachedKeyUnsafe(kid string) *rsa.PublicKey {
	key, exists := cachedKeys[kid]
	if !exists {
		return nil
	}

	expiry, exists := cacheExpiries[kid]
	if !exists || time.Now().After(expiry) {
		return nil
	}

	return key
}

// jwksKey represents a single key in the JWKS response.
type jwksKey struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// jwksResponse represents the JWKS response structure.
type jwksResponse struct {
	Keys []jwksKey `json:"keys"`
}

// fetchJWKS fetches the JWKS from Supabase.
func fetchJWKS(supabaseURL string) (*jwksResponse, error) {
	jwksURL := strings.TrimSuffix(supabaseURL, "/") + "/.well-known/jwks.json"

	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	if len(jwks.Keys) == 0 {
		return nil, fmt.Errorf("no keys found in JWKS")
	}

	return &jwks, nil
}

// findKeyByKid finds a key in the JWKS response matching the given kid.
func findKeyByKid(jwks *jwksResponse, kid string) *jwksKey {
	for i := range jwks.Keys {
		if jwks.Keys[i].Kid == kid {
			return &jwks.Keys[i]
		}
	}
	return nil
}

// buildRSAPublicKey constructs an RSA public key from JWKS key data.
func buildRSAPublicKey(keyData *jwksKey) (*rsa.PublicKey, error) {
	// Decode base64url encoded modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(keyData.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode base64url encoded exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(keyData.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert exponent bytes to int
	var eInt int
	for _, b := range eBytes {
		eInt = eInt<<8 | int(b)
	}

	// Create RSA public key
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}, nil
}
