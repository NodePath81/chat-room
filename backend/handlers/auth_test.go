package handlers

import (
	"bytes"
	"chat-room/tests"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"chat-room/models"
)

func TestRegister(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer tests.CleanupTestDB(db)

	handler := NewAuthHandler(db)

	tests := []struct {
		name       string
		input      map[string]string
		wantStatus int
	}{
		{
			name: "valid registration",
			input: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "duplicate username",
			input: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "missing password",
			input: map[string]string{
				"username": "testuser2",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Register(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Register() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer tests.CleanupTestDB(db)

	handler := NewAuthHandler(db)

	// Create a test user first
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := models.User{
		Username: "testuser",
		Password: string(hashedPassword),
	}
	db.Create(&testUser)

	tests := []struct {
		name       string
		input      map[string]string
		wantStatus int
	}{
		{
			name: "valid login",
			input: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid password",
			input: map[string]string{
				"username": "testuser",
				"password": "wrongpassword",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "user not found",
			input: map[string]string{
				"username": "nonexistent",
				"password": "password123",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing fields",
			input: map[string]string{
				"username": "testuser",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Login() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response LoginResponse
				json.NewDecoder(w.Body).Decode(&response)
				if response.Token == "" {
					t.Error("Expected token in response, got empty string")
				}
				if response.User.Username != tt.input["username"] {
					t.Errorf("Expected username %v, got %v", tt.input["username"], response.User.Username)
				}
			}
		})
	}
}
