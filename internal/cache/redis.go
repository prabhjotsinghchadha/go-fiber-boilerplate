package cache

// Package cache provides a simple Redis client for caching data using Upstash.
// Upstash is a serverless Redis service that uses a REST API instead of the traditional Redis protocol.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Client represents a Redis cache client that connects to Upstash via REST API.
type Client struct {
	url    string // Upstash REST API endpoint URL
	token  string // Authentication token (optional)
	client *http.Client
}

var (
	// DefaultClient is the singleton Redis client instance used throughout the application.
	DefaultClient *Client
	
	// defaultTTL is the default time-to-live for cached values (5 minutes).
	defaultTTL = 5 * time.Minute
)

// Init initializes the default Redis client using environment variables.
// Required: UPSTASH_REDIS_URL
// Optional: UPSTASH_REDIS_TOKEN
func Init() error {
	url := os.Getenv("UPSTASH_REDIS_URL")
	if url == "" {
		return fmt.Errorf("UPSTASH_REDIS_URL environment variable is not set")
	}

	token := os.Getenv("UPSTASH_REDIS_TOKEN")
	if token == "" {
		log.Println("WARNING: UPSTASH_REDIS_TOKEN not set, requests may fail")
	}

	// Create the client with a 10-second timeout for HTTP requests
	DefaultClient = &Client{
		url:    url,
		token:  token,
		client: &http.Client{Timeout: 10 * time.Second},
	}

	log.Println("Redis cache client initialized")
	return nil
}

// GetClient returns the default Redis client instance.
// Returns nil if Init() has not been called yet.
func GetClient() *Client {
	return DefaultClient
}

// upstashRequest represents a request to Upstash REST API.
// Upstash accepts Redis commands as JSON arrays.
type upstashRequest struct {
	Command []string `json:"command"` // Redis command as array, e.g., ["SET", "key", "value", "EX", "300"]
}

// upstashResponse represents a response from Upstash REST API.
type upstashResponse struct {
	Result string `json:"result"`        // The result value (for GET commands)
	Error  string `json:"error,omitempty"` // Error message if the command failed
}

// executeCommand sends a Redis command to Upstash and returns the response.
// This is a helper function that handles all the common HTTP request logic.
func (c *Client) executeCommand(command []string) (*upstashResponse, error) {
	if c == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}

	// Step 1: Create the request body with the Redis command
	reqBody := upstashRequest{Command: command}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Step 2: Create HTTP POST request
	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Step 3: Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Step 4: Send the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Upstash: %w", err)
	}
	defer resp.Body.Close()

	// Step 5: Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Upstash API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Step 6: Parse the JSON response
	var upstashResp upstashResponse
	if err := json.NewDecoder(resp.Body).Decode(&upstashResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Step 7: Check for errors in the response
	if upstashResp.Error != "" {
		return nil, fmt.Errorf("Upstash error: %s", upstashResp.Error)
	}

	return &upstashResp, nil
}

// Set stores a value in Redis with the given key and expiration time.
// If ttl is 0, uses the default TTL (5 minutes).
//
// Example: Set("price:123", "45.67", 5*time.Minute)
func (c *Client) Set(key, value string, ttl time.Duration) error {
	// Use default TTL if none provided
	if ttl == 0 {
		ttl = defaultTTL
	}

	// Convert TTL to seconds (Redis EX command expects seconds)
	ttlSeconds := int(ttl.Seconds())

	// Build Redis command: SET key value EX seconds
	// This tells Redis to set the key-value pair and expire it after ttlSeconds
	command := []string{"SET", key, value, "EX", fmt.Sprintf("%d", ttlSeconds)}

	_, err := c.executeCommand(command)
	return err
}

// Get retrieves a value from Redis by key.
// Returns the value and nil error on success.
// Returns empty string and nil error if the key doesn't exist (cache miss).
// Returns error if something went wrong with the request.
//
// Example: value, err := Get("price:123")
func (c *Client) Get(key string) (string, error) {
	// Build Redis command: GET key
	command := []string{"GET", key}

	resp, err := c.executeCommand(command)
	if err != nil {
		// Check if it's a "key not found" error (this is normal for cache misses)
		if err.Error() == "Upstash error: key not found" || err.Error() == "Upstash error: nil" {
			return "", nil // Cache miss - return empty string, no error
		}
		return "", err // Real error
	}

	return resp.Result, nil
}
