package config

import (
	"embed"
	"os"

	commonconfig "commerce-platform/shared/config"
)

//go:embed base.yaml local.yaml test.yaml
var configFiles embed.FS

func Load() (commonconfig.Config, error) {
	var cfg commonconfig.Config
	profile := os.Getenv("ACTIVE_PROFILE")

	err := commonconfig.Load(configFiles, profile, &cfg)
	return cfg, err
}
