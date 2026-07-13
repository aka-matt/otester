# HTTP / Auth Log Window — Design

**Date:** 2026-07-13
**Status:** Approved (ready for implementation plan)

## Goal

Add an in-app, read-only, multi-line log window that is hidden by default and
opened from **Tools → Open Logs**. Every HTTP request/response cycle — including
the full OAuth token-acquisition flow — writes a log line into this window.

## Context & Constraints

- App is **Wails v2** with a **single native window**. Wails v2 has no supported
  multi-window API, so the "log window" is realized as an **in-app modal**,
  mirroring the existing `frontend/src/components/Base64Convert.vue` pattern
  (`n-modal` toggled from the Tools menu via `v-model:show`).
- **OAuth is currently not wired into the request path.** `oauth.GetAccessToken`
  exists but is never called; `BuildRequest` never adds an `Authorization` header;
  `DoRequest` ignores the `UseOAuth`/`OAuthProfileID` fields the frontend sends.
  This spec **wires OAuth into `SendRequest`** so the auth flow actually runs and
  can be logged.
- **Security (CLAUDE.md §4):** `client_secret` and full `access_token` /
  `Authorization: Bearer …` values must never appear in logs. All log messages
  pass through `internal/security/redact` helpers, which already exist:
  `RedactURL`, `RedactBearerToken`, `RedactAuthorizationHeader`, `RedactHeaders`.
- **No disk persistence:** logs live in memory only (consistent with the project's
  no-persist stance for tokens and history in v1).

## Architecture

### 1. Backend log bus — `internal/logbus/logbus.go`

A small, thread-safe logger injected wherever logging is needed.

```go
type Level string
const (LevelDebug Level = "DEBUG"; LevelInfo Level = "INFO"; LevelWarn Level = "WARN"; LevelError Level = "ERROR")

type Entry struct {
    Time    string `json:"time"`    // RFC3339 / HH:MM:SS formatted
    Level   string `json:"level"`
    Message string `json:"message"`
}

// Logger is the interface consumers depend on (not the concrete app).
type Logger interface {
    Log(level Level, format string, args ...any)
    Debugf(format string, args ...any)
    Infof(format string, args ...any)
    Warnf(format string, args ...any)
    Errorf(format string, args ...any)
}
```

Concrete `*Bus`:
- Holds a **bounded ring buffer** (capacity ~2000 entries; oldest dropped) behind a
  mutex, so opening the window shows history that occurred before it was opened.
- On each append, if a Wails `ctx` is set, also emits the entry live via
  `runtime.EventsEmit(ctx, "log:entry", entry)`.
- `SetContext(ctx context.Context)` — called from `App.Startup` so the bus can emit.
- `Snapshot() []Entry` — returns a copy of the buffer (backs `GetLogs`).
- `Clear()` — empties the buffer (backs `ClearLogs`).
- **Redaction is the caller's responsibility at the call site** (callers pass
  already-redacted values using the `security` helpers). The bus does not attempt
  to parse arbitrary messages; call sites must never pass raw secrets.

### 2. Wiring & log points

- `App` owns a `*logbus.Bus`. `App.Startup` calls `bus.SetContext(ctx)`.
- Injected via constructors: `oauth.NewMicrosoftOAuth(logger)` and
  `httpclient.NewClient(logger)` accept a `logbus.Logger`. (A nil/no-op logger is
  tolerated so existing tests that construct them directly keep working — provide a
  `logbus.Nop()` no-op logger for that.)

**OAuth (`internal/oauth/microsoft.go`)** — log inside `GetAccessToken` / `fetchToken`:
- cache lookup result (hit → `INFO token cache hit`, miss → `INFO token cache miss`)
- token fetch start: `INFO requesting token url=<RedactURL> client_id=<id> scope=<scope>`
  (never the secret)
- success: `INFO token acquired expires_in=<n>s token=<RedactBearerToken>`
- failure: `ERROR token request failed: <status/err>`

**OAuth injection into requests (`app/app.go` `SendRequest`)**:
- When `input.UseOAuth`: resolve profile from config by `input.OAuthProfileID`
  (reuse the lookup already present in `GetTokenStatus`). If not found →
  return `ResponseOutput` with `ErrorCode = OAUTH_PROFILE_NOT_FOUND`.
