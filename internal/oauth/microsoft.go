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
	"otester/internal/logbus"
	"otester/internal/security"
	"golang.org/x/sync/singleflight"
)

type MicrosoftOAuth struct {
	client  *http.Client
	cache   *TokenCache
	sfGroup singleflight.Group
	logger  logbus.Logger
}

func NewMicrosoftOAuth(logger logbus.Logger) *MicrosoftOAuth {
	if logger == nil {
		logger = logbus.Nop()
	}
	return &MicrosoftOAuth{
		client: &http.Client{Timeout: 30 * time.Second},
		cache:  NewTokenCache(),
		logger: logger,
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
		m.logger.Infof("token cache hit for client_id=%s scope=%s", profile.ClientID, profile.Scope)
		return token, true, nil
	}
	m.logger.Infof("token cache miss for client_id=%s scope=%s", profile.ClientID, profile.Scope)

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

		// Cache it for the token's full lifetime as reported by the OAuth
		// server. The refresh_before_expiry_seconds value is NOT the lifetime;
		// it is the slack window before actual expiry at which GetToken /
		// GetTokenStatus should treat the entry as needing a refresh.
		expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

		// Store unredacted token in cache; redaction is only for logging/display
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

	m.logger.Infof("requesting token url=%s client_id=%s scope=%s",
		security.RedactURL(tokenURL), profile.ClientID, profile.Scope)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		m.logger.Errorf("token request failed: status %d", resp.StatusCode)
		return nil, fmt.Errorf("token request failed with status code %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		m.logger.Errorf("token response parse failed: %v", err)
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("access_token missing from token response")
	}

	m.logger.Infof("token acquired expires_in=%ds token=%s",
		tokenResp.ExpiresIn, security.RedactBearerToken(tokenResp.AccessToken))

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

	refreshBefore := time.Duration(profile.RefreshBeforeExpirySeconds) * time.Second
	if refreshBefore == 0 {
		refreshBefore = 60 * time.Second
	}

	token, ok := m.cache.GetStatus(cacheKey, refreshBefore)
	if !ok {
		return &TokenStatus{ProfileID: profileID, HasToken: false, FromCache: false}, nil
	}

	return &TokenStatus{
		ProfileID: profileID,
		HasToken:  true,
		FromCache: true,
		ExpiresAt: &token.ExpiresAt,
	}, nil
}
