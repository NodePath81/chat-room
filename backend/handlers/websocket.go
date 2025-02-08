package handlers

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"chat-room/models"
	"chat-room/store"
	"chat-room/token"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type Client struct {
	UserID    uuid.UUID
	Username  string
	Nickname  string
	AvatarURL string
	Conn      *websocket.Conn
	mu        sync.Mutex
}

type WebSocketMessage struct {
	Content string             `json:"content"`
	Type    models.MessageType `json:"type"`
}

type SessionClients struct {
	Clients map[uuid.UUID][]*Client // map[userID][]*Client
	mu      sync.RWMutex
}

type WebSocketHandler struct {
	sessions     sync.Map // map[uuid.UUID]*SessionClients
	store        store.Store
	tokenManager *token.TokenManager
}

func NewWebSocketHandler(store store.Store, tokenManager *token.TokenManager) *WebSocketHandler {
	return &WebSocketHandler{
		store:        store,
		tokenManager: tokenManager,
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get token from URL parameter
	wsToken := r.URL.Query().Get("token")
	if wsToken == "" {
		http.Error(w, "WebSocket token is required", http.StatusUnauthorized)
		return
	}

	// Verify token
	claims, err := h.tokenManager.VerifyWebSocketToken(wsToken)
	if err != nil {
		http.Error(w, "Invalid or expired WebSocket token", http.StatusUnauthorized)
		return
	}

	sessionID := claims.SessionID
	userID := claims.UserID

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Get user info
	user, err := h.store.GetUsersByIDs(r.Context(), []uuid.UUID{userID})
	if err != nil {
		conn.WriteJSON(map[string]string{"error": "user not found"})
		conn.Close()
		return
	}

	// Create new client
	client := &Client{
		UserID:    user[0].ID,
		Username:  user[0].Username,
		Nickname:  user[0].Nickname,
		AvatarURL: user[0].AvatarURL,
		Conn:      conn,
	}

	// Get or create session clients
	sessionClientsInterface, _ := h.sessions.LoadOrStore(sessionID, &SessionClients{
		Clients: make(map[uuid.UUID][]*Client),
	})
	sessionClients := sessionClientsInterface.(*SessionClients)

	// Add client to session
	sessionClients.mu.Lock()
	if _, exists := sessionClients.Clients[userID]; !exists {
		sessionClients.Clients[userID] = make([]*Client, 0)
	}
	sessionClients.Clients[userID] = append(sessionClients.Clients[userID], client)
	sessionClients.mu.Unlock()

	// Handle messages
	go func() {
		defer h.removeConnection(sessionID, userID, client)

		for {
			var wsMsg WebSocketMessage
			if err := conn.ReadJSON(&wsMsg); err != nil {
				log.Printf("Error reading message: %v", err)
				return
			}

			// Validate message type - only allow text messages
			if wsMsg.Type != models.MessageTypeText {
				log.Printf("Rejected non-text message from user %s: type=%v", user[0].Username, wsMsg.Type)
				conn.WriteJSON(map[string]string{"error": "only text messages are allowed via WebSocket"})
				continue
			}

			log.Printf("Received message from user %s in session %s: %+v", user[0].Username, sessionID, wsMsg)

			message := &models.Message{
				ID:        uuid.New(),
				UserID:    userID,
				Content:   wsMsg.Content,
				Timestamp: time.Now().UTC(),
				SessionID: sessionID,
				Type:      models.MessageTypeText,
			}

			log.Printf("Processing message: %+v", message)

			// Use background context for message handling
			ctx := context.Background()
			// Save message to database
			if err := h.store.CreateMessage(ctx, message); err != nil {
				log.Printf("Error saving message: %v", err)
				continue
			}

			log.Printf("Broadcasting message to session %s", sessionID)
			// Broadcast message
			h.broadcast(sessionID, message)
		}
	}()
}

func (h *WebSocketHandler) removeConnection(sessionID, userID uuid.UUID, client *Client) {
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

func (h *WebSocketHandler) broadcast(sessionID uuid.UUID, message *models.Message) {
	sessionClientsInterface, ok := h.sessions.Load(sessionID)
	if !ok {
		return
	}

	sessionClients := sessionClientsInterface.(*SessionClients)
	sessionClients.mu.RLock()
	defer sessionClients.mu.RUnlock()

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
