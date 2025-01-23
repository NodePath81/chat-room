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

type SessionHandler struct {
	store        store.Store
	tokenManager *token.TokenManager
}

func NewSessionHandler(store store.Store, tokenManager *token.TokenManager) *SessionHandler {
	return &SessionHandler{
		store:        store,
		tokenManager: tokenManager,
	}
}

type CreateSessionRequest struct {
	Name string `json:"name"`
}

type SessionResponse struct {
	*models.Session
	Users []*models.User `json:"users"`
}

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
	users, err := h.store.GetSessionUsers(r.Context(), session.ID)
	if err != nil {
		http.Error(w, "Error fetching session users", http.StatusInternalServerError)
		return
	}

	response := SessionResponse{
		Session: session,
		Users:   users,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SessionHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r)

	sessions, err := h.store.GetUserSessions(r.Context(), userID)
	if err != nil {
		http.Error(w, "Error fetching sessions", http.StatusInternalServerError)
		return
	}

	// Build response with users for each session
	response := make([]SessionResponse, 0, len(sessions))
	for _, session := range sessions {
		users, err := h.store.GetSessionUsers(r.Context(), session.ID)
		if err != nil {
			http.Error(w, "Error fetching session users", http.StatusInternalServerError)
			return
		}
		response = append(response, SessionResponse{
			Session: session,
			Users:   users,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := middleware.GetSessionID(r)
	session, err := h.store.GetSessionByID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Get users in the session
	users, err := h.store.GetSessionUsers(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Error fetching session users", http.StatusInternalServerError)
		return
	}

	response := SessionResponse{
		Session: session,
		Users:   users,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type RoleResponse struct {
	Role string `json:"role"`
}

func (h *SessionHandler) CheckRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := middleware.GetSessionID(r)
	userID := auth.GetUserIDFromContext(r)

	// Get session to check if user is creator
	session, err := h.store.GetSessionByID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Check if user is creator
	if session.CreatorID == userID {
		json.NewEncoder(w).Encode(RoleResponse{Role: "creator"})
		return
	}

	// Get user's role in the session
	role, err := h.store.GetUserSessionRole(r.Context(), userID, sessionID)
	if err != nil {
		json.NewEncoder(w).Encode(RoleResponse{Role: "none"})
		return
	}

	json.NewEncoder(w).Encode(RoleResponse{Role: role})
}

func (h *SessionHandler) JoinSession(w http.ResponseWriter, r *http.Request) {
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

	// Check if session exists
	session, err := h.store.GetSessionByID(r.Context(), claims.SessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Add user to session
	err = h.store.AddUserToSession(r.Context(), userID, session.ID, "member")
	if err != nil {
		http.Error(w, "Failed to join session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully joined session"})
}

type CreateShareLinkRequest struct {
	DurationDays int `json:"durationDays"`
}

type ShareLinkResponse struct {
	Token string `json:"token"`
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

type MembersResponse struct {
	Members []uuid.UUID `json:"members"`
}

func (h *SessionHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sessionID := middleware.GetSessionID(r)
	if sessionID == uuid.Nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	users, err := h.store.GetSessionUsers(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Error fetching session members", http.StatusInternalServerError)
		return
	}

	memberIDs := make([]uuid.UUID, len(users))
	for i, user := range users {
		memberIDs[i] = user.ID
	}

	json.NewEncoder(w).Encode(MembersResponse{Members: memberIDs})
}

func (h *SessionHandler) KickMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := middleware.GetSessionID(r)

	memberID, err := uuid.Parse(r.URL.Query().Get("memberId"))
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Cannot kick the creator
	session, err := h.store.GetSessionByID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if memberID == session.CreatorID {
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Session removed successfully"})
}

type ShareInfoResponse struct {
	SessionName     string `json:"session_name"`
	InviterNickname string `json:"inviter_nickname"`
}

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
	inviter, err := h.store.GetUserByID(r.Context(), session.CreatorID)
	if err != nil {
		http.Error(w, "Inviter not found", http.StatusNotFound)
		return
	}

	response := ShareInfoResponse{
		SessionName:     session.Name,
		InviterNickname: inviter.Nickname,
	}

	json.NewEncoder(w).Encode(response)
}

type GetMessagesResponse struct {
	Messages []models.Message `json:"messages"`
	HasMore  bool             `json:"has_more"`
}

func (h *SessionHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sessionID := middleware.GetSessionID(r)

	// Parse pagination parameters
	limit := 50 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if limit > 100 { // Maximum limit
		limit = 100
	}

	var before time.Time
	if beforeStr := r.URL.Query().Get("before"); beforeStr != "" {
		if parsedBefore, err := time.Parse(time.RFC3339, beforeStr); err == nil {
			before = parsedBefore
		} else {
			before = time.Now().UTC()
		}
	} else {
		before = time.Now().UTC()
	}

	// Get messages using the store interface
	messages, err := h.store.GetMessagesBySessionID(r.Context(), sessionID, limit+1, before)
	if err != nil {
		http.Error(w, "Error fetching messages", http.StatusInternalServerError)
		return
	}

	// Check if there are more messages
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	// Convert to response format
	response := GetMessagesResponse{
		Messages: make([]models.Message, len(messages)),
		HasMore:  hasMore,
	}

	for i, msg := range messages {
		response.Messages[i] = models.Message{
			ID:        msg.ID,
			Content:   msg.Content,
			UserID:    msg.UserID,
			SessionID: msg.SessionID,
			Timestamp: msg.Timestamp,
			Type:      msg.Type,
		}
	}

	json.NewEncoder(w).Encode(response)
}
