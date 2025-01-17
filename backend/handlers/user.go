package handlers

import (
	"encoding/json"
	"net/http"

	"chat-room/models"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	var user models.User
	if err := h.db.Select("id, nickname, avatar_url").First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Only return public user information
	response := struct {
		ID        uint   `json:"id"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatarUrl"`
	}{
		ID:        user.ID,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
