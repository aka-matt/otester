package app

import (
	"context"
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
