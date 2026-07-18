package cache_test

import (
	"testing"

	"github.com/paymentbridge/pcp/internal/infrastructure/cache"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisClient_InvalidHost(t *testing.T) {
	client, err := cache.NewRedisClient("invalid-host-9999", 6379, "", 0)
	assert.Error(t, err)
	assert.Nil(t, client)
}
