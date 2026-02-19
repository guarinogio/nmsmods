package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type NexusConfig struct {
	APIKey string `json:"api_key,omitempty"`
}

type Config struct {
	GamePath      string      `json:"game_path"`
	ActiveProfile string      `json:"active_profile,omitempty"`
	Nexus         NexusConfig `json:"nexus,omitempty"`
}

func LoadConfig(path string) (Config, error) {
	var c Config
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return c, err
	}
	if len(b) == 0 {
		return c, nil
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return c, err
	}
	return c, nil
}

func SaveConfig(path string, c Config) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// Config may contain secrets (Nexus API key), so 0600.
	return os.WriteFile(path, b, 0o600)
}
