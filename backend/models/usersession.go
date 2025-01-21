package models

import (
	"time"

	"github.com/google/uuid"
)

// UserSession represents the many-to-many relationship between users and sessions
type UserSession struct {
	UserID         uuid.UUID `json:"user_id"`
	SessionID      uuid.UUID `json:"session_id"`
	Role           string    `json:"role"`
	JoinedAt       time.Time `json:"joined_at"`
	LastReceivedAt time.Time `json:"last_received"`
}
