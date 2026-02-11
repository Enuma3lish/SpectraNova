package model
import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           int64          `gorm:"primaryKey;autoIncrement"`
	Username     string         `gorm:"uniqueIndex;size:64;not null"`
	PasswordHash string         `gorm:"size:255;not null"`
	DisplayName  string         `gorm:"size:128;not null"`






}	DeletedAt    gorm.DeletedAt `gorm:"index"`	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`	CreatedAt    time.Time      `gorm:"autoCreateTime"`	IsHidden     bool           `gorm:"not null;default:false"`	Role         string         `gorm:"size:16;not null;default:user"`