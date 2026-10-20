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
	cfg.Pool.MaxConns = 10
	cfg.Pool.MinConns = 1

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

func TestNewPostgreClient_WhenEnvInvalid_ReturnsError(t *testing.T) {
	t.Setenv("PGPORT", "not-a-port")

	pool, err := NewPostgreClient(context.Background(), commonconfig.DatabaseConfig{})

	assert.Error(t, err)
	assert.ErrorContains(t, err, "parse database config")
	assert.Nil(t, pool)
}
