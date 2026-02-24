package model

import "time"

type Tag struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Slug      string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relations
	Videos []Video `gorm:"many2many:video_tags"`
}

type UserTagPreference struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	UserID    *uint64   `gorm:"index"`
	TagID     uint64    `gorm:"not null"`
	SessionID *string   `gorm:"type:varchar(100);index"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relations
	User *User `gorm:"foreignKey:UserID"`
	Tag  Tag   `gorm:"foreignKey:TagID"`
}
