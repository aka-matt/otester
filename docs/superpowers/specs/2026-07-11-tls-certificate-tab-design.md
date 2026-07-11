# TLS Certificate Tab Design

Add a third tab ("Certificate") to the right-panel `ResponseViewer`, parallel to the existing "Body" and "Headers" tabs. After every HTTP request — successful, failed, redirected, plain HTTP, or TLS-handshake-failed — this tab displays information about the TLS layer of the connection: the negotiated TLS connection metadata and the full server certificate chain, formatted for debugging.

## Architecture

Three layers, each with one responsibility.

**Backend (`internal/httpclient/` + `internal/model/`)** — capture and shape:
- `internal/httpclient/tls.go` (new) — pure conversion from `*tls.ConnectionState` and `[]*x509.Certificate` into view models. No IO.
- `internal/model/types.go` — extend `ResponseOutput` with `TLS *TLSInfo` (omitempty). Add `TLSInfo`, `TLSConnectionView`, `CertificateView` types.
- `internal/httpclient/response.go` — `ParseResponse` extended to take the originating `*http.Request` so it can build `TLSInfo` from `resp.TLS`.
- `internal/httpclient/client.go` — `DoRequest` plumbs the request into `ParseResponse` and `handleRequestError`. `handleRequestError` builds a `TLSInfo` reflecting whether TLS was attempted and whether it failed.

**Wails bridge** — auto-regenerates `frontend/wailsjs/go/models.ts` on build. No manual edits there.

**Frontend (`frontend/src/`)** — display:
- `frontend/src/types/index.ts` — TS mirror of the Go types.
- `frontend/src/components/CertificateViewer.vue` (new) — presentational only; receives a single `info: TLSInfo | null` prop.
- `frontend/src/components/ResponseViewer.vue` — one new `<n-tab-pane name="certificate" tab="Certificate">` rendering `<CertificateViewer :info="response.tls ?? null" />`.

Data flow per request:

```
RequestEditor → SendRequest → httpclient.DoRequest
                                     ↓
                                resp / err
                                     ↓
              ParseResponse / handleRequestError
                                     ↓
                          build TLSInfo (tls.go)
                                     ↓
                          ResponseOutput (json)
                                     ↓
                       ResponseStore.setResponse
                                     ↓
                    ResponseViewer → CertificateViewer
```

The backend owns the heavy lifting (parsing `x509.Certificate`, computing fingerprints, formatting dates). The frontend is dumb display. This keeps the JSON contract stable and unit-testable without a browser.

**Redirects:** Go's `http.Client` follows redirects transparently. Only the final hop's `*tls.ConnectionState` is attached to the final response. Earlier hops' certs are lost. v1 documents this and shows only the final hop.

## Data Model

All new fields use snake_case JSON to match the codebase convention.

### `TLSConnectionView`

One block per response, describing the TLS handshake itself (not a specific cert).

| Field | Type | Source | Notes |
|---|---|---|---|
| `version` | string | `tls.ConnectionState.Version` | Mapped via `tlsVersionString`: `0x0303` → `"TLS 1.2"`, `0x0304` → `"TLS 1.3"` |
| `cipherSuite` | string | `tls.ConnectionState.CipherSuite` (uint16) | e.g. `"TLS_AES_256_GCM_SHA384"` |
| `cipherSuiteName` | string | mapped from CipherSuite ID | e.g. `"AES-256-GCM"` |
| `negotiatedProtocol` | string | `tls.ConnectionState.NegotiatedProtocol` (omitempty) | ALPN, e.g. `"h2"` |
| `serverName` | string | `tls.ConnectionState.ServerName` | SNI sent; empty string if connecting by IP |
| `resumed` | bool | `tls.ConnectionState.DidResume` | Session resumption indicator |
| `scts` | []string (omitempty) | `tls.ConnectionState.SignedCertificateTimestamps` | base64 of each SCT; available directly from stdlib |
| `ocspStapled` | bool | `tls.ConnectionState.OCSPResponse != nil` | Same on every cert in chain |
| `peerCertificates` | int | `len(tls.ConnectionState.PeerCertificates)` | chain length returned by server |

