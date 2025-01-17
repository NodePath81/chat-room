package models

import (
	"gorm.io/gorm"
)

type Session struct {
	gorm.Model
	Name      string    `json:"name" gorm:"not null"`
	CreatorID uint      `json:"creatorId" gorm:"not null"`
	Users     []User    `gorm:"many2many:user_sessions;"`
	Messages  []Message `gorm:"foreignKey:SessionID"`
}
