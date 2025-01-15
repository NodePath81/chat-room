package handlers

import (
	"chat-room/auth"
	"chat-room/models"
	"chat-room/tests"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebSocket(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer tests.CleanupTestDB(db)

	// Create test users
	user1 := models.User{Username: "user1", Password: "pass1"}
	user2 := models.User{Username: "user2", Password: "pass2"}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	// Create test session
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

	handler := NewWebSocketHandler(db)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), auth.UserIDKey, user1.ID)
		handler.HandleWebSocket(w, r.WithContext(ctx))
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect as user1
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("user1 connection failed: %v", err)
	}
	defer conn1.Close()

	// Connect as user2
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), auth.UserIDKey, user2.ID)
		handler.HandleWebSocket(w, r.WithContext(ctx))
	}))
	defer server2.Close()
	wsURL2 := "ws" + strings.TrimPrefix(server2.URL, "http")
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("user2 connection failed: %v", err)
	}
	defer conn2.Close()

	// Test message
	testMsg := Message{
		Type:      "message",
		Content:   "Hello from user1",
		SessionID: session.ID,
	}

	// User1 sends message
	if err := conn1.WriteJSON(testMsg); err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	// User2 should receive the message
	var received Message
	if err := conn2.ReadJSON(&received); err != nil {
		t.Fatalf("user2 failed to receive message: %v", err)
	}

	if received.Content != testMsg.Content {
		t.Errorf("got message %q, want %q", received.Content, testMsg.Content)
	}
	if received.UserID != user1.ID {
		t.Errorf("got userID %d, want %d", received.UserID, user1.ID)
	}

	// Verify database
	var saved models.Message
	if err := db.First(&saved, "session_id = ? AND user_id = ?", session.ID, user1.ID).Error; err != nil {
		t.Fatalf("message not saved: %v", err)
	}
}
