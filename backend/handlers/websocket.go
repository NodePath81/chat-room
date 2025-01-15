package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"chat-room/models"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type WebSocketHandler struct {
	db       *gorm.DB
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	mutex    sync.Mutex
}

func NewWebSocketHandler(db *gorm.DB) *WebSocketHandler {
	return &WebSocketHandler{
		db: db,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

type Message struct {
	Type      string          `json:"type"`
	Content   string          `json:"content,omitempty"`
	User      json.RawMessage `json:"user,omitempty"`
	SessionID uint            `json:"sessionId"`
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	h.mutex.Lock()
	h.clients[conn] = true
	h.mutex.Unlock()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			h.mutex.Lock()
			delete(h.clients, conn)
			h.mutex.Unlock()
			break
		}

		// Save message to database
		if msg.Type == "message" {
			dbMsg := models.Message{
				Content:   msg.Content,
				SessionID: msg.SessionID,
			}
			if err := h.db.Create(&dbMsg).Error; err != nil {
				log.Printf("Error saving message: %v", err)
				continue
			}
		}

		// Broadcast message to all clients
		h.mutex.Lock()
		for client := range h.clients {
			if err := client.WriteJSON(msg); err != nil {
				log.Printf("Error broadcasting message: %v", err)
				client.Close()
				delete(h.clients, client)
			}
		}
		h.mutex.Unlock()
	}
}
