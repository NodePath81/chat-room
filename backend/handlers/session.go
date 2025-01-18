package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"chat-room/auth"
	"chat-room/models"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type SessionHandler struct {
	db *gorm.DB
}

func NewSessionHandler(db *gorm.DB) *SessionHandler {
	return &SessionHandler{db: db}
}

type CreateSessionRequest struct {
	Name string `json:"name"`
}

func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := auth.GetUserIDFromContext(r)

	// Start a transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	session := models.Session{
		Name:      req.Name,
		CreatorID: userID,
	}

	if err := tx.Create(&session).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	// Add creator to the session
	var user models.User
	if err := tx.First(&user, userID).Error; err != nil {
		tx.Rollback()
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if err := tx.Model(&session).Association("Users").Append(&user); err != nil {
		tx.Rollback()
		http.Error(w, "Failed to add user to session", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Fetch the complete session with users
	if err := h.db.Preload("Users").First(&session, session.ID).Error; err != nil {
		http.Error(w, "Error fetching created session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (h *SessionHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	var sessions []models.Session
	if err := h.db.Preload("Users").Find(&sessions).Error; err != nil {
		http.Error(w, "Error fetching sessions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	var session models.Session

	if err := h.db.Preload("Users").First(&session, sessionID).Error; err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(session)
}

func (h *SessionHandler) checkMembership(userID uint, sessionID string) (bool, error) {
	var exists int64
	err := h.db.Table("user_sessions").
		Where("user_id = ? AND session_id = ?", userID, sessionID).
		Count(&exists).Error
	return exists > 0, err
}

type RoleResponse struct {
	Role string `json:"role"`
}

func (h *SessionHandler) CheckRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := chi.URLParam(r, "id")
	userID := auth.GetUserIDFromContext(r)

	// First check if session exists
	var session models.Session
	if err := h.db.First(&session, sessionID).Error; err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Check if user is creator
	if session.CreatorID == userID {
		json.NewEncoder(w).Encode(RoleResponse{Role: "creator"})
		return
	}

	// Check if user is member
	isMember, err := h.checkMembership(userID, sessionID)
	if err != nil {
		http.Error(w, "Error checking membership", http.StatusInternalServerError)
		return
	}
	if isMember {
		json.NewEncoder(w).Encode(RoleResponse{Role: "member"})
		return
	}

	json.NewEncoder(w).Encode(RoleResponse{Role: "none"})
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

	// Get user ID from auth context
	userID := auth.GetUserIDFromContext(r)

	// Check if user is already a member
	isMember, err := h.checkMembership(userID, strconv.FormatUint(uint64(claims.SessionID), 10))
	if err != nil {
		http.Error(w, "Error checking membership", http.StatusInternalServerError)
		return
	}
	if isMember {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get session
	var session models.Session
	if err := h.db.First(&session, claims.SessionID).Error; err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Get user
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Add user to session
	if err := h.db.Model(&session).Association("Users").Append(&user); err != nil {
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

	sessionID := chi.URLParam(r, "id")
	userID := auth.GetUserIDFromContext(r)

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

	// Convert sessionID from string to uint
	sessionIDUint, err := strconv.ParseUint(sessionID, 10, 32)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Check if user is the creator of the session
	var session models.Session
	if err := h.db.First(&session, sessionIDUint).Error; err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.CreatorID != userID {
		http.Error(w, "Only the session creator can generate share links", http.StatusForbidden)
		return
	}

	// Generate share token
	duration := time.Duration(req.DurationDays) * 24 * time.Hour
	token, err := auth.GenerateSessionShareToken(uint(sessionIDUint), duration)
	if err != nil {
		http.Error(w, "Failed to generate share token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(ShareLinkResponse{Token: token})
}

type MembersResponse struct {
	Members []uint `json:"members"`
}

func (h *SessionHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := chi.URLParam(r, "id")

	var session models.Session
	if err := h.db.Preload("Users").First(&session, sessionID).Error; err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	memberIDs := make([]uint, len(session.Users))
	for i, user := range session.Users {
		memberIDs[i] = user.ID
	}

	json.NewEncoder(w).Encode(MembersResponse{Members: memberIDs})
}

func (h *SessionHandler) KickMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := chi.URLParam(r, "id")
	memberID := r.URL.Query().Get("member")
	userID := auth.GetUserIDFromContext(r)

	// Check if session exists and user is creator
	var session models.Session
	if err := h.db.First(&session, sessionID).Error; err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.CreatorID != userID {
		http.Error(w, "Only the creator can kick members", http.StatusForbidden)
		return
	}

	// Cannot kick the creator
	memberIDUint, err := strconv.ParseUint(memberID, 10, 32)
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}
	if uint(memberIDUint) == session.CreatorID {
		http.Error(w, "Cannot kick the creator", http.StatusBadRequest)
		return
	}

	// Remove the member from the session
	var member models.User
	if err := h.db.First(&member, memberID).Error; err != nil {
		http.Error(w, "Member not found", http.StatusNotFound)
		return
	}

	if err := h.db.Model(&session).Association("Users").Delete(&member); err != nil {
		http.Error(w, "Failed to kick member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Member kicked successfully"})
}

func (h *SessionHandler) RemoveSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := chi.URLParam(r, "id")
	userID := auth.GetUserIDFromContext(r)

	// Check if session exists and user is creator
	var session models.Session
	if err := h.db.First(&session, sessionID).Error; err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.CreatorID != userID {
		http.Error(w, "Only the creator can remove the session", http.StatusForbidden)
		return
	}

	// Start a transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	// Remove all user associations first
	if err := tx.Model(&session).Association("Users").Clear(); err != nil {
		tx.Rollback()
		http.Error(w, "Failed to remove session members", http.StatusInternalServerError)
		return
	}

	// Delete the session
	if err := tx.Delete(&session).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
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
	var session models.Session
	if err := h.db.First(&session, claims.SessionID).Error; err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Get inviter info
	var inviter models.User
	if err := h.db.First(&inviter, session.CreatorID).Error; err != nil {
		http.Error(w, "Inviter not found", http.StatusNotFound)
		return
	}

	response := ShareInfoResponse{
		SessionName:     session.Name,
		InviterNickname: inviter.Nickname,
	}

	json.NewEncoder(w).Encode(response)
}
