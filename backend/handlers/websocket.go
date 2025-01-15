package handlers

import (
	"chat-room/auth"
	"chat-room/config"
	"chat-room/models"
	"log"
	"net/http"
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

type WebSocketHandler struct {
	db       *gorm.DB
	upgrader websocket.Upgrader
	clients  sync.Map // websocket.Conn -> Client
}

type Message struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	UserID    uint   `json:"userId"`
	SessionID uint   `json:"sessionId"`
}

// Add this struct for history messages
type MessageHistory struct {
	Type     string    `json:"type"`
	Messages []Message `json:"messages"`
}

func NewWebSocketHandler(db *gorm.DB) *WebSocketHandler {
	return &WebSocketHandler{
		db: db,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from frontend
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
	defer conn.Close()

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

	// Store connection with both user and session ID
	h.clients.Store(conn, Client{
		UserID:    claims.UserID,
		SessionID: session.ID,
	})
	defer h.clients.Delete(conn)

	// After successful auth, send message history
	var messages []models.Message
	if err := h.db.Where("session_id = ? AND created_at > ?",
		r.URL.Query().Get("sessionId"),
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

		h.broadcast(msg)
	}
}

func (h *WebSocketHandler) broadcast(msg Message) {
	// Track which users have received the message
	sentToUsers := make(map[uint]bool)

	h.clients.Range(func(key, value interface{}) bool {
		conn := key.(*websocket.Conn)
		client := value.(Client)

		// Only send to clients in the same session
		if client.SessionID != msg.SessionID {
			return true
		}

		// Skip if we already sent to this user
		if sentToUsers[client.UserID] {
			return true
		}

		if err := conn.WriteJSON(msg); err != nil {
			h.clients.Delete(conn)
			conn.Close()
		} else {
			sentToUsers[client.UserID] = true
		}
		return true
	})
}
