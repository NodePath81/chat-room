package handlers

import (
	"chat-room/auth"
	"chat-room/middleware"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type SessionTokenResponse struct {
	SessionID string    `json:"session_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type WebSocketTokenResponse struct {
	Token string `json:"token"`
}

// GetSessionToken generates and returns a session token
func (h *SessionHandler) GetSessionToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get session ID from query parameter
	sessionID, err := uuid.Parse(r.URL.Query().Get("session_id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID := auth.GetUserIDFromContext(r)

	// Check if session exists and user is a member
	userSessions, err := h.store.GetUserSessionsBySessionIDAndUserIDs(r.Context(), sessionID, []uuid.UUID{userID})
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	if len(userSessions) == 0 {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Generate session token
	token, err := h.tokenManager.GenerateToken(userSessions[0].SessionID, userSessions[0].Role, 24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to generate session token", http.StatusInternalServerError)
		return
	}

	// Return token in response
	json.NewEncoder(w).Encode(SessionTokenResponse{Token: token, SessionID: sessionID.String(), ExpiresAt: time.Now().Add(24 * time.Hour)})
}

// GetWebSocketToken generates and returns a WebSocket token
func (h *SessionHandler) GetWebSocketToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get session ID from query parameter
	sessionID := middleware.GetSessionID(r)

	// Get user ID from context
	userID := auth.GetUserIDFromContext(r)

	// Generate WebSocket token with 5-minute expiration
	token, err := h.tokenManager.GenerateWebSocketToken(userID, sessionID, 5*time.Minute)
	if err != nil {
		http.Error(w, "Failed to generate WebSocket token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(WebSocketTokenResponse{Token: token})
}
