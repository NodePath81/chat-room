package postgres

import (
	"chat-room/models"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	testDBURL = "postgres://postgres:postgres@localhost:5432/chat_test"
)

type testHelper struct {
	db  *pgxpool.Pool
	ctx context.Context
}

func setupTestDB(t *testing.T) *testHelper {
	ctx := context.Background()

	// Connect to the test database
	pool, err := pgxpool.New(ctx, testDBURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Verify the connection
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	helper := &testHelper{
		db:  pool,
		ctx: ctx,
	}

	// Clean up any existing data
	helper.cleanup(t)

	return helper
}

func (h *testHelper) cleanup(t *testing.T) {
	// Clean up test data
	tables := []string{
		"schema_migrations",
		"messages",
		"user_sessions",
		"sessions",
		"users",
	}
	for _, table := range tables {
		_, err := h.db.Exec(h.ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			t.Errorf("Failed to clean up %s table: %v", table, err)
		}
	}
}

func (h *testHelper) createTestStore(t *testing.T) *Store {
	store, err := New(h.ctx, testDBURL)
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	// Apply migrations
	if err := store.Migrate(h.ctx); err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}

	return store
}

// Test data generators
func generateTestUser() *models.User {
	return &models.User{
		ID:        uuid.New(),
		Username:  fmt.Sprintf("test_user_%s", uuid.NewString()[:8]),
		Password:  "test_password",
		Nickname:  fmt.Sprintf("Test User %s", uuid.NewString()[:8]),
		AvatarURL: "https://example.com/avatar.jpg",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

func generateTestSession(creatorID uuid.UUID) *models.Session {
	return &models.Session{
		ID:        uuid.New(),
		Name:      fmt.Sprintf("test_session_%s", uuid.NewString()[:8]),
		CreatorID: creatorID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

func generateTestMessage(userID, sessionID uuid.UUID) *models.Message {
	return &models.Message{
		ID:        uuid.New(),
		Type:      "text",
		Content:   fmt.Sprintf("Test message %s", uuid.NewString()[:8]),
		UserID:    userID,
		SessionID: sessionID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}
