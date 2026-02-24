package model

import "time"

type Membership struct {
	ID                   uint64     `gorm:"primaryKey;autoIncrement"`
	ChannelID            uint64     `gorm:"not null;uniqueIndex:idx_channel_user"`
	UserID               uint64     `gorm:"not null;uniqueIndex:idx_channel_user"`
	Tier                 int8       `gorm:"not null;default:1"` // 1=free, 2=premium
	Status               string     `gorm:"type:varchar(20);not null;default:'active'"`
	PaddleSubscriptionID *string    `gorm:"type:varchar(50);uniqueIndex"`
	PaddleStatus         *string    `gorm:"type:varchar(20)"`
	StartedAt            time.Time  `gorm:"not null"`
	ExpiresAt            *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time

	// Relations
	Channel Channel `gorm:"foreignKey:ChannelID"`
	User    User    `gorm:"foreignKey:UserID"`
}
