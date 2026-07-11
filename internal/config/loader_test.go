package config

import (
	"os"
	"path/filepath"
	"testing"
)

func chdirToProjectRoot(t *testing.T) func() {
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Failed to determine project root: %v", err)
	}
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to chdir to project root: %v", err)
	}
	return func() { os.Chdir(oldCwd) }
}

func TestLoadConfigFromPathUsesProvidedPath(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "custom-config.json")
	if err := os.WriteFile(configPath, []byte(`{"app":{"name":"custom"}}`), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFromPath failed: %v", err)
	}
	if cfg.ConfigPath != configPath {
		t.Errorf("Expected ConfigPath %q, got %q", configPath, cfg.ConfigPath)
	}
}

func TestLoadConfigFromPathInvalidJSONRetainsCurrentConfig(t *testing.T) {
	tmpDir := t.TempDir()
	validPath := filepath.Join(tmpDir, "valid.json")
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(validPath, []byte(`{"app":{"name":"current"}}`), 0644); err != nil {
		t.Fatalf("Failed to write valid test config: %v", err)
	}
	if err := os.WriteFile(invalidPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid test config: %v", err)
	}

	current, err := LoadConfigFromPath(validPath)
	if err != nil {
		t.Fatalf("LoadConfigFromPath failed: %v", err)
	}

	_, err = LoadConfigFromPath(invalidPath)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
	if got := GetCurrentConfig(); got != current {
		t.Errorf("Expected current config to remain %p, got %p", current, got)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	defer chdirToProjectRoot(t)()

	cfg, err := LoadConfigFromFile()
	if err != nil {
		t.Fatalf("LoadConfigFromFile failed: %v", err)
	}
	if cfg.App.Title != "otester - API Debugger" {
		t.Errorf("Expected title 'otester - API Debugger', got '%s'", cfg.App.Title)
	}
	if len(cfg.Endpoints) != 4 {
		t.Errorf("Expected 4 endpoints, got %d", len(cfg.Endpoints))
	}
}

func TestLoadConfigFromFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir to temp dir: %v", err)
	}

	_, err = LoadConfigFromFile()
	if err == nil {
		t.Fatal("Expected error when config.json not found")
	}
	cfgErr, ok := err.(*ConfigError)
	if !ok {
		t.Fatalf("Expected ConfigError, got %T", err)
	}
	if cfgErr.Code != "CONFIG_FILE_NOT_FOUND" {
		t.Errorf("Expected code CONFIG_FILE_NOT_FOUND, got '%s'", cfgErr.Code)
	}
}

func TestLoadConfigFromFileInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir to temp dir: %v", err)
	}

	_, err = LoadConfigFromFile()
	if err == nil {
		t.Fatal("Expected error when config.json is invalid JSON")
	}
	cfgErr, ok := err.(*ConfigError)
	if !ok {
		t.Fatalf("Expected ConfigError, got %T", err)
	}
	if cfgErr.Code != "CONFIG_PARSE_FAILED" {
		t.Errorf("Expected code CONFIG_PARSE_FAILED, got '%s'", cfgErr.Code)
	}
}

func TestOAuthProfileMasking(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	configContent := `{
		"app": {"name": "test", "title": "Test"},
		"variables": [],
		"oauth_profiles": [
			{
				"id": "oauth-1",
				"name": "Test OAuth",
				"type": "microsoft_client_credentials",
				"org_id_uuid": "org-123",
				"client_id": "client-id",
				"client_secret": "secret-should-be-masked",
				"scope": "read write"
			}
		],
		"endpoint_groups": [],
		"endpoints": []
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir to temp dir: %v", err)
	}

	cfg, err := LoadConfigFromFile()
	if err != nil {
		t.Fatalf("LoadConfigFromFile failed: %v", err)
	}

	if len(cfg.OAuthProfiles) != 1 {
		t.Fatalf("Expected 1 OAuth profile, got %d", len(cfg.OAuthProfiles))
	}
	if cfg.OAuthProfiles[0].ClientSecretMasked != "******" {
		t.Errorf("Expected ClientSecretMasked to be '******', got '%s'", cfg.OAuthProfiles[0].ClientSecretMasked)
	}
}

func TestLoadConfigAndReload(t *testing.T) {
	defer chdirToProjectRoot(t)()

	cfg1, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	cfg2, err := ReloadConfig()
	if err != nil {
		t.Fatalf("ReloadConfig failed: %v", err)
	}

	if cfg1.ConfigPath != cfg2.ConfigPath {
		t.Errorf("ConfigPath mismatch: %s vs %s", cfg1.ConfigPath, cfg2.ConfigPath)
	}
}

func TestGetCurrentConfig(t *testing.T) {
	defer chdirToProjectRoot(t)()

	_, err := LoadConfigFromFile()
	if err != nil {
		t.Fatalf("LoadConfigFromFile failed: %v", err)
	}

	cfg := GetCurrentConfig()
	if cfg == nil {
		t.Fatal("GetCurrentConfig returned nil")
	}
	if cfg.App.Name != "otester" {
		t.Errorf("Expected app name 'otester', got '%s'", cfg.App.Name)
	}
}
