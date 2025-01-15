package handlers

import (
	"chat-room/auth"
	"chat-room/config"
	"chat-room/models"
	"chat-room/tests"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebSocket(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	defer tests.CleanupTestDB(db)

	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}
	config.SetConfig(cfg)

	// Create test users
	user1 := models.User{Username: "user1", Password: "pass1"}
	user2 := models.User{Username: "user2", Password: "pass2"}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	// Create test session with both users
	session := models.Session{
		Name: "Test Room",
		Users: []models.User{
			user1,
			user2,
		},
	}
	if err := db.Create(&session).Error; err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create test messages
	testMessage := models.Message{
		Content:   "Test message",
		UserID:    user1.ID,
		SessionID: session.ID,
	}
	if err := db.Create(&testMessage).Error; err != nil {
		t.Fatalf("failed to create test message: %v", err)
	}

	// Create handler with test-specific upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in test
		},
	}
	handler := NewWebSocketHandler(db)
	handler.upgrader = upgrader // Set the test upgrader

	t.Run("successful connection and message exchange", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.RawQuery = "sessionId=" + strconv.FormatUint(uint64(session.ID), 10)
			handler.HandleWebSocket(w, r)
		}))
		defer server.Close()

		// Connect to WebSocket
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		wsURL += "?sessionId=" + strconv.FormatUint(uint64(session.ID), 10)

		// Add Origin header
		headers := http.Header{}
		headers.Add("Origin", "http://localhost")

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
		if err != nil {
			t.Fatalf("could not open websocket connection: %v", err)
		}
		defer conn.Close()

		// Send auth message
		token, err := auth.GenerateTestToken(user1.ID)
		if err != nil {
			t.Fatalf("could not generate token: %v", err)
		}

		authMsg := map[string]string{
			"type":  "auth",
			"token": token,
		}
		if err := conn.WriteJSON(authMsg); err != nil {
			t.Fatalf("could not send auth message: %v", err)
		}

		// Wait for auth success response
		var response map[string]string
		if err := conn.ReadJSON(&response); err != nil {
			t.Fatalf("could not read auth response: %v", err)
		}
		if response["type"] != "auth_success" {
			t.Fatalf("expected auth_success, got %v", response)
		}

		// Should receive message history
		var history MessageHistory
		if err := conn.ReadJSON(&history); err != nil {
			t.Fatalf("could not read history: %v", err)
		}
		if len(history.Messages) != 1 {
			t.Errorf("expected 1 message in history, got %d", len(history.Messages))
		}

		// Send new message
		newMsg := Message{
			Type:      "message",
			Content:   "Hello from test",
			SessionID: session.ID,
		}
		if err := conn.WriteJSON(newMsg); err != nil {
			t.Fatalf("could not send message: %v", err)
		}

		// Should receive the message back
		var received Message
		if err := conn.ReadJSON(&received); err != nil {
			t.Fatalf("could not read message: %v", err)
		}
		if received.Content != newMsg.Content {
			t.Errorf("expected content %q, got %q", newMsg.Content, received.Content)
		}
		if received.UserID != user1.ID {
			t.Errorf("expected user ID %d, got %d", user1.ID, received.UserID)
		}

		// Verify message was saved
		var savedMsg models.Message
		if err := db.Last(&savedMsg).Error; err != nil {
			t.Fatalf("could not fetch saved message: %v", err)
		}
		if savedMsg.Content != newMsg.Content {
			t.Errorf("expected saved content %q, got %q", newMsg.Content, savedMsg.Content)
		}
	})

	t.Run("unauthorized connection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.RawQuery = "sessionId=" + strconv.FormatUint(uint64(session.ID), 10)
			handler.HandleWebSocket(w, r)
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		wsURL += "?sessionId=" + strconv.FormatUint(uint64(session.ID), 10)

		headers := http.Header{}
		headers.Add("Origin", "http://localhost")

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
		if err != nil {
			t.Fatal("expected successful connection before auth")
		}
		defer conn.Close()

		// Send invalid auth message
		invalidAuthMsg := map[string]string{
			"type":  "auth",
			"token": "invalid-token",
		}
		if err := conn.WriteJSON(invalidAuthMsg); err != nil {
			t.Fatalf("could not send invalid auth message: %v", err)
		}

		// Should receive auth error
		var response map[string]string
		if err := conn.ReadJSON(&response); err != nil {
			t.Fatalf("could not read error response: %v", err)
		}

		// Check for error message instead of type
		if response["error"] != "invalid token" {
			t.Errorf("expected error 'invalid token', got %v", response)
		}
	})
}
