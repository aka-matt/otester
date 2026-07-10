package config

import (
	"encoding/json"
	"testing"

	"otester/internal/model"
)

func TestAppConfigJSON(t *testing.T) {
	cfg := AppConfig{
		Name:                  "otester",
		Title:                 "OTester",
		DefaultTimeoutSeconds: 30,
		MaxResponseBodyBytes:  10485760,
		AllowInsecureTLS:      false,
		PersistRequestHistory: true,
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal AppConfig: %v", err)
	}
	var result AppConfig
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal AppConfig: %v", err)
	}
	if result != cfg {
		t.Errorf("AppConfig mismatch: got %+v, want %+v", result, cfg)
	}
}

func TestVariableJSON(t *testing.T) {
	v := Variable{ID: "var-1", BaseURL: "https://api.example.com", Environment: "production"}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal Variable: %v", err)
	}
	var result Variable
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal Variable: %v", err)
	}
	if result != v {
		t.Errorf("Variable mismatch: got %+v, want %+v", result, v)
	}
}

func TestOAuthProfileJSON(t *testing.T) {
	profile := OAuthProfile{
		ID:                         "oauth-1",
		Name:                       "Test OAuth",
		Type:                       "client_credentials",
		OrgIDUUID:                  "org-123",
		ClientID:                   "client-id",
		ClientSecret:               "secret",
		Scope:                      "read write",
		TokenURL:                   "https://auth.example.com/token",
		RefreshBeforeExpirySeconds: 60,
	}
	data, err := json.Marshal(profile)
	if err != nil {
		t.Fatalf("failed to marshal OAuthProfile: %v", err)
	}
	var result OAuthProfile
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal OAuthProfile: %v", err)
	}
	if result.ClientSecret != profile.ClientSecret {
		t.Errorf("OAuthProfile ClientSecret mismatch")
	}
}

func TestOAuthProfileViewJSON(t *testing.T) {
	view := OAuthProfileView{
		ID:                         "oauth-1",
		Name:                       "Test OAuth",
		Type:                       "client_credentials",
		OrgIDUUID:                  "org-123",
		ClientID:                   "client-id",
		ClientSecretMasked:         "******",
		Scope:                      "read write",
		TokenURL:                   "https://auth.example.com/token",
		RefreshBeforeExpirySeconds: 60,
	}
	data, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("failed to marshal OAuthProfileView: %v", err)
	}
	var result OAuthProfileView
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal OAuthProfileView: %v", err)
	}
	if result.ClientSecretMasked != "******" {
		t.Errorf("OAuthProfileView ClientSecretMasked should always be '******'")
	}
}

func TestEndpointGroupJSON(t *testing.T) {
	group := EndpointGroup{ID: "group-1", Name: "Users API"}
	data, err := json.Marshal(group)
	if err != nil {
		t.Fatalf("failed to marshal EndpointGroup: %v", err)
	}
	var result EndpointGroup
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal EndpointGroup: %v", err)
	}
	if result != group {
		t.Errorf("EndpointGroup mismatch: got %+v, want %+v", result, group)
	}
}

func TestAuthConfigJSON(t *testing.T) {
	auth := AuthConfig{Type: "oauth2", ProfileID: "oauth-1", AllowAuthorizationHeaderOverride: true}
	data, err := json.Marshal(auth)
	if err != nil {
		t.Fatalf("failed to marshal AuthConfig: %v", err)
	}
	var result AuthConfig
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal AuthConfig: %v", err)
	}
	if result != auth {
		t.Errorf("AuthConfig mismatch: got %+v, want %+v", result, auth)
	}
}

func TestEndpointJSON(t *testing.T) {
	ep := Endpoint{
		ID:             "ep-1",
		Name:           "Get Users",
		Description:    "Retrieve all users",
		GroupID:        "group-1",
		Enabled:        true,
		Method:         "GET",
		URL:            "{{base_url}}/users",
		TimeoutSeconds: 30,
		Auth:           AuthConfig{Type: "none"},
		Headers:        []model.KeyValue{{Key: "Accept", Value: "application/json", Enabled: true}},
		QueryParams:    []model.KeyValue{},
		Body:           BodyConfig{Type: "none", Content: ""},
	}
	data, err := json.Marshal(ep)
	if err != nil {
		t.Fatalf("failed to marshal Endpoint: %v", err)
	}
	var result Endpoint
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal Endpoint: %v", err)
	}
	if result.ID != ep.ID || result.Method != ep.Method || result.Enabled != ep.Enabled {
		t.Errorf("Endpoint mismatch: got %+v, want %+v", result, ep)
	}
}

