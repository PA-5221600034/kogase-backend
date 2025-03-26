package models

import (
	"github.com/google/uuid"
)

// ProjectUser represents the join table between projects and users
type ProjectUser struct {
	ProjectID uuid.UUID `json:"project_id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	Role      string    `json:"role" gorm:"not null;default:contributor"` // contributor or admin
	Project   Project   `json:"-" gorm:"foreignKey:ProjectID"`
	User      User      `json:"-" gorm:"foreignKey:UserID"`
}
