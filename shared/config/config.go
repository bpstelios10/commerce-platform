package config

import (
	"fmt"
	"io/fs"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"

	loggerx "commerce-platform/shared/logger"
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

	LogLevel string `yaml:"log-level"`

	Database DatabaseConfig `yaml:"database"`
}

type ProfileConfig interface {
	SetProfile(profile string)
}

func (cfg *Config) SetProfile(profile string) {
	cfg.Profile = profile
}

func (cfg *Config) GetLogLevel(fallback zerolog.Level) zerolog.Level {
	lvl := loggerx.LevelFromEnv("LOG_LEVEL")
	if lvl != zerolog.NoLevel {
		return lvl
	}

	lvl, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil || lvl == zerolog.NoLevel {
		return fallback
	}

	return lvl
}

func Load(files fs.FS, profile string, cfg ProfileConfig) error {
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

func loadFile(files fs.FS, fileName string, cfg any) error {
	data, err := fs.ReadFile(files, fileName)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, cfg)
}
