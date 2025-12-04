package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"boilerplate/internal/cache"

	"github.com/gofiber/fiber/v2"
)

// GraphQLProxy forwards GraphQL requests to Supabase's GraphQL endpoint.
// It preserves the request method, body, and headers (especially Authorization)
// and returns the response from Supabase.
func GraphQLProxy(c *fiber.Ctx) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		log.Println("ERROR: SUPABASE_URL environment variable is not set")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "GraphQL proxy configuration error",
		})
	}

	// Build the target URL
	targetURL := strings.TrimSuffix(supabaseURL, "/") + "/graphql/v1"

	// Read the request body
	body := c.Body()
	if body == nil {
		body = []byte{}
	}

	// Create a new request to Supabase
	req, err := http.NewRequest(c.Method(), targetURL, bytes.NewReader(body))
	if err != nil {
		log.Printf("ERROR: Failed to create request to Supabase: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create proxy request",
		})
	}

	// Copy all headers from the original request
	c.Request().Header.VisitAll(func(key, value []byte) {
		keyStr := string(key)
		valueStr := string(value)
		// Skip hop-by-hop headers that shouldn't be forwarded
		if !isHopByHopHeader(keyStr) {
			req.Header.Set(keyStr, valueStr)
		}
	})

	// Make the request to Supabase
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR: Failed to proxy request to Supabase: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Failed to connect to Supabase",
		})
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read response from Supabase: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Failed to read response from Supabase",
		})
	}

	// Copy response headers (excluding hop-by-hop headers)
	copyResponseHeaders(c, resp)

	// Set status code
	statusCode := resp.StatusCode

	// Log 5xx errors
	if statusCode >= 500 {
		log.Printf("ERROR: Supabase returned 5xx error: %d - %s", statusCode, string(respBody))
	}

	// Inject cached prices if query requests currentPrice
	if statusCode == http.StatusOK && strings.Contains(string(body), "currentPrice") {
		respBody = injectCachedPrices(body, respBody)
	}

	return c.Status(statusCode).Send(respBody)
}

// isHopByHopHeader checks if a header is a hop-by-hop header that shouldn't be forwarded.
func isHopByHopHeader(headerName string) bool {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	headerLower := strings.ToLower(headerName)
	for _, hopHeader := range hopByHopHeaders {
		if strings.ToLower(hopHeader) == headerLower {
			return true
		}
	}
	return false
}

// copyResponseHeaders copies response headers from the Supabase response to the Fiber context.
func copyResponseHeaders(c *fiber.Ctx, resp *http.Response) {
	for key, values := range resp.Header {
		// Skip hop-by-hop headers
		if !isHopByHopHeader(key) {
			for _, value := range values {
				c.Set(key, value)
			}
		}
	}
}

// graphQLRequest represents a GraphQL request body.
type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// graphQLResponse represents a GraphQL response body.
type graphQLResponse struct {
	Data   interface{}   `json:"data,omitempty"`
	Errors []interface{} `json:"errors,omitempty"`
}

// parseGraphQLQuery parses a GraphQL request body and extracts the query string.
// Returns the query string, or empty string if parsing fails.
func parseGraphQLQuery(queryBody []byte) string {
	var req graphQLRequest
	if err := json.Unmarshal(queryBody, &req); err != nil {
		log.Printf("WARNING: Failed to parse GraphQL query: %v", err)
		return ""
	}
	return req.Query
}

// findArtistIDsInResponse extracts artist IDs from the GraphQL response data.
// It looks for an "artists" array and extracts the "id" field from each artist.
// Returns a map of artist ID to their position in the response.
func findArtistIDsInResponse(responseData map[string]interface{}) map[string]int {
	artistIDs := make(map[string]int)

	// Look for the "artists" array in the response
	artists, ok := responseData["artists"].([]interface{})
	if !ok {
		return artistIDs // No artists found
	}

	// Extract ID from each artist
	for i, artist := range artists {
		artistMap, ok := artist.(map[string]interface{})
		if !ok {
			continue
		}

		id, ok := artistMap["id"].(string)
		if ok {
			artistIDs[id] = i // Store the index for later use
		}
	}

	return artistIDs
}

