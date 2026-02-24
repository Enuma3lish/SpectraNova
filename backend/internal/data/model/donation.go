package model

import "time"

type Donation struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement"`
	VideoID             uint64    `gorm:"index;not null"`
	DonorID             uint64    `gorm:"index;not null"`
	CreatorID           uint64    `gorm:"index;not null"`
	Amount              float64   `gorm:"type:decimal(10,2);not null"`
	Currency            string    `gorm:"type:varchar(3);not null;default:'USD'"`
	Message             *string   `gorm:"type:text"`
	PaddleTransactionID *string   `gorm:"type:varchar(50);uniqueIndex"`
	PaddleStatus        string    `gorm:"type:varchar(20);not null;default:'pending'"`
	CreatedAt           time.Time
	UpdatedAt           time.Time

	// Relations
	Video   Video `gorm:"foreignKey:VideoID"`
	Donor   User  `gorm:"foreignKey:DonorID"`
	Creator User  `gorm:"foreignKey:CreatorID"`
}
