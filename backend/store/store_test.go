package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"chat-room/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockStore implements Store interface for testing
type MockStore struct {
	users    map[uuid.UUID]*models.User
	sessions map[uuid.UUID]*models.Session
	messages map[uuid.UUID]*models.Message
	// user_sessions maps user_id -> session_id -> role
	userSessions map[uuid.UUID]map[uuid.UUID]string
}

func NewMockStore() *MockStore {
	return &MockStore{
		users:        make(map[uuid.UUID]*models.User),
		sessions:     make(map[uuid.UUID]*models.Session),
		messages:     make(map[uuid.UUID]*models.Message),
		userSessions: make(map[uuid.UUID]map[uuid.UUID]string),
	}
}

// Mock implementation of Store interface methods
func (m *MockStore) CreateUser(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockStore) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.User, error) {
	var users []*models.User
	for _, id := range ids {
		if user, ok := m.users[id]; ok {
			users = append(users, user)
		}
	}
	return users, nil
}

// Test cases
func TestMockStore_UserOperations(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	// Test CreateUser
	user := &models.User{
		Username: "testuser",
		Password: "password",
		Nickname: "Test User",
	}
	err := store.CreateUser(ctx, user)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.False(t, user.CreatedAt.IsZero())

	// Test GetUsersByIDs
	users, err := store.GetUsersByIDs(ctx, []uuid.UUID{user.ID})
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, user.Username, users[0].Username)

	// Test GetUsersByIDs with non-existent ID
	users, err = store.GetUsersByIDs(ctx, []uuid.UUID{uuid.New()})
	assert.NoError(t, err)
	assert.Len(t, users, 0)
}

func TestMockStore_BatchOperations(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	// Create test users
	userIDs := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		user := &models.User{
			Username: fmt.Sprintf("user%d", i),
			Password: "password",
			Nickname: fmt.Sprintf("User %d", i),
		}
		err := store.CreateUser(ctx, user)
		assert.NoError(t, err)
		userIDs[i] = user.ID
	}

	// Test batch retrieval
	users, err := store.GetUsersByIDs(ctx, userIDs)
	assert.NoError(t, err)
	assert.Len(t, users, 3)

	// Test partial batch retrieval
	nonExistentID := uuid.New()
	users, err = store.GetUsersByIDs(ctx, append(userIDs, nonExistentID))
	assert.NoError(t, err)
	assert.Len(t, users, 3)
}
