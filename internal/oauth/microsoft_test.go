package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
