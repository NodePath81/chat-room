package handlers

import (
	"chat-room/auth"
	"chat-room/models"
	"log"
	"net/http"
	"sync"

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
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	h.clients.Store(conn, userID)
	defer h.clients.Delete(conn)

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		msg.UserID = userID
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
	h.clients.Range(func(key, _ interface{}) bool {
		conn := key.(*websocket.Conn)
		if err := conn.WriteJSON(msg); err != nil {
			h.clients.Delete(conn)
			conn.Close()
		}
		return true
	})
}
