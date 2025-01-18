package models

import (
	"time"

	"gorm.io/gorm"
)

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
)

type Message struct {
	gorm.Model
	Type      MessageType `json:"type"`
	Content   string      `json:"content"`
	UserID    uint        `json:"userId"`
	User      User        `gorm:"foreignKey:UserID"`
	SessionID uint        `json:"sessionId"`
	Session   Session     `gorm:"foreignKey:SessionID"`
	CreatedAt time.Time   `json:"createdAt"`
}
