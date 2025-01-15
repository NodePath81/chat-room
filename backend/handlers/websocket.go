package handlers

import (
	"chat-room/auth"
	"chat-room/config"
	"chat-room/models"
	"log"
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type WebSocketHandler struct {
	db       *gorm.DB
	upgrader websocket.Upgrader
	clients  sync.Map
}

type Message struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	UserID    uint   `json:"userId"`
	SessionID uint   `json:"sessionId"`
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

	// Store connection
	h.clients.Store(conn, claims.UserID)
	defer h.clients.Delete(conn)

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
		userID := value.(uint)

		// Skip if we already sent to this user
		if sentToUsers[userID] {
			return true
		}

		if err := conn.WriteJSON(msg); err != nil {
			h.clients.Delete(conn)
			conn.Close()
		} else {
			sentToUsers[userID] = true
		}
		return true
	})
}
