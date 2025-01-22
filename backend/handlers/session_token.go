package handlers

import (
	"chat-room/auth"
	"chat-room/token"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SessionTokenResponse struct {
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
	_, err = h.store.GetSessionByID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Get user's role in the session
	role, err := h.store.GetUserSessionRole(r.Context(), userID, sessionID)
	if err != nil {
		http.Error(w, "User is not a member of this session", http.StatusForbidden)
		return
	}

	// Generate session token
	token, err := h.tokenManager.GenerateToken(sessionID, role, 24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to generate session token", http.StatusInternalServerError)
		return
	}

	// Set session token in cookie
	cookie := &http.Cookie{
		Name:     "session_token_" + sessionID.String(),
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours
	}
	http.SetCookie(w, cookie)

	// Return token in response
	json.NewEncoder(w).Encode(SessionTokenResponse{Token: token})
}

// RefreshSessionToken refreshes an existing session token
func (h *SessionHandler) RefreshSessionToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Get existing token from cookie
	cookie, err := r.Cookie("session_token_" + sessionID.String())
	if err != nil {
		http.Error(w, "Session token not found", http.StatusUnauthorized)
		return
	}

	// Verify existing token
	claims, err := h.tokenManager.VerifyToken(cookie.Value)
	if err != nil && err != token.ErrExpiredToken {
		http.Error(w, "Invalid session token", http.StatusUnauthorized)
		return
	}

	// Generate new token with same role
	newToken, err := h.tokenManager.GenerateToken(sessionID, claims.Role, 24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to generate session token", http.StatusInternalServerError)
		return
	}

	// Set new token in cookie
	newCookie := &http.Cookie{
		Name:     "session_token_" + sessionID.String(),
		Value:    newToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours
	}
	http.SetCookie(w, newCookie)

	json.NewEncoder(w).Encode(SessionTokenResponse{Token: newToken})
}

// RevokeSessionToken revokes a session token
func (h *SessionHandler) RevokeSessionToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Remove session token cookie
	cookie := &http.Cookie{
		Name:     "session_token_" + sessionID.String(),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // Delete cookie
	}
	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Session token revoked"})
}
