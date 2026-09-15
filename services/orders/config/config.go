package config

import (
	"embed"
	"os"

	commonconfig "commerce-platform/shared/config"
)

//go:embed base.yaml local.yaml test.yaml
var configFiles embed.FS

func Load() (commonconfig.Config, error) {
	profile := os.Getenv("ACTIVE_PROFILE")

	return commonconfig.Load(configFiles, profile)
}
