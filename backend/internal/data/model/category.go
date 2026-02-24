package model

import "time"

type Category struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Slug      string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relations
	Videos []Video `gorm:"foreignKey:CategoryID"`
}
