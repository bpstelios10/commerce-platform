package database

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	commonconfig "commerce-platform/shared/config"

	"github.com/stretchr/testify/assert"
)

type failingFS struct{}

func (failingFS) Open(name string) (fs.File, error) {
	return nil, errors.New("boom")
}

func TestRunMigrations_WhenSourceInvalid_ReturnsError(t *testing.T) {
	cfg := commonconfig.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "products",
		User:     "commerce",
		Password: "commerce",
	}

	err := RunMigrations(cfg, failingFS{})

	assert.Error(t, err)
}

func TestRunMigrations_WhenConnectionStringInvalid_ReturnsError(t *testing.T) {
	cfg := commonconfig.DatabaseConfig{
		Host:     "localhost",
		Port:     -1,
		Name:     "products",
		User:     "commerce",
		Password: "commerce",
	}

	err := RunMigrations(cfg, fstest.MapFS{})

	assert.Error(t, err)
}
