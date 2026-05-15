// Package config handles Agent configuration loading.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all Agent configuration.
type Config struct {
	Server ServerConfig `json:"server"`
	Log    LogConfig    `json:"log"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"` // "json" or "text"
}

// DefaultConfig returns sensible defaults for local development.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 17645,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

// Load reads configuration from the given path.
// If the file does not exist, it returns DefaultConfig.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return cfg, nil
}

// Addr returns the listen address string (e.g. "127.0.0.1:17645").
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// DataDir returns the data root directory.
// If env LOCAL_SERVICE_PANEL_DATA is set, use that; otherwise use the provided default.
func DataDir(fallback string) string {
	if d := os.Getenv("LOCAL_SERVICE_PANEL_DATA"); d != "" {
		return d
	}
	return fallback
}

// ConfigPath returns the full path to the config file within the data directory.
func ConfigPath(dataDir string) string {
	return filepath.Join(dataDir, "config", "agent.json")
}
