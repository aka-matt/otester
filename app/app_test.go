package app

import (
	"errors"
	"testing"

	"otester/internal/config"
)

func TestOpenConfigFileSwitchesActivePathAfterSuccessfulLoad(t *testing.T) {
	a := NewApp()
	a.activeConfigPath = "current.json"
	a.openFileDialog = func() (string, error) { return "selected.json", nil }
	a.loadConfigFromPath = func(path string) (*config.ConfigView, error) {
		return &config.ConfigView{ConfigPath: path}, nil
	}

	cfg, err := a.OpenConfigFile()
	if err != nil {
		t.Fatalf("OpenConfigFile failed: %v", err)
	}
	if cfg == nil || cfg.ConfigPath != "selected.json" {
		t.Fatalf("Expected selected config, got %#v", cfg)
	}
	if a.activeConfigPath != "selected.json" {
		t.Errorf("Expected active path to switch, got %q", a.activeConfigPath)
	}
}

func TestOpenConfigFileCancellationPreservesActivePath(t *testing.T) {
	a := NewApp()
	a.activeConfigPath = "current.json"
	a.openFileDialog = func() (string, error) { return "", nil }

	cfg, err := a.OpenConfigFile()
	if err != nil {
		t.Fatalf("OpenConfigFile failed: %v", err)
	}
	if cfg != nil {
		t.Errorf("Expected nil config on cancellation, got %#v", cfg)
	}
	if a.activeConfigPath != "current.json" {
		t.Errorf("Expected active path to remain current, got %q", a.activeConfigPath)
	}
}

func TestOpenConfigFileLoaderErrorPreservesActivePath(t *testing.T) {
	a := NewApp()
	a.activeConfigPath = "current.json"
	a.openFileDialog = func() (string, error) { return "invalid.json", nil }
	a.loadConfigFromPath = func(string) (*config.ConfigView, error) {
		return nil, errors.New("invalid config")
	}

	cfg, err := a.OpenConfigFile()
	if err == nil {
		t.Fatal("Expected loader error")
	}
	if cfg != nil {
		t.Errorf("Expected nil config on loader error, got %#v", cfg)
	}
	if a.activeConfigPath != "current.json" {
		t.Errorf("Expected active path to remain current, got %q", a.activeConfigPath)
	}
}