// getCachedPrices fetches cached prices from Redis for the given artist IDs.
// Returns a map of artist ID to cached price.
// Only includes prices that were found in the cache.
func getCachedPrices(artistIDs []string) map[string]float64 {
	redisClient := cache.GetClient()
	if redisClient == nil {
		return nil // No cache available
	}

	cachedPrices := make(map[string]float64)

	// Try to get each artist's price from cache
	for _, artistID := range artistIDs {
		cacheKey := "price:" + artistID
		cachedValue, err := redisClient.Get(cacheKey)
		
		// If we got a value and no error, parse it as a float
		if err == nil && cachedValue != "" {
			if price, err := strconv.ParseFloat(cachedValue, 64); err == nil {
				cachedPrices[artistID] = price
				log.Printf("Cache hit for artist %s: %.2f", artistID, price)
			}
		}
	}

	return cachedPrices
}

// injectPriceIntoArtist injects a cached price into an artist object in the response.
// Modifies the artist map in place by adding a "currentPrice" field.
func injectPriceIntoArtist(artistMap map[string]interface{}, price float64) {
	artistMap["currentPrice"] = price
}

// injectCachedPrices is the main function that injects cached prices into a GraphQL response.
// Flow:
//   1. Check if cache is available
//   2. Parse the GraphQL query to see if it requests currentPrice
//   3. Parse the GraphQL response
//   4. Find artist IDs in the response
//   5. Get cached prices from Redis
//   6. Inject cached prices into the response
//   7. Return the modified response
func injectCachedPrices(queryBody, responseBody []byte) []byte {
	// Step 1: Check if cache is available
	redisClient := cache.GetClient()
	if redisClient == nil {
		return responseBody // No cache, return original response
	}

	// Step 2: Check if the query requests currentPrice
	// If not, we don't need to do anything
	query := parseGraphQLQuery(queryBody)
	if !strings.Contains(query, "currentPrice") {
		return responseBody // Query doesn't request price, return original
	}

	// Step 3: Parse the GraphQL response
	var resp graphQLResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		log.Printf("WARNING: Failed to parse GraphQL response: %v", err)
		return responseBody
	}

	// Step 4: Don't modify responses with errors
	if len(resp.Errors) > 0 {
		return responseBody
	}

	// Step 5: Extract response data
	respData, ok := resp.Data.(map[string]interface{})
	if !ok {
		return responseBody // Unexpected response structure
	}

	// Step 6: Find artist IDs in the response
	// This gives us a map of artist ID -> index in the artists array
	artistIDMap := findArtistIDsInResponse(respData)
	if len(artistIDMap) == 0 {
		return responseBody // No artists found
	}

	// Step 7: Get cached prices for all artist IDs
	artistIDs := make([]string, 0, len(artistIDMap))
	for id := range artistIDMap {
		artistIDs = append(artistIDs, id)
	}
	cachedPrices := getCachedPrices(artistIDs)
	if len(cachedPrices) == 0 {
		return responseBody // No cache hits
	}

	// Step 8: Inject cached prices into the response
	artists, ok := respData["artists"].([]interface{})
	if !ok {
		return responseBody // Shouldn't happen, but be safe
	}

	// For each artist that has a cached price, inject it
	for artistID, price := range cachedPrices {
		index, found := artistIDMap[artistID]
		if !found {
			continue // Artist not in response (shouldn't happen)
		}

		artistMap, ok := artists[index].(map[string]interface{})
		if !ok {
			continue // Invalid artist object
		}

		// Inject the cached price
		injectPriceIntoArtist(artistMap, price)
		log.Printf("Injected cached price for artist %s: %.2f", artistID, price)
	}

	// Step 9: Marshal the modified response back to JSON
	modifiedBody, err := json.Marshal(resp)
	if err != nil {
		log.Printf("WARNING: Failed to create modified response: %v", err)
		return responseBody
	}

	return modifiedBody
}

