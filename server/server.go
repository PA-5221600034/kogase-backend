package server

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/atqamz/kogase-backend/config"
	"github.com/atqamz/kogase-backend/controllers"
	"github.com/atqamz/kogase-backend/middleware"
	"github.com/atqamz/kogase-backend/models"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Server represents the main server application
type Server struct {
	Router *gin.Engine
	DB     *gorm.DB
	Config *config.Config
}

// New creates a new server instance
func New() (*Server, error) {
	// Load configuration from environment
	cfg := config.NewConfigFromEnv()

	// Database connection
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	// Configure GORM logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Migrate the schema
	if err := models.MigrateDB(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Set up Gin
	r := gin.Default()

	// Create a new server
	s := &Server{
		Router: r,
		DB:     db,
		Config: cfg,
	}

	// Initialize routes
	s.setupRoutes()

	return s, nil
}

// NewWithConfig creates a new server with custom configuration (useful for testing)
func NewWithConfig(db *gorm.DB, cfg *config.Config) *Server {
	// Set up Gin
	r := gin.Default()

	// Create a new server
	s := &Server{
		Router: r,
		DB:     db,
		Config: cfg,
	}

	// Initialize routes
	s.setupRoutes()

	return s
}

// setupRoutes sets up all the routes
func (s *Server) setupRoutes() {
	// Global middleware
	s.Router.Use(middleware.CORSMiddleware())

	// Swagger documentation
	s.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create controllers
	authController := controllers.NewAuthController(s.DB)
	deviceController := controllers.NewDeviceController(s.DB)
	healthController := controllers.NewHealthController(s.DB)
	projectController := controllers.NewProjectController(s.DB)
	telemetryController := controllers.NewTelemetryController(s.DB)
	userController := controllers.NewUserController(s.DB)

	// API v1 routes
	v1 := s.Router.Group("/api/v1")

	// Auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authController.Login)
		auth.POST("/logout", middleware.AuthMiddleware(s.DB), authController.Logout)
		auth.GET("/me", middleware.AuthMiddleware(s.DB), authController.Me)
	}

	// Device routes
	devices := v1.Group("/devices")
	{
		// API key routes (for game clients)
		apiKeyDevices := devices.Group("")
		apiKeyDevices.Use(middleware.ApiKeyMiddleware(s.DB))
		{
			apiKeyDevices.POST("", deviceController.CreateDevice)
			apiKeyDevices.GET("/:id", deviceController.GetDevice)
			apiKeyDevices.PATCH("/:id", deviceController.UpdateDevice)
		}

		// Auth routes (for dashboard)
		authDevices := devices.Group("")
		authDevices.Use(middleware.AuthMiddleware(s.DB))
		{
			authDevices.GET("", deviceController.GetDevices)
			authDevices.DELETE("/:id", deviceController.DeleteDevice)
		}
	}

	health := v1.Group("/health")
	{
		health.GET("", healthController.GetHealth)
		health.GET("/apikey", middleware.ApiKeyMiddleware(s.DB), healthController.GetHealthWithApiKey)
	}

	// Project routes
	projects := v1.Group("/projects")
	{
		projects.GET("/apikey", middleware.ApiKeyMiddleware(s.DB), projectController.GetProjectWithApiKey)
		projects.POST("", projectController.CreateProject)

		authProjects := projects.Group("")
		authProjects.Use(middleware.AuthMiddleware(s.DB))
		{
			authProjects.GET("/:id", projectController.GetProject)
			authProjects.GET("", projectController.GetProjects)
			authProjects.PATCH("/:id", projectController.UpdateProject)
			authProjects.DELETE("/:id", projectController.DeleteProject)
			authProjects.POST("/:id/apikey", projectController.RegenerateApiKey)
		}
	}

	// Telemetry Collection routes (API key required)
	telemetry := v1.Group("/telemetry")
	telemetry.Use(middleware.ApiKeyMiddleware(s.DB))
	{
		// Events
		telemetry.POST("/events", telemetryController.RecordEvent)
		telemetry.POST("/events/batch", telemetryController.RecordEvents)

		// Sessions
		telemetry.POST("/session/start", telemetryController.StartSession)
		telemetry.POST("/session/end", telemetryController.EndSession)

		// Acquisition
		telemetry.POST("/install", telemetryController.RecordInstall)
	}

	// User Management routes (authenticated)
	users := v1.Group("/users")
	{
		users.POST("", userController.CreateUser)
		users.GET("/:id", userController.GetUser)
		users.GET("", userController.GetUsers)
		users.PATCH("/:id", userController.UpdateUser)
		users.DELETE("/:id", userController.DeleteUser)
	}
}

// Run starts the server
func (s *Server) Run() error {
	return s.Router.Run(":" + s.Config.Port)
}

// The helper functions below can stay for backwards compatibility
// but they should now be used through the config package

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
