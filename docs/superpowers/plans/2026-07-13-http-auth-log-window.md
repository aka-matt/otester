# HTTP / Auth Log Window Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a hidden, read-only in-app log window (Tools → Open Logs) that receives a log line for every HTTP request/response and for the full OAuth token-acquisition flow.

**Architecture:** A backend `internal/logbus` package keeps a bounded in-memory ring buffer of log entries and streams each new entry to the frontend via the Wails event `log:entry`. The logger is injected into the OAuth client and HTTP client, which emit redacted log lines at each step. OAuth is wired into `SendRequest` (previously dead code) so an `Authorization: Bearer` header is injected and the auth flow actually runs. The frontend keeps a Pinia store primed from `GetLogs()` and appended live from the event, rendered in an `n-modal` read-only textarea mirroring the existing `Base64Convert.vue`.

**Tech Stack:** Go, Wails v2 (`runtime.EventsEmit`), Vue 3 + TypeScript, Pinia, naive-ui, Vitest, Go `testing`.

## Global Constraints

- **Never log secrets:** `client_secret` and full `access_token` / `Authorization: Bearer …` must never appear in logs or UI. Use existing helpers in `internal/security/redact.go`: `RedactURL`, `RedactBearerToken`, `RedactAuthorizationHeader`, `RedactHeaders`. (CLAUDE.md §4)
- **In-memory only:** logs are never written to disk (CLAUDE.md — no persistence in v1).
- **No direct frontend HTTP:** all requests, incl. OAuth, go through the Go backend (CLAUDE.md §"No direct API calls from frontend").
- **Wails v2 is single-window:** the log window is an in-app modal, not a native OS window.
- **Follow existing DI style:** the `App` struct uses injected function fields (`loadConfigFromPath`, `openFileDialog`) for testability. New seams follow the same pattern.
- **Module path:** Go imports are rooted at `otester/…`.

---

### Task 1: Backend log bus (`internal/logbus`)

**Files:**
- Create: `internal/logbus/logbus.go`
- Test: `internal/logbus/logbus_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `type Entry struct { Time string; Level string; Message string }` with JSON tags `time`,`level`,`message`.
  - `type Level string` with consts `LevelDebug`, `LevelInfo`, `LevelWarn`, `LevelError` (values `"DEBUG"`,`"INFO"`,`"WARN"`,`"ERROR"`).
  - `type Logger interface { Log(Level, string, ...any); Debugf(string, ...any); Infof(string, ...any); Warnf(string, ...any); Errorf(string, ...any) }`
  - `func New(capacity int) *Bus` — `*Bus` implements `Logger`.
  - `func Nop() Logger` — no-op logger (safe when unused).
  - Methods on `*Bus`: `SetEmit(fn func(Entry))`, `Snapshot() []Entry`, `Clear()`, `SetClock(fn func() string)` (test seam for timestamps).
  - `Bus` also implements the four `…f` helpers by delegating to `Log`.

- [ ] **Step 1: Write the failing test**

```go
package logbus

import "testing"

func TestRingBufferDropsOldestBeyondCapacity(t *testing.T) {
	b := New(2)
	b.SetClock(func() string { return "T" })
	b.Infof("first")
	b.Infof("second")
	b.Infof("third")

	got := b.Snapshot()
	if len(got) != 2 {
		t.Fatalf("want 2 entries, got %d", len(got))
	}
	if got[0].Message != "second" || got[1].Message != "third" {
		t.Fatalf("want [second third], got [%s %s]", got[0].Message, got[1].Message)
	}
	if got[0].Level != "INFO" || got[0].Time != "T" {
		t.Fatalf("unexpected level/time: %+v", got[0])
	}
}

func TestSnapshotReturnsCopy(t *testing.T) {
	b := New(4)
	b.Infof("one")
	snap := b.Snapshot()
	snap[0].Message = "mutated"
	if b.Snapshot()[0].Message != "one" {
		t.Fatal("Snapshot must return a copy, not the backing slice")
	}
}

func TestClearEmptiesBuffer(t *testing.T) {
	b := New(4)
	b.Infof("one")
	b.Clear()
	if len(b.Snapshot()) != 0 {
		t.Fatal("Clear must empty the buffer")
	}
}

func TestEmitCalledPerEntry(t *testing.T) {
	b := New(4)
	var emitted []Entry
	b.SetEmit(func(e Entry) { emitted = append(emitted, e) })
	b.Warnf("hi %d", 7)
	if len(emitted) != 1 || emitted[0].Message != "hi 7" || emitted[0].Level != "WARN" {
		t.Fatalf("emit not called correctly: %+v", emitted)
	}
}

func TestNopLoggerIsSafe(t *testing.T) {
	l := Nop()
	l.Infof("no panic")
	l.Log(LevelError, "still no panic")
}

