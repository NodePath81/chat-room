package middleware

import (
	"chat-room/config"
	"chat-room/handlers"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAuthMiddleware(t *testing.T) {
	// Setup test config
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}
	// Set the global config for testing
	config.SetConfig(cfg)

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{
			name:       "valid token",
			token:      generateTestToken(1),
			wantStatus: http.StatusOK,
		},
		{
			name:       "no token",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token format",
			token:      "invalid-token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "expired token",
			token:      generateExpiredToken(1),
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that always returns 200 OK
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if userID is in context for valid tokens
				if tt.wantStatus == http.StatusOK {
					userID := GetUserID(r)
					if userID != 1 {
						t.Errorf("GetUserID() = %v, want %v", userID, 1)
					}
				}
				w.WriteHeader(http.StatusOK)
			})

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Test the middleware
			handler := AuthMiddleware(nextHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("AuthMiddleware() status = %v, want %v", rr.Code, tt.wantStatus)
			}
		})
	}
}

func generateTestToken(userID uint) string {
	claims := handlers.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(config.GetConfig().JWTSecret))
	return tokenString
}

func generateExpiredToken(userID uint) string {
	claims := handlers.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour * 2)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(config.GetConfig().JWTSecret))
	return tokenString
}
