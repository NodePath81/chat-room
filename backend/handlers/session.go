package handlers

import (
	"encoding/json"
	"net/http"

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
		Name: req.Name,
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

	if err := h.db.First(&session, sessionID).Error; err != nil {
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

func (h *SessionHandler) JoinSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	userID := auth.GetUserIDFromContext(r)

	// Check membership using helper
	isMember, err := h.checkMembership(userID, sessionID)
	if err != nil {
		http.Error(w, "Error checking membership", http.StatusInternalServerError)
		return
	}
	if isMember {
		w.WriteHeader(http.StatusOK)
		return
	}

	var session models.Session
	if err := h.db.First(&session, sessionID).Error; err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Get existing user and add to session
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if err := h.db.Model(&session).Association("Users").Append(&user); err != nil {
		http.Error(w, "Failed to join session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *SessionHandler) CheckSessionMembership(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	userID := auth.GetUserIDFromContext(r)

	// Check membership using helper
	isMember, err := h.checkMembership(userID, sessionID)
	if err != nil {
		http.Error(w, "Error checking membership", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Not a member", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
}
