package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production restrict origins
	},
}

type Client struct {
	ID         uint
	Role       string // "user" or "merchant"
	Conn       *websocket.Conn
	Send       chan []byte
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Manager struct {
	mu             sync.RWMutex
	merchantClients map[uint]*Client // merchantID -> client
	userClients     map[uint]*Client // userID -> client
}

func NewManager() *Manager {
	return &Manager{
		merchantClients: make(map[uint]*Client),
		userClients:     make(map[uint]*Client),
	}
}

// RegisterMerchant upgrades the HTTP connection for a merchant
func (m *Manager) RegisterMerchant(w http.ResponseWriter, r *http.Request, merchantID uint) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	client := &Client{
		ID:   merchantID,
		Role: "merchant",
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	m.mu.Lock()
	m.merchantClients[merchantID] = client
	m.mu.Unlock()

	log.Printf("Merchant %d connected via WebSocket", merchantID)

	go client.writePump()
	go func() {
		client.readPump()
		m.mu.Lock()
		delete(m.merchantClients, merchantID)
		m.mu.Unlock()
		log.Printf("Merchant %d disconnected", merchantID)
	}()
}

// RegisterUser upgrades the HTTP connection for a user
func (m *Manager) RegisterUser(w http.ResponseWriter, r *http.Request, userID uint) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	client := &Client{
		ID:   userID,
		Role: "user",
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	m.mu.Lock()
	m.userClients[userID] = client
	m.mu.Unlock()

	log.Printf("User %d connected via WebSocket", userID)

	go client.writePump()
	go func() {
		client.readPump()
		m.mu.Lock()
		delete(m.userClients, userID)
		m.mu.Unlock()
		log.Printf("User %d disconnected", userID)
	}()
}

// BroadcastToMerchant sends a typed message to a specific merchant if connected
func (m *Manager) BroadcastToMerchant(merchantID uint, msgType string, payload interface{}) {
	m.mu.RLock()
	client, ok := m.merchantClients[merchantID]
	m.mu.RUnlock()

	if !ok {
		return // merchant not connected — notification saved to DB anyway
	}

	msg := WSMessage{Type: msgType, Payload: payload}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case client.Send <- data:
	default:
		log.Printf("merchant %d send buffer full, dropping message", merchantID)
	}
}

// BroadcastToUser sends a typed message to a specific user if connected
func (m *Manager) BroadcastToUser(userID uint, msgType string, payload interface{}) {
	m.mu.RLock()
	client, ok := m.userClients[userID]
	m.mu.RUnlock()

	if !ok {
		return
	}

	msg := WSMessage{Type: msgType, Payload: payload}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case client.Send <- data:
	default:
		log.Printf("user %d send buffer full, dropping message", userID)
	}
}

// IsMerchantOnline returns whether a merchant has an active WS connection
func (m *Manager) IsMerchantOnline(merchantID uint) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.merchantClients[merchantID]
	return ok
}

func (c *Client) writePump() {
	defer c.Conn.Close()
	for data := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("ws write error for %s %d: %v", c.Role, c.ID, err)
			return
		}
	}
}

func (c *Client) readPump() {
	defer close(c.Send)
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		// Client-to-server messages handled via REST/GraphQL; WS is push-only for now
	}
}
