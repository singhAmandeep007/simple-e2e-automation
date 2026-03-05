package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Agent        AgentConfig        `yaml:"agent"`
	ControlPlane ControlPlaneConfig `yaml:"control_plane"`
	Log          LogConfig          `yaml:"log"`
}

type AgentConfig struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type ControlPlaneConfig struct {
	WSURL string `yaml:"ws_url"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, just use empty config + defaults
			// so it can be started purely via CLI flags (e.g., from Agent UI)
			cfg.ControlPlane.WSURL = "ws://localhost:4000/ws"
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	// Defaults
	if cfg.ControlPlane.WSURL == "" {
		cfg.ControlPlane.WSURL = "ws://localhost:4000/ws"
	}
	return cfg, nil
}
