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
	UserID    uuid.UUID   `json:"user_id"`
	SessionID uuid.UUID   `json:"session_id"`
	Timestamp time.Time   `json:"timestamp"`
}
