package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Device struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primary_key"`
	ProjectID       uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	Identifier      string         `json:"identifier" gorm:"not null"`           // Client-generated device identifier
	Platform        string         `json:"platform" gorm:"not null"`             // iOS, Android, Windows, etc.
	PlatformVersion string         `json:"platform_version" gorm:"not null"`     // e.g., "10.0", "Android 11"
	AppVersion      string         `json:"app_version" gorm:"not null"`          // App version
	FirstSeen       time.Time      `json:"first_seen" gorm:"not null"`           // First session timestamp
	LastSeen        time.Time      `json:"last_seen" gorm:"not null"`            // Last session timestamp
	IpAddress       string         `json:"ip_address,omitempty" gorm:"not null"` // Hashed/anonymized IP address
	Country         string         `json:"country,omitempty"`                    // Country based on IP (optional)
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
	Project         Project        `json:"-" gorm:"foreignKey:ProjectID;references:ID"`
	Events          []Event        `json:"events,omitempty" gorm:"foreignKey:DeviceID;references:ID"`
}

func (device *Device) BeforeCreate(_ *gorm.DB) error {
	if device.ID == uuid.Nil {
		device.ID = uuid.New()
	}

	now := time.Now()
	if device.FirstSeen.IsZero() {
		device.FirstSeen = now
	}
	if device.LastSeen.IsZero() {
		device.LastSeen = now
	}

	return nil
}
