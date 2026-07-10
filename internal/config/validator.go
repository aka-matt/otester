package config

func ValidateConfig(cfg *ConfigView) *ValidationResult {
	result := &ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}

	// Basic validation
	if cfg == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "config is nil")
		return result
	}

	if cfg.App.Title == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "app.title is required")
	}

	// Endpoint ID uniqueness
	endpointIDs := make(map[string]bool)
	for _, ep := range cfg.Endpoints {
		if endpointIDs[ep.ID] {
			result.Valid = false
			result.Errors = append(result.Errors, "duplicate endpoint id: "+ep.ID)
		}
		endpointIDs[ep.ID] = true
	}

	// OAuth profile ID uniqueness
	profileIDs := make(map[string]bool)
	for _, p := range cfg.OAuthProfiles {
		if profileIDs[p.ID] {
			result.Valid = false
			result.Errors = append(result.Errors, "duplicate oauth profile id: "+p.ID)
		}
		profileIDs[p.ID] = true
	}

	// Endpoint group references
	groupIDs := make(map[string]bool)
	for _, g := range cfg.EndpointGroups {
		groupIDs[g.ID] = true
	}
	for _, ep := range cfg.Endpoints {
		if ep.GroupID != "" && !groupIDs[ep.GroupID] {
			result.Errors = append(result.Errors, "endpoint "+ep.ID+" references non-existent group: "+ep.GroupID)
		}
	}

	// OAuth profile references
	for _, ep := range cfg.Endpoints {
		if ep.Auth.Type == "oauth2" && ep.Auth.ProfileID != "" {
			if !profileIDs[ep.Auth.ProfileID] {
				result.Valid = false
				result.Errors = append(result.Errors, "endpoint "+ep.ID+" references non-existent oauth profile: "+ep.Auth.ProfileID)
			}
		}
	}

	// Method validation
	validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true, "HEAD": true, "OPTIONS": true}
	for _, ep := range cfg.Endpoints {
		if !validMethods[ep.Method] {
			result.Valid = false
			result.Errors = append(result.Errors, "endpoint "+ep.ID+" has invalid method: "+ep.Method)
		}
	}

	// Warnings
	for _, p := range cfg.OAuthProfiles {
		if p.ClientSecretMasked == "replace-with-secret" || p.ClientSecretMasked == "" {
			result.Warnings = append(result.Warnings, "oauth profile "+p.ID+" may have placeholder client_secret")
		}
	}

	return result
}
