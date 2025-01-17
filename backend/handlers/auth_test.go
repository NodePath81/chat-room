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
		name          string
		input         map[string]string
		wantStatus    int
		checkNickname bool
		setup         func() // Setup function to run before the test
	}{
		{
			name: "valid registration with nickname",
			input: map[string]string{
				"username": "testuser",
				"password": "password123",
				"nickname": "TestNick",
			},
			wantStatus:    http.StatusCreated,
			checkNickname: true,
		},
		{
			name: "valid registration without nickname",
			input: map[string]string{
				"username": "testuser2",
				"password": "password123",
			},
			wantStatus:    http.StatusCreated,
			checkNickname: true,
		},
		{
			name: "duplicate username",
			input: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			wantStatus:    http.StatusBadRequest,
			checkNickname: false,
		},
		{
			name: "duplicate nickname",
			input: map[string]string{
				"username": "testuser3",
				"password": "password123",
				"nickname": "TestNick",
			},
			wantStatus:    http.StatusBadRequest,
			checkNickname: false,
			setup: func() {
				// Create a user with the same nickname
				user := models.User{
					Username: "existinguser",
					Password: "somepassword",
					Nickname: "TestNick",
				}
				db.Create(&user)
			},
		},
		{
			name: "missing password",
			input: map[string]string{
				"username": "testuser3",
			},
			wantStatus:    http.StatusBadRequest,
			checkNickname: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Register(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Register() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.checkNickname && w.Code == http.StatusCreated {
				var user models.User
				if err := db.Where("username = ?", tt.input["username"]).First(&user).Error; err != nil {
					t.Errorf("Failed to find created user: %v", err)
				}

				// Check if nickname is set
				if tt.input["nickname"] != "" {
					if user.Nickname != tt.input["nickname"] {
						t.Errorf("Expected nickname %v, got %v", tt.input["nickname"], user.Nickname)
					}
				} else {
					if user.Nickname == "" {
						t.Error("Expected auto-generated nickname, got empty string")
					}
				}

				// Verify nickname uniqueness
				var count int64
				if err := db.Model(&models.User{}).Where("nickname = ?", user.Nickname).Count(&count).Error; err != nil {
					t.Errorf("Failed to count nicknames: %v", err)
				}
				if count != 1 {
					t.Errorf("Expected nickname to be unique, found %d occurrences", count)
				}
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
		Nickname: "TestNick",
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
				if response.User.ID != testUser.ID {
					t.Errorf("Expected user ID %v, got %v", testUser.ID, response.User.ID)
				}
				if response.User.Nickname != testUser.Nickname {
					t.Errorf("Expected nickname %v, got %v", testUser.Nickname, response.User.Nickname)
				}
			}
		})
	}
}

func TestCheckUsernameAvailability(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer tests.CleanupTestDB(db)

	handler := NewAuthHandler(db)

	// Create a test user
	testUser := models.User{
		Username: "existinguser",
		Password: "password123",
		Nickname: "ExistingNick",
	}
	db.Create(&testUser)

	tests := []struct {
		name          string
		username      string
		wantStatus    int
		wantAvailable bool
	}{
		{
			name:          "available username",
			username:      "newuser",
			wantStatus:    http.StatusOK,
			wantAvailable: true,
		},
		{
			name:          "taken username",
			username:      "existinguser",
			wantStatus:    http.StatusOK,
			wantAvailable: false,
		},
		{
			name:          "empty username",
			username:      "",
			wantStatus:    http.StatusOK,
			wantAvailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/auth/check-username?username="+tt.username, nil)
			w := httptest.NewRecorder()

			handler.CheckUsernameAvailability(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CheckUsernameAvailability() status = %v, want %v", w.Code, tt.wantStatus)
			}

			var response AvailabilityResponse
			json.NewDecoder(w.Body).Decode(&response)

			if response.Available != tt.wantAvailable {
				t.Errorf("CheckUsernameAvailability() available = %v, want %v", response.Available, tt.wantAvailable)
			}

			if !response.Available && response.Message == "" {
				t.Error("Expected error message for unavailable username")
			}
		})
	}
}

func TestCheckNicknameAvailability(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer tests.CleanupTestDB(db)

	handler := NewAuthHandler(db)

	// Create a test user
	testUser := models.User{
		Username: "testuser",
		Password: "password123",
		Nickname: "ExistingNick",
	}
	db.Create(&testUser)

	tests := []struct {
		name          string
		nickname      string
		wantStatus    int
		wantAvailable bool
	}{
		{
			name:          "available nickname",
			nickname:      "NewNick",
			wantStatus:    http.StatusOK,
			wantAvailable: true,
		},
		{
			name:          "taken nickname",
			nickname:      "ExistingNick",
			wantStatus:    http.StatusOK,
			wantAvailable: false,
		},
		{
			name:          "empty nickname",
			nickname:      "",
			wantStatus:    http.StatusOK,
			wantAvailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/auth/check-nickname?nickname="+tt.nickname, nil)
			w := httptest.NewRecorder()

			handler.CheckNicknameAvailability(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CheckNicknameAvailability() status = %v, want %v", w.Code, tt.wantStatus)
			}

			var response AvailabilityResponse
			json.NewDecoder(w.Body).Decode(&response)

			if response.Available != tt.wantAvailable {
				t.Errorf("CheckNicknameAvailability() available = %v, want %v", response.Available, tt.wantAvailable)
			}

			if !response.Available && response.Message == "" {
				t.Error("Expected error message for unavailable nickname")
			}
		})
	}
}
