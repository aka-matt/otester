package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"otester/internal/config"
	"otester/internal/httpclient"
	"otester/internal/model"
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

func (a *App) SendRequest(input *model.RequestInput) (*model.ResponseOutput, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app context not initialized")
	}
	return a.httpClient.DoRequest(a.ctx, input)
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

func (a *App) OpenConfigDirectory() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	configPath := filepath.Join(cwd, "config.json")
	dir := filepath.Dir(configPath)
	return openDirectory(dir)
}

func (a *App) OpenLogDirectory() error {
	// For now, open the current working directory as logs would be there
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	return openDirectory(cwd)
}

func openDirectory(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}
