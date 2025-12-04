package realtime

// Package realtime handles Supabase Realtime subscriptions to listen for database changes.
// When prices change in the artist_metrics table, we cache them in Redis and broadcast to WebSocket clients.

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"boilerplate/internal/cache"
	"boilerplate/internal/handlers"

	"github.com/gorilla/websocket"
)

// PriceUpdate represents a price update from the artist_metrics table.
// This is the message format we send to WebSocket clients.
type PriceUpdate struct {
	ArtistID string  `json:"artist_id"` // The ID of the artist
	Price    float64 `json:"price"`     // The new price
	Event    string  `json:"event"`     // Type of change: INSERT, UPDATE, or DELETE
}

// Init initializes the Supabase Realtime client.
// Currently a placeholder - no initialization needed for raw WebSocket approach.
func Init() error {
	return nil
}

// SubscribeToPrices is the main entry point for subscribing to price updates.
// It sets up the connection to Supabase Realtime and listens for changes.
func SubscribeToPrices() {
	// Step 1: Get configuration from environment variables
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		log.Println("WARNING: SUPABASE_URL or SUPABASE_ANON_KEY not set, skipping Realtime subscription")
		return
	}

	// Step 2: Make sure Redis cache is initialized
	if cache.GetClient() == nil {
		if err := cache.Init(); err != nil {
			log.Printf("ERROR: Failed to initialize Redis cache: %v", err)
			// Continue without cache - we can still broadcast updates
		}
	}

	// Step 3: Make sure WebSocket hub is initialized
	if handlers.GetHub() == nil {
		handlers.InitHub()
	}

	// Step 4: Start the WebSocket subscription
	subscribeViaWebSocket(supabaseURL, supabaseKey)
}

// buildRealtimeURL converts a Supabase HTTP URL to a WebSocket URL for Realtime.
// Example: https://xxx.supabase.co -> wss://xxx.supabase.co/realtime/v1/websocket
func buildRealtimeURL(supabaseURL string) string {
	// Replace https:// with wss:// (WebSocket Secure) or http:// with ws://
	realtimeURL := strings.Replace(supabaseURL, "https://", "wss://", 1)
	realtimeURL = strings.Replace(realtimeURL, "http://", "ws://", 1)
	
	// Add the Realtime WebSocket endpoint path
	realtimeURL = strings.TrimSuffix(realtimeURL, "/") + "/realtime/v1/websocket"
	
	return realtimeURL
}

// connectToRealtime establishes a WebSocket connection to Supabase Realtime.
// Returns the connection and the full URL with API key, or an error.
func connectToRealtime(supabaseURL, supabaseKey string) (*websocket.Conn, string, error) {
	// Step 1: Build the WebSocket URL
	realtimeURL := buildRealtimeURL(supabaseURL)

	// Step 2: Parse the URL and add the API key as a query parameter
	// Supabase requires the API key in the URL for authentication
	u, err := url.Parse(realtimeURL)
	if err != nil {
		return nil, "", err
	}

	// Add API key to URL query parameters
	q := u.Query()
	q.Set("apikey", supabaseKey)
	u.RawQuery = q.Encode()
	fullURL := u.String()

	log.Printf("Connecting to Supabase Realtime at %s", fullURL)

	// Step 3: Dial (connect) to the WebSocket server
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(fullURL, nil)
	if err != nil {
		return nil, "", err
	}

	log.Println("Connected to Supabase Realtime successfully")
	return conn, fullURL, nil
}

// subscribeToTable sends a subscription message to Supabase to listen for changes
// on a specific database table.
func subscribeToTable(conn *websocket.Conn, tableName string) error {
	// Build the subscription message
	// Supabase uses Phoenix channels protocol - "phx_join" means "join this channel"
	subscribeMsg := map[string]interface{}{
		"topic":   "realtime:public:" + tableName, // Channel name: realtime:public:artist_metrics
		"event":   "phx_join",                      // Event type: join the channel
		"payload": map[string]interface{}{},        // Empty payload for join
		"ref":     "1",                             // Reference ID for this message
	}

	// Send the subscription message as JSON
	if err := conn.WriteJSON(subscribeMsg); err != nil {
		return err
	}

	log.Printf("Subscribed to table: %s", tableName)
	return nil
}

// extractPriceFromRecord extracts the artist_id and price from a database record.
// Returns empty values if the record doesn't have the expected structure.
func extractPriceFromRecord(record map[string]interface{}) (artistID string, price float64, ok bool) {
	// Try to get artist_id
	artistIDValue, found := record["artist_id"]
	if !found {
		return "", 0, false
	}

	// Convert artist_id to string (it might be different types)
	artistID, ok = artistIDValue.(string)
	if !ok {
		return "", 0, false
	}

	// Try to get price
	priceValue, found := record["price"]
	if !found {
		return "", 0, false
	}

	// Convert price to float64 (handle different number types)
	switch v := priceValue.(type) {
	case float64:
		price = v
	case float32:
		price = float64(v)
	case int:
		price = float64(v)
	case int64:
		price = float64(v)
	default:
		return "", 0, false // Unknown type
	}

	return artistID, price, true
}

