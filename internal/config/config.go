package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultAPIBase = "https://link-ai.tech"
	configDirName  = ".linkai"
	configFileName = "config.json"
)

// AppUser is a logged-in user record stored in config.
type AppUser struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	AccountNo string `json:"account_no,omitempty"`
}

// Config is the CLI configuration stored in ~/.linkai/config.json.
type Config struct {
	DeviceID string   `json:"device_id,omitempty"`
	User     *AppUser `json:"user,omitempty"`
	apiBase  string   // runtime only, not persisted
}

// APIBase returns the effective API base URL:
// LINKAI_API_BASE env var takes precedence, then DefaultAPIBase.
func (c *Config) APIBase() string {
	if c.apiBase != "" {
		return c.apiBase
	}
	return DefaultAPIBase
}

// ConfigDir returns the path to the config directory (~/.linkai/).
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}
	return filepath.Join(home, configDirName), nil
}

func configPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// Load reads the config file from disk. Returns a default config if the file doesn't exist.
// The API base is resolved from LINKAI_API_BASE env var at runtime and never persisted.
func Load() (*Config, error) {
	p, err := configPath()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	data, err := os.ReadFile(p)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// Resolve API base from env var at runtime (not stored on disk)
	if base := os.Getenv("LINKAI_API_BASE"); base != "" {
		cfg.apiBase = strings.TrimRight(base, "/")
	}

	return cfg, nil
}

// Save writes the config file to disk (device_id and user only; api_base is never persisted).
func Save(cfg *Config) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(p, data, 0600)
}

// EnsureDeviceID returns the config's DeviceID, generating and persisting one
// if it is not yet set. The config file is updated in place.
func EnsureDeviceID(cfg *Config) (string, error) {
	if cfg.DeviceID != "" {
		return cfg.DeviceID, nil
	}
	id, err := generateDeviceID()
	if err != nil {
		return "", fmt.Errorf("failed to generate device ID: %w", err)
	}
	cfg.DeviceID = id
	if err := Save(cfg); err != nil {
		return "", fmt.Errorf("failed to save device ID: %w", err)
	}
	return id, nil
}

func generateDeviceID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	s := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", s[0:8], s[8:12], s[12:16], s[16:20], s[20:32]), nil
}
