package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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

type UpdateNicknameRequest struct {
	Nickname string `json:"nickname"`
}

type UpdateUsernameRequest struct {
	Username string `json:"username"`
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := chi.URLParam(r, "id")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "User not found"})
		return
	}

	// Return user information
	response := map[string]interface{}{
		"id":        user.ID,
		"username":  user.Username,
		"nickname":  user.Nickname,
		"avatarUrl": user.AvatarURL,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse user ID from URL
	userID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid user ID"})
		return
	}

	// Parse request body
	var req UpdateUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request format"})
		return
	}

	if req.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Username cannot be empty"})
		return
	}

	// Check if username is already taken
	var count int64
	if err := h.db.Model(&models.User{}).Where("username = ? AND id != ?", req.Username, userID).Count(&count).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Error checking username availability"})
		return
	}

	if count > 0 {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"message": "Username is already taken"})
		return
	}

	// Update username
	result := h.db.Model(&models.User{}).Where("id = ?", userID).Update("username", req.Username)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Error updating username"})
		return
	}

	if result.RowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "User not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Username updated successfully"})
}

func (h *UserHandler) UpdateNickname(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse user ID from URL
	userID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid user ID"})
		return
	}

	// Parse request body
	var req UpdateNicknameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request format"})
		return
	}

	if req.Nickname == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Nickname cannot be empty"})
		return
	}

	// Check if nickname is already taken
	var count int64
	if err := h.db.Model(&models.User{}).Where("nickname = ? AND id != ?", req.Nickname, userID).Count(&count).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Error checking nickname availability"})
		return
	}

	if count > 0 {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"message": "Nickname is already taken"})
		return
	}

	// Update nickname
	result := h.db.Model(&models.User{}).Where("id = ?", userID).Update("nickname", req.Nickname)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Error updating nickname"})
		return
	}

	if result.RowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "User not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Nickname updated successfully"})
}
