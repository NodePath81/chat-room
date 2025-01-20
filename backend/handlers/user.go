package handlers

import (
	"encoding/json"
	"net/http"

	"chat-room/auth"
	"chat-room/store"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHandler struct {
	store store.Store
}

func NewUserHandler(store store.Store) *UserHandler {
	return &UserHandler{store: store}
}

type UpdateNicknameRequest struct {
	Nickname string `json:"nickname"`
}

type UpdateUsernameRequest struct {
	Username string `json:"username"`
}

type UserSessionResponse struct {
	SessionID uuid.UUID `json:"sessionId"`
	Role      string    `json:"role"`
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.store.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

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

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Username cannot be empty", http.StatusBadRequest)
		return
	}

	// Check if username is already taken
	exists, err := h.store.CheckUsernameExists(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "Error checking username availability", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, "Username is already taken", http.StatusConflict)
		return
	}

	// Get current user
	user, err := h.store.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Update username
	user.Username = req.Username
	if err := h.store.UpdateUser(r.Context(), user); err != nil {
		http.Error(w, "Error updating username", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Username updated successfully"})
}

func (h *UserHandler) UpdateNickname(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateNicknameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Nickname == "" {
		http.Error(w, "Nickname cannot be empty", http.StatusBadRequest)
		return
	}

	// Check if nickname is already taken
	exists, err := h.store.CheckNicknameExists(r.Context(), req.Nickname)
	if err != nil {
		http.Error(w, "Error checking nickname availability", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, "Nickname is already taken", http.StatusConflict)
		return
	}

	// Get current user
	user, err := h.store.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Update nickname
	user.Nickname = req.Nickname
	if err := h.store.UpdateUser(r.Context(), user); err != nil {
		http.Error(w, "Error updating nickname", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Nickname updated successfully"})
}

func (h *UserHandler) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := auth.GetUserIDFromContext(r)

	// Get all sessions for the user
	sessions, err := h.store.GetUserSessions(r.Context(), userID)
	if err != nil {
		http.Error(w, "Error fetching sessions", http.StatusInternalServerError)
		return
	}

	// Build response with session IDs and roles
	response := make([]UserSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		role, err := h.store.GetUserSessionRole(r.Context(), userID, session.ID)
		if err != nil {
			http.Error(w, "Error fetching session roles", http.StatusInternalServerError)
			return
		}
		response = append(response, UserSessionResponse{
			SessionID: session.ID,
			Role:      role,
		})
	}

	json.NewEncoder(w).Encode(response)
}
