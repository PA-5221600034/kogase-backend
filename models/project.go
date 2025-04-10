package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key"`
	Name      string         `json:"name" gorm:"not null"`
	ApiKey    string         `json:"api_key,omitempty" gorm:"unique;not null"`
	OwnerID   uuid.UUID      `json:"owner_id" gorm:"type:uuid;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Owner     User           `json:"owner,omitempty" gorm:"foreignKey:OwnerID;references:ID"`
	Devices   []Device       `json:"devices,omitempty" gorm:"foreignKey:ProjectID;references:ID"`
	Events    []Event        `json:"events,omitempty" gorm:"foreignKey:ProjectID;references:ID"`
}

func (project *Project) BeforeCreate(tx *gorm.DB) error {
	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}

	if project.ApiKey == "" {
		project.ApiKey = uuid.New().String()
	}

	return nil
}
