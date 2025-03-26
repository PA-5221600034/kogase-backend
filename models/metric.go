package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MetricType represents the type of metric being tracked
type MetricType string

const (
	DailyActiveUsers   MetricType = "dau"
	MonthlyActiveUsers MetricType = "mau"
	NewUsers           MetricType = "new_users"
	SessionCount       MetricType = "session_count"
	SessionLength      MetricType = "session_length"
	EventCount         MetricType = "event_count"
	RetentionRate      MetricType = "retention_rate"
)

// PeriodType represents the time period for metrics aggregation
type PeriodType string

const (
	Hourly  PeriodType = "hourly"
	Daily   PeriodType = "daily"
	Weekly  PeriodType = "weekly"
	Monthly PeriodType = "monthly"
	Yearly  PeriodType = "yearly"
	Total   PeriodType = "total"
)

// Dimensions is a JSON object for metric dimensions (platform, country, etc)
type Dimensions map[string]string

// Value makes the Dimensions struct implement the driver.Valuer interface.
func (d Dimensions) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan makes the Dimensions struct implement the sql.Scanner interface.
func (d *Dimensions) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &d)
}

// Metric represents a pre-aggregated metric
type Metric struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key"`
	ProjectID   uuid.UUID      `json:"project_id" gorm:"type:uuid;not null"`
	MetricType  MetricType     `json:"metric_type" gorm:"not null;type:varchar(20)"`
	Period      PeriodType     `json:"period" gorm:"not null;type:varchar(10)"`
	PeriodStart time.Time      `json:"period_start" gorm:"not null"`
	Value       float64        `json:"value" gorm:"not null"`
	Dimensions  Dimensions     `json:"dimensions" gorm:"type:jsonb;default:'{}'"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Project     Project        `json:"-" gorm:"foreignKey:ProjectID"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (metric *Metric) BeforeCreate(_ *gorm.DB) error {
	if metric.ID == uuid.Nil {
		metric.ID = uuid.New()
	}
	return nil
}
