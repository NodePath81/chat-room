package handlers

import (
	"chat-room/auth"
	"chat-room/models"
	"chat-room/tests"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebSocket(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer tests.CleanupTestDB(db)

	wsHandler := NewWebSocketHandler(db)
	server := httptest.NewServer(http.HandlerFunc(wsHandler.HandleWebSocket))
	defer server.Close()

	// Create test session
	session := models.Session{Name: "Test Session"}
	db.Create(&session)

	// Create test user
	user := models.User{Username: "testuser"}
	db.Create(&user)

	// Generate valid token
	token, err := auth.GenerateTestToken(user.ID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	t.Run("multiple connections same user", func(t *testing.T) {
		// Convert ws:// to ws:// for test server
		wsURL := strings.Replace(server.URL, "http://", "ws://", 1)
		wsURL = fmt.Sprintf("%s?sessionId=%d", wsURL, session.ID)

		// Create first connection
		conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn1.Close()

		// Authenticate first connection
		err = conn1.WriteJSON(map[string]string{"token": token})
		if err != nil {
			t.Fatalf("Failed to send auth message: %v", err)
		}

		// Create second connection
		conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect second client: %v", err)
		}
		defer conn2.Close()

		// Authenticate second connection
		err = conn2.WriteJSON(map[string]string{"token": token})
		if err != nil {
			t.Fatalf("Failed to send auth message for second connection: %v", err)
		}

		// Send message from first connection
		testMessage := "Hello from conn1"
		err = conn1.WriteJSON(map[string]string{"content": testMessage})
		if err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		// Both connections should receive the message
		for _, conn := range []*websocket.Conn{conn1, conn2} {
			var received map[string]interface{}
			err = conn.ReadJSON(&received)
			if err != nil {
				t.Fatalf("Failed to read message: %v", err)
			}

			if content, ok := received["content"].(string); !ok || content != testMessage {
				t.Errorf("Expected message %q, got %v", testMessage, received)
			}
		}
	})
}
