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
	SessionID uuid.UUID `json:"session_id"`
	Role      string    `json:"role"`
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := auth.GetUserIDFromContext(r)

	user, err := h.store.GetUsersByIDs(r.Context(), []uuid.UUID{userID})
	if err != nil || len(user) == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	response := user[0]

	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) PostFetchUsersByIDs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req BatchIDsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	users, err := h.store.GetUsersByIDs(r.Context(), req.IDs)
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
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
	user, err := h.store.GetUsersByIDs(r.Context(), []uuid.UUID{userID})
	if err != nil || len(user) == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Update username
	user[0].Username = req.Username
	if err := h.store.UpdateUser(r.Context(), user[0]); err != nil {
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
	user, err := h.store.GetUsersByIDs(r.Context(), []uuid.UUID{userID})
	if err != nil || len(user) == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Update nickname
	user[0].Nickname = req.Nickname
	if err := h.store.UpdateUser(r.Context(), user[0]); err != nil {
		http.Error(w, "Error updating nickname", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Nickname updated successfully"})
}
