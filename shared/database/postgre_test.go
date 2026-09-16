package database

import (
	"context"
	"testing"

	commonconfig "commerce-platform/shared/config"

	"github.com/stretchr/testify/assert"
)

func TestNewPostgreClient_WhenConfigValid_BuildsPoolWithExpectedConnConfig(t *testing.T) {
	cfg := commonconfig.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "products",
		User:     "commerce",
		Password: "commerce",
	}

	pool, err := NewPostgreClient(context.Background(), cfg)
	assert.NoError(t, err)
	assert.NotNil(t, pool)
	defer pool.Close()

	connConfig := pool.Config().ConnConfig
	assert.Equal(t, cfg.Host, connConfig.Host)
	assert.Equal(t, uint16(cfg.Port), connConfig.Port)
	assert.Equal(t, cfg.Name, connConfig.Database)
	assert.Equal(t, cfg.User, connConfig.User)
	assert.Equal(t, cfg.Password, connConfig.Password)
}

func TestNewPostgreClient_WhenConnectionStringInvalid_ReturnsError(t *testing.T) {
	cfg := commonconfig.DatabaseConfig{
		Host:     "localhost",
		Port:     -1,
		Name:     "products",
		User:     "commerce",
		Password: "commerce",
	}

	pool, err := NewPostgreClient(context.Background(), cfg)

	assert.Error(t, err)
	assert.Nil(t, pool)
}
