package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Session represents a user session
type Session struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key"`
	ProjectID uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	DeviceID  uuid.UUID      `json:"device_id" gorm:"type:uuid;not null"`
	BeginAt   time.Time      `json:"begin_at" gorm:"not null"`
	EndAt     time.Time      `json:"end_at"`
	Duration  time.Duration  `json:"duration"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Project   Project        `json:"-" gorm:"foreignKey:ProjectID;references:ID"`
	Device    Device         `json:"-" gorm:"foreignKey:DeviceID;references:ID"`
}

func (session *Session) BeforeCreate(_ *gorm.DB) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}

	// Set begin at if not set
	now := time.Now()
	if session.BeginAt.IsZero() {
		session.BeginAt = now
	}

	return nil
}
