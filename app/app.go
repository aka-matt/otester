package app

import (
	"context"

	"otester/internal/config"
)

type App struct {
	ctx context.Context
}

type AppInfo struct {
	Version string
	Name    string
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
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
