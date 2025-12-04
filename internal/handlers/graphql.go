package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

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

	// Set status code and return response
	statusCode := resp.StatusCode

	// Log 5xx errors
	if statusCode >= 500 {
		log.Printf("ERROR: Supabase returned 5xx error: %d - %s", statusCode, string(respBody))
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

