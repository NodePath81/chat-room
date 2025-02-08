// Package handlers provides HTTP handlers for the chat application.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"chat-room/auth"
	"chat-room/middleware"
	"chat-room/models"
	"chat-room/store"
	"chat-room/token"

	"github.com/google/uuid"
)

// SessionHandler manages HTTP requests for session-related operations.
type SessionHandler struct {
	store        store.Store
	tokenManager *token.TokenManager
}

// NewSessionHandler creates a new session handler with the given store and token manager.
func NewSessionHandler(store store.Store, tokenManager *token.TokenManager) *SessionHandler {
	return &SessionHandler{
		store:        store,
		tokenManager: tokenManager,
	}
}

// Request/Response types
type (
	// CreateSessionRequest represents the request body for creating a new session.
	CreateSessionRequest struct {
		Name string `json:"name"`
	}

	// SessionResponse represents a session with its member users.
	SessionResponse struct {
		*models.Session
		Users []*models.User `json:"users"`
	}

	// BatchIDsRequest represents a request containing a list of UUIDs.
	BatchIDsRequest struct {
		IDs []uuid.UUID `json:"ids"`
	}

	// RoleResponse represents a user's role in a session.
	RoleResponse struct {
		Role string `json:"role"`
	}

	// CreateShareLinkRequest represents the request body for creating a share link.
	CreateShareLinkRequest struct {
		DurationDays int `json:"durationDays"`
	}

	// ShareLinkResponse represents the response containing a share token.
	ShareLinkResponse struct {
		Token string `json:"token"`
	}

	// ShareInfoResponse represents information about a shared session.
	ShareInfoResponse struct {
		SessionName     string `json:"session_name"`
		InviterNickname string `json:"inviter_nickname"`
	}
)

// parsePaginationLimit parses and validates the limit query parameter.
// defaultLimit is used if no limit is provided.
// maxLimit is the maximum allowed limit.
func parsePaginationLimit(r *http.Request, defaultLimit, maxLimit int) int {
	limit := defaultLimit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return limit
}

// parsePaginationBefore parses the before query parameter.
// Returns the current time if no valid before parameter is provided.
func parsePaginationBefore(r *http.Request) time.Time {
	if beforeStr := r.URL.Query().Get("before"); beforeStr != "" {
		if parsedBefore, err := time.Parse(time.RFC3339, beforeStr); err == nil {
			return parsedBefore
		}
	}
	return time.Now().UTC()
}

// GetMessageIDsBySessionID returns message IDs for a session with pagination.
// Route: GET /api/sessions/messages/ids
// Query parameters:
//   - limit: maximum number of messages to return (default: 50, max: 100)
//   - before: timestamp to get messages before (default: now)
//
// Response: {"message_ids": ["uuid1", "uuid2", ...], "has_more": bool}
func (h *SessionHandler) GetMessageIDsBySessionID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := middleware.GetSessionID(r)

	limit := parsePaginationLimit(r, 50, 100)
	before := parsePaginationBefore(r)

	messageIDs, err := h.store.GetMessageIDsBySessionID(r.Context(), sessionID, limit+1, before)
	if err != nil {
		http.Error(w, "Error fetching message IDs", http.StatusInternalServerError)
		return
	}

	hasMore := len(messageIDs) > limit
	if hasMore {
		messageIDs = messageIDs[:limit]
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message_ids": messageIDs,
		"has_more":    hasMore,
	})
}

// PostFetchMessages retrieves messages by their IDs.
// Route: POST /api/messages/batch
// Request: {"ids": ["uuid1", "uuid2", ...]}
// Response: {"messages": [{"id": "uuid", "content": "text", ...}]}
func (h *SessionHandler) PostFetchMessages(w http.ResponseWriter, r *http.Request) {
	var req BatchIDsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		http.Error(w, "No message IDs provided", http.StatusBadRequest)
		return
	}

	messages, err := h.store.GetMessagesByIDs(r.Context(), req.IDs)
	if err != nil {
		http.Error(w, "Error fetching messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
	})
}

// CreateSession creates a new session.
// Route: POST /api/sessions
// Request: {"name": "session name"}
// Response: {"id": "uuid", "name": "name", "creator_id": "uuid", "users": [...]}
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	userID := auth.GetUserIDFromContext(r)

	session := &models.Session{
		ID:        uuid.New(),
		Name:      req.Name,
		CreatorID: userID,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.store.CreateSession(r.Context(), session); err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	// Get users in the session
	userIDs, err := h.store.GetUserIDsBySessionID(r.Context(), session.ID)
	if err != nil {
		http.Error(w, "Error fetching session users", http.StatusInternalServerError)
		return
	}

	users, err := h.store.GetUsersByIDs(r.Context(), userIDs)
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	response := SessionResponse{
		Session: session,
		Users:   users,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSession retrieves a single session by ID
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := middleware.GetSessionID(r)

	session, err := h.store.GetSessionByID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Get users in the session
	userIDs, err := h.store.GetUserIDsBySessionID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Error fetching session users", http.StatusInternalServerError)
		return
	}

	users, err := h.store.GetUsersByIDs(r.Context(), userIDs)
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	response := SessionResponse{
		Session: session,
		Users:   users,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SessionHandler) CheckRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionClaims := middleware.GetSessionClaims(r)
	json.NewEncoder(w).Encode(RoleResponse{Role: sessionClaims.Role})
}

func (h *SessionHandler) CreateShareLink(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := middleware.GetSessionID(r)

	// Parse request body
	var req CreateShareLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate duration
	if req.DurationDays <= 0 || req.DurationDays > 30 {
		http.Error(w, "Duration must be between 1 and 30 days", http.StatusBadRequest)
		return
	}

	// Generate share token
	token, err := auth.GenerateSessionShareToken(sessionID, time.Duration(req.DurationDays)*24*time.Hour)
	if err != nil {
		http.Error(w, "Error generating share token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(ShareLinkResponse{Token: token})
}

// GetShareInfo returns information about a shared session
func (h *SessionHandler) GetShareInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get token from query params
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// Validate token
	claims, err := auth.ValidateSessionShareToken(token)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Get session info
	session, err := h.store.GetSessionByID(r.Context(), claims.SessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Get inviter info
	users, err := h.store.GetUsersByIDs(r.Context(), []uuid.UUID{session.CreatorID})
	if err != nil || len(users) == 0 {
		http.Error(w, "Inviter not found", http.StatusNotFound)
		return
	}
	inviter := users[0]

	response := ShareInfoResponse{
		SessionName:     session.Name,
		InviterNickname: inviter.Nickname,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *SessionHandler) RemoveSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := middleware.GetSessionID(r)

	// Start a transaction to remove the session and all its associations
	tx, err := h.store.BeginTx(r.Context())
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Delete the session (this will cascade delete all associations)
	if err := tx.DeleteSession(r.Context(), sessionID); err != nil {
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
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
	json.NewEncoder(w).Encode(map[string]string{"message": "Session removed successfully"})
}
