package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
// Using a struct for configuration allows for easy testing and dependency injection
type Config struct {
	// Server settings
	Port          int
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	Environment   string
	AllowOrigins  []string
	
	// Auth settings
	JWTSecret     string
	JWTExpiration time.Duration
	
	// Database settings
	DBType        string // "memory", "postgres", etc.
	DBConnection  string
	
	// Feature flags for gradual rollout
	Features      map[string]bool
}

// DefaultConfig creates a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Port:          8080,
		ReadTimeout:   10 * time.Second,
		WriteTimeout:  10 * time.Second,
		Environment:   "development",
		AllowOrigins:  []string{"*"},
		JWTSecret:     "dev-secret-change-in-production",
		JWTExpiration: 24 * time.Hour,
		DBType:        "memory",
		Features:      map[string]bool{},
	}
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	config := DefaultConfig()
	
	// Server settings
	if port := os.Getenv("PORT"); port != "" {
		if portNum, err := strconv.Atoi(port); err == nil {
			config.Port = portNum
		}
	}
	
	if env := os.Getenv("ENV"); env != "" {
		config.Environment = env
	}
	
	// Security settings
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.JWTSecret = jwtSecret
	} else if config.Environment == "production" {
		// Fail fast in production if no JWT secret is provided
		fmt.Println("WARNING: No JWT_SECRET set in production environment!")
	}
	
	if jwtExp := os.Getenv("JWT_EXPIRATION"); jwtExp != "" {
		if expHours, err := strconv.Atoi(jwtExp); err == nil {
			config.JWTExpiration = time.Duration(expHours) * time.Hour
		}
	}
	
	// Database settings
	if dbType := os.Getenv("DB_TYPE"); dbType != "" {
		config.DBType = dbType
	}
	
	if dbConn := os.Getenv("DB_CONNECTION"); dbConn != "" {
		config.DBConnection = dbConn
	}
	
	// Load feature flags
	// Format: FEATURE_X=true,FEATURE_Y=false
	for _, key := range []string{
		"FEATURE_ARCHITECTURE_CANVAS",
		"FEATURE_STORY_FLOW",
		"FEATURE_TASK_HUB",
		"FEATURE_REVIEW_QUEUE",
		"FEATURE_DESIGN_ASSISTANT",
	} {
		if value := os.Getenv(key); value != "" {
			config.Features[key] = value == "true"
		} else {
			// Default all features to false in production, true in development
			config.Features[key] = config.Environment != "production"
		}
	}
	
	return config
}

// IsDevelopment returns true if the application is in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if the application is in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsFeatureEnabled checks if a feature flag is enabled
func (c *Config) IsFeatureEnabled(feature string) bool {
	enabled, exists := c.Features[feature]
	return exists && enabled
}

// GetAddress returns the full server address with port
func (c *Config) GetAddress() string {
	return fmt.Sprintf(":%d", c.Port)
}
