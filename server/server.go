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
	analyticsController := controllers.NewAnalyticsController(s.DB)
	authController := controllers.NewAuthController(s.DB)
	deviceController := controllers.NewDeviceController(s.DB)
	eventController := controllers.NewEventController(s.DB)
	healthController := controllers.NewHealthController(s.DB)
	projectController := controllers.NewProjectController(s.DB)
	sessionController := controllers.NewSessionController(s.DB)
	userController := controllers.NewUserController(s.DB)

	// API v1 routes
	v1 := s.Router.Group("/api/v1")

	// Analytics routes
	analytics := v1.Group("/analytics")
	analytics.Use(middleware.AuthMiddleware(s.DB))
	{
		analytics.GET("", analyticsController.GetAnalytics)
	}

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
			apiKeyDevices.POST("", deviceController.CreateOrUpdateDevice)
			apiKeyDevices.GET("/:id", deviceController.GetDevice)
		}

		// Auth routes (for dashboard)
		authDevices := devices.Group("")
		authDevices.Use(middleware.AuthMiddleware(s.DB))
		{
			authDevices.GET("", deviceController.GetDevices)
			authDevices.DELETE("/:id", deviceController.DeleteDevice)
		}
	}

	// Telemetry Collection routes (API key required)
	events := v1.Group("/events")
	events.Use(middleware.ApiKeyMiddleware(s.DB))
	{
		events.POST("", eventController.RecordEvent)
		events.POST("/batch", eventController.RecordEvents)
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

	// Session routes
	sessions := v1.Group("/sessions")
	{
		apiSessions := sessions.Group("")
		apiSessions.Use(middleware.ApiKeyMiddleware(s.DB))
		{
			apiSessions.POST("/begin", sessionController.BeginSession)
			apiSessions.POST("/end", sessionController.EndSession)
		}

		authSessions := sessions.Group("")
		authSessions.Use(middleware.AuthMiddleware(s.DB))
		{
			authSessions.GET("", sessionController.GetSessions)
			authSessions.GET("/:id", sessionController.GetSession)
		}
	}

	// User routes
	users := v1.Group("/users")
	{
		users.POST("", userController.CreateUser)

		authUsers := users.Group("")
		authUsers.Use(middleware.AuthMiddleware(s.DB))
		{
			authUsers.GET("", userController.GetUsers)
			authUsers.GET("/:id", userController.GetUser)
			authUsers.PATCH("/:id", userController.UpdateUser)
			authUsers.DELETE("/:id", userController.DeleteUser)
		}
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
