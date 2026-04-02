package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	// Server settings
	Port        string
	Environment string

	// Database settings
	DatabaseURL         string
	DBMaxOpenConns       int32
	DBMaxIdleConns       int32
	DBConnMaxLifetime    time.Duration
	MigrateDatabaseURL   string

	// JWT settings
	JWTSecret            string
	JWTExpiryHours       int
	RefreshTokenExpiryDays int

	// Security settings
	BCryptCost            int
	RateLimitRequestsPerMinute int
	CORSAllowedOrigins    []string

	// Logging
	LogLevel             string
}

// Load reads configuration from environment variables using Viper
func Load() (*Config, error) {
	v := viper.New()

	// Enable environment variable override
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults(v)

	// Read .env file if present (for local development)
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("./cmd/server")
	v.AddConfigPath("../..")

	// Ignore error if .env doesn't exist
	_ = v.ReadInConfig()

	// Validate required fields
	if err := validate(v); err != nil {
		return nil, err
	}

	return fromViper(v), nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("PORT", "8080")
	v.SetDefault("ENV", "production")
	v.SetDefault("LOG_LEVEL", "info")

	// Database defaults
	v.SetDefault("DB_MAX_OPEN_CONNS", int32(25))
	v.SetDefault("DB_MAX_IDLE_CONNS", int32(5))
	v.SetDefault("DB_CONN_MAX_LIFETIME", "30m")

	// JWT defaults
	v.SetDefault("JWT_EXPIRY_HOURS", 24)
	v.SetDefault("REFRESH_TOKEN_EXPIRY_DAYS", 7)

	// Security defaults
	v.SetDefault("BCRYPT_COST", 12)
	v.SetDefault("RATE_LIMIT_REQUESTS_PER_MINUTE", 100)
	v.SetDefault("CORS_ALLOWED_ORIGINS", []string{"*"})
}

// validate checks that required configuration values are present
func validate(v *viper.Viper) error {
	var errs []string

	// JWT secret is required in production
	if v.GetString("ENV") == "production" {
		if v.GetString("JWT_SECRET") == "" {
			errs = append(errs, "JWT_SECRET is required in production")
		} else if len(v.GetString("JWT_SECRET")) < 32 {
			errs = append(errs, "JWT_SECRET must be at least 32 characters")
		}
	}

	// JWT secret is optional but warned in development
	if v.GetString("ENV") != "production" && v.GetString("JWT_SECRET") == "" {
		errs = append(errs, "JWT_SECRET is not set, using insecure default (development only)")
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// fromViper creates a Config from Viper instance
func fromViper(v *viper.Viper) *Config {
	// Parse conn max lifetime
	connMaxLifetime := v.GetDuration("DB_CONN_MAX_LIFETIME")
	if connMaxLifetime == 0 {
		connMaxLifetime = 30 * time.Minute
	}

	// Parse CORS origins
	corsOrigins := v.GetStringSlice("CORS_ALLOWED_ORIGINS")
	if len(corsOrigins) == 0 {
		corsOrigins = []string{"*"}
	}

	return &Config{
		Port:        v.GetString("PORT"),
		Environment: v.GetString("ENV"),
		DatabaseURL: v.GetString("DATABASE_URL"),
		DBMaxOpenConns: v.GetInt32("DB_MAX_OPEN_CONNS"),
		DBMaxIdleConns: v.GetInt32("DB_MAX_IDLE_CONNS"),
		DBConnMaxLifetime: connMaxLifetime,
		MigrateDatabaseURL: v.GetString("MIGRATE_DATABASE_URL"),
		JWTSecret:   v.GetString("JWT_SECRET"),
		JWTExpiryHours: v.GetInt("JWT_EXPIRY_HOURS"),
		RefreshTokenExpiryDays: v.GetInt("REFRESH_TOKEN_EXPIRY_DAYS"),
		BCryptCost:  v.GetInt("BCRYPT_COST"),
		RateLimitRequestsPerMinute: v.GetInt("RATE_LIMIT_REQUESTS_PER_MINUTE"),
		CORSAllowedOrigins: corsOrigins,
		LogLevel:   v.GetString("LOG_LEVEL"),
	}
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}
