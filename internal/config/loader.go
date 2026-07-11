package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"otester/internal/model"
)

var currentConfig *ConfigView

func LoadConfigFromFile() (*ConfigView, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(cwd, "config.json")
	return LoadConfigFromPath(configPath)
}

func LoadConfigFromPath(configPath string) (*ConfigView, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &ConfigError{Code: model.ErrConfigFileNotFound, Message: "config.json not found in " + configPath}
		}
		return nil, &ConfigError{Code: model.ErrConfigReadFailed, Message: err.Error()}
	}

	var rawCfg map[string]interface{}
	if err := json.Unmarshal(data, &rawCfg); err != nil {
		return nil, &ConfigError{Code: model.ErrConfigParseFailed, Message: err.Error()}
	}

	appCfg := parseAppConfig(rawCfg["app"])
	variables := parseVariables(rawCfg["variables"])
	oauthProfiles := parseOAuthProfiles(rawCfg["oauth_profiles"])
	endpointGroups := parseEndpointGroups(rawCfg["endpoint_groups"])
	endpoints := parseEndpoints(rawCfg["endpoints"], oauthProfiles)

	cfg := &ConfigView{
		App:            appCfg,
		Variables:      variables,
		OAuthProfiles:  oauthProfiles,
		EndpointGroups: endpointGroups,
		Endpoints:      endpoints,
		ConfigPath:     configPath,
	}

	currentConfig = cfg
	return cfg, nil
}

func LoadConfig() (*ConfigView, error) {
	return LoadConfigFromFile()
}

func ReloadConfig() (*ConfigView, error) {
	return LoadConfigFromFile()
}

func GetCurrentConfig() *ConfigView {
	return currentConfig
}

type ConfigError struct {
	Code    model.ErrorCode
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}

func parseAppConfig(raw interface{}) AppConfig {
	if raw == nil {
		return AppConfig{}
	}
	cfg, ok := raw.(map[string]interface{})
	if !ok {
		return AppConfig{}
	}
	return AppConfig{
		Name:                  getString(cfg, "name"),
		Title:                 getString(cfg, "title"),
		DefaultTimeoutSeconds: getInt(cfg, "defaultTimeoutSeconds", 30),
		MaxResponseBodyBytes:  getInt64(cfg, "maxResponseBodyBytes", 10485760),
		AllowInsecureTLS:      getBool(cfg, "allowInsecureTLS", false),
		PersistRequestHistory: getBool(cfg, "persistRequestHistory", false),
	}
}

func parseVariables(raw interface{}) []Variable {
	if raw == nil {
		return []Variable{}
	}
	list, ok := raw.([]interface{})
	if !ok {
		return []Variable{}
	}
	variables := make([]Variable, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		variables = append(variables, Variable{
			ID:          getString(m, "id"),
			BaseURL:     getString(m, "base_url"),
			Environment: getString(m, "environment"),
		})
	}
	return variables
}

func parseOAuthProfiles(raw interface{}) []OAuthProfileView {
	if raw == nil {
		return []OAuthProfileView{}
	}
	list, ok := raw.([]interface{})
	if !ok {
		return []OAuthProfileView{}
	}
	profiles := make([]OAuthProfileView, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		profiles = append(profiles, OAuthProfileView{
			ID:                         getString(m, "id"),
			Name:                       getString(m, "name"),
			Type:                       getString(m, "type"),
			OrgIDUUID:                  getString(m, "org_id_uuid"),
			ClientID:                   getString(m, "client_id"),
			ClientSecretMasked:         "******",
			Scope:                      getString(m, "scope"),
			TokenURL:                   getString(m, "token_url"),
			RefreshBeforeExpirySeconds: getInt(m, "refreshBeforeExpirySeconds", 60),
		})
	}
	return profiles
}

func parseEndpointGroups(raw interface{}) []EndpointGroup {
	if raw == nil {
		return []EndpointGroup{}
	}
	list, ok := raw.([]interface{})
	if !ok {
		return []EndpointGroup{}
	}
	groups := make([]EndpointGroup, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		groups = append(groups, EndpointGroup{
			ID:   getString(m, "id"),
			Name: getString(m, "name"),
		})
	}
	return groups
}

func parseEndpoints(raw interface{}, oauthProfiles []OAuthProfileView) []EndpointView {
	if raw == nil {
		return []EndpointView{}
	}
	list, ok := raw.([]interface{})
	if !ok {
		return []EndpointView{}
	}

	oauthProfileMap := make(map[string]bool)
	for _, p := range oauthProfiles {
		oauthProfileMap[p.ID] = true
	}

	endpoints := make([]EndpointView, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		auth := parseAuthConfig(m["auth"])
		hasOAuth := auth.Type == "oauth2" && oauthProfileMap[auth.ProfileID]

		endpoint := EndpointView{
			ID:             getString(m, "id"),
			Name:           getString(m, "name"),
			Description:    getString(m, "description"),
			GroupID:        getString(m, "group_id"),
			Enabled:        getBool(m, "enabled", true),
			Method:         getString(m, "method"),
			URL:            getString(m, "url"),
			TimeoutSeconds: getInt(m, "timeoutSeconds", 0),
			Auth:           auth,
			Headers:        parseKeyValues(m["headers"]),
			QueryParams:    parseKeyValues(m["queryParameters"]),
			Body:           parseBodyConfig(m["body"]),
			HasOAuth:       hasOAuth,
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints
}

func parseAuthConfig(raw interface{}) AuthConfig {
	if raw == nil {
		return AuthConfig{Type: "none"}
	}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return AuthConfig{Type: "none"}
	}
	return AuthConfig{
		Type:                             getString(m, "type"),
		ProfileID:                        getString(m, "profileId"),
		AllowAuthorizationHeaderOverride: getBool(m, "allowAuthorizationHeaderOverride", false),
	}
}

func parseKeyValues(raw interface{}) []model.KeyValue {
	if raw == nil {
		return []model.KeyValue{}
	}
	list, ok := raw.([]interface{})
	if !ok {
		return []model.KeyValue{}
	}
	items := make([]model.KeyValue, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		items = append(items, model.KeyValue{
			Key:     getString(m, "key"),
			Value:   getString(m, "value"),
			Enabled: getBool(m, "enabled", true),
		})
	}
	return items
}

func parseBodyConfig(raw interface{}) BodyConfig {
	if raw == nil {
		return BodyConfig{Type: "none", Content: ""}
	}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return BodyConfig{Type: "none", Content: ""}
	}
	return BodyConfig{
		Type:    getString(m, "type"),
		Content: getString(m, "content"),
	}
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string, defaultVal int) int {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return defaultVal
}

func getInt64(m map[string]interface{}, key string, defaultVal int64) int64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return int64(f)
		}
	}
	return defaultVal
}

func getBool(m map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultVal
}
