package config

import (
	"embed"
	"os"

	commonconfig "commerce-platform/shared/config"
)

//go:embed base.yaml local.yaml test.yaml
var configFiles embed.FS

type Config struct {
	commonconfig.Config `yaml:",inline"`

	Products struct {
		GrpcClient string `yaml:"grpc-client"`
	} `yaml:"products"`
}

func Load() (Config, error) {
	var cfg Config
	profile := os.Getenv("ACTIVE_PROFILE")

	err := commonconfig.Load(configFiles, profile, &cfg)
	return cfg, err
}
