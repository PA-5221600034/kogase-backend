package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EventType represents the type of telemetry event
type EventType string

const (
	SessionStart EventType = "kogase_predefined_session_start"
	SessionEnd   EventType = "kogase_predefined_session_end"
	Install      EventType = "kogase_predefined_install"
)

// Parameters is a JSON object for event parameters
type Parameters map[string]interface{}

// Value makes the Parameters struct implement the driver.Valuer interface.
// This method is used to convert the Parameters to a value that can be stored in the database.
func (p Parameters) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan makes the Parameters struct implement the sql.Scanner interface.
// This method is used to convert a value from the database to a Parameters.
func (p *Parameters) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &p)
}

// Event represents a telemetry event
type Event struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key"`
	ProjectID  uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	DeviceID   uuid.UUID      `json:"device_id" gorm:"type:uuid;not null"`
	EventType  EventType      `json:"event_type" gorm:"not null;type:varchar(50)"`
	EventName  string         `json:"event_name" gorm:"not null"`                // For custom events
	Parameters Parameters     `json:"parameters" gorm:"type:jsonb;default:'{}'"` // JSON parameters
	Timestamp  time.Time      `json:"timestamp" gorm:"not null"`                 // When event occurred (client-side)
	ReceivedAt time.Time      `json:"received_at" gorm:"not null"`               // When event was received by server
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	Project    Project        `json:"-" gorm:"foreignKey:ProjectID;references:ID"`
	Device     Device         `json:"-" gorm:"foreignKey:DeviceID;references:ID"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (event *Event) BeforeCreate(_ *gorm.DB) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	// Set received timestamp if not set
	if event.ReceivedAt.IsZero() {
		event.ReceivedAt = time.Now()
	}

	// Set event timestamp to now if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	return nil
}
