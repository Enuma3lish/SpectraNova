package model

import (
	"time"

	"gorm.io/gorm"
)

type Video struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement"`
	UserID         uint64         `gorm:"index;not null"`
	CategoryID     uint64         `gorm:"index;not null"`
	Title          string         `gorm:"type:varchar(200);not null"`
	Description    *string        `gorm:"type:text"`
	VideoURL       string         `gorm:"type:varchar(500);not null"`
	ThumbnailURL   *string        `gorm:"type:varchar(500)"`
	Duration       uint32         `gorm:"not null;default:0"`
	ViewsMember    uint64         `gorm:"not null;default:0"`
	ViewsNonMember uint64         `gorm:"not null;default:0"`
	AccessTier     int8           `gorm:"not null;default:0"` // 0=public, 1=subscriber, 2=premium
	IsPublished    bool           `gorm:"not null;default:true"`
	IsHidden       bool           `gorm:"not null;default:false"`
	CreatedAt      time.Time      `gorm:"index"`
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	// Relations
	User     User     `gorm:"foreignKey:UserID"`
	Category Category `gorm:"foreignKey:CategoryID"`
	Tags     []Tag    `gorm:"many2many:video_tags"`
}
