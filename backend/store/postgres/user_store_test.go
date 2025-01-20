package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStore(t *testing.T) {
	helper := setupTestDB(t)
	defer helper.cleanup(t)

	store := helper.createTestStore(t)
	defer store.Close()

	t.Run("CreateUser", func(t *testing.T) {
		user := generateTestUser()
		err := store.CreateUser(helper.ctx, user)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("GetUserByID", func(t *testing.T) {
		// Create a test user
		user := generateTestUser()
		err := store.CreateUser(helper.ctx, user)
		require.NoError(t, err)

		// Get the user by ID
		retrieved, err := store.GetUserByID(helper.ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Username, retrieved.Username)
		assert.Equal(t, user.Password, retrieved.Password)
		assert.Equal(t, user.Nickname, retrieved.Nickname)
		assert.Equal(t, user.AvatarURL, retrieved.AvatarURL)
	})

	t.Run("GetUserByUsername", func(t *testing.T) {
		// Create a test user
		user := generateTestUser()
		err := store.CreateUser(helper.ctx, user)
		require.NoError(t, err)

		// Get the user by username
		retrieved, err := store.GetUserByUsername(helper.ctx, user.Username)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Username, retrieved.Username)
		assert.Equal(t, user.Password, retrieved.Password)
		assert.Equal(t, user.Nickname, retrieved.Nickname)
		assert.Equal(t, user.AvatarURL, retrieved.AvatarURL)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		// Create a test user
		user := generateTestUser()
		err := store.CreateUser(helper.ctx, user)
		require.NoError(t, err)

		// Update user fields
		user.Nickname = "Updated Nickname"
		user.AvatarURL = "https://example.com/new-avatar.jpg"
		originalUpdatedAt := user.UpdatedAt

		// Wait a moment to ensure UpdatedAt will be different
		time.Sleep(time.Millisecond * 10)

		err = store.UpdateUser(helper.ctx, user)
		require.NoError(t, err)

		// Get the updated user
		retrieved, err := store.GetUserByID(helper.ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Nickname, retrieved.Nickname)
		assert.Equal(t, user.AvatarURL, retrieved.AvatarURL)
		assert.True(t, retrieved.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("DeleteUser", func(t *testing.T) {
		// Create a test user
		user := generateTestUser()
		err := store.CreateUser(helper.ctx, user)
		require.NoError(t, err)

		// Delete the user
		err = store.DeleteUser(helper.ctx, user.ID)
		require.NoError(t, err)

		// Try to get the deleted user
		retrieved, err := store.GetUserByID(helper.ctx, user.ID)
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("CheckUsernameExists", func(t *testing.T) {
		// Create a test user
		user := generateTestUser()
		err := store.CreateUser(helper.ctx, user)
		require.NoError(t, err)

		// Check if username exists
		exists, err := store.CheckUsernameExists(helper.ctx, user.Username)
		require.NoError(t, err)
		assert.True(t, exists)

		// Check non-existent username
		exists, err = store.CheckUsernameExists(helper.ctx, "nonexistent_user")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("CheckNicknameExists", func(t *testing.T) {
		// Create a test user
		user := generateTestUser()
		err := store.CreateUser(helper.ctx, user)
		require.NoError(t, err)

		// Check if nickname exists
		exists, err := store.CheckNicknameExists(helper.ctx, user.Nickname)
		require.NoError(t, err)
		assert.True(t, exists)

		// Check non-existent nickname
		exists, err = store.CheckNicknameExists(helper.ctx, "Nonexistent User")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}
