package handlers

import (
	"bytes"
	"chat-room/auth"
	"chat-room/config"
	"chat-room/models"
	"chat-room/tests"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestSessionHandler(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer tests.CleanupTestDB(db)

	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}
	config.SetConfig(cfg)

	handler := NewSessionHandler(db)

	// Create test user
	user := models.User{Username: "testuser", Password: "password"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	t.Run("create session", func(t *testing.T) {
		req := CreateSessionRequest{Name: "Test Session"}
		body, _ := json.Marshal(req)

		r := httptest.NewRequest("POST", "/api/sessions", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		// Set user ID in context
		ctx := r.Context()
		ctx = context.WithValue(ctx, auth.UserIDKey, user.ID)
		r = r.WithContext(ctx)

		handler.CreateSession(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response models.Session
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Name != req.Name {
			t.Errorf("expected session name %q, got %q", req.Name, response.Name)
		}

		// Verify creator was added to session
		var session models.Session
		if err := db.Preload("Users").First(&session, response.ID).Error; err != nil {
			t.Fatalf("failed to fetch session: %v", err)
		}
		if len(session.Users) != 1 {
			t.Errorf("expected 1 user in session, got %d", len(session.Users))
		}
		if session.Users[0].ID != user.ID {
			t.Errorf("expected user ID %d, got %d", user.ID, session.Users[0].ID)
		}
	})

	t.Run("join session", func(t *testing.T) {
		// Create test session
		session := models.Session{Name: "Join Test"}
		if err := db.Create(&session).Error; err != nil {
			t.Fatalf("failed to create test session: %v", err)
		}

		// Fix: Use strconv.FormatUint for proper string conversion
		sessionIDStr := strconv.FormatUint(uint64(session.ID), 10)
		r := httptest.NewRequest("POST", "/api/sessions/"+sessionIDStr+"/join", nil)
		w := httptest.NewRecorder()

		// Setup chi router context for URL parameters
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", sessionIDStr)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

		// Set user ID in context
		r = r.WithContext(context.WithValue(r.Context(), auth.UserIDKey, user.ID))

		handler.JoinSession(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}

		// Verify user was added to session
		var updatedSession models.Session
		if err := db.Preload("Users").First(&updatedSession, session.ID).Error; err != nil {
			t.Fatalf("failed to fetch updated session: %v", err)
		}
		if len(updatedSession.Users) != 1 {
			t.Errorf("expected 1 user in session, got %d", len(updatedSession.Users))
		}
		if updatedSession.Users[0].ID != user.ID {
			t.Errorf("expected user ID %d, got %d", user.ID, updatedSession.Users[0].ID)
		}
	})
}
