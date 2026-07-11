package app

import (
	"context"

	"otester/internal/config"
	"otester/internal/httpclient"
	"otester/internal/oauth"
)

type App struct {
	ctx        context.Context
	oauth      *oauth.MicrosoftOAuth
	httpClient *httpclient.Client
}

type AppInfo struct {
	Version string
	Name    string
}

func NewApp() *App {
	return &App{
		oauth:      oauth.NewMicrosoftOAuth(),
		httpClient: httpclient.NewClient(),
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	config.LoadConfigFromFile()
}

func (a *App) DomReady(ctx context.Context) {
}

func (a *App) BeforeClose(ctx context.Context) bool {
	return false
}

func (a *App) Shutdown(ctx context.Context) {
}

func (a *App) GetAppInfo() (*AppInfo, error) {
	return &AppInfo{
		Version: "1.0.0",
		Name:    "otester",
	}, nil
}

func (a *App) LoadConfig() (*config.ConfigView, error) {
	return config.LoadConfigFromFile()
}

func (a *App) ReloadConfig() (*config.ConfigView, error) {
	return config.LoadConfigFromFile()
}

func (a *App) ValidateConfig() (*config.ValidationResult, error) {
	cfg, err := config.LoadConfigFromFile()
	if err != nil {
		return &config.ValidationResult{Valid: false, Errors: []string{err.Error()}}, nil
	}
	result := config.ValidateConfig(cfg)
	return result, nil
}

func (a *App) SendRequest(ctx context.Context, input *config.OAuthProfile) (string, bool, error) {
	return a.oauth.GetAccessToken(ctx, input)
}

func (a *App) CancelRequest(requestID string) error {
	return a.httpClient.CancelRequest(requestID)
}

func (a *App) ClearTokenCache() {
	a.oauth.ClearCache()
}

func (a *App) GetTokenStatus(profileID string, profile *config.OAuthProfile) (*oauth.TokenStatus, error) {
	return a.oauth.GetTokenStatus(profileID, profile)
}
