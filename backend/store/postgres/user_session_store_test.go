package postgres

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSessionStore(t *testing.T) {
	helper := setupTestDB(t)
	defer helper.cleanup(t)

	store := helper.createTestStore(t)
	defer store.Close()

	// Create test users and a session
	user1 := generateTestUser()
	user2 := generateTestUser()
	err := store.CreateUser(helper.ctx, user1)
	require.NoError(t, err)
	err = store.CreateUser(helper.ctx, user2)
	require.NoError(t, err)

	session := generateTestSession(user1.ID)
	err = store.CreateSession(helper.ctx, session)
	require.NoError(t, err)

	t.Run("AddUserToSession", func(t *testing.T) {
		err := store.AddUserToSession(helper.ctx, user2.ID, session.ID, "member")
		require.NoError(t, err)

		// Verify user was added
		users, err := store.GetSessionUsers(helper.ctx, session.ID)
		require.NoError(t, err)
		assert.Len(t, users, 2) // Including the creator

		found := false
		for _, user := range users {
			if user.ID == user2.ID {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("GetUserSessions", func(t *testing.T) {
		// Create another session
		session2 := generateTestSession(user1.ID)
		err := store.CreateSession(helper.ctx, session2)
		require.NoError(t, err)

		err = store.AddUserToSession(helper.ctx, user2.ID, session2.ID, "member")
		require.NoError(t, err)

		// Get sessions for user2
		userSessions, err := store.GetUserSessions(helper.ctx, user2.ID)
		require.NoError(t, err)
		assert.Len(t, userSessions, 2)

		// Verify session IDs and roles
		sessionMap := make(map[uuid.UUID]string)
		for _, us := range userSessions {
			sessionMap[us.SessionID] = us.Role
			assert.Equal(t, user2.ID, us.UserID)
			assert.False(t, us.JoinedAt.IsZero())
		}

		assert.Equal(t, "member", sessionMap[session.ID])
		assert.Equal(t, "member", sessionMap[session2.ID])
	})

	t.Run("GetSessionUsers", func(t *testing.T) {
		users, err := store.GetSessionUsers(helper.ctx, session.ID)
		require.NoError(t, err)
		assert.Len(t, users, 2)

		userIDs := make(map[uuid.UUID]bool)
		for _, u := range users {
			userIDs[u.ID] = true
		}
		assert.True(t, userIDs[user1.ID])
		assert.True(t, userIDs[user2.ID])
	})

	t.Run("GetUserSessionRole", func(t *testing.T) {
		// Check creator role
		role, err := store.GetUserSessionRole(helper.ctx, user1.ID, session.ID)
		require.NoError(t, err)
		assert.Equal(t, "creator", role)

		// Check member role
		role, err = store.GetUserSessionRole(helper.ctx, user2.ID, session.ID)
		require.NoError(t, err)
		assert.Equal(t, "member", role)
	})

	t.Run("RemoveUserFromSession", func(t *testing.T) {
		err := store.RemoveUserFromSession(helper.ctx, user2.ID, session.ID)
		require.NoError(t, err)

		// Verify user was removed
		users, err := store.GetSessionUsers(helper.ctx, session.ID)
		require.NoError(t, err)
		assert.Len(t, users, 1) // Only creator remains

		// Verify user's sessions no longer include this session
		userSessions, err := store.GetUserSessions(helper.ctx, user2.ID)
		require.NoError(t, err)
		for _, us := range userSessions {
			assert.NotEqual(t, session.ID, us.SessionID)
		}
	})

	t.Run("NonExistentUserSession", func(t *testing.T) {
		// Try to get role for non-existent user-session relationship
		role, err := store.GetUserSessionRole(helper.ctx, uuid.New(), session.ID)
		assert.Error(t, err)
		assert.Empty(t, role)

		// Try to get sessions for non-existent user
		userSessions, err := store.GetUserSessions(helper.ctx, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, userSessions)

		// Try to get users for non-existent session
		users, err := store.GetSessionUsers(helper.ctx, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, users)
	})
}
