package logging_test

import (
	"testing"

	"github.com/paymentbridge/pcp/internal/infrastructure/logging"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger_Production(t *testing.T) {
	logger, err := logging.NewLogger("info", "json")
	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewLogger_Console(t *testing.T) {
	logger, err := logging.NewLogger("debug", "console")
	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewLogger_InvalidLevelFallback(t *testing.T) {
	logger, err := logging.NewLogger("invalid_level", "json")
	assert.NoError(t, err)
	assert.NotNil(t, logger)
}
