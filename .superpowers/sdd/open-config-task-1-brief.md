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