func TestEmitNilIsNoop(t *testing.T) {
	b := New(4) // no SetEmit
	b.Infof("buffered but not emitted") // must not panic
	if len(b.Snapshot()) != 1 {
		t.Fatal("entry should still buffer when emit is nil")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/logbus/`
Expected: FAIL — `undefined: New` / package has no test target yet.

- [ ] **Step 3: Write minimal implementation**

```go
package logbus

import (
	"fmt"
	"sync"
	"time"
)

type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

type Entry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

type Logger interface {
	Log(level Level, format string, args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type Bus struct {
	mu       sync.Mutex
	capacity int
	entries  []Entry
	emit     func(Entry)
	clock    func() string
}

func New(capacity int) *Bus {
	if capacity < 1 {
		capacity = 1
	}
	return &Bus{
		capacity: capacity,
		entries:  make([]Entry, 0, capacity),
		clock:    func() string { return time.Now().Format("15:04:05.000") },
	}
}

func (b *Bus) SetEmit(fn func(Entry)) { b.mu.Lock(); b.emit = fn; b.mu.Unlock() }
func (b *Bus) SetClock(fn func() string) { b.mu.Lock(); b.clock = fn; b.mu.Unlock() }

func (b *Bus) Log(level Level, format string, args ...any) {
	e := Entry{Level: string(level), Message: fmt.Sprintf(format, args...)}
	b.mu.Lock()
	e.Time = b.clock()
	b.entries = append(b.entries, e)
	if len(b.entries) > b.capacity {
		b.entries = b.entries[len(b.entries)-b.capacity:]
	}
	emit := b.emit
	b.mu.Unlock()
	if emit != nil {
		emit(e)
	}
}

func (b *Bus) Debugf(f string, a ...any) { b.Log(LevelDebug, f, a...) }
func (b *Bus) Infof(f string, a ...any)  { b.Log(LevelInfo, f, a...) }
func (b *Bus) Warnf(f string, a ...any)  { b.Log(LevelWarn, f, a...) }
func (b *Bus) Errorf(f string, a ...any) { b.Log(LevelError, f, a...) }

func (b *Bus) Snapshot() []Entry {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]Entry, len(b.entries))
	copy(out, b.entries)
	return out
}

func (b *Bus) Clear() {
	b.mu.Lock()
	b.entries = b.entries[:0]
	b.mu.Unlock()
}

type nop struct{}

func Nop() Logger { return nop{} }

func (nop) Log(Level, string, ...any) {}
func (nop) Debugf(string, ...any)     {}
func (nop) Infof(string, ...any)      {}
func (nop) Warnf(string, ...any)      {}
func (nop) Errorf(string, ...any)     {}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/logbus/`
Expected: PASS (all tests).

- [ ] **Step 5: Commit**

```bash
git add internal/logbus/logbus.go internal/logbus/logbus_test.go
git commit -m "feat(logbus): in-memory ring-buffer logger with emit hook"
```

---

### Task 2: Inject logger into HTTP client and log request/response

**Files:**
- Modify: `internal/httpclient/client.go`
- Test: `internal/httpclient/client_test.go` (add one test; keep existing)

**Interfaces:**
- Consumes: `logbus.Logger`, `logbus.Nop`, `security.RedactURL`, `security.RedactHeaders` (Task 1 + existing).
- Produces:
  - `func NewClient(logger logbus.Logger) *Client` — **signature change** (was `NewClient()`).
  - `Client` gains field `logger logbus.Logger`.
  - Log lines emitted from `DoRequest`.

- [ ] **Step 1: Update existing callers/tests to the new constructor signature first**

In `internal/httpclient/client_test.go`, every `NewClient()` becomes `NewClient(logbus.Nop())`. Add import `"otester/internal/logbus"`. (Do the same for any other `NewClient()` call found via `grep -rn "NewClient(" --include=*.go .` — currently `app/app.go`, handled in Task 4.)

- [ ] **Step 2: Write the failing test**

Add to `internal/httpclient/client_test.go`:

```go
func TestDoRequestLogsRequestAndResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	bus := logbus.New(50)
	c := NewClient(bus)
	_, err := c.DoRequest(context.Background(), &model.RequestInput{
		RequestID:      "r1",
		Method:         "GET",
		URL:            srv.URL,
		TimeoutSeconds: 5,
	})
	if err != nil {
		t.Fatalf("DoRequest error: %v", err)
	}

	var joined string
	for _, e := range bus.Snapshot() {
		joined += e.Level + " " + e.Message + "\n"
	}
	if !strings.Contains(joined, "→ GET") {
		t.Fatalf("missing request log line, got:\n%s", joined)
	}
	if !strings.Contains(joined, "← 200") {
		t.Fatalf("missing response log line, got:\n%s", joined)
	}
}
```

Ensure imports in the test file include `"context"`, `"net/http"`, `"net/http/httptest"`, `"strings"`, `"otester/internal/logbus"`, `"otester/internal/model"` (add any missing).

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/httpclient/ -run TestDoRequestLogsRequestAndResponse`
Expected: FAIL — `NewClient(bus)` too many arguments / missing log lines.

- [ ] **Step 4: Implement the constructor change and log points**

In `internal/httpclient/client.go`:

Add imports:
```go
"otester/internal/logbus"
"otester/internal/security"
```

Add field and constructor:
```go
type Client struct {
	client           *http.Client
	cancelFuncs      map[string]context.CancelFunc
	mu               sync.RWMutex
	allowInsecureTLS bool
	logger           logbus.Logger
}

func NewClient(logger logbus.Logger) *Client {
	if logger == nil {
		logger = logbus.Nop()
	}
	return &Client{
		client:      &http.Client{Timeout: 30 * time.Second},
		cancelFuncs: make(map[string]context.CancelFunc),
		logger:      logger,
	}
}
```

In `DoRequest`, immediately after `req, err := BuildRequest(...)` succeeds (before `start := time.Now()`), add:
```go
	c.logger.Infof("→ %s %s", req.Method, security.RedactURL(req.URL.String()))
	for k, v := range security.RedactHeaders(req.Header) {
		c.logger.Debugf("  header %s: %s", k, strings.Join(v, ", "))
	}
```
After the first successful `resp` (inside `if err == nil {` block, before `return ParseResponse(...)`), add:
```go
	c.logger.Infof("← %d %s (%d ms)", resp.StatusCode, resp.Status, duration.Milliseconds())
```
In the insecure-TLS retry branch, right after entering `if c.getAllowInsecureTLS() && isTLSError(err) {`, add:
```go
	c.logger.Warnf("TLS validation failed, retrying with insecure TLS: %s", err.Error())
```
And after a successful `retryResp` (before building `output`), add:
```go
	c.logger.Infof("← %d %s (%d ms, insecure TLS)", retryResp.StatusCode, retryResp.Status, (duration + retryDuration).Milliseconds())
```
At the two `handleRequestError(...)` return sites, precede each with:
```go
	c.logger.Errorf("request failed: %v", err) // use retryErr in the retry branch
```
(In the retry-failed branch log `retryErr`; in the final branch log `err`.)

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/httpclient/`
Expected: PASS (new test + all existing).

- [ ] **Step 6: Commit**

```bash
git add internal/httpclient/client.go internal/httpclient/client_test.go
git commit -m "feat(httpclient): log request/response/retry via logbus"
```

---

### Task 3: Inject logger into OAuth client and log the token flow

**Files:**
- Modify: `internal/oauth/microsoft.go`
- Test: `internal/oauth/microsoft_test.go` (create)

**Interfaces:**
- Consumes: `logbus.Logger`, `logbus.Nop`, `security.RedactURL`, `security.RedactBearerToken`.
- Produces:
  - `func NewMicrosoftOAuth(logger logbus.Logger) *MicrosoftOAuth` — **signature change** (was `NewMicrosoftOAuth()`).
  - `MicrosoftOAuth` gains field `logger logbus.Logger`.
  - Log lines emitted from `GetAccessToken` / `fetchToken`.

- [ ] **Step 1: Write the failing test**

Create `internal/oauth/microsoft_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/oauth/ -run TestGetAccessTokenLogsFetchAndSuccessWithoutSecret`
Expected: FAIL — `NewMicrosoftOAuth(bus)` too many arguments.

- [ ] **Step 3: Implement the constructor change and log points**

In `internal/oauth/microsoft.go`:

Add imports `"otester/internal/logbus"` and `"otester/internal/security"`.

```go
type MicrosoftOAuth struct {
	client  *http.Client
	cache   *TokenCache
	sfGroup singleflight.Group
	logger  logbus.Logger
}

func NewMicrosoftOAuth(logger logbus.Logger) *MicrosoftOAuth {
	if logger == nil {
		logger = logbus.Nop()
	}
	return &MicrosoftOAuth{
		client: &http.Client{Timeout: 30 * time.Second},
		cache:  NewTokenCache(),
		logger: logger,
	}
}
```

In `GetAccessToken`, after computing `cacheKey`/`refreshBefore`, replace the cache-hit block to log:
```go
	if token, ok := m.cache.Get(cacheKey, refreshBefore); ok {
		m.logger.Infof("token cache hit for client_id=%s scope=%s", profile.ClientID, profile.Scope)
		return token, true, nil
	}
	m.logger.Infof("token cache miss for client_id=%s scope=%s", profile.ClientID, profile.Scope)
```

In `fetchToken`, at the top (after building `form`, before `m.client.Do`), add:
```go
	m.logger.Infof("requesting token url=%s client_id=%s scope=%s",
		security.RedactURL(tokenURL), profile.ClientID, profile.Scope)
```
On the non-2xx branch, before returning the error:
```go
	m.logger.Errorf("token request failed: status %d", resp.StatusCode)
```
On decode error, before returning:
```go
	m.logger.Errorf("token response parse failed: %v", err)
```
After the successful decode and non-empty `AccessToken` check, before `return &tokenResp, nil`:
```go
	m.logger.Infof("token acquired expires_in=%ds token=%s",
		tokenResp.ExpiresIn, security.RedactBearerToken(tokenResp.AccessToken))
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/oauth/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/oauth/microsoft.go internal/oauth/microsoft_test.go
git commit -m "feat(oauth): log token cache/fetch/success without leaking secrets"
```

---

### Task 4: Add raw OAuth-profile loader (unmasked secret)

**Files:**
- Modify: `internal/config/loader.go`
- Test: `internal/config/loader_test.go` (add one test)

**Interfaces:**
- Consumes: existing `parse*` helpers.
- Produces:
  - `func LoadOAuthProfilesFromPath(configPath string) ([]OAuthProfile, error)` — returns profiles with the **real** `ClientSecret` (never sent to the frontend; used only for token fetching in `SendRequest`).

**Why:** `LoadConfigFromPath` returns `ConfigView` where `ClientSecretMasked = "******"`. Real token fetching needs the actual secret, so a backend-only loader is required.

- [ ] **Step 1: Write the failing test**

Add to `internal/config/loader_test.go`:

```go
func TestLoadOAuthProfilesFromPathReturnsRealSecret(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	json := `{"oauth_profiles":[{"id":"p1","client_id":"cid","client_secret":"real-secret","scope":"s","token_url":"https://t/","refreshBeforeExpirySeconds":30}]}`
	if err := os.WriteFile(path, []byte(json), 0o600); err != nil {
		t.Fatal(err)
	}

	profiles, err := LoadOAuthProfilesFromPath(path)
	if err != nil {
		t.Fatalf("LoadOAuthProfilesFromPath error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("want 1 profile, got %d", len(profiles))
	}
	p := profiles[0]
	if p.ID != "p1" || p.ClientID != "cid" || p.ClientSecret != "real-secret" ||
		p.Scope != "s" || p.TokenURL != "https://t/" || p.RefreshBeforeExpirySeconds != 30 {
		t.Fatalf("unexpected profile: %+v", p)
	}
}
```

Ensure test imports include `"os"`, `"path/filepath"` (add if missing).

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run TestLoadOAuthProfilesFromPathReturnsRealSecret`
Expected: FAIL — `undefined: LoadOAuthProfilesFromPath`.

- [ ] **Step 3: Implement**

Add to `internal/config/loader.go`:

```go
func LoadOAuthProfilesFromPath(configPath string) ([]OAuthProfile, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &ConfigError{Code: model.ErrConfigFileNotFound, Message: "config.json not found in " + configPath}
		}
		return nil, &ConfigError{Code: model.ErrConfigReadFailed, Message: err.Error()}
	}
	var rawCfg map[string]interface{}
	if err := json.Unmarshal(data, &rawCfg); err != nil {
		return nil, &ConfigError{Code: model.ErrConfigParseFailed, Message: err.Error()}
	}
	list, ok := rawCfg["oauth_profiles"].([]interface{})
	if !ok {
		return []OAuthProfile{}, nil
	}
	profiles := make([]OAuthProfile, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		profiles = append(profiles, OAuthProfile{
			ID:                         getString(m, "id"),
			Name:                       getString(m, "name"),
			Type:                       getString(m, "type"),
			OrgIDUUID:                  getString(m, "org_id_uuid"),
			ClientID:                   getString(m, "client_id"),
			ClientSecret:               getString(m, "client_secret"),
			Scope:                      getString(m, "scope"),
			TokenURL:                   getString(m, "token_url"),
			RefreshBeforeExpirySeconds: getInt(m, "refreshBeforeExpirySeconds", 60),
		})
	}
	return profiles, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/config/loader.go internal/config/loader_test.go
git commit -m "feat(config): load raw OAuth profiles with real client_secret"
```

---

### Task 5: Wire logbus + OAuth into `App`, add `GetLogs`/`ClearLogs` bindings

**Files:**
- Modify: `app/app.go`, `main.go`
- Test: `app/app_test.go` (add tests)

**Interfaces:**
- Consumes: `logbus.New`, `logbus.Nop`, `NewClient(logger)` (Task 2), `NewMicrosoftOAuth(logger)` (Task 3), `LoadOAuthProfilesFromPath` (Task 4), existing `model` error codes, `wailsruntime.EventsEmit`.
- Produces:
  - `App` gains fields: `logs *logbus.Bus`, and seams `getAccessToken func(context.Context, *config.OAuthProfile) (string, bool, error)`, `doRequest func(context.Context, *model.RequestInput) (*model.ResponseOutput, error)`, `loadOAuthProfiles func(string) ([]config.OAuthProfile, error)`.
  - `func (a *App) GetLogs() []logbus.Entry`
  - `func (a *App) ClearLogs()`
  - `SendRequest` now performs OAuth injection when `input.UseOAuth`.
  - Event name emitted: `"log:entry"` (payload: `logbus.Entry`).

- [ ] **Step 1: Write the failing tests**

Add to `app/app_test.go` (add imports `"context"`, `"strings"`, `"otester/internal/logbus"`, `"otester/internal/model"`):

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./app/`
Expected: FAIL — unknown fields `logs`,`getAccessToken`,`doRequest`,`loadOAuthProfiles`; undefined `GetLogs`/`ClearLogs`; `SendRequest` doesn't inject.

- [ ] **Step 3: Update `NewApp` and `App` struct**

In `app/app.go`, add imports `"otester/internal/logbus"` and (already present) `"otester/internal/model"`. Extend the struct and constructor:

```go
type App struct {
	ctx                context.Context
	oauth              *oauth.MicrosoftOAuth
	httpClient         *httpclient.Client
	logs               *logbus.Bus
	activeConfigPath   string
	loadConfigFromPath func(string) (*config.ConfigView, error)
	loadOAuthProfiles  func(string) ([]config.OAuthProfile, error)
	openFileDialog     func() (string, error)
	getAccessToken     func(context.Context, *config.OAuthProfile) (string, bool, error)
	doRequest          func(context.Context, *model.RequestInput) (*model.ResponseOutput, error)

	mu               sync.RWMutex
	allowInsecureTLS bool
}

func NewApp() *App {
	logs := logbus.New(2000)
	oauthClient := oauth.NewMicrosoftOAuth(logs)
	httpClient := httpclient.NewClient(logs)
	a := &App{
		oauth:              oauthClient,
		httpClient:         httpClient,
		logs:               logs,
		loadConfigFromPath: config.LoadConfigFromPath,
		loadOAuthProfiles:  config.LoadOAuthProfilesFromPath,
	}
	a.getAccessToken = oauthClient.GetAccessToken
	a.doRequest = httpClient.DoRequest
	a.openFileDialog = func() (string, error) {
		return wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
			Filters: []wailsruntime.FileFilter{{DisplayName: "JSON files", Pattern: "*.json"}},
		})
	}
	return a
}
```

- [ ] **Step 4: Emit events on Startup, add bindings**

In `Startup`, at the very top (after `a.ctx = ctx`), wire the emitter:
```go
	a.ctx = ctx
	a.logs.SetEmit(func(e logbus.Entry) {
		wailsruntime.EventsEmit(ctx, "log:entry", e)
	})
	a.logs.Infof("otester started")
