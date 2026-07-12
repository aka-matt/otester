Task 1: complete (commits 303ebc6..a37e3e2, review clean)
Task 2: complete (commits a37e3e2..9577471, review clean)

## TLS Certificate Tab Branch (current)

Step4 branch starting commit: f654f45 (docs: add TLS certificate tab design)
Plan: docs/superpowers/plans/2026-07-11-tls-certificate-tab.md

Task 1: complete (commits f654f45..565e8db, review clean)
Task 2: complete (commits 565e8db..da55ce2, review clean)
  - Deviations: KeyUsageNonRepudiation→ContentCommitment (Go 1.26 rename); tlsVersionString default="unknown" (followed test over impl); daysUntilExpiry uses math.Round (test precision)
  - Minor ledger: tls.go + tls_test.go lack trailing newline; flagged for final review
Task 3: complete (commits da55ce2..337691e, review clean)
  - Deviation: trimmed crypto/x509/pkix + encoding/base64 from tls.go imports (brief listed unused imports; placeholder was deleted)
Task 4: complete (commits 337691e..1441845, review clean)
  - Minor ledger: tls.go + tls_test.go still lack trailing newline (carry-over from Task 2)
  - Minor ledger: formatFingerprint param name 'der' is misleading when called with hash digests (carry-over from Task 2)
Task 5: complete (commits 1441845..73d844, review clean)
  - Minor ledger: trailing-newline nit continues
  - Minor ledger: BuildTLSInfo leaves Certificates/Connection nil; BuildTLSInfoFromError uses explicit empty slice (semantically equivalent, slight inconsistency)
  - Minor ledger: hardcoded "URL scheme is not HTTPS" error message would be misleading for HTTPS→HTTP downgrade redirect (out of scope per spec)
Task 6: complete (commits 73d844..82e33f0, review clean)
  - Deviation: implementer updated 3 TestHandleRequestError_* tests for new signature (brief Step 3 didn't mention them); passed nil for req
  - Minor ledger: handleRequestError default branch has slightly redundant if/else
Task 7: complete (commits 82e33f0..0ca04f8)
  - Orchestrator performed the work directly: Haiku subagent failed silently on two consecutive dispatches. The task was a simple type append (verified by vue-tsc), so no reviewer was dispatched.
Task 8: complete (commits 0ca04f8..8339e38, review clean)
  - Deviations: vi.mock('naive-ui') to stub useMessage (matches MainView.test.ts); button labels disambiguated to Copy SHA-256/Copy SHA-1/Copy Public Key; rollup dep installed for vitest
  - Minor ledger: CertificateViewer.test.ts:3 imports nextTick unused (carry-over from brief)
Task 9: complete (commits 8339e38..f3f601f, review found 1 High + 1 Low)
  - Deviation: added display-directive="show" to Certificate pane (brief's tests need panel mounted at mount time)
  - Deviation: added vi.mock('naive-ui') to ResponseViewer.test.ts (CertificateViewer uses useMessage)
  - HIGH fix: ResponseOutput interface missing tls?: TLSInfo field — orchestrator added directly (f3f601f)
  - Low: no-TLS test only reasserts tab title (could be tighter, but matches brief's exact assertions)

## Final state
- All 10 tasks complete (9 implementation + 1 verification)
- 13 commits on step4 branch since f654f45 (design spec)
- All Go tests green (`go test ./...`)
- Frontend build green (`npm run build`)
- All new frontend tests pass (16/16: CertificateViewer 13 + ResponseViewer 3)
- Pre-existing EndpointList.test.ts failures confirmed unrelated
- Whole-branch review verdict: APPROVED FOR MERGE

## TLS Insecure Retry Branch (current)
Step6 branch starting commit: 1f1bf4d (docs: add TLS insecure-retry design spec)
Plan: docs/superpowers/plans/2026-07-11-tls-insecure-retry.md

Task 1: complete (commits 1f1bf4d..7d9e5d1, review clean; deviation: ValidationSkipped has omitempty to satisfy the omit test)
Task 2: complete (commits 7d9e5d1..74df955, review clean)
Task 3: complete (commits 74df955..591c720, review clean; fix commit excludes context timeouts from isTLSError)
Task 4: complete (commits 591c720..0c08776, review clean; minor: CRLF churn, dead-code observation on App.allowInsecureTLS mirror — both non-blocking)
Task 5: complete (commits 0c08776..9588e2a, review clean)
Task 6: complete (commits 9588e2a..691b8a9, review clean; concerns: pre-existing EndpointList failures + WSL rollup dep, both environment-only)
Task 7: complete (final verification: 6 backend packages OK; frontend 21/21 tests pass; vue-tsc clean via npm run build; go vet clean; 8 commits on step5 since 1f1bf4d)

### Whole-branch review fixes (Important findings)

- **Commit hash:** `10ec38a55203d5d8d1c20a2c7f58972072d1c3e5`
- **Commit message:** `fix(httpclient): preserve POST body and add timeout to insecure-TLS retry`
- **One-line summary:** Two changes — (1) the TLS retry now re-populates `req.Body` from `GetBody()` before `retryClient.Do`, fixing silently dropped POST/PUT/PATCH bodies on retry; (2) `InsecureTLSClient(timeout time.Duration)` sets `Timeout` so a hung insecure connection cannot outlive the request's deadline.

### Tests run

- `go test ./internal/httpclient/ -v` — all PASS, including new `TestClient_DoRequest_PostBody_PreservedAcrossTLSRetry`. Targeted retry/POST tests:
  - `TestClient_DoRequest_TLSHandshakeFailure` PASS
  - `TestClient_DoRequest_PostWithBody` PASS
  - `TestClient_DoRequest_TLSHandshakeFailure_FlagOn` PASS
  - `TestClient_DoRequest_TLSOK_FlagOn_NoRetry` PASS
  - `TestClient_DoRequest_NonTLSError_NoRetry` PASS
  - `TestClient_DoRequest_PostBody_PreservedAcrossTLSRetry` PASS (the new test; server captures the body via `io.ReadAll(r.Body)` and asserts it equals the input)
- `go vet ./...` — clean (no output).
- `go build ./...` — clean (no output).
- Final test status: `PASS` — `ok otester/internal/httpclient 5.613s`.

### Change details

- `internal/httpclient/client.go`: the retry block now clones the request, then if `retryReq.GetBody != nil` calls `retryReq.GetBody()` and assigns the result to `retryReq.Body` before `retryClient.Do`. The retry client is constructed via `InsecureTLSClient(time.Duration(input.TimeoutSeconds) * time.Second)`.
- `internal/httpclient/tls.go`: `InsecureTLSClient(timeout time.Duration)` now sets `Timeout: timeout` on the returned `*http.Client` to match `NewClient`'s policy.
- `internal/httpclient/client_test.go`: added `TestClient_DoRequest_PostBody_PreservedAcrossTLSRetry` — POSTs `{"name":"retry-body"}` JSON to an `httptest.NewTLSServer` reached via `127.0.0.1` (cert is valid for "example.com" so the first attempt fails x509 hostname validation). With `SetAllowInsecureTLS(true)`, the retry must succeed AND the server-side handler must observe the original body (locked via the handler storing `r.Method` and `io.ReadAll(r.Body)` under a mutex for the test to assert).
