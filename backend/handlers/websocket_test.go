package handlers

import (
	"chat-room/models"
	"chat-room/tests"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebSocketConnection(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer tests.CleanupTestDB(db)

	handler := NewWebSocketHandler(db)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect to WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("could not open websocket connection: %v", err)
	}
	defer ws.Close()

	// Create test session
	session := models.Session{Name: "Test Room"}
	db.Create(&session)

	// Test sending message
	t.Run("send message", func(t *testing.T) {
		message := Message{
			Type:      "message",
			Content:   "Hello, World!",
			SessionID: session.ID,
		}

		err := ws.WriteJSON(message)
		if err != nil {
			t.Fatalf("could not send message: %v", err)
		}

		// Read response
		var response Message
		err = ws.ReadJSON(&response)
		if err != nil {
			t.Fatalf("could not read response: %v", err)
		}

		if response.Type != "message" || response.Content != "Hello, World!" {
			t.Errorf("unexpected response: %v", response)
		}
	})

	// Verify message was saved to database
	t.Run("verify message in database", func(t *testing.T) {
		var message models.Message
		err := db.Where("session_id = ?", session.ID).First(&message).Error
		if err != nil {
			t.Errorf("message not found in database: %v", err)
		}

		if message.Content != "Hello, World!" {
			t.Errorf("unexpected message content: got %v, want Hello, World!", message.Content)
		}
	})
}
