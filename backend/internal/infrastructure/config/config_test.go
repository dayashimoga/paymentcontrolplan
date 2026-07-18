package config_test

import (
	"testing"

	"github.com/paymentbridge/pcp/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	cfg, err := config.Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Database.Host)
}

func TestDatabaseConfig_DSN(t *testing.T) {
	dbCfg := config.DatabaseConfig{
		User:     "user",
		Password: "password",
		Host:     "host",
		Port:     5432,
		Name:     "dbname",
		SSLMode:  "disable",
	}
	expected := "postgres://user:password@host:5432/dbname?sslmode=disable"
	assert.Equal(t, expected, dbCfg.DSN())
}
