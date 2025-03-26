package config

// Config holds the configuration for the application
type Config struct {
	// Database connection details
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT settings
	JWTSecret     string
	JWTExpiration string

	// Server settings
	Port string
}

// NewConfigFromEnv creates a new Config from environment variables
func NewConfigFromEnv() *Config {
	return &Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", "postgres"),
		DBName:        getEnv("DB_NAME", "kogase"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		JWTSecret:     getEnv("JWT_SECRET", "kogase-jwt-secret"),
		JWTExpiration: getEnv("JWT_EXPIRATION", "24h"),
		Port:          getEnv("PORT", "8080"),
	}
}