```

Add the two binding methods (anywhere among the exported methods):
```go
func (a *App) GetLogs() []logbus.Entry {
	return a.logs.Snapshot()
}

func (a *App) ClearLogs() {
	a.logs.Clear()
}
```

- [ ] **Step 5: Rewrite `SendRequest` to inject OAuth and use the seams**

Replace the existing `SendRequest` body with:
```go
func (a *App) SendRequest(input *model.RequestInput) (*model.ResponseOutput, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app context not initialized")
	}

	usedOAuth := false
	fromCache := false
	if input.UseOAuth {
		profiles, err := a.loadOAuthProfiles(a.activeConfigPath)
		if err != nil {
			return &model.ResponseOutput{RequestID: input.RequestID, ErrorCode: string(model.ErrConfigReadFailed), ErrorMessage: err.Error()}, nil
		}
		var profile *config.OAuthProfile
		for i := range profiles {
			if profiles[i].ID == input.OAuthProfileID {
				profile = &profiles[i]
				break
			}
		}
		if profile == nil {
			a.logs.Errorf("oauth profile %q not found", input.OAuthProfileID)
			return &model.ResponseOutput{RequestID: input.RequestID, ErrorCode: string(model.ErrOAuthProfileNotFound), ErrorMessage: "oauth profile not found: " + input.OAuthProfileID}, nil
		}
		token, cached, err := a.getAccessToken(a.ctx, profile)
		if err != nil {
			return &model.ResponseOutput{RequestID: input.RequestID, ErrorCode: string(model.ErrOAuthTokenRequestFailed), ErrorMessage: err.Error()}, nil
		}
		input.Headers = append(input.Headers, model.KeyValue{Key: "Authorization", Value: "Bearer " + token, Enabled: true})
		usedOAuth = true
		fromCache = cached
	}

	out, err := a.doRequest(a.ctx, input)
	if err != nil {
		return nil, err
	}
	if out != nil {
		out.UsedOAuth = usedOAuth
		out.TokenFromCache = fromCache
	}
	return out, nil
}
```

Confirm `model.KeyValue` has fields `Key`,`Value`,`Enabled` (it does — used in `request_builder.go`). Leave `main.go`'s `app.NewApp()` call unchanged (signature is the same).

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./app/ ./internal/...`
Expected: PASS (all packages).

