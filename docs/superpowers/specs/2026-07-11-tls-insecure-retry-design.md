# TLS Insecure Retry Design

Extend the existing Certificate tab so that when the target site has a problematic certificate (expired, self-signed, hostname mismatch, untrusted CA, etc.) AND `config.json` has `allow_insecure_tls: true`, the request still completes and the Certificate tab surfaces the original validation error alongside the server's certificate chain. The existing config flag — already declared in the schema and called out in CLAUDE.md §24 item 7 — is wired up to actually take effect; today it is silently ignored.

This is a follow-up to `2026-07-11-tls-certificate-tab-design.md`, which established the `TLSInfo` data model and the Certificate tab itself. This spec only extends that contract.

## Goals & non-goals

**Goals**
- A request to a server with a bad TLS certificate succeeds (status code, headers, body all returned) when `allow_insecure_tls: true`.
- The Certificate tab makes the insecure fallback visible: a warning banner naming the original error, plus the cert chain that the server actually presented (captured from the retry that succeeded).
- When `allow_insecure_tls: false` (the default), behaviour is unchanged: a TLS validation error returns `status: "handshake_failed"` and no response body.
- A config reload (via `ReloadConfig` / `OpenConfigFile`) updates the flag without restarting the app.

**Non-goals**
- No per-request or per-endpoint toggle. The flag is global.
- No UI control to flip the flag from the running app. The flag lives in `config.json` only.
- No new error codes. The retry is transparent: the response to the user is a normal 200/4xx/5xx from the server, with the cert problem disclosed in the TLS panel.
- No proxy, mTLS, or client-cert changes.
- No changes to the OAuth path — token requests go through the same `httpclient.DoRequest` so they inherit the same retry automatically; that is the desired behaviour.

## Architecture

Three layers, each with one responsibility.

**Model (`internal/model/`)** — extend the contract:
- `TLSInfo` gains two new fields: `ValidationSkipped bool` and `OriginalError string`. `Status` retains its current string values; `"ok"` is reused for both a clean connection and a retry-succeeded connection (distinguished by the new fields).

**Backend (`internal/httpclient/` + `app/`)** — wire the retry:
- `httpclient.Client` gains `allowInsecureTLS bool` (RWMutex-protected) and a setter `SetAllowInsecureTLS(bool)`.
- `httpclient.NewClient()` returns a client with the flag off by default.
- `httpclient.DoRequest` tries the normal shared `*http.Client`. If `Do` errors AND the error message contains `tls:` or `x509:` AND the flag is on, retry once using a per-call `http.Client` whose `Transport.TLSClientConfig` is `&tls.Config{InsecureSkipVerify: true}`. On retry success, the returned `TLSInfo` is built from the retry's `resp.TLS` with `ValidationSkipped=true` and `OriginalError=<first attempt message>`. On retry failure, build `TLSInfo` with `Status="handshake_failed"` using the second error.
- `app.App` holds an `allowInsecureTLS bool` (mutex-protected) and calls `a.httpClient.SetAllowInsecureTLS` from `LoadConfig` / `OpenConfigFile` / `ReloadConfig` whenever the active config changes. The `SendRequest` signature is unchanged.

**Frontend (`frontend/src/`)** — surface the warning:
- `src/types/index.ts` mirrors the two new fields on `TLSInfo`.
- `components/CertificateViewer.vue` renders a new warning banner above the existing `ok` branch when `info.validationSkipped === true`. The banner uses `--warning-color` (amber) — not `--danger-color` (red), because the request itself succeeded. Below it, the full cert chain and connection block render exactly as today.

Data flow per request:

```
RequestEditor → app.SendRequest → httpclient.DoRequest
                                           ↓
                              c.client.Do(req)  ← first attempt, validates cert
                                           ↓
                          ┌──────────────┴───────────────┐
                       no error                       error w/ "tls:"|"x509:"
                          ↓                                  ↓
                  BuildTLSInfo(resp)         allow_insecure_tls == true?
                                                       ↓ yes          ↓ no
                                       client{InsecureSkipVerify}.Do(req)
                                                ↓                ↓
                                         retry resp         handleRequestError
                                                ↓                ↓
                                  BuildTLSInfo with          TLSInfo
                                  ValidationSkipped=true,    status="handshake_failed"
                                  OriginalError=<1st err>
                                                ↓
                                       ResponseOutput (json)
                                                ↓
                                     ResponseStore.setResponse
                                                ↓
                                ResponseViewer → CertificateViewer
                                                ↓
                                  banner (if ValidationSkipped) +
                                  existing cert chain & connection block
```

The retry uses a fresh `http.Client` rather than mutating the shared client's transport. Mutating shared state mid-request risks racing with other in-flight requests; constructing a small `http.Client` per retry is local and bounded.

## Detailed behaviour

### TLSInfo status values (after this change)

| Status | Meaning | When |
|---|---|---|
| `no_tls_attempted` | URL was `http://` or the request never reached TLS | URL scheme not `https`, or context cancelled before TLS |
| `handshake_failed` | TLS handshake or validation failed and was not retried (either no flag, or the retry also failed) | `Do` returns an error matching the regex, flag is off OR retry also fails |
| `ok` | TLS handshake and validation succeeded — or succeeded on retry after a first-attempt failure | Clean path, **or** retry path |

The two new fields on `ok` distinguish those cases:
- `ValidationSkipped=false, OriginalError=""` → clean TLS.
- `ValidationSkipped=true, OriginalError="<msg>"` → retry succeeded; the first attempt failed with `<msg>`.

### App-level plumbing

