package models

import (
	"log"

	"gorm.io/gorm"
)

// MigrateDB handles the database migration process
func MigrateDB(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(&Project{})
	if err != nil {
		log.Printf("Failed to migrate Project table: %v", err)
		return err
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Printf("Failed to migrate user tables: %v", err)
		return err
	}

	err = db.AutoMigrate(&AuthToken{})
	if err != nil {
		log.Printf("Failed to migrate AuthToken table: %v", err)
		return err
	}

	err = db.AutoMigrate(&Device{})
	if err != nil {
		log.Printf("Failed to migrate Device table: %v", err)
		return err
	}

	err = db.AutoMigrate(&Event{})
	if err != nil {
		log.Printf("Failed to migrate Event table: %v", err)
		return err
	}

	log.Println("Database migration completed successfully")
	return nil
}