- [ ] **Step 7: Verify the whole backend builds**

Run: `go build ./...`
Expected: no output (success).

- [ ] **Step 8: Commit**

```bash
git add app/app.go main.go app/app_test.go
git commit -m "feat(app): wire logbus + OAuth into SendRequest; add GetLogs/ClearLogs"
```

---

### Task 6: Hand-add Wails JS bindings for `GetLogs`/`ClearLogs`

**Files:**
- Modify: `frontend/wailsjs/go/app/App.js`, `frontend/wailsjs/go/app/App.d.ts`

**Interfaces:**
- Consumes: Go methods from Task 5.
- Produces: JS/TS bindings `GetLogs()` and `ClearLogs()` importable from `../../wailsjs/go/app/App`.

**Note:** These files are normally regenerated by `wails build`/`wails generate module`. We hand-add them so the frontend + tests compile without a Wails toolchain in this environment. `logbus.Entry` isn't in the generated `models`, so type `GetLogs` return as a locally-defined shape via `any[]` in the `.d.ts` (the frontend store maps it to its own `LogEntry` type).

- [ ] **Step 1: Add to `App.js`**

Append:
```js
export function GetLogs() {
  return window['go']['app']['App']['GetLogs']();
}

export function ClearLogs() {
  return window['go']['app']['App']['ClearLogs']();
}
```