// handlePriceUpdate processes a price update from Supabase Realtime.
// It caches the price in Redis and broadcasts it to all WebSocket clients.
func handlePriceUpdate(payload map[string]interface{}) {
	// Step 1: Extract the event type (INSERT, UPDATE, or DELETE)
	eventType := ""
	if evt, ok := payload["eventType"].(string); ok {
		eventType = evt
	} else if evt, ok := payload["event"].(string); ok {
		eventType = evt
	} else {
		log.Println("WARNING: Could not find event type in payload")
		return
	}

	// Step 2: Only process INSERT and UPDATE events (ignore DELETE)
	if eventType != "INSERT" && eventType != "UPDATE" {
		return
	}

	// Step 3: Extract the new record data
	// Supabase sends the new record in the "new" field (or sometimes "record")
	var newRecord map[string]interface{}
	if new, ok := payload["new"].(map[string]interface{}); ok {
		newRecord = new
	} else if record, ok := payload["record"].(map[string]interface{}); ok {
		newRecord = record
	} else {
		log.Println("WARNING: Could not find record data in payload")
		return
	}

	// Step 4: Extract artist_id and price from the record
	artistID, price, ok := extractPriceFromRecord(newRecord)
	if !ok {
		log.Println("WARNING: Could not extract artist_id or price from record")
		return
	}

	// Step 5: Cache the price in Redis
	// Cache key format: "price:artist123"
	redisClient := cache.GetClient()
	if redisClient != nil {
		cacheKey := "price:" + artistID
		priceString := formatPrice(price)
		
		if err := redisClient.Set(cacheKey, priceString, 5*time.Minute); err != nil {
			log.Printf("ERROR: Failed to cache price in Redis: %v", err)
		} else {
			log.Printf("Cached price for artist %s: %.2f", artistID, price)
		}
	}

	// Step 6: Create the update message for WebSocket clients
	update := PriceUpdate{
		ArtistID: artistID,
		Price:    price,
		Event:    eventType,
	}

	// Step 7: Broadcast the update to all connected WebSocket clients
	hub := handlers.GetHub()
	if hub != nil {
		message, err := json.Marshal(update)
		if err != nil {
			log.Printf("ERROR: Failed to create message: %v", err)
			return
		}
		
		hub.Broadcast(message)
		log.Printf("Broadcasted price update: artist_id=%s, price=%.2f", artistID, price)
	}
}

// formatPrice converts a float64 price to a string for storage in Redis.
func formatPrice(price float64) string {
	// Use JSON marshaling to ensure consistent formatting
	b, _ := json.Marshal(price)
	return string(b)
}

// listenForUpdates listens for messages from Supabase Realtime and processes them.
// This function runs in a loop until the connection is closed.
func listenForUpdates(conn *websocket.Conn, supabaseURL, supabaseKey string) {
	log.Println("Listening for database changes...")

	for {
		// Read a message from the WebSocket connection
		var message map[string]interface{}
		if err := conn.ReadJSON(&message); err != nil {
			log.Printf("ERROR: Connection lost: %v", err)
			log.Println("Attempting to reconnect in 5 seconds...")
			
			// Wait 5 seconds before reconnecting
			time.Sleep(5 * time.Second)
			
			// Reconnect in a new goroutine (don't block)
			go subscribeViaWebSocket(supabaseURL, supabaseKey)
			return
		}

		// Check what type of message we received
		event, _ := message["event"].(string)
		
		// If it's a database change event, process it
		if event == "postgres_changes" {
			payload, ok := message["payload"].(map[string]interface{})
			if ok {
				handlePriceUpdate(payload)
			}
		}
		// Other events (like "phx_reply" for subscription confirmation) are ignored
	}
}

// subscribeViaWebSocket is the main function that orchestrates the Realtime subscription.
// It connects, subscribes, and listens for updates.
func subscribeViaWebSocket(supabaseURL, supabaseKey string) {
	// Step 1: Connect to Supabase Realtime WebSocket
	conn, _, err := connectToRealtime(supabaseURL, supabaseKey)
	if err != nil {
		log.Printf("ERROR: Failed to connect to Supabase Realtime: %v", err)
		log.Println("Please ensure:")
		log.Println("  1. SUPABASE_URL and SUPABASE_ANON_KEY are set correctly")
		log.Println("  2. Supabase Realtime is enabled for the artist_metrics table")
		return
	}
	defer conn.Close() // Make sure we close the connection when done

	// Step 2: Subscribe to the artist_metrics table
	if err := subscribeToTable(conn, "artist_metrics"); err != nil {
		log.Printf("ERROR: Failed to subscribe to table: %v", err)
		return
	}

	// Step 3: Start listening for updates (this blocks forever)
	listenForUpdates(conn, supabaseURL, supabaseKey)
}