### `CertificateView`

One per cert in the chain. Aim: every meaningful field from `*x509.Certificate`, without dumping binary blobs by default. `RawDER` and `PEM` are kept for the Copy actions.

| Field | Type | Source / mapping |
|---|---|---|
| `position` | string | `"leaf"` / `"intermediate"` / `"root"` (computed from index vs total) |
| `subject` | string | `cert.Subject` formatted via `String()` (RFC 2253) |
| `issuer` | string | `cert.Issuer` formatted via `String()` |
| `serialNumber` | string | `cert.SerialNumber.Text(16)` — uppercase hex |
| `version` | int | `cert.Version` |
| `signatureAlgorithm` | string | e.g. `"SHA256-RSA"` — derived from `cert.SignatureAlgorithm` |
| `notBefore` | string | `cert.NotBefore.UTC().Format(time.RFC3339)` |
| `notAfter` | string | `cert.NotAfter.UTC().Format(time.RFC3339)` |
| `isExpired` | bool | `time.Now().After(cert.NotAfter)` |
| `isNotYetValid` | bool | `time.Now().Before(cert.NotBefore)` |
| `daysToExpiry` | int | `int(cert.NotAfter.Sub(time.Now()) / 24h)` — negative if expired |
| `subjectKeyId` | string | hex of `cert.SubjectKeyId` (empty if nil) |
| `authorityKeyId` | string | hex of `cert.AuthorityKeyId` (empty if nil) |
| `sans` | []string | All SANs as `"DNS:..."`, `"IP:..."`, `"email:..."`, `"URI:..."` |
| `keyAlgorithm` | string | `"RSA"` / `"ECDSA"` / `"Ed25519"` — derived via type switch on `cert.PublicKey` |
| `keySize` | int | `rsa.N.BitLen()` / `ecdsa.Curve.Params().BitSize` / 0 for Ed25519 |
| `publicKeyPem` | string | PEM-encoded SubjectPublicKeyInfo |
| `fingerprintSha1` | string | colon-separated uppercase hex of `sha1(cert.Raw)` |
| `fingerprintSha256` | string | colon-separated uppercase hex of `sha256(cert.Raw)` |
| `isCa` | bool | `cert.BasicConstraintsValid && cert.IsCA` |
| `maxPathLength` | int | `cert.MaxPathLen`; `-1` if `cert.MaxPathLenZero` not set |
| `keyUsage` | []string | Bitmask expanded: `DigitalSignature`, `KeyEncipherment`, etc. |
| `extendedKeyUsage` | []string | E.g. `ServerAuth`, `ClientAuth` |
| `crlDistributionPoints` | []string (omitempty) | from `cert.CRLDistributionPoints` |
| `policies` | []string (omitempty) | OIDs from `cert.PolicyIdentifiers` |
| `rawDer` | string | base64 of `cert.Raw` |
| `pem` | string | full PEM block (header + base64 + footer) |
| `signatureBytes` | string | base64 of `cert.Signature` |

### `TLSInfo`