- [ ] **Step 2: Add to `App.d.ts`**

Append:
```ts
export function GetLogs():Promise<Array<{time:string;level:string;message:string}>>;

export function ClearLogs():Promise<void>;
```

- [ ] **Step 3: Commit**

```bash
git add frontend/wailsjs/go/app/App.js frontend/wailsjs/go/app/App.d.ts
git commit -m "chore(wailsjs): add GetLogs/ClearLogs bindings"
```

---

### Task 7: Frontend logs store

**Files:**
- Create: `frontend/src/stores/logs.ts`
- Modify: `frontend/src/types/index.ts` (add `LogEntry`)
- Test: `frontend/src/stores/logs.test.ts`

**Interfaces:**
- Consumes: nothing (pure store).
- Produces:
  - `type LogEntry = { time: string; level: string; message: string }` in `types/index.ts`.
  - `useLogsStore()` with reactive `entries: LogEntry[]`, `text` (computed joined string), and actions `append(e: LogEntry)`, `prime(list: LogEntry[])`, `clear()`. Client-side cap of 2000.

- [ ] **Step 1: Add the type**

In `frontend/src/types/index.ts`, add:
```ts
export type LogEntry = {
  time: string
  level: string
  message: string
}
```

- [ ] **Step 2: Write the failing test**

