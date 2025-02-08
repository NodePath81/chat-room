// Package store provides interfaces for data persistence operations.
package store

import (
	"context"
	"time"

	"chat-room/models"

	"github.com/google/uuid"
)

// UserStore defines operations for managing user data.
type UserStore interface {
	// CreateUser creates a new user in the store.
	// If user.ID is nil, it will be generated.
	// If user.CreatedAt is zero, it will be set to current time.
	CreateUser(ctx context.Context, user *models.User) error

	// GetUsersByIDs retrieves multiple users by their IDs.
	// Returns a slice of users in no particular order.
	// If some IDs don't exist, they will be omitted from the result.
	GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.User, error)

	// GetUserByUsername retrieves a user by their username.
	// Returns ErrNotFound if the user doesn't exist.
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)

	// UpdateUser updates an existing user's information.
	// Only updates username, nickname, and avatar_url fields.
	UpdateUser(ctx context.Context, user *models.User) error

	// DeleteUser removes a user from the store.
	// This operation is irreversible.
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// CheckUsernameExists checks if a username is already taken.
	CheckUsernameExists(ctx context.Context, username string) (bool, error)

	// CheckNicknameExists checks if a nickname is already taken.
	CheckNicknameExists(ctx context.Context, nickname string) (bool, error)
}

// SessionStore defines operations for managing chat sessions.
type SessionStore interface {
	// CreateSession creates a new chat session.
	// If session.ID is nil, it will be generated.
	// If session.CreatedAt is zero, it will be set to current time.
	CreateSession(ctx context.Context, session *models.Session) error

	// GetSessionByID retrieves a session by its ID.
	// Returns ErrNotFound if the session doesn't exist.
	GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error)

	// UpdateSession updates an existing session's information.
	// Only updates the name field.
	UpdateSession(ctx context.Context, session *models.Session) error

	// DeleteSession removes a session and all its associated data.
	// This operation is irreversible.
	DeleteSession(ctx context.Context, id uuid.UUID) error
}

// MessageStore defines operations for managing chat messages.
type MessageStore interface {
	// CreateMessage creates a new message in a session.
	// If message.ID is nil, it will be generated.
	// If message.Timestamp is zero, it will be set to current time.
	CreateMessage(ctx context.Context, message *models.Message) error

	// DeleteMessage removes a message from the store.
	// This operation is irreversible.
	DeleteMessage(ctx context.Context, id uuid.UUID) error

	// GetMessageIDsBySessionID retrieves message IDs for a session.
	// Returns IDs ordered by timestamp DESC, limited by the limit parameter.
	// Only returns messages with timestamp before the specified time.
	GetMessageIDsBySessionID(ctx context.Context, sessionID uuid.UUID, limit int, before time.Time) ([]uuid.UUID, error)

	// GetMessagesByIDs retrieves multiple messages by their IDs.
	// Returns a slice of messages in no particular order.
	// If some IDs don't exist, they will be omitted from the result.
	GetMessagesByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Message, error)
}

// UserSessionStore defines operations for managing user-session relationships.
type UserSessionStore interface {
	// AddUserToSession adds a user to a session with the specified role.
	// The joined_at timestamp will be set to current time.
	AddUserToSession(ctx context.Context, userID, sessionID uuid.UUID, role string) error

	// RemoveUserFromSession removes a user from a session.
	// This operation is irreversible.
	RemoveUserFromSession(ctx context.Context, userID, sessionID uuid.UUID) error

	// GetSessionIDsByUserID retrieves all session IDs a user is a member of.
	GetSessionIDsByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)

	// GetUserIDsBySessionID retrieves all user IDs that are members of a session.
	GetUserIDsBySessionID(ctx context.Context, sessionID uuid.UUID) ([]uuid.UUID, error)

	// GetUserSessionsBySessionIDAndUserIDs retrieves all user sessions by session ID and user IDs.
	GetUserSessionsBySessionIDAndUserIDs(ctx context.Context, sessionID uuid.UUID, userIDs []uuid.UUID) ([]*models.UserSession, error)
}

// Store combines all sub-stores into a single interface.
// It provides transaction support and manages the lifecycle of the store.
type Store interface {
	UserStore
	SessionStore
	MessageStore
	UserSessionStore

	// BeginTx starts a new transaction.
	// The transaction must be committed or rolled back.
	BeginTx(ctx context.Context) (Transaction, error)

	// Close releases any resources held by the store.
	Close()
}

// Transaction represents a database transaction.
// It combines all store operations that can be executed within a transaction.
type Transaction interface {
	UserStore
	SessionStore
	MessageStore
	UserSessionStore

	// Commit commits the transaction.
	Commit() error

	// Rollback aborts the transaction.
	Rollback() error
}
