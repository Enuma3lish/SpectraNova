package model

import "time"

type ViewRecord struct {
	ID       uint64    `gorm:"primaryKey;autoIncrement"`
	VideoID  uint64    `gorm:"index;not null"`
	UserID   *uint64   `gorm:"index"`
	IsMember bool      `gorm:"not null;default:false"`
	ViewedAt time.Time `gorm:"index;not null"`

	// Relations
	Video Video `gorm:"foreignKey:VideoID"`
	User  *User `gorm:"foreignKey:UserID"`
}
