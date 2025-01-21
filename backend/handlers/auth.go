package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"chat-room/auth"
	"chat-room/models"
	"chat-room/store"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	store store.Store
}

func NewAuthHandler(store store.Store) *AuthHandler {
	return &AuthHandler{store: store}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       uuid.UUID `json:"id"`
		Nickname string    `json:"nickname"`
	} `json:"user"`
}

// Add JWT claims struct
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	log.Printf("Received registration request for username: %s, nickname: %s", req.Username, req.Nickname)

	// Validate required fields
	if req.Username == "" || req.Password == "" || req.Nickname == "" {
		log.Print("Missing required fields")
		http.Error(w, "Username, password, and nickname are required", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := h.store.BeginTx(r.Context())
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Check username availability
	exists, err := tx.CheckUsernameExists(r.Context(), req.Username)
	if err != nil {
		log.Printf("Error checking username availability: %v", err)
		http.Error(w, "Error checking username availability", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Username is already taken", http.StatusConflict)
		return
	}

	// Check nickname availability
	exists, err = tx.CheckNicknameExists(r.Context(), req.Nickname)
	if err != nil {
		log.Printf("Error checking nickname availability: %v", err)
		http.Error(w, "Error checking nickname availability", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Nickname is already taken", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	// Create user
	user := &models.User{
		ID:       uuid.New(),
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
	}

	log.Printf("Attempting to create user with username: %s, nickname: %s", user.Username, user.Nickname)

	if err := tx.CreateUser(r.Context(), user); err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Error saving user", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully created user with ID: %s", user.ID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Find user by username
	user, err := h.store.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := LoginResponse{
		Token: token,
		User: struct {
			ID       uuid.UUID `json:"id"`
			Nickname string    `json:"nickname"`
		}{
			ID:       user.ID,
			Nickname: user.Nickname,
		},
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (h *AuthHandler) CheckUsernameAvailability(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	exists, err := h.store.CheckUsernameExists(r.Context(), username)
	if err != nil {
		http.Error(w, "Error checking username availability", http.StatusInternalServerError)
		return
	}

	if exists {
		w.WriteHeader(http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) CheckNicknameAvailability(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	nickname := r.URL.Query().Get("nickname")
	if nickname == "" {
		http.Error(w, "Nickname is required", http.StatusBadRequest)
		return
	}

	exists, err := h.store.CheckNicknameExists(r.Context(), nickname)
	if err != nil {
		http.Error(w, "Error checking nickname availability", http.StatusInternalServerError)
		return
	}

	if exists {
		w.WriteHeader(http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
}
