### Task 2: Extend `BuildTLSInfo` signature and add a retry helper

**Files:**
- Modify: `internal/httpclient/tls.go` (extend `BuildTLSInfo` signature, add `InsecureTLSClient()` factory)
- Modify: `internal/httpclient/tls_test.go` (one new test for the new signature)

**Interfaces (after this task):**
- `func BuildTLSInfo(resp *http.Response, req *http.Request, validationSkipped bool, originalError string) *model.TLSInfo`
- `func InsecureTLSClient() *http.Client` — fresh client with `InsecureSkipVerify: true`

- [ ] **Step 1: Update existing `BuildTLSInfo` callers and signature**

In `internal/httpclient/tls.go`, change the signature and body of `BuildTLSInfo`:

```go
// BuildTLSInfo returns the TLSInfo for a successful (or non-TLS) HTTP response.
// When resp.TLS is nil (plain HTTP), status is "no_tls_attempted".
// When validationSkipped is true (the request succeeded via a retry with
// InsecureSkipVerify after a first attempt failed), the returned TLSInfo
// carries ValidationSkipped=true and OriginalError=<first attempt message>.
func BuildTLSInfo(resp *http.Response, req *http.Request, validationSkipped bool, originalError string) *model.TLSInfo {
	host := ""
	if req != nil && req.URL != nil {
		host = req.URL.Host
	}
	if resp.TLS == nil {
		return &model.TLSInfo{
			Status:     "no_tls_attempted",
			Error:      "URL scheme is not HTTPS",
			TargetHost: host,
		}
	}
	return &model.TLSInfo{
		Status:              "ok",
		TargetHost:          host,
		AttemptedServerName: resp.TLS.ServerName,
		Connection:          buildConnectionView(resp.TLS),
		Certificates:        buildCertificateViews(resp.TLS.PeerCertificates),
		ValidationSkipped:   validationSkipped,
		OriginalError:       originalError,
	}
}
```

Add a new exported factory `InsecureTLSClient` at the end of the file:

```go
// InsecureTLSClient returns a fresh *http.Client whose Transport skips
// certificate validation. It is intended for one-shot retry attempts
// after a TLS handshake failure; the returned client must not be shared
// across goroutines because each call constructs a new Transport.
func InsecureTLSClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
```

(`crypto/tls` and `net/http` are already imported in `tls.go`.)

- [ ] **Step 2: Update existing callers**

The existing `BuildTLSInfo` callers all pass `(resp, req)`. Update them to pass `false, ""` for the two new parameters:

- `internal/httpclient/response.go` — `ParseResponse` calls `BuildTLSInfo(resp, req)` twice. Replace each with `BuildTLSInfo(resp, req, false, "")`.

- [ ] **Step 3: Update existing tests**

The existing tests in `tls_test.go` that call `BuildTLSInfo` need the new arguments:

- `TestBuildTLSInfo_NonHTTPS`: change `BuildTLSInfo(resp, req)` → `BuildTLSInfo(resp, req, false, "")`. Add assertion `if got.ValidationSkipped { t.Error("ValidationSkipped should be false for plain HTTP") }`.
- `TestBuildTLSInfo_HTTPS`: change `BuildTLSInfo(resp, req)` → `BuildTLSInfo(resp, req, false, "")`. Add assertion `if got.ValidationSkipped { t.Error("ValidationSkipped should be false for clean TLS") }`.

Run `go test ./internal/httpclient/ -run 'TestBuildTLSInfo' -v` and confirm both still pass before moving on.

- [ ] **Step 4: Write a failing test for the new signature**

Append to `internal/httpclient/tls_test.go`:

```go
func TestBuildTLSInfo_ValidationSkipped(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Use a client that trusts the server cert so the call succeeds.
	rootCAs := x509.NewCertPool()
	rootCAs.AddCert(server.Certificate())

	resp, err := (&http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: rootCAs}},
	}).Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	got := BuildTLSInfo(resp, req, true, "x509: certificate signed by unknown authority")

	if got.Status != "ok" {
		t.Fatalf("Status = %q, want ok", got.Status)
	}
	if !got.ValidationSkipped {
		t.Error("ValidationSkipped should be true")
	}
	if got.OriginalError != "x509: certificate signed by unknown authority" {
		t.Errorf("OriginalError = %q, want x509: ...", got.OriginalError)
	}
	if len(got.Certificates) < 1 {
		t.Error("Certificates should be populated from resp.TLS")
	}
}
```

Add `"crypto/tls"` and `"crypto/x509"` to the test imports if not already present.

- [ ] **Step 5: Run test to verify it fails (compile error)**

Run: `go test ./internal/httpclient/ -run 'TestBuildTLSInfo_ValidationSkipped' -v`
Expected: FAIL with a compile error: `too many arguments in call to BuildTLSInfo`.

- [ ] **Step 6: Run test to verify it passes**

Run: `go test ./internal/httpclient/ -run 'TestBuildTLSInfo' -v`
Expected: PASS for all `BuildTLSInfo` tests (including the existing `NonHTTPS` / `HTTPS` ones after the signature update in Step 3, plus this new one).

- [ ] **Step 7: Run full httpclient tests**

Run: `go test ./internal/httpclient/ -v`
Expected: All tests pass.

- [ ] **Step 8: Commit**

```bash
git add internal/httpclient/tls.go internal/httpclient/tls_test.go internal/httpclient/response.go
git commit -m "feat(httpclient): extend BuildTLSInfo with insecure-retry fields"
```

---

