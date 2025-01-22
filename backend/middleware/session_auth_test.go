package middleware

import (
	"chat-room/token"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionAuth(t *testing.T) {
	// Create token manager for testing
	serverKey := "12345678901234567890123456789012"
	tokenManager, err := token.NewManager(serverKey)
	require.NoError(t, err)

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetSessionClaims(r)
		require.NotNil(t, claims)
		w.WriteHeader(http.StatusOK)
	})

	t.Run("ValidToken", func(t *testing.T) {
		// Generate valid token
		groupID := uuid.New()
		validToken, err := tokenManager.GenerateToken(groupID, "member", time.Hour)
		require.NoError(t, err)

		// Create request with token
		req := httptest.NewRequest("GET", "/?token="+validToken, nil)
		rec := httptest.NewRecorder()

		// Test middleware
		handler := NewSessionAuth(tokenManager)(testHandler)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("MissingToken", func(t *testing.T) {
		// Create request without token
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		// Test middleware
		handler := NewSessionAuth(tokenManager)(testHandler)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		// Generate expired token
		groupID := uuid.New()
		expiredToken, err := tokenManager.GenerateToken(groupID, "member", -time.Hour)
		require.NoError(t, err)

		// Create request with expired token
		req := httptest.NewRequest("GET", "/?token="+expiredToken, nil)
		rec := httptest.NewRecorder()

		// Test middleware
		handler := NewSessionAuth(tokenManager)(testHandler)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		// Create request with invalid token
		req := httptest.NewRequest("GET", "/?token=invalid-token", nil)
		rec := httptest.NewRecorder()

		// Test middleware
		handler := NewSessionAuth(tokenManager)(testHandler)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestRequireRole(t *testing.T) {
	// Create token manager for testing
	serverKey := "12345678901234567890123456789012"
	tokenManager, err := token.NewManager(serverKey)
	require.NoError(t, err)

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("AllowedRole", func(t *testing.T) {
		// Generate token with allowed role
		groupID := uuid.New()
		token, err := tokenManager.GenerateToken(groupID, "admin", time.Hour)
		require.NoError(t, err)

		// Create request with token
		req := httptest.NewRequest("GET", "/?token="+token, nil)
		rec := httptest.NewRecorder()

		// Test middleware chain
		handler := NewSessionAuth(tokenManager)(RequireRole("admin")(testHandler))
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("ForbiddenRole", func(t *testing.T) {
		// Generate token with non-allowed role
		groupID := uuid.New()
		token, err := tokenManager.GenerateToken(groupID, "member", time.Hour)
		require.NoError(t, err)

		// Create request with token
		req := httptest.NewRequest("GET", "/?token="+token, nil)
		rec := httptest.NewRecorder()

		// Test middleware chain
		handler := NewSessionAuth(tokenManager)(RequireRole("admin")(testHandler))
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("MultipleAllowedRoles", func(t *testing.T) {
		// Generate token with one of allowed roles
		groupID := uuid.New()
		token, err := tokenManager.GenerateToken(groupID, "moderator", time.Hour)
		require.NoError(t, err)

		// Create request with token
		req := httptest.NewRequest("GET", "/?token="+token, nil)
		rec := httptest.NewRecorder()

		// Test middleware chain with multiple allowed roles
		handler := NewSessionAuth(tokenManager)(RequireRole("admin", "moderator")(testHandler))
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
