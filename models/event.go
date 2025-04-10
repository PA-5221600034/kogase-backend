package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Payloads map[string]interface{}

func (p Payloads) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Payloads) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &p)
}

type Event struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key"`
	ProjectID  uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	DeviceID   uuid.UUID      `json:"device_id" gorm:"type:uuid;not null"`
	EventType  string         `json:"event_type" gorm:"not null;type:varchar(50)"`
	EventName  string         `json:"event_name" gorm:"not null"`              // For custom events
	Payloads   Payloads       `json:"payloads" gorm:"type:jsonb;default:'{}'"` // JSON payloads
	Timestamp  time.Time      `json:"timestamp" gorm:"not null"`               // When event occurred (client-side)
	ReceivedAt time.Time      `json:"received_at" gorm:"not null"`             // When event was received by server
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	Project    Project        `json:"-" gorm:"foreignKey:ProjectID;references:ID"`
	Device     Device         `json:"-" gorm:"foreignKey:DeviceID;references:ID"`
}

func (event *Event) BeforeCreate(_ *gorm.DB) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	if event.ReceivedAt.IsZero() {
		event.ReceivedAt = time.Now()
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	return nil
}
