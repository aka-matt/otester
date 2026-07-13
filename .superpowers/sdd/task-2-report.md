# Task 2 Report: Inject logger into HTTP client and log request/response/retry

## What was built

`internal/httpclient/client.go` now accepts a `logbus.Logger` via `NewClient(logger logbus.Logger) *Client` and emits redacted log lines from `DoRequest` at every meaningful transition:

- After `BuildRequest` succeeds: `c.logger.Infof("→ %s %s", req.Method, security.RedactURL(req.URL.String()))` plus a `Debugf` per header via `security.RedactHeaders`.
- After the first successful response: `c.logger.Infof("← %d %s (%d ms)", resp.StatusCode, resp.Status, duration.Milliseconds())`.
- Inside the insecure-TLS retry branch on TLS failure: `c.logger.Warnf("TLS validation failed, retrying with insecure TLS: %s", err.Error())`.
- After a successful insecure-TLS retry: `c.logger.Infof("← %d %s (%d ms, insecure TLS)", retryResp.StatusCode, retryResp.Status, totalDuration.Milliseconds())`.
- At each `handleRequestError` return site: `c.logger.Errorf("request failed: %v", err)` — `retryErr` on the retry-failure branch, `err` on the final-failure branch.

The `Client` struct gains a private `logger logbus.Logger` field. The constructor defends against nil by substituting `logbus.Nop()`. URL and header values are passed through `security.RedactURL` and `security.RedactHeaders`, so `client_secret` / `access_token` query parameters and `Authorization: Bearer <token>` headers never appear in logs.

The `app/app.go` call site for `httpclient.NewClient()` was updated transitively to `httpclient.NewClient(logbus.Nop())` so `go build ./...` stays green until Task 5 wires the real logger in. Every existing test in `internal/httpclient/client_test.go` was updated to `NewClient(logbus.Nop())`, plus one new test (`TestDoRequestLogsRequestAndResponse`) was added per the brief.

## TDD Evidence

### RED — failing test before implementation

Command: `go test ./internal/httpclient/ -run TestDoRequestLogsRequestAndResponse`

```
internal/httpclient/client_test.go:22:22: too many arguments in call to NewClient
        have (logbus.Logger)
        want ()
internal/httpclient/client_test.go:38:22: too many arguments in call to NewClient
... (14 sites total)
FAIL    otester/internal/httpclient [build failed]
FAIL
```

The build refused to compile because `NewClient` did not yet accept the new `logbus.Logger` argument — exactly the failure mode the brief predicted.

### GREEN — implementation and full test suite pass

Commands and results after implementation:

- `go test ./internal/httpclient/ -run TestDoRequestLogsRequestAndResponse -v`
  ```
  === RUN   TestDoRequestLogsRequestAndResponse
  --- PASS: TestDoRequestLogsRequestAndResponse (0.00s)
  PASS
  ok      otester/internal/httpclient   0.009s
  ```

- `go test ./internal/httpclient/ -v` (last 80 lines shown)
  ```
  ... (all 30+ tests pass)
  --- PASS: TestIsTLSError (0.00s)
  PASS
  ok      otester/internal/httpclient   5.776s
  ```

- `go test ./...`
  ```
  ?       otester  [no test files]
  ok      otester/app             0.006s
  ok      otester/internal/config 0.029s
  ok      otester/internal/httpclient 5.671s
  ok      otester/internal/logbus 0.005s
  ok      otester/internal/model  (cached)
  ok      otester/internal/oauth  (cached)
  ok      otester/internal/security (cached)
  ```

- `go build ./...` — silent (success).
- `go vet ./internal/httpclient/` — silent (success).

## Files changed

```
 app/app.go                         |   3 +-
 internal/httpclient/client.go      |  23 +-
 internal/httpclient/client_test.go |  60 +-
```

- `app/app.go` — transitional call-site fix: import `logbus` and pass `logbus.Nop()` so the build stays green. Task 5 will swap this for the real `*logbus.Bus`.
- `internal/httpclient/client.go` — added `logbus` and `security` imports, added `logger logbus.Logger` field, rewrote `NewClient` to accept the logger (with nil → `Nop()` fallback), and added log points in `DoRequest` as described above.
- `internal/httpclient/client_test.go` — added `logbus` import, updated every `NewClient()` call to `NewClient(logbus.Nop())`, and added `TestDoRequestLogsRequestAndResponse` per the brief.

## Commit

`3f92936 feat(httpclient): log request/response/retry via logbus`

Only the three files changed for this task were staged; pre-existing line-ending churn on `app/app_test.go`, the `config/...` Go files, `internal/config/loader.go`, `frontend/*`, `schemas/config.schema.json`, `wails.json`, etc., was deliberately left unstaged per the brief.

## Self-review

- **Completeness** — every log point in the brief is implemented in the exact location it specifies (request line + per-header debug lines after `BuildRequest`, response line in the success branch, TLS-failed warn at the top of the retry branch, insecure-TLS response line after the retry succeeds, and the two `Errorf` log lines at the two `handleRequestError` return sites). No extra log points added.
- **Quality** — single defensive check at construction (`if logger == nil`) mirrors the test-side discipline. URL/header redaction always goes through the `security` package; no raw `Authorization` value can leak. Test output is pristine (`PASS` only).
- **YAGNI** — no new fields, no new methods, no new helpers. No `fmt.Sprintf` shims, no extra struct copy, no goroutine wiring. The brief's log lines are emitted in the precise order it dictates, so no premature abstraction around them.
- **Concerns / Notes**
  - The headers loop emits one `Debugf` per header key, which (per the brief) is correct. If `Accept-Encoding` and friends cause noisy debug lines, the user can disable that level in the UI later. Not changed here.
  - `app/app.go` is updated transitively with `logbus.Nop()` as the brief preferred. Task 5 will replace this with a real `*logbus.Bus` instance. The change is one line of diff and is reversible.
  - Working tree contained pre-existing CRLF/LF churn on many files (frontend, config loader, app_test, etc.). Per the brief, those were left untouched and only the three in-scope files were staged.
