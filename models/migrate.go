package models

import (
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

	err = db.AutoMigrate(&Session{})
	if err != nil {
		log.Printf("Failed to migrate Session table: %v", err)
		return err
	}

	err = db.AutoMigrate(&Event{})
	if err != nil {
		log.Printf("Failed to migrate Event table: %v", err)
		return err
	}

	// Create default admin user if it doesn't exist
	if err := createDefaultAdminUser(db); err != nil {
		log.Printf("Warning: Failed to create default admin user: %v", err)
		// Continue anyway, not a critical error
	}

	log.Println("Database migration completed successfully")
	return nil
}

// createDefaultAdminUser creates a default admin user if no users exist in the database
func createDefaultAdminUser(db *gorm.DB) error {
	// Check if any users exist
	var count int64
	if err := db.Model(&User{}).Count(&count).Error; err != nil {
		return err
	}

	// Only create default admin if no users exist
	if count == 0 {
		// Default admin credentials - these should be changed after first login
		defaultEmail := "admin@kogase.io"
		defaultPassword := "Admin@123" // This should be a secure password

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		defaultAdmin := User{
			ID:        uuid.New(),
			Email:     defaultEmail,
			Password:  string(hashedPassword),
			Name:      "Admin",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Create the admin user
		if err := db.Create(&defaultAdmin).Error; err != nil {
			return err
		}

		log.Printf("Created default admin user: %s", defaultEmail)

		// // Optionally, create a default project for the admin
		// defaultProject := Project{
		// 	ID:        uuid.New(),
		// 	Name:      "Default Project",
		// 	ApiKey:    uuid.New().String(),
		// 	OwnerID:   defaultAdmin.ID,
		// 	CreatedAt: time.Now(),
		// 	UpdatedAt: time.Now(),
		// }

		// if err := db.Create(&defaultProject).Error; err != nil {
		// 	return err
		// }

		// log.Printf("Created default project for admin user")
	}

	return nil
}
