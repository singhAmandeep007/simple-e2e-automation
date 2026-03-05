// Package config provides YAML-based configuration loading for the Control Plane.
// All top-level fields have sane defaults applied when omitted from the config file.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the root configuration structure for the Control Plane.
type Config struct {
	Server ServerConfig `yaml:"server"`
	DB     DBConfig     `yaml:"db"`
	CORS   CORSConfig   `yaml:"cors"`
	Log    LogConfig    `yaml:"log"`
}

// ServerConfig sets the HTTP listener host and port.
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// DBConfig specifies the filesystem path to the SQLite database file.
type DBConfig struct {
	Path string `yaml:"path"`
}

// CORSConfig holds allowed origins for cross-origin HTTP requests.
type CORSConfig struct {
	AllowOrigins []string `yaml:"allow_origins"`
}

// LogConfig controls application log verbosity.
type LogConfig struct {
	Level string `yaml:"level"`
}

// Load reads the YAML config file at the given path and applies defaults.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Apply defaults
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 4000
	}
	if cfg.DB.Path == "" {
		cfg.DB.Path = "data/data.db"
	}

	return cfg, nil
}
