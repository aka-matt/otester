package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"otester/internal/config"
	"otester/internal/model"
)

func newTestApp() *App {
	a := NewApp()
	a.ctx = context.Background()
	return a
}

func TestSendRequestInjectsBearerHeaderOnOAuth(t *testing.T) {
	a := newTestApp()
	a.loadOAuthProfiles = func(string) ([]config.OAuthProfile, error) {
		return []config.OAuthProfile{{ID: "p1", ClientID: "c", ClientSecret: "s"}}, nil
	}
	a.getAccessToken = func(context.Context, *config.OAuthProfile) (string, bool, error) {
		return "tok-123", true, nil
	}
	var captured *model.RequestInput
	a.doRequest = func(_ context.Context, in *model.RequestInput) (*model.ResponseOutput, error) {
		captured = in
		return &model.ResponseOutput{RequestID: in.RequestID, StatusCode: 200}, nil
	}

	out, err := a.SendRequest(&model.RequestInput{RequestID: "r1", Method: "GET", URL: "https://x/", UseOAuth: true, OAuthProfileID: "p1"})
	if err != nil {
		t.Fatalf("SendRequest error: %v", err)
	}
	if !out.UsedOAuth || !out.TokenFromCache {
		t.Fatalf("want UsedOAuth+TokenFromCache, got %+v", out)
	}
	var authVal string
	for _, h := range captured.Headers {
		if strings.EqualFold(h.Key, "Authorization") {
			authVal = h.Value
		}
	}
	if authVal != "Bearer tok-123" {
		t.Fatalf("want Authorization 'Bearer tok-123', got %q", authVal)
	}
}

func TestSendRequestProfileNotFound(t *testing.T) {
	a := newTestApp()
	a.loadOAuthProfiles = func(string) ([]config.OAuthProfile, error) { return []config.OAuthProfile{}, nil }
	a.doRequest = func(context.Context, *model.RequestInput) (*model.ResponseOutput, error) {
		t.Fatal("doRequest must not be called when profile is missing")
		return nil, nil
	}

	out, err := a.SendRequest(&model.RequestInput{RequestID: "r1", UseOAuth: true, OAuthProfileID: "missing"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ErrorCode != string(model.ErrOAuthProfileNotFound) {
		t.Fatalf("want OAUTH_PROFILE_NOT_FOUND, got %q", out.ErrorCode)
	}
}

func TestSendRequestTokenFailure(t *testing.T) {
	a := newTestApp()
	a.loadOAuthProfiles = func(string) ([]config.OAuthProfile, error) {
		return []config.OAuthProfile{{ID: "p1"}}, nil
	}
	a.getAccessToken = func(context.Context, *config.OAuthProfile) (string, bool, error) {
		return "", false, errors.New("boom")
	}
	out, err := a.SendRequest(&model.RequestInput{RequestID: "r1", UseOAuth: true, OAuthProfileID: "p1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ErrorCode != string(model.ErrOAuthTokenRequestFailed) {
		t.Fatalf("want OAUTH_TOKEN_REQUEST_FAILED, got %q", out.ErrorCode)
	}
}

func TestGetLogsAndClearLogs(t *testing.T) {
	a := newTestApp()
	a.logs.Infof("hello")
	if len(a.GetLogs()) != 1 {
		t.Fatalf("want 1 log entry, got %d", len(a.GetLogs()))
	}
	a.ClearLogs()
	if len(a.GetLogs()) != 0 {
		t.Fatal("ClearLogs must empty the buffer")
	}
}

func TestOpenConfigFileSwitchesActivePathAfterSuccessfulLoad(t *testing.T) {
	a := NewApp()
	a.activeConfigPath = "current.json"
	a.openFileDialog = func() (string, error) { return "selected.json", nil }
	a.loadConfigFromPath = func(path string) (*config.ConfigView, error) {
		return &config.ConfigView{ConfigPath: path}, nil
	}

	cfg, err := a.OpenConfigFile()
	if err != nil {
		t.Fatalf("OpenConfigFile failed: %v", err)
	}
	if cfg == nil || cfg.ConfigPath != "selected.json" {
		t.Fatalf("Expected selected config, got %#v", cfg)
	}
	if a.activeConfigPath != "selected.json" {
		t.Errorf("Expected active path to switch, got %q", a.activeConfigPath)
	}
}

func TestOpenConfigFileCancellationPreservesActivePath(t *testing.T) {
	a := NewApp()
	a.activeConfigPath = "current.json"
	a.openFileDialog = func() (string, error) { return "", nil }

	cfg, err := a.OpenConfigFile()
	if err != nil {
		t.Fatalf("OpenConfigFile failed: %v", err)
	}
	if cfg != nil {
		t.Errorf("Expected nil config on cancellation, got %#v", cfg)
	}
	if a.activeConfigPath != "current.json" {
		t.Errorf("Expected active path to remain current, got %q", a.activeConfigPath)
	}
}

func TestOpenConfigFileLoaderErrorPreservesActivePath(t *testing.T) {
	a := NewApp()
	a.activeConfigPath = "current.json"
	a.openFileDialog = func() (string, error) { return "invalid.json", nil }
	a.loadConfigFromPath = func(string) (*config.ConfigView, error) {
		return nil, errors.New("invalid config")
	}

	cfg, err := a.OpenConfigFile()
	if err == nil {
		t.Fatal("Expected loader error")
	}
	if cfg != nil {
		t.Errorf("Expected nil config on loader error, got %#v", cfg)
	}
	if a.activeConfigPath != "current.json" {
		t.Errorf("Expected active path to remain current, got %q", a.activeConfigPath)
	}
}
