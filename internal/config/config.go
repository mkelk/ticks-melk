package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const (
	DefaultVersion  = 1
	DefaultIDLength = 3
)

// Config defines project configuration stored in .tick/config.json.
type Config struct {
	Version  int `json:"version"`
	IDLength int `json:"id_length"`
}

// Default returns the default config.
func Default() Config {
	return Config{
		Version:  DefaultVersion,
		IDLength: DefaultIDLength,
	}
}

// Load reads config from disk and applies defaults for zero values.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, fmt.Errorf("config not found: %w", err)
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Version == 0 {
		cfg.Version = DefaultVersion
	}
	if cfg.IDLength == 0 {
		cfg.IDLength = DefaultIDLength
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Save writes a config to disk.
func Save(path string, cfg Config) error {
	if cfg.Version == 0 {
		cfg.Version = DefaultVersion
	}
	if cfg.IDLength == 0 {
		cfg.IDLength = DefaultIDLength
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// Validate ensures config values are within supported ranges.
func (c Config) Validate() error {
	if c.Version != DefaultVersion {
		return fmt.Errorf("unsupported config version: %d", c.Version)
	}
	if c.IDLength < 3 || c.IDLength > 4 {
		return fmt.Errorf("id_length must be 3 or 4, got %d", c.IDLength)
	}
	return nil
}
