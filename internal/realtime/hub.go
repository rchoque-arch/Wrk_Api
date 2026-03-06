package realtime

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Message defines the structure of data sent to clients
type Message struct {
	Type      string      `json:"type"`      // TASK_UPDATED, TASK_CREATED, etc.
	ProjectID string      `json:"projectId"` // Room ID
	Payload   interface{} `json:"payload"`
}

// Client represents a connected websocket user
type Client struct {
	Hub       *Hub
	Conn      *websocket.Conn
	Send      chan Message
	UserID    string
	ProjectID string // The project/room they are currently viewing
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients map: ProjectID -> map[Client]bool
	rooms map[string]map[*Client]bool

	// Inbound messages from the clients (if any)
	broadcast chan Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	mutex sync.RWMutex
}

// GlobalHub is the singleton instance of the realtime Hub.
var GlobalHub = NewHub()

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rooms:      make(map[string]map[*Client]bool),
	}
}

// Register returns the channel used to register clients.
func (h *Hub) Register() chan *Client {
	return h.register
}

// Run starts the Hub main loop to handle registration and broadcasting.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)
		case client := <-h.unregister:
			h.handleUnregister(client)
		case message := <-h.broadcast:
			h.handleBroadcast(message)
		}
	}
}

func (h *Hub) handleRegister(client *Client) {
	h.mutex.Lock()
	if h.rooms[client.ProjectID] == nil {
		h.rooms[client.ProjectID] = make(map[*Client]bool)
	}
	h.rooms[client.ProjectID][client] = true
	h.mutex.Unlock()
	log.Printf("User %s joined room %s", client.UserID, client.ProjectID)
}

func (h *Hub) handleUnregister(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	clients, ok := h.rooms[client.ProjectID]
	if !ok {
		return
	}
	if _, exists := clients[client]; exists {
		delete(clients, client)
		close(client.Send)
		log.Printf("User %s left room %s", client.UserID, client.ProjectID)
	}
	if len(clients) == 0 {
		delete(h.rooms, client.ProjectID)
	}
}

func (h *Hub) handleBroadcast(message Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	clients := h.rooms[message.ProjectID]
	for client := range clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(clients, client)
		}
	}
}

// BroadcastEvent is a helper to send notifications from REST handlers
func (h *Hub) BroadcastEvent(projectID string, eventType string, payload interface{}) {
	msg := Message{
		Type:      eventType,
		ProjectID: projectID,
		Payload:   payload,
	}
	h.broadcast <- msg
}
