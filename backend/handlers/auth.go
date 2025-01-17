package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"chat-room/auth"
	"chat-room/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
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
		ID       uint   `json:"id"`
		Nickname string `json:"nickname"`
	} `json:"user"`
}

// Add JWT claims struct
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// Add new response type for availability checks
type AvailabilityResponse struct {
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request format"})
		return
	}

	log.Printf("Received registration request for username: %s, nickname: %s", req.Username, req.Nickname)

	// Start transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %v", tx.Error)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Database error"})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		log.Print("Missing required fields")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Username and password are required"})
		return
	}

	// Check both username and nickname uniqueness in a single query
	var count int64
	if err := tx.Model(&models.User{}).
		Where("username = ? OR nickname = ?", req.Username, req.Nickname).
		Count(&count).Error; err != nil {
		log.Printf("Error checking availability: %v", err)
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Error checking availability"})
		return
	}

	log.Printf("Found %d existing users with same username/nickname", count)

	if count > 0 {
		tx.Rollback()
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Username or nickname is already taken"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Error hashing password"})
		return
	}

	user := &models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
	}

	log.Printf("Attempting to create user with username: %s, nickname: %s", user.Username, user.Nickname)

	// Create user within transaction
	if err := tx.Create(user).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Error creating user: " + err.Error()})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Error saving user"})
		return
	}

	log.Printf("Successfully created user with ID: %d", user.ID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request format"})
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Username and password are required"})
		return
	}

	// Find user by username
	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid credentials"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid credentials"})
		return
	}

	// Generate JWT token using auth package
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Error generating token"})
		return
	}

	// Create response
	response := LoginResponse{
		Token: token,
		User: struct {
			ID       uint   `json:"id"`
			Nickname string `json:"nickname"`
		}{
			ID:       user.ID,
			Nickname: user.Nickname,
		},
	}

	// Write response
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		// Log the error but don't try to write another response since headers are already sent
		println("Error encoding response:", err.Error())
	}
}

func (h *AuthHandler) CheckUsernameAvailability(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var count int64
	if err := h.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if count > 0 {
		w.WriteHeader(http.StatusConflict) // 409 Conflict for taken username
		return
	}

	w.WriteHeader(http.StatusOK) // 200 OK for available username
}

func (h *AuthHandler) CheckNicknameAvailability(w http.ResponseWriter, r *http.Request) {
	nickname := r.URL.Query().Get("nickname")
	if nickname == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var count int64
	if err := h.db.Model(&models.User{}).Where("nickname = ?", nickname).Count(&count).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if count > 0 {
		w.WriteHeader(http.StatusConflict) // 409 Conflict for taken nickname
		return
	}

	w.WriteHeader(http.StatusOK) // 200 OK for available nickname
}
