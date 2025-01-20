package models

import (
	"time"

	"github.com/google/uuid"
)

// UserSession represents the many-to-many relationship between users and sessions
type UserSession struct {
	UserID         uuid.UUID `json:"userId"`
	SessionID      uuid.UUID `json:"sessionId"`
	Role           string    `json:"role"`
	JoinedAt       time.Time `json:"joinedAt"`
	LastReceivedAt time.Time `json:"lastReceivedAt"`
}
