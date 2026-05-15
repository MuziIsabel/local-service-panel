// Package settings provides configuration storage for the Agent UI.
// Settings are persisted as a JSON file in the data directory.
package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Settings holds all user-configurable settings.
type Settings struct {
	// Server settings
	Host string `json:"host"`
	Port int    `json:"port"`

	// UI preferences
	Theme string `json:"theme"` // "light" or "dark"

	// Version tracking
	SchemaVersion int `json:"schemaVersion"`
}

// Defaults returns the default settings.
func Defaults() *Settings {
	return &Settings{
		Host:          "127.0.0.1",
		Port:          17645,
		Theme:         "light",
		SchemaVersion: 1,
	}
}

// Store persists settings to a JSON file.
type Store struct {
	mu       sync.RWMutex
	settings *Settings
	filePath string
}

// NewStore creates or loads a settings store from the given data directory.
func NewStore(dataDir string) (*Store, error) {
	dir := filepath.Join(dataDir, "config")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create settings dir: %w", err)
	}

	path := filepath.Join(dir, "settings.json")
	s := &Store{
		settings: Defaults(),
		filePath: path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// First time - save defaults.
			if err := s.save(); err != nil {
				return nil, err
			}
			return s, nil
		}
		return nil, fmt.Errorf("read settings: %w", err)
	}

	var loaded Settings
	if err := json.Unmarshal(data, &loaded); err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}

	// Merge loaded settings into defaults (to handle new fields gracefully).
	defaults := Defaults()
	if loaded.Host != "" {
		defaults.Host = loaded.Host
	}
	if loaded.Port != 0 {
		defaults.Port = loaded.Port
	}
	if loaded.Theme != "" {
		defaults.Theme = loaded.Theme
	}
	if loaded.SchemaVersion != 0 {
		defaults.SchemaVersion = loaded.SchemaVersion
	}
	s.settings = defaults

	return s, nil
}

// Get returns a copy of the current settings.
func (s *Store) Get() *Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := *s.settings
	return &cp
}

// Update applies partial updates to settings and saves.
func (s *Store) Update(updates map[string]interface{}) (*Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Apply partial updates
	if v, ok := updates["host"]; ok {
		if host, ok := v.(string); ok && host != "" {
			s.settings.Host = host
		}
	}
	if v, ok := updates["port"]; ok {
		if port, ok := v.(float64); ok && port > 0 {
			s.settings.Port = int(port)
		}
	}
	if v, ok := updates["theme"]; ok {
		if theme, ok := v.(string); ok && (theme == "light" || theme == "dark") {
			s.settings.Theme = theme
		}
	}

	if err := s.save(); err != nil {
		return nil, err
	}

	cp := *s.settings
	return &cp, nil
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}
	return nil
}
