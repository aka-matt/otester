# Open Config File Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Load a user-selected JSON configuration file for the current app run.

**Architecture:** Add an explicit-path loader that updates global config only after successful parsing. `App` owns the active path and uses Wails' native open-file dialog; the frontend invokes it from the renamed File menu item and atomically replaces its Pinia config on success.

**Tech Stack:** Go, Wails v2 runtime dialog, Vue 3, Pinia, Vitest.

## Global Constraints

- A selected path is in-memory only; application restart returns to working-directory `config.json`.
- Cancellation and all failed loads retain the existing active configuration.
- Only JSON files are selectable; configuration remains read-only.
- Reload Config re-reads the current active configuration path.

### Task 1: Path-based backend config loading and dialog API

**Files:**
- Modify: `internal/config/loader.go`, `internal/config/loader_test.go`, `app/app.go`
- Create: `app/app_test.go`

- [ ] Write failing tests for `LoadConfigFromPath(path)` returning a `ConfigView` with that exact path, and invalid JSON retaining `GetCurrentConfig()`.
- [ ] Run `go test ./internal/config` and verify the new tests fail because `LoadConfigFromPath` is absent.
- [ ] Implement `LoadConfigFromPath`: move current parse logic into it; assign `currentConfig` only after successful parse. Keep `LoadConfigFromFile` as a working-directory wrapper.
- [ ] Add `activeConfigPath string` to `App`, initialize it to the working-directory config path at startup, and make Load/Reload use `config.LoadConfigFromPath(activeConfigPath)`.
- [ ] Implement `OpenConfigFile() (*config.ConfigView, error)` using `runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Filters: []runtime.FileFilter{{DisplayName: "JSON files", Pattern: "*.json"}}})`; empty selected path returns `(nil, nil)`. Load selected path first, then assign `activeConfigPath` only on success.
- [ ] Unit-test the path-switch helper through an injectable loader/dialog seam: success switches path, cancellation preserves it, and loader error preserves it.
- [ ] Run `go test ./internal/config ./app` and commit `feat: load runtime config file`.

### Task 2: Frontend menu integration and generated bindings

**Files:**
- Modify: `frontend/src/views/MainView.vue`, `frontend/src/stores/config.ts`
- Regenerate: `frontend/wailsjs/go/app/App.js`, `frontend/wailsjs/go/app/App.d.ts`
- Create/Modify: `frontend/src/views/MainView.test.ts`

- [ ] Write a failing menu test that expects `Open Config File`, mocks `OpenConfigFile`, and verifies a returned config replaces `configStore.config`; test cancellation leaves it unchanged and errors show Naive UI error feedback.
- [ ] Run `npm test -- src/views/MainView.test.ts` and verify RED.
- [ ] Add `OpenConfigFile(): Promise<ConfigView | null>` to frontend Wails typing and regenerate bindings with the project Wails generator rather than hand-editing generated files.
- [ ] Rename menu option/key, call `window.go.app.OpenConfigFile()`, assign successful config through a store method that reconciles selected variables, and clear response; cancellation does nothing and catch reports error.
- [ ] Run `npm test && npm run build`, then `go test ./...`; report the known hard-coded-path failures separately if unchanged. Commit `feat: open config file from menu`.

## Plan Self-Review

- Task 1 covers atomic backend path replacement, cancellation, failure retention, and reload semantics.
- Task 2 covers the visible menu rename, frontend state update, user feedback, and regenerated bindings.
- All interfaces use `OpenConfigFile() (*config.ConfigView, error)` / `Promise<ConfigView | null>` consistently.