Create `frontend/src/stores/logs.test.ts`:
```ts
import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useLogsStore } from './logs'

describe('logs store', () => {
  beforeEach(() => setActivePinia(createPinia()))

  it('primes, appends, and joins text', () => {
    const s = useLogsStore()
    s.prime([{ time: '1', level: 'INFO', message: 'a' }])
    s.append({ time: '2', level: 'WARN', message: 'b' })
    expect(s.entries.length).toBe(2)
    expect(s.text).toContain('INFO')
    expect(s.text).toContain('a')
    expect(s.text).toContain('b')
  })

  it('clears', () => {
    const s = useLogsStore()
    s.append({ time: '1', level: 'INFO', message: 'a' })
    s.clear()
    expect(s.entries.length).toBe(0)
  })

  it('caps at 2000 entries', () => {
    const s = useLogsStore()
    for (let i = 0; i < 2100; i++) s.append({ time: String(i), level: 'INFO', message: 'm' })
    expect(s.entries.length).toBe(2000)
    expect(s.entries[0].time).toBe('100')
  })
})
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd frontend && npx vitest run src/stores/logs.test.ts`
Expected: FAIL — cannot resolve `./logs`.

- [ ] **Step 4: Implement the store**

Create `frontend/src/stores/logs.ts`:
```ts
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { LogEntry } from '../types'

const MAX_ENTRIES = 2000

export const useLogsStore = defineStore('logs', () => {
  const entries = ref<LogEntry[]>([])

  function cap() {
    if (entries.value.length > MAX_ENTRIES) {
      entries.value.splice(0, entries.value.length - MAX_ENTRIES)
    }
  }

  function append(e: LogEntry) {
    entries.value.push(e)
    cap()
  }

  function prime(list: LogEntry[]) {
    entries.value = list.slice(-MAX_ENTRIES)
  }

  function clear() {
    entries.value = []
  }

  const text = computed(() =>
    entries.value.map((e) => `${e.time} [${e.level}] ${e.message}`).join('\n'),
  )

  return { entries, text, append, prime, clear }
})
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd frontend && npx vitest run src/stores/logs.test.ts`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/stores/logs.ts frontend/src/stores/logs.test.ts frontend/src/types/index.ts
git commit -m "feat(frontend): logs Pinia store with append/prime/clear + cap"
```

---

### Task 8: LogsViewer modal component

**Files:**
- Create: `frontend/src/components/LogsViewer.vue`
- Test: `frontend/src/components/LogsViewer.test.ts`

**Interfaces:**
- Consumes: `useLogsStore` (Task 7), `ClearLogs` from `../../wailsjs/go/app/App` (Task 6), naive-ui `NModal`,`NInput`,`NButton`,`NSpace`.
- Produces: `LogsViewer.vue` — props `{ show: boolean }`, emits `update:show`; read-only textarea bound to `logsStore.text`, Clear + Copy buttons.

- [ ] **Step 1: Write the failing test**

Create `frontend/src/components/LogsViewer.test.ts`:
```ts
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import LogsViewer from './LogsViewer.vue'
import { useLogsStore } from '../stores/logs'

const { clearLogs } = vi.hoisted(() => ({ clearLogs: vi.fn(() => Promise.resolve()) }))
vi.mock('../../wailsjs/go/app/App', () => ({ ClearLogs: clearLogs }))

vi.mock('naive-ui', () => ({
  NModal: { name: 'NModal', props: ['show'], template: '<div><slot /></div>' },
  NInput: { name: 'NInput', props: ['value', 'readonly'], template: '<textarea :readonly="readonly" :value="value"></textarea>' },
  NButton: { name: 'NButton', emits: ['click'], template: '<button @click="$emit(\'click\')"><slot /></button>' },
  NSpace: { name: 'NSpace', template: '<div><slot /></div>' },
  useMessage: () => ({ success: vi.fn(), error: vi.fn() }),
}))

