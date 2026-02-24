package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement"`
	Username    string         `gorm:"type:varchar(50);uniqueIndex;not null"`
	DisplayName string         `gorm:"type:varchar(100);not null"`
	Password    string         `gorm:"type:varchar(255);not null"`
	AvatarURL   *string        `gorm:"type:varchar(500)"`
	Role        string         `gorm:"type:varchar(20);not null;default:'user'"`
	IsHidden    bool           `gorm:"not null;default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relations
	Channel        *Channel          `gorm:"foreignKey:UserID"`
	Videos         []Video           `gorm:"foreignKey:UserID"`
	Memberships    []Membership      `gorm:"foreignKey:UserID"`
	TagPreferences []UserTagPreference `gorm:"foreignKey:UserID"`
}
