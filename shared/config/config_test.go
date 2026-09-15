package config

import (
	"commerce-platform/shared/config/testdata"
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type extendedConfig struct {
	Config `yaml:",inline"`

	Products struct {
		GRPCClient string `yaml:"grpc-client"`
	} `yaml:"products"`
}

func TestLoad_WhenProfileIsEmpty_LoadsBaseConfigAndUsesDefaultProfile(t *testing.T) {
	var cfg extendedConfig

	err := Load(testdata.Files, "", &cfg)

	require.NoError(t, err)
	assert.Equal(t, "default", cfg.Profile)
	assert.Equal(t, "base", cfg.Environment)
	assert.Equal(t, 8080, cfg.Server.HTTPPort)
	assert.Equal(t, 10, cfg.Server.GracefulShutdown.Timeout)
	assert.Equal(t, "localhost:8092", cfg.Products.GRPCClient)
}

func TestLoad_WhenProfileIsSet_OverlaysProfileConfig(t *testing.T) {
	var cfg extendedConfig

	err := Load(testdata.Files, "test", &cfg)

	require.NoError(t, err)
	assert.Equal(t, "test", cfg.Profile)
	assert.Equal(t, "test", cfg.Environment)
	assert.Equal(t, 8080, cfg.Server.HTTPPort)
	assert.Equal(t, 3, cfg.Server.GracefulShutdown.Timeout)
	assert.Equal(t, "products:9092", cfg.Products.GRPCClient)
}

func TestLoad_WhenProfileFileDoesNotExist_ReturnsError(t *testing.T) {
	var cfg extendedConfig

	err := Load(testdata.Files, "missing", &cfg)

	assert.Error(t, err)
	assert.Equal(t, "missing", cfg.Profile)
}

func TestLoad_WhenBaseFileDoesNotExist_ReturnsError(t *testing.T) {
	var cfg extendedConfig

	err := Load(embed.FS{}, "", &cfg)

	assert.Error(t, err)
}
