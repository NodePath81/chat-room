package token

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenManager(t *testing.T) {
	// Create a test server key (32 bytes for AES-256)
	serverKey := "12345678901234567890123456789012"

	t.Run("NewManager", func(t *testing.T) {
		manager, err := NewManager(serverKey)
		require.NoError(t, err)
		assert.NotNil(t, manager)

	})

	t.Run("GenerateAndVerifyToken", func(t *testing.T) {
		manager, err := NewManager(serverKey)
		require.NoError(t, err)

		groupID := uuid.New()
		role := "member"
		duration := time.Hour

		// Generate token
		token, err := manager.GenerateToken(groupID, role, duration)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token
		claims, err := manager.VerifyToken(token)
		require.NoError(t, err)
		assert.Equal(t, 1, claims.Version)
		assert.Equal(t, groupID, claims.GroupID)
		assert.Equal(t, role, claims.Role)
		assert.True(t, claims.ExpiresAt > time.Now().Unix())
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		manager, err := NewManager(serverKey)
		require.NoError(t, err)

		// Generate token with negative duration (already expired)
		token, err := manager.GenerateToken(uuid.New(), "member", -time.Hour)
		require.NoError(t, err)

		// Verify expired token
		_, err = manager.VerifyToken(token)
		assert.ErrorIs(t, err, ErrExpiredToken)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		manager, err := NewManager(serverKey)
		require.NoError(t, err)

		// Test with invalid base64
		_, err = manager.VerifyToken("invalid-token")
		assert.ErrorIs(t, err, ErrInvalidToken)

		// Test with empty token
		_, err = manager.VerifyToken("")
		assert.ErrorIs(t, err, ErrInvalidToken)

		// Test with tampered token
		token, err := manager.GenerateToken(uuid.New(), "member", time.Hour)
		require.NoError(t, err)
		tamperedToken := token[:len(token)-1] + "X"
		_, err = manager.VerifyToken(tamperedToken)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("DifferentKeys", func(t *testing.T) {
		manager1, err := NewManager(serverKey)
		require.NoError(t, err)

		differentKey := "12345678901234567890123456789013"
		manager2, err := NewManager(differentKey)
		require.NoError(t, err)

		// Generate token with first manager
		token, err := manager1.GenerateToken(uuid.New(), "member", time.Hour)
		require.NoError(t, err)

		// Try to verify with second manager
		_, err = manager2.VerifyToken(token)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
}
