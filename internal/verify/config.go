package verify

import (
	"path/filepath"

	"github.com/pengelbrecht/ticks/internal/config"
)

// Config is an alias to config.VerificationConfig for backwards compatibility.
type Config = config.VerificationConfig

// ContextConfig is an alias to config.ContextConfig for backwards compatibility.
type ContextConfig = config.ContextConfig

// Default values re-exported from config package.
const (
	DefaultContextMaxTokens       = config.DefaultContextMaxTokens
	DefaultContextAutoRefreshDays = config.DefaultContextAutoRefreshDays
	DefaultContextTimeout         = config.DefaultContextTimeout
)

// LoadConfig loads verification configuration from .tick/config.json in the given directory.
// Returns nil config (not error) if file doesn't exist.
// Returns error only for malformed JSON.
func LoadConfig(dir string) (*Config, error) {
	configPath := filepath.Join(dir, ".tick", "config.json")

	cfg, err := config.LoadOrDefault(configPath)
	if err != nil {
		return nil, err
	}

	return cfg.Verification, nil
}

// LoadContextConfig loads context configuration from .tick/config.json in the given directory.
// Returns nil config (not error) if file doesn't exist (defaults will be applied via getter methods).
// Returns error only for malformed JSON or invalid config values.
func LoadContextConfig(dir string) (*ContextConfig, error) {
	configPath := filepath.Join(dir, ".tick", "config.json")

	cfg, err := config.LoadOrDefault(configPath)
	if err != nil {
		return nil, err
	}

	return cfg.Context, nil
}
