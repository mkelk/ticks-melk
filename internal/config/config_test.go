package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.Version != DefaultVersion {
		t.Fatalf("expected version %d, got %d", DefaultVersion, cfg.Version)
	}
	if cfg.IDLength != DefaultIDLength {
		t.Fatalf("expected id_length %d, got %d", DefaultIDLength, cfg.IDLength)
	}
}

func TestLoadMissingConfig(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "config.json"))
	if err == nil {
		t.Fatalf("expected error for missing config")
	}
}

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Version != DefaultVersion {
		t.Fatalf("expected default version %d, got %d", DefaultVersion, cfg.Version)
	}
	if cfg.IDLength != DefaultIDLength {
		t.Fatalf("expected default id_length %d, got %d", DefaultIDLength, cfg.IDLength)
	}
}

func TestValidateRejectsInvalidIDLength(t *testing.T) {
	cfg := Default()
	cfg.IDLength = 2
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error for invalid id_length")
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := Config{Version: DefaultVersion, IDLength: 4}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.IDLength != 4 {
		t.Fatalf("expected id_length 4, got %d", loaded.IDLength)
	}
}
