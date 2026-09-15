package config

import (
	"embed"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed base.yaml local.yaml test.yaml
var configFiles embed.FS

type Config struct {
	Profile string

	Server struct {
		HTTPPort int `yaml:"http-port"`
		GRPCPort int `yaml:"grpc-port"`
	} `yaml:"server"`

	Environment string `yaml:"environment"`
}

func Load() (Config, error) {
	var cfg Config

	if err := loadFile("base.yaml", &cfg); err != nil {
		return cfg, err
	}

	profile := os.Getenv("ACTIVE_PROFILE")
	if profile == "" {
		cfg.Profile = "default"
		return cfg, nil
	}

	cfg.Profile = profile
	if err := loadFile(fmt.Sprintf("%s.yaml", profile), &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func loadFile(path string, cfg *Config) error {
	data, err := configFiles.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, cfg)
}
