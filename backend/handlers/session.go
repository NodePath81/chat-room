package handlers

import (
	"encoding/json"
	"net/http"

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

	session := models.Session{
		Name: req.Name,
	}

	if err := h.db.Create(&session).Error; err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(session)
}

func (h *SessionHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	var sessions []models.Session
	if err := h.db.Find(&sessions).Error; err != nil {
		http.Error(w, "Error fetching sessions", http.StatusInternalServerError)
		return
	}

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
