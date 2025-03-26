package integration

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/atqamz/kogase-backend/config"
	"github.com/atqamz/kogase-backend/models"
	"github.com/atqamz/kogase-backend/server"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testDB     *gorm.DB
	testRouter *gin.Engine
	testConfig *config.Config
	testServer *server.Server
)

func TestMain(m *testing.M) {
	// Setup test environment
	setupTestEnv()

	// Run tests
	exitCode := m.Run()

	// Cleanup test environment
	cleanupTestEnv()

	os.Exit(exitCode)
}

func setupTestEnv() {
	// Load test configuration
	testConfig = &config.Config{
		DBHost:        "localhost",
		DBPort:        "5432",
		DBUser:        "postgres",
		DBPassword:    "postgres",
		DBName:        "kogase_test",
		DBSSLMode:     "disable",
		JWTSecret:     "test_secret_key",
		JWTExpiration: "24h",
		Port:          "8080",
	}

	// Environment variables can override the default test config
	if os.Getenv("TEST_DB_HOST") != "" {
		testConfig.DBHost = os.Getenv("TEST_DB_HOST")
	}
	if os.Getenv("TEST_DB_PORT") != "" {
		testConfig.DBPort = os.Getenv("TEST_DB_PORT")
	}
	if os.Getenv("TEST_DB_USER") != "" {
		testConfig.DBUser = os.Getenv("TEST_DB_USER")
	}
	if os.Getenv("TEST_DB_PASSWORD") != "" {
		testConfig.DBPassword = os.Getenv("TEST_DB_PASSWORD")
	}
	if os.Getenv("TEST_DB_NAME") != "" {
		testConfig.DBName = os.Getenv("TEST_DB_NAME")
	}

	// Connect to the database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		testConfig.DBHost, testConfig.DBPort, testConfig.DBUser, testConfig.DBPassword, testConfig.DBName, testConfig.DBSSLMode)

	// Customize logger to reduce noise during tests
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent, // Silent for tests
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	var err error
	testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate database schema for testing
	err = models.MigrateDB(testDB)
	if err != nil {
		log.Fatalf("Failed to migrate test database: %v", err)
	}

	// Set up test server
	gin.SetMode(gin.TestMode)
	testServer = server.NewWithConfig(testDB, testConfig)
	testRouter = testServer.Router
}

func cleanupTestEnv() {
	// Get the underlying SQL DB
	sqlDB, err := testDB.DB()
	if err != nil {
		log.Printf("Error getting SQL DB: %v", err)
		return
	}

	// Drop all tables in the test database
	testDB.Exec("DROP SCHEMA public CASCADE")
	testDB.Exec("CREATE SCHEMA public")

	// Close the database connection
	sqlDB.Close()
}
