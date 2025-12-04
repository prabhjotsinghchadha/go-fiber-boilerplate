package handlers

// Package handlers provides WebSocket functionality for real-time communication.
// The Hub pattern is used to manage multiple WebSocket connections and broadcast messages to all clients.

import (
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Hub is the central manager for all WebSocket connections.
// It uses channels to safely handle client registration, unregistration, and message broadcasting
// from multiple goroutines (threads).
//
// Architecture:
//   - clients: Map of all active WebSocket connections
//   - register: Channel for new clients to join
//   - unregister: Channel for clients to leave
//   - broadcast: Channel for messages to send to all clients
//   - mu: Mutex (lock) to prevent race conditions when accessing the clients map
type Hub struct {
	// clients stores all active WebSocket connections.
	// The boolean value is just a placeholder (we only care about the keys).
	clients map[*websocket.Conn]bool

	// broadcast is a channel that receives messages to send to all connected clients.
	// When a message is sent here, the hub will forward it to every client.
	broadcast chan []byte

	// register is a channel for new clients to join the hub.
	// When a client connects, it sends itself through this channel.
	register chan *websocket.Conn

	// unregister is a channel for clients to leave the hub.
	// When a client disconnects, it sends itself through this channel.
	unregister chan *websocket.Conn

	// mu is a read-write mutex to safely access the clients map from multiple goroutines.
	// This prevents race conditions (data corruption) when multiple threads access the map at once.
	mu sync.RWMutex
}

var (
	// DefaultHub is the singleton WebSocket hub instance used throughout the application.
	// Only one hub exists, and all WebSocket connections use it.
	DefaultHub *Hub
)

// InitHub creates and starts the default WebSocket hub.
// This should be called once when the application starts.
func InitHub() {
	// Create a new hub with empty clients map and channels
	DefaultHub = &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte, 256), // Buffer up to 256 messages
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}

	// Start the hub's main loop in a separate goroutine (background thread)
	// This loop runs forever, handling client connections and message broadcasting
	go DefaultHub.Run()

	log.Println("WebSocket hub initialized")
}

// GetHub returns the default WebSocket hub instance.
func GetHub() *Hub {
	return DefaultHub
}

// Run is the hub's main event loop that runs forever.
// It listens for three types of events:
//   1. New clients registering (joining)
//   2. Clients unregistering (leaving)
//   3. Messages to broadcast to all clients
//
// This function runs in a separate goroutine and blocks forever.
func (h *Hub) Run() {
	for {
		// The select statement waits for one of these events to happen
		select {
		// Case 1: A new client wants to join
		case conn := <-h.register:
			// Lock the clients map before modifying it (thread safety)
			h.mu.Lock()
			h.clients[conn] = true // Add the new client
			h.mu.Unlock()
			log.Printf("WebSocket client connected. Total clients: %d", len(h.clients))

		// Case 2: A client wants to leave
		case conn := <-h.unregister:
			// Lock the clients map before modifying it
			h.mu.Lock()
			if _, exists := h.clients[conn]; exists {
				delete(h.clients, conn) // Remove the client
				conn.Close()            // Close the WebSocket connection
				log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))
			}
			h.mu.Unlock()

		// Case 3: A message needs to be broadcast to all clients
		case message := <-h.broadcast:
			// Use read lock since we're only reading from the map
			h.mu.RLock()
			// Send the message to every connected client
			for conn := range h.clients {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					// If we can't send to a client, they're probably disconnected
					log.Printf("Error sending message to client: %v", err)
					// Remove the broken connection (we'll clean it up on next unregister)
					delete(h.clients, conn)
					conn.Close()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected WebSocket clients.
// This is the main way to send real-time updates to all connected users.
//
// Example: hub.Broadcast([]byte(`{"artist_id": "123", "price": 45.67}`))
func (h *Hub) Broadcast(message []byte) {
	if h == nil {
		return // Hub not initialized, ignore
	}

	// Try to send the message to the broadcast channel
	// If the channel is full, drop the message (non-blocking)
	select {
	case h.broadcast <- message:
		// Message sent successfully, hub will broadcast it
	default:
		// Channel is full, drop this message to prevent blocking
		log.Println("Broadcast channel full, dropping message")
	}
}

// WebSocketHandler handles individual WebSocket connections.
// This function is called by Fiber for each new WebSocket connection.
//
// Flow:
//   1. Client connects via WebSocket
//   2. Register client with the hub
//   3. Listen for messages from the client
//   4. When client disconnects, unregister them
func WebSocketHandler(c *websocket.Conn) {
	// Get the hub instance
	hub := GetHub()
	if hub == nil {
		log.Println("ERROR: WebSocket hub not initialized")
		c.Close()
		return
	}

	// Step 1: Register this client with the hub
	// This adds the client to the hub's clients map
	hub.register <- c

	// Step 2: Make sure we unregister when this function exits (client disconnects)
	// The defer statement runs this code when the function ends
	defer func() {
		hub.unregister <- c
	}()

	// Step 3: Listen for messages from this client
	// This loop runs until the client disconnects
	for {
		// Read a message from the client
		messageType, msg, err := c.ReadMessage()
		if err != nil {
			// Client disconnected or error occurred
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break // Exit the loop, which will trigger the defer (unregister)
		}

		// For now, just echo the message back to the client
		// TODO: Later, parse JSON messages like {"subscribe": "prices:artist123"}
		//       to allow clients to subscribe to specific updates
		if messageType == websocket.TextMessage {
			log.Printf("Received message from client: %s", string(msg))
			
			// Echo the message back to the client
			if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("Error writing message: %v", err)
				break // Exit if we can't write
			}
		}
	}
}

// UpgradeWebSocket is a middleware function that checks if an HTTP request
// is trying to upgrade to a WebSocket connection.
// This is required by Fiber to handle WebSocket upgrades.
func UpgradeWebSocket(c *fiber.Ctx) error {
	// Check if this is a WebSocket upgrade request
	if websocket.IsWebSocketUpgrade(c) {
		// Allow the request to proceed to the WebSocket handler
		c.Locals("allowed", true)
		return c.Next()
	}
	// Not a WebSocket request, return an error
	return fiber.ErrUpgradeRequired
}