Wrapper attached to `ResponseOutput` as `TLS *TLSInfo` (omitempty so HTTP-only responses don't carry the field).

| Field | Type | Notes |
|---|---|---|
| `status` | string | One of `"ok"`, `"handshake_failed"`, `"no_tls_attempted"` |
| `error` | string (omitempty) | Populated when `status != "ok"` |
| `targetHost` | string | `host:port` from the request URL — always populated for context |
| `attemptedServerName` | string (omitempty) | SNI we sent |
| `connection` | `*TLSConnectionView` (omitempty) | Populated when `status == "ok"` |
| `certificates` | `[]CertificateView` | Empty when `status != "ok"` |

### Status semantics

- `"ok"` — full TLS handshake succeeded; chain captured
- `"handshake_failed"` — TLS was attempted but failed; `error` populated, `attemptedServerName` set, `certificates` empty
- `"no_tls_attempted"` — request never reached TLS stage (plain HTTP, DNS failure, connection refused, cancelled, timeout); `error` carries the original error message

## Go Implementation

### New file: `internal/httpclient/tls.go`

Public surface:

```
func BuildTLSInfo(resp *http.Response, req *http.Request) *model.TLSInfo
func BuildTLSInfoFromError(req *http.Request, err error) *model.TLSInfo
```

Plus unexported helpers, each unit-tested in isolation:

```
func buildConnectionView(cs *tls.ConnectionState) *model.TLSConnectionView
func buildCertificateViews(certs []*x509.Certificate) []model.CertificateView
func buildCertificateView(cert *x509.Certificate, position string) model.CertificateView
func tlsVersionString(version uint16) string
func cipherSuiteName(id uint16) string
func keyUsageToStrings(usage x509.KeyUsage) []string
func extKeyUsageToStrings(usage []x509.ExtKeyUsage) []string
func formatFingerprint(der []byte, sep string) string
func sanToStrings(cert *x509.Certificate) []string  // accepts the cert directly to handle all SAN types
func positionLabel(idx, total int) string
func publicKeyPEM(pub any) (string, error)
func certToPEM(der []byte) string
func daysUntilExpiry(notAfter time.Time) int
```

`BuildTLSInfo`:
1. If `resp.TLS == nil` → return `&model.TLSInfo{Status: "no_tls_attempted", TargetHost: req.URL.Host, Error: "URL scheme is not HTTPS"}`
2. Else → build full info with `Status: "ok"`

`BuildTLSInfoFromError`:
1. If `errors.Is(err, context.DeadlineExceeded)` or `errors.Is(err, context.Canceled)` → `Status: "no_tls_attempted"`, error is the original message
2. Else if `strings.Contains(err.Error(), "tls:")` or `strings.Contains(err.Error(), "x509:")` → `Status: "handshake_failed"`, `AttemptedServerName` from host portion of `req.URL.Host`
3. Else → `Status: "no_tls_attempted"`, error is the original message

Always populates `TargetHost` from `req.URL.Host`.

### Changes to `internal/httpclient/response.go`

`ParseResponse` signature change:

```go
// before
func ParseResponse(resp *http.Response, requestID string, duration time.Duration) (*model.ResponseOutput, error)
// after
func ParseResponse(resp *http.Response, req *http.Request, requestID string, duration time.Duration) (*model.ResponseOutput, error)
```

Inside, after building the existing `ResponseOutput`:

```go
output.TLS = BuildTLSInfo(resp, req)
```

The error path inside `ParseResponse` (body read failure) also sets `output.TLS = BuildTLSInfo(resp, req)` — same logic, the response exists, so `resp.TLS` is informative even when body read failed.

### Changes to `internal/httpclient/client.go`

In `DoRequest`:

```go
// success
resp, err := c.client.Do(req)
if err != nil {
    return handleRequestError(input.RequestID, reqCtx.Err(), err, duration, req)
}
// ...
return ParseResponse(resp, req, input.RequestID, duration)
```

`handleRequestError` signature change:

```go
// before
func handleRequestError(requestID string, ctxErr error, err error, duration time.Duration) (*model.ResponseOutput, error)
// after
func handleRequestError(requestID string, ctxErr error, err error, duration time.Duration, req *http.Request) (*model.ResponseOutput, error)
```

Inside, before returning the error response, build `output.TLS = BuildTLSInfoFromError(req, err)`.

## Frontend Implementation

### `frontend/src/types/index.ts`

Add the three interfaces — `TLSInfo`, `TLSConnectionView`, `CertificateView` — matching the Go JSON shape exactly. JSDoc on each interface for IDE help.

### `frontend/src/components/CertificateViewer.vue` (new)

Props: `{ info: TLSInfo | null }`.

Four render branches based on `info?.status`:

**1. `info === null`** — empty state:

> "No request sent yet."

**2. `info.status === 'no_tls_attempted'`** — info panel (not error styling):

- Neutral info icon
- Heading: "No TLS"
- Body: `info.error` if set, else default "This request did not use HTTPS — certificate info is not applicable."
- Sub-line: `Target: <info.targetHost>`

**3. `info.status === 'handshake_failed'`** — error panel:

- Warning icon
- Heading: "TLS handshake failed"
- Body: `info.error`
- Sub-line: `Target: <host>`, `SNI attempted: <name>`

**4. `info.status === 'ok'`** — the full panel:

- **TLS Connection block** at the top — acrylic card, two-column key-value list:
  - Version | Cipher suite (ID + human name)
  - Negotiated protocol (ALPN) | Server name (SNI)
  - Resumed session (yes/no) | Peer certificates count
  - OCSP stapled (yes/no) | SCTs count (if any)
- **Certificate chain accordion** — `<n-collapse>` with one `<n-collapse-item>` per cert:
  - Title format: `"<position> — <CN>"` where CN is the Common Name parsed from Subject, with a fallback to the full Subject string
  - Leaf expanded by default; intermediates and root collapsed

Per-cert sections (all visible in the expanded leaf):

1. **Subject & Issuer** — Subject DN, Issuer DN, Serial Number, Version
2. **Validity** — Not Before, Not After, Days to expiry, status pill
3. **Subject Alternative Names** — list with type prefix (DNS / IP / email / URI)
4. **Key** — Algorithm, Size, Public Key PEM in `<pre>` block with Copy button
5. **Fingerprints** — SHA-1, SHA-256 with Copy buttons
6. **Extensions** — Basic Constraints, Key Usage (badges), Extended Key Usage (badges), CRL Distribution Points (list), Policies (list)
7. **Raw** — collapsed by default: full PEM block, Signature (base64)

**Copy buttons** — `navigator.clipboard.writeText(text)` with `useMessage().success("Copied")` toast. Same pattern as `copyBody` in `ResponseViewer.vue`.

**Validity pill colors** (color + text — color is never the only signal):
- Valid → `var(--success-color)`
- Expiring within 30 days → `var(--warning-color)` with text "Expires soon"
- Expired → `var(--danger-color)` with text "Expired"
- Not yet valid → `var(--warning-color)` with text "Not yet valid"

**Layout:**
- Tab content scrolls within the right-panel scroll container (no internal scroll on `CertificateViewer`)
- Long PEM: `<pre style="overflow-x:auto;white-space:pre-wrap">`
- Acrylic card treatment on the connection block matches the existing response-overview card

**Accessibility:**
- `n-collapse` is keyboard-navigable by default (Naive UI handles it)
- Copy buttons have `aria-label` (e.g. "Copy SHA-256 fingerprint")
- Validity status text is always present alongside the pill color

### `frontend/src/components/ResponseViewer.vue` change

Add one `<n-tab-pane>` after the Headers pane:

```vue
<n-tab-pane name="certificate" tab="Certificate">
  <CertificateViewer :info="response.tls ?? null" />
</n-tab-pane>
```

No other changes to this file.

## Testing Strategy

### Go tests (`internal/httpclient/tls_test.go`, new)

| Test | Asserts |
|---|---|
| `TestBuildTLSInfo_NonHTTPS` | `resp.TLS == nil` → status `no_tls_attempted`, `Connection == nil`, `Certificates` empty, `TargetHost` populated |
| `TestBuildTLSInfo_HTTPS` | `httptest.NewTLSServer` → status `ok`, `Certificates` len matches chain, `Connection.Version` matches TLS 1.2/1.3 |
| `TestBuildTLSInfoFromError_TLSHandshake` | error string `"tls: handshake failure"` → status `handshake_failed`, `AttemptedServerName` set |
| `TestBuildTLSInfoFromError_X509Error` | error string `"x509: certificate is valid for other.example"` → status `handshake_failed` |
| `TestBuildTLSInfoFromError_ConnectionRefused` | error string `"dial tcp: connection refused"` → status `no_tls_attempted` |
| `TestBuildTLSInfoFromError_DNSFailure` | error string `"no such host"` → status `no_tls_attempted` |
| `TestBuildTLSInfoFromError_Timeout` | `errors.Is(err, context.DeadlineExceeded)` → status `no_tls_attempted`, error preserved |
| `TestFormatFingerprint_SHA1` | Known input → known colon-separated hex |
| `TestFormatFingerprint_SHA256` | Same |
| `TestPositionLabel` | `(0, 3)` → `"leaf"`; `(1, 3)` → `"intermediate"`; `(2, 3)` → `"root"` |
| `TestKeyUsageToStrings` | Known bitmask → expected list |
| `TestSANToStrings` | Mixed DNS/IP/email/URI SANs → expected prefixed strings |
| `TestCipherSuiteName` | Known IDs (TLS_AES_128_GCM_SHA256 etc.) → expected names; unknown ID → `"0xNNNN"` fallback |
| `TestTLSVersionString` | `0x0303` → `"TLS 1.2"`; `0x0304` → `"TLS 1.3"` |
| `TestPublicKeyPEM` | Built-in test key → valid PEM with `BEGIN PUBLIC KEY` header; `error == nil` |
| `TestCertToPEM` | DER bytes → PEM with `BEGIN CERTIFICATE` header |
| `TestDaysUntilExpiry` | Past time → negative; future time → positive; boundary 0 → 0 |
| `TestKeyAlgorithm_RSA` | Cert built with `rsa.GenerateKey` → `KeyAlgorithm: "RSA"`, correct `KeySize` |
| `TestKeyAlgorithm_ECDSA` | Cert built with `ecdsa.GenerateKey` → `KeyAlgorithm: "ECDSA"`, correct curve-derived size |
| `TestKeyAlgorithm_Ed25519` | Cert built with `ed25519.GenerateKey` → `KeyAlgorithm: "Ed25519"`, `KeySize: 0` |
| `TestDaysUntilExpiry_Past` | Time before now → negative |
| `TestDaysUntilExpiry_Future` | Time after now → positive |

### Go tests (`internal/httpclient/client_test.go`, updates)

Existing tests `TestClient_DoRequest_Success`, `TestClient_DoRequest_Timeout`, `TestClient_DoRequest_Cancellation`, `TestClient_DoRequest_QueryParams`, `TestClient_DoRequest_Headers`, `TestClient_DoRequest_PostWithBody` — keep them passing. The first two gain `TLS != nil` assertions:

- `TestClient_DoRequest_Success` — switch the test server to `httptest.NewTLSServer`, assert `output.TLS.Status == "ok"` and `len(output.TLS.Certificates) >= 1`
- `TestClient_DoRequest_Timeout` — assert `output.TLS.Status == "no_tls_attempted"`
- New: `TestClient_DoRequest_TLSHandshakeFailure` — use `httptest.NewTLSServer`, then set the URL to `127.0.0.1:port` while the server's cert is for `other.example` → assert `output.TLS.Status == "handshake_failed"` and that `output.TLS.Error` contains `"x509:"`

### Frontend tests (`frontend/src/components/CertificateViewer.test.ts`, new)

| Test | Asserts |
|---|---|
| `renders empty state when info is null` | Text "No request sent yet" visible |
| `renders no-TLS panel for no_tls_attempted` | "No TLS" heading, target host sub-line |
| `renders handshake-failed panel` | "TLS handshake failed" heading, error message visible, danger styling applied |
| `renders connection block for ok status` | Version, cipher suite, ALPN, SNI fields all visible |
| `renders accordion with one item per cert` | Three `<n-collapse-item>` for a 3-cert fixture |
| `leaf expanded by default` | First item's content visible; others collapsed |
| `validity pill - expired` | Renders "Expired" with danger CSS class |
| `validity pill - valid` | Renders "Valid" with success CSS class |
| `validity pill - expiring soon` | Renders "Expires soon" with warning CSS class |
| `validity pill - not yet valid` | Renders "Not yet valid" with warning CSS class |
| `copy SHA-256 button calls clipboard.writeText` | Mocked `navigator.clipboard.writeText` called with the colon-separated SHA-256 string |
| `copy PEM button calls clipboard.writeText` | Same, with the PEM string |
| `copy SHA-1 button calls clipboard.writeText` | Same, with SHA-1 string |
| `all SANs rendered with type prefix` | DNS, IP, email, URI variants all visible |

### Frontend tests (`frontend/src/components/ResponseViewer.test.ts`, new file or extension)

| Test | Asserts |
|---|---|
| `Certificate tab pane exists` | Third `<n-tab-pane name="certificate">` rendered |
| `Certificate tab passes null when response.tls is undefined` | `CertificateViewer` receives `null` |
| `Certificate tab passes info when response.tls is set` | `CertificateViewer` receives the `TLSInfo` object |

## Acceptance Criteria

The feature is done when all of these are true:

1. Sending a request to `https://www.example.com` shows a Certificate tab with at least the leaf cert of the actual chain returned by example.com.
2. The leaf cert's panel shows: Subject, Issuer, Validity (with status), SANs, Key algorithm + size, SHA-1 + SHA-256 fingerprints, Key Usage, Extended Key Usage, Basic Constraints, Public Key PEM, and Raw PEM.
3. The chain accordion shows all returned certificates in order (leaf first), labeled by position.
4. Sending a request to `http://example.com` (plain HTTP) shows a Certificate tab with an informational "No TLS" panel that does not look like an error.
5. A request to a server with an untrusted / mismatched cert shows the Certificate tab with "TLS handshake failed" and the underlying error message.
6. Sending to an unreachable host (DNS fail, connection refused) shows "No TLS" with the underlying connection error — not a fake cert error.
7. A cancelled or timed-out request does not show fake TLS info; it shows the actual reason.
8. Each cert panel has working Copy buttons for SHA-1, SHA-256, and PEM (verified in test).
9. Validity status is visually distinct (color + text) for: valid / expiring-within-30-days / expired / not-yet-valid.
10. All new Go functions have ≥ 1 unit test; all new Vue states have ≥ 1 component test.
11. Existing tests in `client_test.go`, `response_test.go`, `ResponseViewer.test.ts`, `MainView.test.ts` still pass without modification to their assertions about non-TLS fields.
12. `go test ./...` passes; `npm test` passes; `npm run build` succeeds.

## Out of Scope (v1)

- Multi-hop redirect cert chain — only the final hop's certs are shown
- Saving/exporting `.crt` files to disk (Copy buttons only)
- Cert pinning / pinning violation detection
- Showing OAuth tokens or `client_secret` in the cert panel (the data path doesn't allow it, but noted for future scope clarity)

## Risks

- *Risk:* TLS handshake error strings vary across Go versions. *Mitigation:* substring match for `"tls:"` / `"x509:"` is stable across Go 1.18+; tests cover both substrings.
- *Risk:* Base64 raw DER for a 10-cert chain is large. *Mitigation:* < 100 KB even pathologically; Wails JSON pipe handles it. Drop `RawDER` and recompute on demand if we observe lag (YAGNI for v1).
- *Risk:* `n-collapse` semantics differ in jsdom vs browser. *Mitigation:* tests assert on rendered text and class names, not internal collapse state.

## File Touch List

Backend (Go):
- `internal/model/types.go` — add `TLSInfo`, `TLSConnectionView`, `CertificateView` types; add `TLS` field to `ResponseOutput`
- `internal/httpclient/tls.go` — new file, conversion helpers
- `internal/httpclient/tls_test.go` — new file, unit tests
- `internal/httpclient/response.go` — `ParseResponse` signature change, build `TLSInfo`
- `internal/httpclient/client.go` — pass `req` to `ParseResponse` and `handleRequestError`
- `internal/httpclient/response_test.go` — update `TestParseResponse_Success` and `TestParseResponse_TruncatedBody` for new signature; add SNI-mismatch TLS failure test
- `internal/httpclient/client_test.go` — update `TestClient_DoRequest_Success` to use TLS server + assertions; update `TestClient_DoRequest_Timeout`; add `TestClient_DoRequest_TLSHandshakeFailure`

Frontend (Vue):
- `frontend/src/types/index.ts` — mirror types
- `frontend/src/components/CertificateViewer.vue` — new file
- `frontend/src/components/CertificateViewer.test.ts` — new file
- `frontend/src/components/ResponseViewer.vue` — add one `<n-tab-pane>` + import

Auto-generated:
- `frontend/wailsjs/go/models.ts` — regenerated by Wails on next build; not hand-edited