func TestEndpointViewJSON(t *testing.T) {
	view := EndpointView{
		ID:             "ep-1",
		Name:           "Get Users",
		Description:    "Retrieve all users",
		GroupID:        "group-1",
		Enabled:        true,
		Method:         "GET",
		URL:            "{{base_url}}/users",
		TimeoutSeconds: 30,
		Auth:           AuthConfig{Type: "oauth2", ProfileID: "oauth-1"},
		Headers:        []model.KeyValue{{Key: "Accept", Value: "application/json", Enabled: true}},
		QueryParams:    []model.KeyValue{},
		Body:           BodyConfig{Type: "none", Content: ""},
		HasOAuth:       true,
	}
	data, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("failed to marshal EndpointView: %v", err)
	}
	var result EndpointView
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal EndpointView: %v", err)
	}
	if result.HasOAuth != view.HasOAuth {
		t.Errorf("EndpointView HasOAuth mismatch")
	}
}

func TestBodyConfigJSON(t *testing.T) {
	body := BodyConfig{Type: "json", Content: `{"name":"test"}`}
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal BodyConfig: %v", err)
	}
	var result BodyConfig
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal BodyConfig: %v", err)
	}
	if result != body {
		t.Errorf("BodyConfig mismatch: got %+v, want %+v", result, body)
	}
}

func TestKeyValueItemJSON(t *testing.T) {
	item := KeyValueItem{Key: "debug", Value: "true", Enabled: false}
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("failed to marshal KeyValueItem: %v", err)
	}
	var result KeyValueItem
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal KeyValueItem: %v", err)
	}
	if result != item {
		t.Errorf("KeyValueItem mismatch: got %+v, want %+v", result, item)
	}
}

func TestConfigViewJSON(t *testing.T) {
	view := ConfigView{
		App: AppConfig{
			Name: "otester", Title: "OTester", DefaultTimeoutSeconds: 30,
			MaxResponseBodyBytes: 10485760, AllowInsecureTLS: false, PersistRequestHistory: true,
		},
		Variables:      []Variable{{ID: "var-1", BaseURL: "https://api.example.com", Environment: "prod"}},
		OAuthProfiles:  []OAuthProfileView{{ID: "oauth-1", Name: "Test", Type: "client_credentials", OrgIDUUID: "org-1", ClientID: "cid", ClientSecretMasked: "******", Scope: "read"}},
		EndpointGroups: []EndpointGroup{{ID: "group-1", Name: "Users"}},
		Endpoints:      []EndpointView{{ID: "ep-1", Name: "Get Users", Enabled: true, Method: "GET", URL: "/users", Auth: AuthConfig{Type: "none"}, Body: BodyConfig{Type: "none"}}},
		ConfigPath:     "/path/to/config.yaml",
	}
	data, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("failed to marshal ConfigView: %v", err)
	}
	var result ConfigView
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal ConfigView: %v", err)
	}
	if result.ConfigPath != view.ConfigPath || len(result.Endpoints) != len(view.Endpoints) {
		t.Errorf("ConfigView mismatch: got %+v, want %+v", result, view)
	}
}

func TestValidationResultJSON(t *testing.T) {
	vr := ValidationResult{
		Valid:    false,
		Errors:   []string{"missing required field", "invalid value"},
		Warnings: []string{"deprecated field used"},
	}
	data, err := json.Marshal(vr)
	if err != nil {
		t.Fatalf("failed to marshal ValidationResult: %v", err)
	}
	var result ValidationResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal ValidationResult: %v", err)
	}
	if result.Valid != vr.Valid || len(result.Errors) != len(vr.Errors) {
		t.Errorf("ValidationResult mismatch: got %+v, want %+v", result, vr)
	}
}

func TestValidationResultValidJSON(t *testing.T) {
	vr := ValidationResult{Valid: true, Errors: nil, Warnings: nil}
	data, err := json.Marshal(vr)
	if err != nil {
		t.Fatalf("failed to marshal ValidationResult: %v", err)
	}
	var result ValidationResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal ValidationResult: %v", err)
	}
	if !result.Valid {
		t.Errorf("ValidationResult Valid should be true")
	}
}
