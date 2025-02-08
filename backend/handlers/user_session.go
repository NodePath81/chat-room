package handlers

import (
	"encoding/json"
	"net/http"

	"chat-room/auth"
	"chat-room/middleware"
	"chat-room/store"

	"github.com/google/uuid"
)

// UserSessionHandler manages HTTP requests for user-session relationship operations.
type UserSessionHandler struct {
	store store.Store
}

// NewUserSessionHandler creates a new user session handler with the given store.
func NewUserSessionHandler(store store.Store) *UserSessionHandler {
	return &UserSessionHandler{store: store}
}

// GetSessionIDsByUserID returns all session IDs that the user is a member of.
// Route: GET /api/sessions/ids
// Response: {"session_ids": ["uuid1", "uuid2", ...]}
func (h *UserSessionHandler) GetSessionIDsByUserID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := auth.GetUserIDFromContext(r)

	sessionIDs, err := h.store.GetSessionIDsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Error fetching session IDs", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_ids": sessionIDs,
	})
}

// GetUserIDsBySessionID returns all user IDs in a session.
// Route: GET /api/sessions/users/ids
// Response: {"user_ids": ["uuid1", "uuid2", ...]}
func (h *UserSessionHandler) GetUserIDsBySessionID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := middleware.GetSessionID(r)

	userIDs, err := h.store.GetUserIDsBySessionID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Error fetching user IDs", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_ids": userIDs,
	})
}

// JoinSession adds a user to a session using a share token.
// Route: POST /api/sessions/join
// Query parameters:
//   - token: share token for the session
//
// Response: {"message": "Successfully joined session"}
func (h *UserSessionHandler) JoinSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Share token is required", http.StatusBadRequest)
		return
	}

	// Validate share token
	claims, err := auth.ValidateSessionShareToken(token)
	if err != nil {
		http.Error(w, "Invalid or expired share token", http.StatusUnauthorized)
		return
	}

	userID := auth.GetUserIDFromContext(r)

	// Add user to session
	err = h.store.AddUserToSession(r.Context(), userID, claims.SessionID, "member")
	if err != nil {
		http.Error(w, "Failed to join session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully joined session"})
}

// LeaveSession removes a normal user from a session.
// Route: POST /api/sessions/leave
// Response: {"message": "Successfully left session"}
func (h *UserSessionHandler) LeaveSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionClaims := middleware.GetSessionClaims(r)
	userID := auth.GetUserIDFromContext(r)

	if sessionClaims.Role != "member" {
		http.Error(w, "Only normal users can leave sessions", http.StatusForbidden)
		return
	}

	err := h.store.RemoveUserFromSession(r.Context(), userID, sessionClaims.GroupID)
	if err != nil {
		http.Error(w, "Failed to leave session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully left session"})
}

// KickMember removes a user from a session (creator only).
// Route: POST /api/sessions/kick
// Query parameters:
//   - memberId: ID of the user to kick
//
// Response: {"message": "Member kicked successfully"}
func (h *UserSessionHandler) KickMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := middleware.GetSessionID(r)

	memberID, err := uuid.Parse(r.URL.Query().Get("memberId"))
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Cannot kick the creator
	userSessions, err := h.store.GetUserSessionsBySessionIDAndUserIDs(r.Context(), sessionID, []uuid.UUID{memberID})
	if err != nil || len(userSessions) == 0 {
		http.Error(w, "User/Session not found", http.StatusNotFound)
		return
	}
	userSession := userSessions[0]

	if userSession.Role == "creator" {
		http.Error(w, "Cannot kick the creator", http.StatusBadRequest)
		return
	}

	// Remove the member from the session
	err = h.store.RemoveUserFromSession(r.Context(), memberID, sessionID)
	if err != nil {
		http.Error(w, "Failed to kick member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Member kicked successfully"})
}
