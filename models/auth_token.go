package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthToken struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key"`
	UserID     uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	Token      string         `json:"token" gorm:"not null"`
	ExpiresAt  time.Time      `json:"expires_at" gorm:"not null"`
	LastUsedAt time.Time      `json:"last_used_at" gorm:"not null"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	User       User           `json:"-" gorm:"foreignKey:UserID;references:ID"`
}

func (token *AuthToken) BeforeCreate(_ *gorm.DB) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}

	if token.LastUsedAt.IsZero() {
		token.LastUsedAt = time.Now()
	}

	return nil
}
