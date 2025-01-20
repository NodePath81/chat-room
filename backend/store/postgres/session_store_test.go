package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionStore(t *testing.T) {
	helper := setupTestDB(t)
	defer helper.cleanup(t)

	store := helper.createTestStore(t)
	defer store.Close()

	// Create a test user for sessions
	user := generateTestUser()
	err := store.CreateUser(helper.ctx, user)
	require.NoError(t, err)

	t.Run("CreateSession", func(t *testing.T) {
		session := generateTestSession(user.ID)
		err := store.CreateSession(helper.ctx, session)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, session.ID)
		assert.False(t, session.CreatedAt.IsZero())
		assert.False(t, session.UpdatedAt.IsZero())
	})

	t.Run("GetSessionByID", func(t *testing.T) {
		// Create a test session
		session := generateTestSession(user.ID)
		err := store.CreateSession(helper.ctx, session)
		require.NoError(t, err)

		// Get the session by ID
		retrieved, err := store.GetSessionByID(helper.ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, session.ID, retrieved.ID)
		assert.Equal(t, session.Name, retrieved.Name)
		assert.Equal(t, session.CreatorID, retrieved.CreatorID)
		assert.WithinDuration(t, session.CreatedAt, retrieved.CreatedAt, time.Second)
		assert.WithinDuration(t, session.UpdatedAt, retrieved.UpdatedAt, time.Second)
	})

	t.Run("UpdateSession", func(t *testing.T) {
		// Create a test session
		session := generateTestSession(user.ID)
		err := store.CreateSession(helper.ctx, session)
		require.NoError(t, err)

		// Update session fields
		session.Name = "Updated Session Name"
		originalUpdatedAt := session.UpdatedAt

		// Wait a moment to ensure UpdatedAt will be different
		time.Sleep(time.Millisecond * 10)

		err = store.UpdateSession(helper.ctx, session)
		require.NoError(t, err)

		// Get the updated session
		retrieved, err := store.GetSessionByID(helper.ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, session.Name, retrieved.Name)
		assert.True(t, retrieved.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("DeleteSession", func(t *testing.T) {
		// Create a test session
		session := generateTestSession(user.ID)
		err := store.CreateSession(helper.ctx, session)
		require.NoError(t, err)

		// Delete the session
		err = store.DeleteSession(helper.ctx, session.ID)
		require.NoError(t, err)

		// Try to get the deleted session
		retrieved, err := store.GetSessionByID(helper.ctx, session.ID)
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("GetNonExistentSession", func(t *testing.T) {
		// Try to get a session with a non-existent ID
		retrieved, err := store.GetSessionByID(helper.ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})
}
