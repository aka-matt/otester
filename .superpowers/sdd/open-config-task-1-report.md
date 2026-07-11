# Task 1: Runtime config loading report

## Delivered

- Added `config.LoadConfigFromPath(path)`, which parses the specified JSON file and only replaces `currentConfig` after a successful parse.
- Retained `LoadConfigFromFile()` as the working-directory `config.json` wrapper.
- Added `App.activeConfigPath`, initialized during startup to the working-directory configuration path. `LoadConfig` and `ReloadConfig` now load this active path.
- Added `App.OpenConfigFile()`, which opens Wails' JSON-only file dialog, leaves state unchanged on cancellation or any error, and switches the active path only after a successful load.
- Added injectable loader and dialog seams for app unit tests.

## Tests

- `TestLoadConfigFromPathUsesProvidedPath`
- `TestLoadConfigFromPathInvalidJSONRetainsCurrentConfig`
- `TestOpenConfigFileSwitchesActivePathAfterSuccessfulLoad`
- `TestOpenConfigFileCancellationPreservesActivePath`
- `TestOpenConfigFileLoaderErrorPreservesActivePath`

The existing config test helper now locates the repository root relative to the package instead of using a hard-coded developer-machine path. Its endpoint expectation was also updated to match the checked-in four-endpoint fixture.

## Verification

```
GOCACHE=/private/tmp/otester-gocache go test ./internal/config ./app
```

Result: both packages passed.
