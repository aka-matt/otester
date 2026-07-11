package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"otester/internal/config"
	"otester/internal/security"
	"golang.org/x/sync/singleflight"
)

type MicrosoftOAuth struct {
	client  *http.Client
	cache   *TokenCache
	sfGroup singleflight.Group
}

func NewMicrosoftOAuth() *MicrosoftOAuth {
	return &MicrosoftOAuth{
		client: &http.Client{Timeout: 30 * time.Second},
		cache:  NewTokenCache(),
	}
}

func (m *MicrosoftOAuth) GetAccessToken(ctx context.Context, profile *config.OAuthProfile) (string, bool, error) {
	tokenURL := profile.TokenURL
	if tokenURL == "" {
		tokenURL = fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token",
			url.PathEscape(profile.OrgIDUUID))
	}

	cacheKey := BuildCacheKey(tokenURL, profile.ClientID, profile.Scope)
	refreshBefore := time.Duration(profile.RefreshBeforeExpirySeconds) * time.Second
	if refreshBefore == 0 {
		refreshBefore = 60 * time.Second
	}

	// Try cache first
	if token, ok := m.cache.Get(cacheKey, refreshBefore); ok {
		return token, true, nil
	}

	// Use singleflight to prevent concurrent refreshes
	result, err, _ := m.sfGroup.Do(cacheKey, func() (interface{}, error) {
		// Double-check cache after acquiring singleflight
		if token, ok := m.cache.Get(cacheKey, refreshBefore); ok {
			return tokenFromCacheResult{token: token, fromCache: true}, nil
		}

		// Fetch new token
		tokenResp, err := m.fetchToken(ctx, profile, tokenURL)
		if err != nil {
			return nil, err
		}

		// Cache it
		expiresAt := time.Now().Add(time.Duration(profile.RefreshBeforeExpirySeconds) * time.Second)
		if profile.RefreshBeforeExpirySeconds == 0 {
			expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		}

		m.cache.Set(cacheKey, &CachedToken{
			AccessToken: tokenResp.AccessToken,
			ExpiresAt:   expiresAt,
		})

		return tokenFromCacheResult{token: tokenResp.AccessToken, fromCache: false}, nil
	})

	if err != nil {
		return "", false, err
	}

	r := result.(tokenFromCacheResult)
	return r.token, r.fromCache, nil
}

type tokenFromCacheResult struct {
	token     string
	fromCache bool
}

func (m *MicrosoftOAuth) fetchToken(ctx context.Context, profile *config.OAuthProfile, tokenURL string) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", profile.ClientID)
	form.Set("scope", profile.Scope)
	form.Set("client_secret", profile.ClientSecret)
	form.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("access_token missing from token response")
	}

	_ = security.RedactJSON("") // Use the security package for potential future redaction

	return &tokenResp, nil
}

func (m *MicrosoftOAuth) ClearCache() {
	m.cache.Clear()
}

func (m *MicrosoftOAuth) GetTokenStatus(profileID string, profile *config.OAuthProfile) (*TokenStatus, error) {
	tokenURL := profile.TokenURL
	if tokenURL == "" {
		tokenURL = fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token",
			url.PathEscape(profile.OrgIDUUID))
	}

	cacheKey := BuildCacheKey(tokenURL, profile.ClientID, profile.Scope)

	m.cache.mu.RLock()
	token, ok := m.cache.tokens[cacheKey]
	m.cache.mu.RUnlock()

	if !ok {
		return &TokenStatus{ProfileID: profileID, HasToken: false}, nil
	}

	return &TokenStatus{
		ProfileID: profileID,
		HasToken:  true,
		ExpiresAt: &token.ExpiresAt,
	}, nil
}