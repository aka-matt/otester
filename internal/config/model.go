package config

import (
	"otester/internal/model"
)

type AppConfig struct {
	Name                  string `json:"name"`
	Title                 string `json:"title"`
	DefaultTimeoutSeconds int    `json:"defaultTimeoutSeconds"`
	MaxResponseBodyBytes  int64  `json:"maxResponseBodyBytes"`
	AllowInsecureTLS      bool   `json:"allowInsecureTLS"`
	PersistRequestHistory bool   `json:"persistRequestHistory"`
}

type Variable struct {
	ID          string `json:"id"`
	BaseURL     string `json:"base_url"`
	Environment string `json:"environment"`
}

type OAuthProfile struct {
	ID                         string `json:"id"`
	Name                       string `json:"name"`
	Type                       string `json:"type"`
	OrgIDUUID                  string `json:"org_id_uuid"`
	ClientID                   string `json:"client_id"`
	ClientSecret               string `json:"client_secret"`
	Scope                      string `json:"scope"`
	TokenURL                   string `json:"token_url,omitempty"`
	RefreshBeforeExpirySeconds int    `json:"refreshBeforeExpirySeconds"`
}

type OAuthProfileView struct {
	ID                         string `json:"id"`
	Name                       string `json:"name"`
	Type                       string `json:"type"`
	OrgIDUUID                  string `json:"org_id_uuid"`
	ClientID                   string `json:"client_id"`
	ClientSecretMasked         string `json:"clientSecretMasked"` // always "******"
	Scope                      string `json:"scope"`
	TokenURL                   string `json:"token_url,omitempty"`
	RefreshBeforeExpirySeconds int    `json:"refreshBeforeExpirySeconds"`
}

type EndpointGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AuthConfig struct {
	Type                              string `json:"type"` // "none" or "oauth2"
	ProfileID                         string `json:"profileId,omitempty"`
	AllowAuthorizationHeaderOverride  bool   `json:"allowAuthorizationHeaderOverride"`
}

type Endpoint struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description,omitempty"`
	GroupID        string        `json:"groupId,omitempty"`
	Enabled        bool          `json:"enabled"`
	Method         string        `json:"method"`
	URL            string        `json:"url"`
	TimeoutSeconds int           `json:"timeoutSeconds,omitempty"`
	Auth           AuthConfig    `json:"auth"`
	Headers        []model.KeyValue `json:"headers,omitempty"`
	QueryParams    []model.KeyValue `json:"queryParameters,omitempty"`
	Body           BodyConfig    `json:"body"`
}

type EndpointView struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description,omitempty"`
	GroupID        string        `json:"groupId,omitempty"`
	Enabled        bool          `json:"enabled"`
	Method         string        `json:"method"`
	URL            string        `json:"url"`
	TimeoutSeconds int           `json:"timeoutSeconds,omitempty"`
	Auth           AuthConfig    `json:"auth"`
	Headers        []model.KeyValue `json:"headers,omitempty"`
	QueryParams    []model.KeyValue `json:"queryParameters,omitempty"`
	Body           BodyConfig    `json:"body"`
	HasOAuth       bool          `json:"hasOAuth"`
}

type BodyConfig struct {
	Type    string `json:"type"` // "none", "json", "text", "x-www-form-urlencoded", "raw"
	Content string `json:"content"`
}

type KeyValueItem struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

type ConfigView struct {
	App            AppConfig           `json:"app"`
	Variables      []Variable          `json:"variables"`
	OAuthProfiles  []OAuthProfileView  `json:"oauthProfiles"`
	EndpointGroups []EndpointGroup     `json:"endpointGroups"`
	Endpoints      []EndpointView      `json:"endpoints"`
	ConfigPath     string              `json:"configPath"`
}

type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}
