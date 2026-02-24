package model

import (
	"time"

	"gorm.io/gorm"
)

type Channel struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement"`
	UserID     uint64         `gorm:"uniqueIndex;not null"`
	MonthlyFee float64        `gorm:"type:decimal(10,2);not null;default:0"`
	IsHidden   bool           `gorm:"not null;default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`

	// Relations
	User        User         `gorm:"foreignKey:UserID"`
	Memberships []Membership `gorm:"foreignKey:ChannelID"`
}
