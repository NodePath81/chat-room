package handlers

import (
	"chat-room/auth"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type Client struct {
	UserID   uint
	Username string
	Conn     *websocket.Conn
	mu       sync.Mutex
}

type Message struct {
	UserID    uint      `json:"userId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"timestamp"`
	SessionID uint      `json:"sessionId"`
	Type      string    `json:"type"`
	MsgType   string    `json:"msgType"`
}

type SessionClients struct {
	Clients map[uint][]*Client // map[userID][]*Client
	mu      sync.RWMutex
}

type WebSocketHandler struct {
	sessions sync.Map // map[uint]*SessionClients
	db       *gorm.DB
}

func NewWebSocketHandler(db *gorm.DB) *WebSocketHandler {
	return &WebSocketHandler{
		db: db,
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID, err := strconv.ParseUint(r.URL.Query().Get("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Wait for authentication message
	var authMsg struct {
		Token string `json:"token"`
	}
	if err := conn.ReadJSON(&authMsg); err != nil {
		conn.Close()
		return
	}

	claims, err := auth.ParseToken(authMsg.Token)
	if err != nil {
		conn.WriteJSON(map[string]string{"error": "invalid token"})
		conn.Close()
		return
	}

	// Get or create session clients
	sessionClientsInterface, _ := h.sessions.LoadOrStore(uint(sessionID), &SessionClients{
		Clients: make(map[uint][]*Client),
	})
	sessionClients := sessionClientsInterface.(*SessionClients)

	// Create new client
	client := &Client{
		UserID: claims.UserID,
		Conn:   conn,
	}

	// Add client to session
	sessionClients.mu.Lock()
	if _, exists := sessionClients.Clients[claims.UserID]; !exists {
		sessionClients.Clients[claims.UserID] = make([]*Client, 0)
	}
	sessionClients.Clients[claims.UserID] = append(sessionClients.Clients[claims.UserID], client)
	sessionClients.mu.Unlock()

	// Send message history
	history, err := h.GetSessionMessages(uint(sessionID))
	if err == nil {
		conn.WriteJSON(map[string]interface{}{
			"type":     "history",
			"messages": history,
		})
	}

	// Handle messages
	go func() {
		defer h.removeConnection(uint(sessionID), claims.UserID, client)

		for {
			var msg struct {
				Content string `json:"content"`
				Type    string `json:"type"`
				MsgType string `json:"msgType"`
			}
			if err := conn.ReadJSON(&msg); err != nil {
				log.Printf("Error reading message: %v", err)
				return
			}

			log.Printf("Received message: %+v", msg)

			message := Message{
				UserID:    claims.UserID,
				Content:   msg.Content,
				CreatedAt: time.Now().UTC(),
				SessionID: uint(sessionID),
				Type:      msg.Type,
				MsgType:   msg.MsgType,
			}

			// Set default type if not specified
			if message.Type == "" {
				message.Type = "message"
			}
			if message.MsgType == "" {
				message.MsgType = "text"
			}

			log.Printf("Processing message: %+v", message)

			// Save message to database
			if err := h.SaveMessage(uint(sessionID), message); err != nil {
				log.Printf("Error saving message: %v", err)
				continue
			}

			log.Printf("Broadcasting message to session %d", sessionID)
			// Broadcast message
			h.broadcast(uint(sessionID), message)
		}
	}()
}

func (h *WebSocketHandler) removeConnection(sessionID, userID uint, client *Client) {
	sessionClientsInterface, ok := h.sessions.Load(sessionID)
	if !ok {
		return
	}

	sessionClients := sessionClientsInterface.(*SessionClients)
	sessionClients.mu.Lock()
	defer sessionClients.mu.Unlock()

	clients := sessionClients.Clients[userID]
	for i, c := range clients {
		if c == client {
			// Remove this specific client
			sessionClients.Clients[userID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}

	// If no more clients for this user in this session, remove the user
	if len(sessionClients.Clients[userID]) == 0 {
		delete(sessionClients.Clients, userID)
	}

	// If no more clients in the session, remove the session
	if len(sessionClients.Clients) == 0 {
		h.sessions.Delete(sessionID)
	}

	client.Conn.Close()
}

func (h *WebSocketHandler) broadcast(sessionID uint, message Message) {
	sessionClientsInterface, ok := h.sessions.Load(sessionID)
	if !ok {
		return
	}

	sessionClients := sessionClientsInterface.(*SessionClients)
	sessionClients.mu.RLock()
	defer sessionClients.mu.RUnlock()

	// Set default message type if not specified
	if message.Type == "" {
		message.Type = "message"
	}

	// Broadcast to all clients in the session
	for _, clients := range sessionClients.Clients {
		for _, client := range clients {
			client.mu.Lock()
			if err := client.Conn.WriteJSON(message); err != nil {
				log.Printf("Error broadcasting to client: %v", err)
			}
			client.mu.Unlock()
		}
	}
}

func (h *WebSocketHandler) GetSessionMessages(sessionID uint) ([]Message, error) {
	var messages []Message
	err := h.db.Where("session_id = ?", sessionID).Order("created_at asc").Find(&messages).Error
	return messages, err
}

func (h *WebSocketHandler) SaveMessage(sessionID uint, message Message) error {
	return h.db.Create(&message).Error
}
