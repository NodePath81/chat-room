package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username  string    `gorm:"uniqueIndex;not null"`
	Password  string    `gorm:"not null"`
	AvatarURL string    `gorm:"type:text"`
	Sessions  []Session `gorm:"many2many:user_sessions;"`
}
