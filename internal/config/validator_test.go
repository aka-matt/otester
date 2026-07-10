package config

import "testing"

func TestValidateConfig_ValidConfig(t *testing.T) {
	cfg := &ConfigView{
		App: AppConfig{Title: "test"},
		Endpoints: []EndpointView{{ID: "ep1", Method: "GET"}},
	}
	result := ValidateConfig(cfg)
	if !result.Valid {
		t.Errorf("expected valid config, got errors: %v", result.Errors)
	}
}

func TestValidateConfig_DuplicateEndpointID(t *testing.T) {
	cfg := &ConfigView{
		App: AppConfig{Title: "test"},
		Endpoints: []EndpointView{
			{ID: "ep1", Method: "GET"},
			{ID: "ep1", Method: "POST"},
		},
	}
	result := ValidateConfig(cfg)
	if result.Valid {
		t.Error("expected invalid config due to duplicate endpoint ID")
	}
}

func TestValidateConfig_NilConfig(t *testing.T) {
	result := ValidateConfig(nil)
	if result.Valid {
		t.Error("expected invalid config for nil config")
	}
	if len(result.Errors) == 0 || result.Errors[0] != "config is nil" {
		t.Errorf("expected 'config is nil' error, got: %v", result.Errors)
	}
}

func TestValidateConfig_MissingTitle(t *testing.T) {
	cfg := &ConfigView{
		App:       AppConfig{Title: ""},
		Endpoints: []EndpointView{{ID: "ep1", Method: "GET"}},
	}
	result := ValidateConfig(cfg)
	if result.Valid {
		t.Error("expected invalid config due to missing title")
	}
}

func TestValidateConfig_DuplicateOAuthProfileID(t *testing.T) {
	cfg := &ConfigView{
		App: AppConfig{Title: "test"},
		OAuthProfiles: []OAuthProfileView{
			{ID: "prof1", Name: "Profile 1"},
			{ID: "prof1", Name: "Profile 2"},
		},
		Endpoints: []EndpointView{{ID: "ep1", Method: "GET"}},
	}
	result := ValidateConfig(cfg)
	if result.Valid {
		t.Error("expected invalid config due to duplicate oauth profile ID")
	}
}

func TestValidateConfig_InvalidGroupReference(t *testing.T) {
	cfg := &ConfigView{
		App: AppConfig{Title: "test"},
		Endpoints: []EndpointView{
			{ID: "ep1", Method: "GET", GroupID: "non-existent-group"},
		},
	}
	result := ValidateConfig(cfg)
	// This is a warning, not an error
	found := false
	for _, err := range result.Errors {
		if err == "endpoint ep1 references non-existent group: non-existent-group" {
			found = true
		}
	}
	if !found {
		t.Error("expected warning for non-existent group reference")
	}
}

func TestValidateConfig_InvalidOAuthProfileReference(t *testing.T) {
	cfg := &ConfigView{
		App: AppConfig{Title: "test"},
		Endpoints: []EndpointView{
			{ID: "ep1", Method: "GET", Auth: AuthConfig{Type: "oauth2", ProfileID: "non-existent-profile"}},
		},
	}
	result := ValidateConfig(cfg)
	if result.Valid {
		t.Error("expected invalid config due to non-existent oauth profile reference")
	}
}

func TestValidateConfig_InvalidMethod(t *testing.T) {
	cfg := &ConfigView{
		App:       AppConfig{Title: "test"},
		Endpoints: []EndpointView{{ID: "ep1", Method: "INVALID"}},
	}
	result := ValidateConfig(cfg)
	if result.Valid {
		t.Error("expected invalid config due to invalid method")
	}
}

func TestValidateConfig_PlaceholderSecretWarning(t *testing.T) {
	cfg := &ConfigView{
		App: AppConfig{Title: "test"},
		OAuthProfiles: []OAuthProfileView{
			{ID: "prof1", Name: "Profile 1", ClientSecretMasked: "replace-with-secret"},
		},
		Endpoints: []EndpointView{{ID: "ep1", Method: "GET"}},
	}
	result := ValidateConfig(cfg)
	if !result.Valid {
		t.Errorf("expected valid config, got errors: %v", result.Errors)
	}
	if len(result.Warnings) == 0 {
		t.Error("expected warning for placeholder client_secret")
	}
	found := false
	for _, w := range result.Warnings {
		if w == "oauth profile prof1 may have placeholder client_secret" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected specific placeholder warning, got: %v", result.Warnings)
	}
}

func TestValidateConfig_EmptySecretWarning(t *testing.T) {
	cfg := &ConfigView{
		App: AppConfig{Title: "test"},
		OAuthProfiles: []OAuthProfileView{
			{ID: "prof1", Name: "Profile 1", ClientSecretMasked: ""},
		},
		Endpoints: []EndpointView{{ID: "ep1", Method: "GET"}},
	}
	result := ValidateConfig(cfg)
	if !result.Valid {
		t.Errorf("expected valid config, got errors: %v", result.Errors)
	}
	if len(result.Warnings) == 0 {
		t.Error("expected warning for empty client_secret")
	}
}

func TestValidateConfig_ValidMethods(t *testing.T) {
	validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	for _, method := range validMethods {
		cfg := &ConfigView{
			App:       AppConfig{Title: "test"},
			Endpoints: []EndpointView{{ID: "ep1", Method: method}},
		}
		result := ValidateConfig(cfg)
		if result.Valid {
			continue
		}
		t.Errorf("method %s should be valid but got errors: %v", method, result.Errors)
	}
}
