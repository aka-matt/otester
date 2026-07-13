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

