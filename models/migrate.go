package models

import (
	"log"

	"gorm.io/gorm"
)

// MigrateDB handles the database migration process
func MigrateDB(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Auto migrate models
	err := db.AutoMigrate(
		&User{},
		&AuthToken{},
		&Project{},
		&Device{},
		&Event{},
		&Metric{},
	)

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
		return err
	}

	log.Println("Database migration completed successfully")
	return nil
}
