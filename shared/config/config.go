package config

import (
	"embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type Config struct {
	Profile string

	Server struct {
		HTTPPort         int `yaml:"http-port"`
		GRPCPort         int `yaml:"grpc-port"`
		GracefulShutdown struct {
			Timeout int `yaml:"timeout"`
		} `yaml:"graceful-shutdown"`
	} `yaml:"server"`

	Environment string `yaml:"environment"`

	Database DatabaseConfig `yaml:"database"`
}

func (cfg *Config) SetProfile(profile string) {
	cfg.Profile = profile
}

type ProfileConfig interface {
	SetProfile(profile string)
}

func Load(files embed.FS, profile string, cfg ProfileConfig) error {
	if err := loadFile(files, "base.yaml", cfg); err != nil {
		return err
	}

	if profile == "" {
		cfg.SetProfile("default")
		return nil
	}

	cfg.SetProfile(profile)
	return loadFile(files, fmt.Sprintf("%s.yaml", profile), cfg)
}

func loadFile(files embed.FS, fileName string, cfg any) error {
	data, err := files.ReadFile(fileName)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, cfg)
}
