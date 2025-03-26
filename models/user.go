package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user of the dashboard
type User struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key"`
	Email     string         `json:"email" gorm:"unique;not null"`
	Password  string         `json:"-" gorm:"not null"` // Password is not exposed in JSON responses
	Name      string         `json:"name" gorm:"not null"`
	Role      string         `json:"role" gorm:"not null;default:developer"` // 'admin' or 'developer'
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Projects  []Project      `json:"projects,omitempty" gorm:"many2many:project_users;"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (user *User) BeforeCreate(_ *gorm.DB) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	return nil
}
