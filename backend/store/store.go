package store

import (
	"context"
	"time"

	"chat-room/models"

	"github.com/google/uuid"
)

// UserStore handles user-related operations
type UserStore interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	CheckNicknameExists(ctx context.Context, nickname string) (bool, error)
}

// SessionStore handles session-related operations
type SessionStore interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error)
	UpdateSession(ctx context.Context, session *models.Session) error
	DeleteSession(ctx context.Context, id uuid.UUID) error
}

// MessageStore handles message-related operations
type MessageStore interface {
	CreateMessage(ctx context.Context, message *models.Message) error
	GetMessagesBySessionID(ctx context.Context, sessionID uuid.UUID, limit int, before time.Time) ([]*models.Message, error)
	DeleteMessage(ctx context.Context, id uuid.UUID) error
}

// UserSessionStore handles operations at the intersection of users and sessions
type UserSessionStore interface {
	AddUserToSession(ctx context.Context, userID, sessionID uuid.UUID, role string) error
	RemoveUserFromSession(ctx context.Context, userID, sessionID uuid.UUID) error
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)
	GetSessionUsers(ctx context.Context, sessionID uuid.UUID) ([]*models.User, error)
	GetUserSessionRole(ctx context.Context, userID, sessionID uuid.UUID) (string, error)
}

// Store combines all sub-stores into a single interface
type Store interface {
	UserStore
	SessionStore
	MessageStore
	UserSessionStore
	BeginTx(ctx context.Context) (Transaction, error)
	Close()
}

// Transaction represents a database transaction
type Transaction interface {
	UserStore
	SessionStore
	MessageStore
	UserSessionStore
	Commit() error
	Rollback() error
}
