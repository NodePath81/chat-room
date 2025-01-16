package handlers

import (
	"chat-room/auth"
	"chat-room/config"
	"chat-room/models"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type Client struct {
	UserID    uint
	SessionID uint
}

type SessionClients struct {
	clients map[*websocket.Conn]Client
	mu      sync.RWMutex
}

type WebSocketHandler struct {
	db       *gorm.DB
	upgrader websocket.Upgrader
	sessions sync.Map // map[uint]*SessionClients
}

type Message struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	UserID    uint   `json:"userId"`
	SessionID uint   `json:"sessionId"`
}

type MessageHistory struct {
	Type     string    `json:"type"`
	Messages []Message `json:"messages"`
}

func NewWebSocketHandler(db *gorm.DB) *WebSocketHandler {
	return &WebSocketHandler{
		db: db,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return r.Header.Get("Origin") == "http://localhost:3000"
			},
		},
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get sessionId from query params
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	// Upgrade connection first
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}
	defer h.removeConnection(sessionID, conn)

	// Wait for auth message
	var authMsg struct {
		Type  string `json:"type"`
		Token string `json:"token"`
	}
	if err := conn.ReadJSON(&authMsg); err != nil {
		log.Printf("auth message error: %v", err)
		return
	}

	if authMsg.Type != "auth" || authMsg.Token == "" {
		conn.WriteJSON(map[string]string{"error": "unauthorized"})
		return
	}

	// Parse and validate token
	claims := &auth.Claims{}
	parsedToken, err := jwt.ParseWithClaims(authMsg.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().JWTSecret), nil
	})

	if err != nil || !parsedToken.Valid {
		conn.WriteJSON(map[string]string{"error": "invalid token"})
		return
	}

	// Send auth success
	conn.WriteJSON(map[string]string{"type": "auth_success"})

	// Verify user is member of the session
	var session models.Session
	if err := h.db.Preload("Users", "id = ?", claims.UserID).First(&session, sessionID).Error; err != nil {
		conn.WriteJSON(map[string]string{"error": "session not found or not a member"})
		return
	}
	if len(session.Users) == 0 {
		conn.WriteJSON(map[string]string{"error": "not a member of this session"})
		return
	}

	// Add connection to session clients
	h.addConnection(session.ID, conn, Client{
		UserID:    claims.UserID,
		SessionID: session.ID,
	})

	// After successful auth, send message history
	var messages []models.Message
	if err := h.db.Where("session_id = ? AND created_at > ?",
		sessionID,
		time.Now().Add(-24*time.Hour), // Get last 24 hours of messages
	).Order("created_at asc").Find(&messages).Error; err != nil {
		log.Printf("error fetching message history: %v", err)
	} else {
		// Convert and send history
		history := make([]Message, len(messages))
		for i, msg := range messages {
			history[i] = Message{
				Type:      "message",
				Content:   msg.Content,
				UserID:    msg.UserID,
				SessionID: msg.SessionID,
			}
		}

		conn.WriteJSON(MessageHistory{
			Type:     "history",
			Messages: history,
		})
	}

	// Handle messages
	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		msg.UserID = claims.UserID
		if msg.Type == "message" {
			if err := h.db.Create(&models.Message{
				Content:   msg.Content,
				UserID:    msg.UserID,
				SessionID: msg.SessionID,
			}).Error; err != nil {
				log.Printf("error saving message: %v", err)
				continue
			}
		}

		h.broadcast(session.ID, msg)
	}
}

func (h *WebSocketHandler) addConnection(sessionID uint, conn *websocket.Conn, client Client) {
	sessionClientsInterface, _ := h.sessions.LoadOrStore(sessionID, &SessionClients{
		clients: make(map[*websocket.Conn]Client),
	})

	sessionClients := sessionClientsInterface.(*SessionClients)
	sessionClients.mu.Lock()
	defer sessionClients.mu.Unlock()

	sessionClients.clients[conn] = client
}

func (h *WebSocketHandler) removeConnection(sessionID string, conn *websocket.Conn) {
	// Convert string sessionID to uint for sync.Map lookup
	sid, _ := strconv.ParseUint(sessionID, 10, 64)
	if sessionClientsInterface, ok := h.sessions.Load(uint(sid)); ok {
		sessionClients := sessionClientsInterface.(*SessionClients)
		sessionClients.mu.Lock()
		defer sessionClients.mu.Unlock()

		delete(sessionClients.clients, conn)
		conn.Close()

		// If session is empty, remove it
		if len(sessionClients.clients) == 0 {
			h.sessions.Delete(sessionID)
		}
	}
}

func (h *WebSocketHandler) broadcast(sessionID uint, msg Message) {
	if sessionClientsInterface, ok := h.sessions.Load(sessionID); ok {
		sessionClients := sessionClientsInterface.(*SessionClients)
		sessionClients.mu.RLock()
		defer sessionClients.mu.RUnlock()

		sentToUsers := make(map[uint]bool)

		for conn, client := range sessionClients.clients {
			// Skip if already sent to this user
			if sentToUsers[client.UserID] {
				continue
			}

			if err := conn.WriteJSON(msg); err != nil {
				// Convert uint to string properly for removeConnection
				go h.removeConnection(strconv.FormatUint(uint64(sessionID), 10), conn)
			} else {
				sentToUsers[client.UserID] = true
			}
		}
	}
}
