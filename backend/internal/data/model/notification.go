package model

import (
	"time"

	"gorm.io/datatypes"
)

type Notification struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement"`
	UserID    uint64         `gorm:"index;not null"`
	Type      string         `gorm:"type:varchar(30);not null"`
	Title     string         `gorm:"type:varchar(200);not null"`
	Message   *string        `gorm:"type:text"`
	Payload   datatypes.JSON `gorm:"type:json"`
	IsRead    bool           `gorm:"not null;default:false"`
	CreatedAt time.Time      `gorm:"index"`

	// Relations
	User User `gorm:"foreignKey:UserID"`
}
