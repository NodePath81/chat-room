package models

import (
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
)

type Message struct {
	ID        uuid.UUID   `json:"id"`
	Type      MessageType `json:"type"`
	Content   string      `json:"content"`
	UserID    uuid.UUID   `json:"userId"`
	SessionID uuid.UUID   `json:"sessionId"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}
