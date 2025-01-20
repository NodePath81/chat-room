package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	// Do not expose password in JSON
	Password  string
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