describe('LogsViewer', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders log text read-only', () => {
    const store = useLogsStore()
    store.append({ time: '1', level: 'INFO', message: 'hello-log' })
    const wrapper = mount(LogsViewer, { props: { show: true } })
    const ta = wrapper.find('textarea')
    expect(ta.attributes('readonly')).toBeDefined()
    expect(ta.element.getAttribute('value')).toContain('hello-log')
  })

  it('Clear button calls backend and empties the store', async () => {
    const store = useLogsStore()
    store.append({ time: '1', level: 'INFO', message: 'x' })
    const wrapper = mount(LogsViewer, { props: { show: true } })
    await wrapper.findAll('button')[0].trigger('click')
    expect(clearLogs).toHaveBeenCalled()
    expect(store.entries.length).toBe(0)
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd frontend && npx vitest run src/components/LogsViewer.test.ts`
Expected: FAIL — cannot resolve `./LogsViewer.vue`.

- [ ] **Step 3: Implement the component**

Create `frontend/src/components/LogsViewer.vue`:
```vue
<template>
  <n-modal
    :show="show"
    preset="card"
    title="Logs"
    class="logs-modal"
    :style="{ width: '80vw', maxWidth: '1100px' }"
    :bordered="false"
    size="huge"
    @update:show="(v: boolean) => emit('update:show', v)"
  >
    <div class="logs-body">
      <n-input
        :value="logsStore.text"
        type="textarea"
        class="logs-textarea"
        readonly
        placeholder="HTTP and auth activity appears here…"
        :autosize="false"
      />
      <n-space class="logs-actions">
        <n-button size="small" @click="onClear">Clear</n-button>
        <n-button size="small" @click="onCopy">Copy</n-button>
      </n-space>
    </div>
  </n-modal>
</template>

<script setup lang="ts">
import { NModal, NInput, NButton, NSpace, useMessage } from 'naive-ui'
import { useLogsStore } from '../stores/logs'
import { ClearLogs } from '../../wailsjs/go/app/App'

defineProps<{ show: boolean }>()
const emit = defineEmits<{ (e: 'update:show', value: boolean): void }>()

const logsStore = useLogsStore()
const message = useMessage()

async function onClear() {
  try {
    await ClearLogs()
  } catch {
    /* backend may be unavailable in dev; still clear the view */
  }
  logsStore.clear()
}

async function onCopy() {
  try {
    await navigator.clipboard.writeText(logsStore.text)
    message.success('Logs copied')
  } catch {
    message.error('Copy failed')
  }
}
</script>

<style scoped>
.logs-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 60vh;
}

.logs-textarea {
  flex: 1;
}

.logs-textarea :deep(.n-input__textarea-el) {
  height: 100%;
  resize: none;
  font-family: 'Consolas', 'Menlo', monospace;
  font-size: 12px;
}

.logs-actions {
  justify-content: flex-end;
}
</style>
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd frontend && npx vitest run src/components/LogsViewer.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/LogsViewer.vue frontend/src/components/LogsViewer.test.ts
git commit -m "feat(frontend): read-only LogsViewer modal with Clear/Copy"
```

---

### Task 9: Subscribe to `log:entry` in App.vue

**Files:**
- Modify: `frontend/src/App.vue`

**Interfaces:**
- Consumes: `useLogsStore` (Task 7), `GetLogs` (Task 6), `EventsOn` from `../wailsjs/runtime/runtime`.
- Produces: live log accumulation regardless of whether the window is open.

- [ ] **Step 1: Wire subscription in `onMounted`**

Edit `frontend/src/App.vue` script:
```ts
import { onMounted } from 'vue'
import { NConfigProvider, NMessageProvider, NDialogProvider } from 'naive-ui'
import MainView from './views/MainView.vue'
import { useThemeStore } from './stores/theme'
import { useConfigStore } from './stores/config'
import { useLogsStore } from './stores/logs'
import { GetLogs } from '../wailsjs/go/app/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import type { LogEntry } from './types'

const themeStore = useThemeStore()
const configStore = useConfigStore()
const logsStore = useLogsStore()

onMounted(async () => {
  themeStore.init()
  EventsOn('log:entry', (e: LogEntry) => logsStore.append(e))
  try {
    const existing = await GetLogs()
    if (existing) logsStore.prime(existing as LogEntry[])
  } catch {
    /* backend not ready in dev/test; ignore */
  }
  await configStore.loadConfig()
})
```

- [ ] **Step 2: Verify the frontend still builds/tests**

Run: `cd frontend && npx vitest run`
Expected: PASS (all existing + new tests). If `App.vue` has no dedicated test, this confirms no regressions.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/App.vue
git commit -m "feat(frontend): subscribe to log:entry and prime logs on mount"
```

---

### Task 10: Add "Open Logs" to Tools menu + mount LogsViewer

**Files:**
- Modify: `frontend/src/views/MainView.vue`
- Test: `frontend/src/views/MainView.test.ts` (add one test)

**Interfaces:**
- Consumes: `LogsViewer` (Task 8).
- Produces: Tools → Open Logs opens the modal.

- [ ] **Step 1: Write the failing test**

Add to `frontend/src/views/MainView.test.ts`. First extend the naive-ui mock so `NModal`, `NInput`, `NButton`, `NSpace`, and `useMessage` exist (the LogsViewer child needs them). Update the existing `vi.mock('naive-ui', …)` block to:
```ts
vi.mock('naive-ui', () => ({
  NDropdown: {
    name: 'NDropdown', props: ['options'], emits: ['select'],
    template: '<div><slot /></div>',
  },
  NModal: { name: 'NModal', props: ['show'], template: '<div><slot /></div>' },
  NInput: { name: 'NInput', props: ['value'], template: '<textarea :value="value"></textarea>' },
  NButton: { name: 'NButton', template: '<button><slot /></button>' },
  NSpace: { name: 'NSpace', template: '<div><slot /></div>' },
  useMessage: () => message,
}))
```
Also extend the App mock so `ClearLogs`/`GetLogs` are defined:
```ts
vi.mock('../../wailsjs/go/app/App', () => ({
  OpenConfigFile: openConfigFile,
  ClearLogs: vi.fn(() => Promise.resolve()),
  GetLogs: vi.fn(() => Promise.resolve([])),
}))
```
Then add the test:
```ts
it('Tools menu contains an Open Logs item', () => {
  const wrapper = mount(MainView, { global: { plugins: [createPinia()] } })
  const dropdowns = wrapper.findAllComponents({ name: 'NDropdown' })
  const tools = dropdowns.find((d) =>
    (d.props('options') as Array<{ key: string }>).some((o) => o.key === 'clear-tokens'),
  )
  expect(tools).toBeTruthy()
  const keys = (tools!.props('options') as Array<{ key: string }>).map((o) => o.key)
  expect(keys).toContain('open-logs')
})
```
(If the existing tests mount `MainView` without a Pinia plugin, keep their setup; this new test provides its own `createPinia()` because `LogsViewer` uses the logs store. Ensure `createPinia` is imported — it already is in this test file.)

- [ ] **Step 2: Run test to verify it fails**

Run: `cd frontend && npx vitest run src/views/MainView.test.ts`
Expected: FAIL — `open-logs` not in Tools options.

- [ ] **Step 3: Implement the menu item + modal**

In `frontend/src/views/MainView.vue`:

Add the import near the other component imports:
```ts
import LogsViewer from '../components/LogsViewer.vue'
```
Add a ref beside `showBase64`:
```ts
const showLogs = ref(false)
```
Add the menu item to `toolsMenuOptions`:
```ts
const toolsMenuOptions = [
  { label: 'base64 Convert', key: 'base64-convert' },
  { label: 'Validate Config', key: 'validate-config' },
  { label: 'Clear Token Cache', key: 'clear-tokens' },
  { label: 'Open Logs', key: 'open-logs' },
]
```
Add a case in `handleMenuSelect`:
```ts
    case 'open-logs':
      showLogs.value = true
      break
```
Mount the component in the template, next to `<Base64Convert … />`:
```html
    <Base64Convert v-model:show="showBase64" />
    <LogsViewer v-model:show="showLogs" />
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd frontend && npx vitest run src/views/MainView.test.ts`
Expected: PASS (new test + existing MainView tests).

- [ ] **Step 5: Full frontend test sweep**

Run: `cd frontend && npx vitest run`
Expected: PASS (whole suite).

- [ ] **Step 6: Commit**

```bash
git add frontend/src/views/MainView.vue frontend/src/views/MainView.test.ts
git commit -m "feat(frontend): add Tools > Open Logs and mount LogsViewer"
```

---

### Task 11: Full build + manual verification

**Files:** none (verification only).

- [ ] **Step 1: Backend build + tests**

Run: `go build ./... && go test ./...`
Expected: all PASS.

- [ ] **Step 2: Frontend build + tests**

Run: `cd frontend && npm run build && npx vitest run`
Expected: build succeeds, all tests PASS.

- [ ] **Step 3: Manual smoke (documented, requires Wails/desktop)**

Follow the `run` skill / `wails dev` to launch, then verify:
1. Tools menu shows **Open Logs**; clicking it opens the modal.
2. The textarea is read-only (typing does nothing).
3. Send a request → a `→ METHOD …` and `← STATUS …` line appear.
4. For an OAuth endpoint, `token cache miss`/`requesting token`/`token acquired` lines appear and **no** `client_secret` or full token text is visible.
5. Close and reopen the window → prior lines are still present (buffer/prime works).
6. Clear empties the view; Copy copies the text.

- [ ] **Step 4: Final commit (if any doc updates)**

```bash
git add -A
git commit -m "chore: verify HTTP/auth log window end-to-end" || echo "nothing to commit"
```

---

## Self-Review

**Spec coverage:**
- Backend log bus / ring buffer / emit / redaction → Tasks 1–3 ✓
- OAuth wired into `SendRequest` + Bearer injection + `UsedOAuth`/`TokenFromCache` → Task 5 ✓
- Raw secret loader (gap discovered vs. spec) → Task 4 ✓ (spec said "resolve profile"; masked secret made a new loader necessary — documented in Task 4)
- `GetLogs`/`ClearLogs` bindings + `log:entry` event → Tasks 5, 6 ✓
- Frontend store primed + live → Tasks 7, 9 ✓
- `LogsViewer` read-only modal with Clear/Copy → Task 8 ✓
- Tools → Open Logs + mount → Task 10 ✓
- Testing (Go + Vitest) → included per task; sweep in Task 11 ✓
- Out-of-scope items (filtering/search/export/native window) → not implemented ✓

**Placeholder scan:** No TBD/TODO/"handle edge cases"; all code blocks are concrete.

**Type consistency:** `logbus.Entry{Time,Level,Message}` matches the `.d.ts` shape and the frontend `LogEntry` type. `NewClient(logger)` / `NewMicrosoftOAuth(logger)` signatures are consistent across Tasks 2, 3, 5. Seam field names (`getAccessToken`, `doRequest`, `loadOAuthProfiles`) are consistent between the struct definition and `SendRequest`/tests in Task 5. `LoadOAuthProfilesFromPath` return type `[]config.OAuthProfile` matches the seam and usage.
