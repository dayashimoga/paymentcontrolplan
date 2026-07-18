// Package config provides application configuration via Viper.
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Log      LogConfig
}

type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	Issuer     string
	Expiration time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

// Load reads configuration from environment variables with PCP_ prefix.
func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("PCP")
	v.AutomaticEnv()

	// Server defaults
	v.SetDefault("SERVER_HOST", "0.0.0.0")
	v.SetDefault("SERVER_PORT", 8080)
	v.SetDefault("SERVER_READ_TIMEOUT", "15s")
	v.SetDefault("SERVER_WRITE_TIMEOUT", "15s")
	v.SetDefault("SERVER_SHUTDOWN_TIMEOUT", "30s")

	// Database defaults
	v.SetDefault("DATABASE_HOST", "localhost")
	v.SetDefault("DATABASE_PORT", 5432)
	v.SetDefault("DATABASE_USER", "pcp")
	v.SetDefault("DATABASE_PASSWORD", "pcp_secret")
	v.SetDefault("DATABASE_NAME", "pcp")
	v.SetDefault("DATABASE_SSL_MODE", "disable")

	// Redis defaults
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", 6379)
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)

	// JWT defaults
	v.SetDefault("JWT_SECRET", "change-me-in-production")
	v.SetDefault("JWT_ISSUER", "pcp")
	v.SetDefault("JWT_EXPIRATION", "24h")

	// Log defaults
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")

	readTimeout, _ := time.ParseDuration(v.GetString("SERVER_READ_TIMEOUT"))
	writeTimeout, _ := time.ParseDuration(v.GetString("SERVER_WRITE_TIMEOUT"))
	shutdownTimeout, _ := time.ParseDuration(v.GetString("SERVER_SHUTDOWN_TIMEOUT"))
	jwtExpiration, _ := time.ParseDuration(v.GetString("JWT_EXPIRATION"))

	return &Config{
		Server: ServerConfig{
			Host: v.GetString("SERVER_HOST"), Port: v.GetInt("SERVER_PORT"),
			ReadTimeout: readTimeout, WriteTimeout: writeTimeout, ShutdownTimeout: shutdownTimeout,
		},
		Database: DatabaseConfig{
			Host: v.GetString("DATABASE_HOST"), Port: v.GetInt("DATABASE_PORT"),
			User: v.GetString("DATABASE_USER"), Password: v.GetString("DATABASE_PASSWORD"),
			Name: v.GetString("DATABASE_NAME"), SSLMode: v.GetString("DATABASE_SSL_MODE"),
		},
		Redis: RedisConfig{
			Host: v.GetString("REDIS_HOST"), Port: v.GetInt("REDIS_PORT"),
			Password: v.GetString("REDIS_PASSWORD"), DB: v.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			Secret: v.GetString("JWT_SECRET"), Issuer: v.GetString("JWT_ISSUER"),
			Expiration: jwtExpiration,
		},
		Log: LogConfig{
			Level: v.GetString("LOG_LEVEL"), Format: v.GetString("LOG_FORMAT"),
		},
	}, nil
}