- Call `oauth.GetAccessToken(ctx, profile)`. On error → map to
  `OAUTH_TOKEN_REQUEST_FAILED` (and `OAUTH_ACCESS_TOKEN_MISSING` where applicable).
- On success, append `Authorization: Bearer <token>` (enabled) to `input.Headers`,
  and set `ResponseOutput.UsedOAuth = true`, `TokenFromCache = <fromCache>`.
  (`UsedOAuth`/`TokenFromCache` fields already exist on the model.)

**HTTP client (`internal/httpclient/client.go`)** — log inside `DoRequest`:
- start: `INFO → <METHOD> <RedactURL>` plus redacted headers (`RedactHeaders`)
- insecure-TLS retry path: `WARN TLS validation failed, retrying with insecure TLS`
- response: `INFO ← <status> <bytes>B <duration>` — or `ERROR request failed: <err>`

### 3. Backend bindings (`app/app.go`)

- `GetLogs() []logbus.Entry` — returns `bus.Snapshot()` to prime the window on open.
- `ClearLogs()` — calls `bus.Clear()`.
- Event emitted to frontend: `"log:entry"` (single `Entry` payload).

### 4. Frontend

- **`frontend/src/stores/logs.ts`** (Pinia): `entries: Entry[]`, `append(e)`,
  `prime(list)`, `clear()`. Caps client-side length (~2000) to match backend.
- **Subscription** registered in `App.vue` `onMounted`: call `GetLogs()` to prime,
  then `EventsOn('log:entry', e => logsStore.append(e))`. This runs regardless of
  whether the window is open, so logs accumulate in the background.
- **`frontend/src/components/LogsViewer.vue`**: an `n-modal` mirroring
  `Base64Convert.vue`. Contains a **read-only** `n-input type="textarea"` (monospace)
  bound to the joined log text, auto-scrolled to the newest line, plus **Clear**
  (calls `ClearLogs()` + `logsStore.clear()`) and **Copy** buttons. Props `show`,
  emits `update:show`.
- **`frontend/src/views/MainView.vue`**:
  - add `{ label: 'Open Logs', key: 'open-logs' }` to `toolsMenuOptions`
  - add `case 'open-logs': showLogs.value = true` in `handleMenuSelect`
  - add `const showLogs = ref(false)` and `<LogsViewer v-model:show="showLogs" />`
- **Types**: add a `LogEntry` type in `frontend/src/types/index.ts`.
- **Wails bindings**: `frontend/wailsjs/go/app/App.{js,d.ts}` are committed and
  regenerate on `wails build`. If not regenerating in the working environment,
  hand-add `GetLogs`/`ClearLogs` declarations to match.

## Data Flow

```
SendRequest(input)
  └─ if UseOAuth: resolve profile → oauth.GetAccessToken ──┐ (logs cache/fetch/success/fail)
  │                                    inject Bearer header │
  └─ httpClient.DoRequest ─────────────────────────────────┘ (logs → request, ← response/retry/error)
        every logbus.Log(...) → ring buffer  ── EventsEmit("log:entry") ──▶ App.vue EventsOn
                                     ▲                                            └─ logsStore.append
                                     └── GetLogs() snapshot ◀── LogsViewer opens ─┘ (prime)
```

## Error Handling

- OAuth failures during `SendRequest` return a `ResponseOutput` carrying the
  appropriate existing `OAUTH_*` error code and message (no HTTP request is sent).
- Bus emit with a nil ctx is a silent no-op (before `Startup`); entries still buffer.
- Copy uses the browser clipboard API; failure shows a naive-ui message, non-fatal.

## Testing

**Go**
- `internal/logbus/logbus_test.go`: ring-buffer cap (oldest dropped), `Snapshot`
  returns a copy, `Clear` empties, no-op logger is safe, emit is skipped when ctx nil.
- `app/app_test.go`: new `SendRequest` OAuth branch — token injected as Bearer
  header, `UsedOAuth`/`TokenFromCache` set, profile-not-found and token-failure map
  to the right error codes. Use a fake oauth/httpclient seam consistent with the
  existing test style.

**Frontend (Vitest)**
- `stores/logs.test.ts`: append/prime/clear and length cap.
- `components/LogsViewer.test.ts`: renders entries read-only, Clear button calls the
  binding + store clear, respects `show`/`update:show`.

## Out of Scope (YAGNI)

- Log level filtering / search UI, log export to file, and a separate native OS
  window are explicitly out of scope for this change.
