package config

import (
	"embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Profile string

	Server struct {
		HTTPPort int `yaml:"http-port"`
		GRPCPort int `yaml:"grpc-port"`
	} `yaml:"server"`

	Environment string `yaml:"environment"`
}

func Load(files embed.FS, profile string) (Config, error) {
	var cfg Config

	if err := loadFile(files, "base.yaml", &cfg); err != nil {
		return cfg, err
	}

	if profile == "" {
		cfg.Profile = "default"
		return cfg, nil
	}

	cfg.Profile = profile
	if err := loadFile(files, fmt.Sprintf("%s.yaml", profile), &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func loadFile(files embed.FS, name string, cfg *Config) error {
	data, err := files.ReadFile(name)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, cfg)
}
