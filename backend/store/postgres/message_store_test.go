package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageStore(t *testing.T) {
	helper := setupTestDB(t)
	defer helper.cleanup(t)

	store := helper.createTestStore(t)
	defer store.Close()

	// Create a test user and session for messages
	user := generateTestUser()
	err := store.CreateUser(helper.ctx, user)
	require.NoError(t, err)

	session := generateTestSession(user.ID)
	err = store.CreateSession(helper.ctx, session)
	require.NoError(t, err)

	t.Run("CreateMessage", func(t *testing.T) {
		message := generateTestMessage(user.ID, session.ID)
		err := store.CreateMessage(helper.ctx, message)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, message.ID)
		assert.False(t, message.CreatedAt.IsZero())
		assert.False(t, message.UpdatedAt.IsZero())
	})

	t.Run("GetMessagesBySessionID", func(t *testing.T) {
		// Create multiple test messages
		message1 := generateTestMessage(user.ID, session.ID)
		message2 := generateTestMessage(user.ID, session.ID)
		message3 := generateTestMessage(user.ID, session.ID)

		err := store.CreateMessage(helper.ctx, message1)
		require.NoError(t, err)
		time.Sleep(time.Millisecond * 10) // Ensure different timestamps

		err = store.CreateMessage(helper.ctx, message2)
		require.NoError(t, err)
		time.Sleep(time.Millisecond * 10)

		err = store.CreateMessage(helper.ctx, message3)
		require.NoError(t, err)

		// Get messages with pagination
		messages, err := store.GetMessagesBySessionID(helper.ctx, session.ID, 2, time.Now().UTC().Add(time.Second))
		require.NoError(t, err)
		assert.Len(t, messages, 2)
		assert.True(t, messages[0].CreatedAt.After(messages[1].CreatedAt))

		// Get messages before the second message's timestamp
		messages, err = store.GetMessagesBySessionID(helper.ctx, session.ID, 2, message2.CreatedAt)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.Equal(t, message1.ID, messages[0].ID)
	})

	t.Run("DeleteMessage", func(t *testing.T) {
		// Create a test message
		message := generateTestMessage(user.ID, session.ID)
		err := store.CreateMessage(helper.ctx, message)
		require.NoError(t, err)

		// Delete the message
		err = store.DeleteMessage(helper.ctx, message.ID)
		require.NoError(t, err)

		// Try to get messages after deletion
		messages, err := store.GetMessagesBySessionID(helper.ctx, session.ID, 1, time.Now().UTC().Add(time.Second))
		require.NoError(t, err)
		for _, msg := range messages {
			assert.NotEqual(t, message.ID, msg.ID)
		}
	})

	t.Run("GetMessagesFromNonExistentSession", func(t *testing.T) {
		messages, err := store.GetMessagesBySessionID(helper.ctx, uuid.New(), 10, time.Now().UTC())
		require.NoError(t, err)
		assert.Empty(t, messages)
	})

	t.Run("GetMessagesWithDifferentLimits", func(t *testing.T) {
		// Create 5 messages
		for i := 0; i < 5; i++ {
			message := generateTestMessage(user.ID, session.ID)
			err := store.CreateMessage(helper.ctx, message)
			require.NoError(t, err)
			time.Sleep(time.Millisecond * 10)
		}

		// Test different limit values
		limits := []int{1, 3, 5, 10}
		for _, limit := range limits {
			messages, err := store.GetMessagesBySessionID(helper.ctx, session.ID, limit, time.Now().UTC().Add(time.Second))
			require.NoError(t, err)
			assert.LessOrEqual(t, len(messages), limit)

			// Verify messages are ordered by creation time (descending)
			for i := 1; i < len(messages); i++ {
				assert.True(t, messages[i-1].CreatedAt.After(messages[i].CreatedAt))
			}
		}
	})
}
