package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"otester/internal/config"
	"otester/internal/logbus"
)

func TestGetAccessTokenLogsFetchAndSuccessWithoutSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(TokenResponse{AccessToken: "supersecrettoken123456", ExpiresIn: 3600})
	}))
	defer srv.Close()

	bus := logbus.New(50)
	o := NewMicrosoftOAuth(bus)
	profile := &config.OAuthProfile{
		ID: "p1", ClientID: "client-abc", ClientSecret: "the-real-secret",
		Scope: "api://x/.default", TokenURL: srv.URL,
	}

	tok, fromCache, err := o.GetAccessToken(context.Background(), profile)
	if err != nil {
		t.Fatalf("GetAccessToken error: %v", err)
	}
	if tok != "supersecrettoken123456" || fromCache {
		t.Fatalf("unexpected token/fromCache: %q %v", tok, fromCache)
	}

	var joined string
	for _, e := range bus.Snapshot() {
		joined += e.Message + "\n"
	}
	if strings.Contains(joined, "the-real-secret") {
		t.Fatalf("client_secret leaked into logs:\n%s", joined)
	}
	if strings.Contains(joined, "supersecrettoken123456") {
		t.Fatalf("full access_token leaked into logs:\n%s", joined)
	}
	if !strings.Contains(joined, "requesting token") || !strings.Contains(joined, "token acquired") {
		t.Fatalf("missing expected log lines:\n%s", joined)
	}
}

// Regression: previously GetAccessToken stored ExpiresAt = now +
// RefreshBeforeExpirySeconds, which is the refresh window, not the token
// lifetime. After the fix, ExpiresAt is now + ExpiresIn (the actual lifetime
// reported by the OAuth server), and GetTokenStatus reports HasToken=true
// for the full lifetime minus the refresh window.
func TestGetTokenStatusReportsCachedTokenAcrossLifetime(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(TokenResponse{AccessToken: "tok", ExpiresIn: 3600})
	}))
	defer srv.Close()

	o := NewMicrosoftOAuth(logbus.New(10))
	profile := &config.OAuthProfile{
		ID:                         "p1",
		ClientID:                   "cid",
		ClientSecret:               "secret",
		Scope:                      "s",
		TokenURL:                   srv.URL,
		RefreshBeforeExpirySeconds: 60,
	}

	if _, _, err := o.GetAccessToken(context.Background(), profile); err != nil {
		t.Fatalf("GetAccessToken error: %v", err)
	}

	st, err := o.GetTokenStatus("p1", profile)
	if err != nil {
		t.Fatalf("GetTokenStatus error: %v", err)
	}
	if !st.HasToken {
		t.Fatalf("HasToken must be true after a successful fetch: %+v", st)
	}
	if !st.FromCache {
		t.Errorf("FromCache must be true for a re-read from the cache: %+v", st)
	}
	if st.ExpiresAt == nil {
		t.Fatal("ExpiresAt must be populated when HasToken=true")
	}
	if delta := time.Until(*st.ExpiresAt); delta < 59*time.Minute {
		t.Errorf("ExpiresAt must reflect the real lifetime (ExpiresIn=3600s), got remaining=%s", delta)
	}
}

// When the cached entry's remaining lifetime drops below
// refresh_before_expiry_seconds, GetTokenStatus must report HasToken=false
// so the next call to GetAccessToken re-fetches.
func TestGetTokenStatusHonorsRefreshWindow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(TokenResponse{AccessToken: "tok", ExpiresIn: 3600})
	}))
	defer srv.Close()

	o := NewMicrosoftOAuth(logbus.New(10))
	profile := &config.OAuthProfile{
		ID:                         "p1",
		ClientID:                   "cid",
		ClientSecret:               "secret",
		Scope:                      "s",
		TokenURL:                   srv.URL,
		RefreshBeforeExpirySeconds: 60,
	}
	if _, _, err := o.GetAccessToken(context.Background(), profile); err != nil {
		t.Fatalf("GetAccessToken error: %v", err)
	}

	// Force the cached entry to be effectively expired by shortening its
	// expiry to within the refresh window.
	key := BuildCacheKey(profile.TokenURL, profile.ClientID, profile.Scope)
	o.cache.Set(key, &CachedToken{AccessToken: "tok", ExpiresAt: time.Now().Add(10 * time.Second)})

	st, err := o.GetTokenStatus("p1", profile)
	if err != nil {
		t.Fatalf("GetTokenStatus error: %v", err)
	}
	if st.HasToken {
		t.Errorf("HasToken must be false when the cached entry is within refresh window: %+v", st)
	}
}