`app.App.Startup` already calls `LoadConfig`. After this change:
- `LoadConfig` and `OpenConfigFile` and `ReloadConfig` (when added) all update `a.allowInsecureTLS` from `cfg.App.AllowInsecureTLS` and call `a.httpClient.SetAllowInsecureTLS(...)`.
- The mutex is a `sync.RWMutex` on `App`, mirroring the one on `httpclient.Client`.

A user can therefore reload the active config file at runtime (future feature, out of scope here) and the flag flips without restart. For now the plumbing is in place but the UI trigger for reload is not added by this spec.

### Frontend banner

When `info.validationSkipped === true`, render above the existing `ok` content:

```
┌──────────────────────────────────────────────────────────┐
│ ⚠ TLS certificate validation was skipped                │
│   (allow_insecure_tls: true)                             │
│   Original error: x509: certificate signed by unknown …  │
└──────────────────────────────────────────────────────────┘
[ TLS Connection block — unchanged ]
[ Certificate Chain (n) — unchanged ]
```

Style:
- Border + tinted background using `--warning-color` (matches the existing amber palette already used for the truncation warning and the `Expires soon` pill).
- Text uses `var(--text-primary)` for the heading and `var(--text-secondary)` for the meta lines.
- Banner is only rendered inside the `ok` branch (status `"ok"`). It is not added to `handshake_failed` (where the existing red error panel already conveys the issue) or `no_tls_attempted`.

### Error classification — what triggers a retry?

`tls:` and `x509:` substring match on the error message. This catches:
- `x509: certificate signed by unknown authority`
- `x509: certificate has expired or is not yet valid`
- `x509: hostname does not match certificate`
- `tls: failed to verify certificate: …`
- `tls: handshake failure`

It does **not** catch:
- `context deadline exceeded` / `context canceled` — these are timeouts/cancellations, not cert errors.
- TCP-level failures (`connection refused`, `no such host`) — the request never reached TLS.
- HTTP-level errors with status 4xx/5xx — those are returned normally without ever producing an error from `Do`.

This matches the existing classification logic in `BuildTLSInfoFromError`, which is the source of truth for "was this a TLS error?".

## Testing

### Backend (`internal/httpclient/`)

New cases in `client_test.go`, using `httptest.NewTLSServer` (self-signed cert):

1. **Flag off, server self-signed** → `DoRequest` returns a `ResponseOutput` with `TLS.Status == "handshake_failed"`, `TLS.ValidationSkipped == false`, `ErrorCode` set, body empty. Mirrors today's behaviour.
2. **Flag on, server self-signed** → `DoRequest` returns a `ResponseOutput` with `TLS.Status == "ok"`, `TLS.ValidationSkipped == true`, `TLS.OriginalError` non-empty and contains `x509:`, the response has a populated body and a non-zero `StatusCode`, and `TLS.Certificates` has at least one entry (the self-signed server cert).
3. **Flag on, server with a trusted cert** — `httptest.NewTLSServer` whose certificate the test client transport trusts (configure `RootCAs` to a pool containing the server's leaf, so the first attempt succeeds) → `TLS.Status == "ok"`, `TLS.ValidationSkipped == false`, `TLS.OriginalError == ""`. Confirms the retry is not taken when the first attempt succeeds.
4. **Flag on, server unreachable** (TCP `connection refused`) → first attempt errors with a non-TLS message; no retry is attempted; `TLS.Status == "handshake_failed"` is not appropriate here — fall through to the existing `no_tls_attempted` / `connection_failed` path. Verify that the retry is not taken.

A new helper test in `tls_test.go` for `BuildTLSInfo` with the new parameters — assert that `ValidationSkipped=true` produces a `TLSInfo` whose `Certificates` come from the provided `resp.TLS.PeerCertificates` and whose `OriginalError` matches the supplied string.

Existing tests in `internal/model/` get one new case for the new fields (default zero values round-trip correctly through JSON).

### Frontend (`frontend/src/components/`)

`CertificateViewer.test.ts` gains two cases:
1. Render an info with `status: "ok"`, `validationSkipped: true`, `originalError: "x509: …"` → the warning banner is visible, and the cert chain still renders.
2. Render an info with `status: "ok"`, `validationSkipped: false`, no `originalError` → no banner; cert chain renders unchanged.

Existing 13 cases (chain rendering, validity pills, copy buttons) keep passing without modification — the banner is purely additive and lives above the existing `ok` branch.

### Wails bridge

`frontend/wailsjs/go/models.ts` regenerates automatically when `wails build` or `wails dev` next runs. The hand-written mirror in `src/types/index.ts` is updated by hand and must match. A type-check (`npx vue-tsc --noEmit`) is the contract test.

## Files touched

- `internal/model/types.go` — add `ValidationSkipped bool` and `OriginalError string` to `TLSInfo`.
- `internal/model/types_test.go` — JSON round-trip for new fields.
- `internal/httpclient/client.go` — flag field + setter + retry branch in `DoRequest`.
- `internal/httpclient/client_test.go` — four new cases (above).
- `internal/httpclient/tls.go` — extend `BuildTLSInfo` signature with `validationSkipped bool` + `originalError string`; populate them on the returned `TLSInfo`.
- `internal/httpclient/tls_test.go` — one new case for the new signature.
- `app/app.go` — track flag, call setter on config load/open/reload.
- `frontend/src/types/index.ts` — mirror new fields on `TLSInfo`.
- `frontend/src/components/CertificateViewer.vue` — render the banner inside the `ok` branch.
- `frontend/src/components/CertificateViewer.test.ts` — two new cases.
- (auto) `frontend/wailsjs/go/models.ts` — regenerated by Wails.