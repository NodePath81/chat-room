package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	Content   string `json:"content"`
	UserID    uint   `json:"userId"`
	User      User   `gorm:"foreignKey:UserID"`
	SessionID uint   `json:"sessionId"`
	Session   Session `gorm:"foreignKey:SessionID"`
	CreatedAt time.Time `json:"createdAt"`
} 