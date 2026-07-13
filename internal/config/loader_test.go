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

func TestLoadOAuthProfilesFromPathReturnsRealSecret(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	json := `{"oauth_profiles":[{"id":"p1","client_id":"cid","client_secret":"real-secret","scope":"s","token_url":"https://t/","refresh_before_expiry_seconds":30}]}`
	if err := os.WriteFile(path, []byte(json), 0o600); err != nil {
		t.Fatal(err)
	}

	profiles, err := LoadOAuthProfilesFromPath(path)
	if err != nil {
		t.Fatalf("LoadOAuthProfilesFromPath error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("want 1 profile, got %d", len(profiles))
	}
	p := profiles[0]
	if p.ID != "p1" || p.ClientID != "cid" || p.ClientSecret != "real-secret" ||
		p.Scope != "s" || p.TokenURL != "https://t/" || p.RefreshBeforeExpirySeconds != 30 {
		t.Fatalf("unexpected profile: %+v", p)
	}
}

// Regression: ensure snake_case keys (as defined by schemas/config.schema.json and
// the bundled build/bin/config.json sample) are read for every section — app,
// variables, oauth_profiles, endpoints. Previously the loader only read camelCase
// keys, which silently left the OAuth profile fields empty and made endpoint
// auth.profile_id fail to match any profile.
func TestLoadConfigFromPathReadsSnakeCaseKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	js := `{
		"app": {
			"name": "snake",
			"title": "Snake",
			"default_timeout_seconds": 45,
			"max_response_body_bytes": 2097152,
			"allow_insecure_tls": true,
			"persist_request_history": true
		},
		"variables": [
			{"id": "dev", "base_url": "https://dev/", "environment": "dev"}
		],
		"oauth_profiles": [
			{
				"id": "ms",
				"name": "MS",
				"type": "microsoft_client_credentials",
				"org_id_uuid": "11111111-2222-3333-4444-555555555555",
				"client_id": "cid-snake",
				"client_secret": "secret-snake",
				"scope": "https://graph.microsoft.com/.default",
				"token_url": "https://login.microsoftonline.com/abc/oauth2/v2.0/token",
				"refresh_before_expiry_seconds": 90
			}
		],
		"endpoint_groups": [{"id": "g1", "name": "Group 1"}],
		"endpoints": [
			{
				"id": "ep1",
				"name": "Protected",
				"description": "needs auth",
				"group_id": "g1",
				"enabled": true,
				"method": "GET",
				"url": "https://dev/x",
				"timeout_seconds": 12,
				"auth": {
					"type": "oauth2",
					"profile_id": "ms",
					"allow_authorization_header_override": true
				},
				"headers": [{"key": "Accept", "value": "application/json", "enabled": true}],
				"query_parameters": [{"key": "q", "value": "1", "enabled": true}],
				"body": {"type": "none", "content": ""}
			}
		]
	}`
	if err := os.WriteFile(path, []byte(js), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfigFromPath(path)
	if err != nil {
		t.Fatalf("LoadConfigFromPath error: %v", err)
	}

	// App
	if cfg.App.DefaultTimeoutSeconds != 45 {
		t.Errorf("default_timeout_seconds: want 45, got %d", cfg.App.DefaultTimeoutSeconds)
	}
	if cfg.App.MaxResponseBodyBytes != 2097152 {
		t.Errorf("max_response_body_bytes: want 2097152, got %d", cfg.App.MaxResponseBodyBytes)
	}
	if !cfg.App.AllowInsecureTLS {
		t.Errorf("allow_insecure_tls: want true")
	}
	if !cfg.App.PersistRequestHistory {
		t.Errorf("persist_request_history: want true")
	}

	// OAuth profile
	if len(cfg.OAuthProfiles) != 1 {
		t.Fatalf("oauth_profiles: want 1, got %d", len(cfg.OAuthProfiles))
	}
	p := cfg.OAuthProfiles[0]
	if p.ID != "ms" || p.ClientID != "cid-snake" || p.Scope != "https://graph.microsoft.com/.default" {
		t.Errorf("oauth profile fields wrong: %+v", p)
	}
	if p.ClientSecretMasked != "******" {
		t.Errorf("client secret must be masked, got %q", p.ClientSecretMasked)
	}

	// Endpoint — auth.profile_id must match and HasOAuth must be true
	if len(cfg.Endpoints) != 1 {
		t.Fatalf("endpoints: want 1, got %d", len(cfg.Endpoints))
	}
	ep := cfg.Endpoints[0]
	if ep.GroupID != "g1" {
		t.Errorf("group_id: want g1, got %q", ep.GroupID)
	}
	if ep.TimeoutSeconds != 12 {
		t.Errorf("timeout_seconds: want 12, got %d", ep.TimeoutSeconds)
	}
	if ep.Auth.Type != "oauth2" {
		t.Errorf("auth.type: want oauth2, got %q", ep.Auth.Type)
	}
	if ep.Auth.ProfileID != "ms" {
		t.Errorf("auth.profile_id: want ms, got %q", ep.Auth.ProfileID)
	}
	if !ep.Auth.AllowAuthorizationHeaderOverride {
		t.Errorf("allow_authorization_header_override: want true")
	}
	if !ep.HasOAuth {
		t.Errorf("HasOAuth must be true when profile_id matches a known oauth_profiles entry")
	}
	if len(ep.QueryParams) != 1 || ep.QueryParams[0].Key != "q" {
		t.Errorf("query_parameters not loaded: %+v", ep.QueryParams)
	}
}
