### Task 2: Backend tls.go — primitive formatters

**Files:**
- Create: `internal/httpclient/tls.go` (start with package + helpers, no top-level functions yet)
- Create: `internal/httpclient/tls_test.go` (start with tests for the helpers)

**Interfaces:**
- Produces:
  - `func tlsVersionString(version uint16) string`
  - `func cipherSuiteName(id uint16) string`
  - `func formatFingerprint(der []byte, sep string) string`
  - `func positionLabel(idx, total int) string`
  - `func certToPEM(der []byte) string`
  - `func daysUntilExpiry(notAfter time.Time) int`
  - `func keyUsageToStrings(usage x509.KeyUsage) []string`
  - `func extKeyUsageToStrings(usage []x509.ExtKeyUsage) []string`

- [ ] **Step 1: Write the failing tests**

Create `internal/httpclient/tls_test.go`:

```go
package httpclient

import (
	"crypto/x509"
	"testing"
	"time"
)

func TestTLSVersionString(t *testing.T) {
	cases := []struct {
		in   uint16
		want string
	}{
		{0x0303, "TLS 1.2"},
		{0x0304, "TLS 1.3"},
		{0x0302, "unknown"},
	}
	for _, c := range cases {
		if got := tlsVersionString(c.in); got != c.want {
			t.Errorf("tlsVersionString(0x%04x) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCipherSuiteName(t *testing.T) {
	cases := []struct {
		in   uint16
		want string
	}{
		{0x1301, "TLS_AES_128_GCM_SHA256"},
		{0x1302, "TLS_AES_256_GCM_SHA384"},
		{0x1303, "TLS_CHACHA20_POLY1305_SHA256"},
		{0x0000, "0x0000"},
	}
	for _, c := range cases {
		if got := cipherSuiteName(c.in); got != c.want {
			t.Errorf("cipherSuiteName(0x%04x) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatFingerprint(t *testing.T) {
	der := []byte{0x01, 0x02, 0x03, 0xab, 0xcd}
	got := formatFingerprint(der, ":")
	want := "01:02:03:AB:CD"
	if got != want {
		t.Errorf("formatFingerprint colon = %q, want %q", got, want)
	}
	gotSpace := formatFingerprint(der, " ")
	wantSpace := "01 02 03 AB CD"
	if gotSpace != wantSpace {
		t.Errorf("formatFingerprint space = %q, want %q", gotSpace, wantSpace)
	}
}

func TestPositionLabel(t *testing.T) {
	cases := []struct {
		idx, total int
		want       string
	}{
		{0, 3, "leaf"},
		{1, 3, "intermediate"},
		{2, 3, "root"},
		{0, 1, "leaf"},
		{1, 2, "root"},
		{0, 2, "leaf"},
	}
	for _, c := range cases {
		if got := positionLabel(c.idx, c.total); got != c.want {
			t.Errorf("positionLabel(%d, %d) = %q, want %q", c.idx, c.total, got, c.want)
		}
	}
}

func TestCertToPEM(t *testing.T) {
	der := []byte("not really der but enough for header check")
	pem := certToPEM(der)
	if pem == "" {
		t.Fatal("expected non-empty PEM")
	}
	if !contains(pem, "-----BEGIN CERTIFICATE-----") {
		t.Errorf("expected BEGIN CERTIFICATE header in %q", pem)
	}
	if !contains(pem, "-----END CERTIFICATE-----") {
		t.Errorf("expected END CERTIFICATE footer in %q", pem)
	}
}

func TestDaysUntilExpiry_Past(t *testing.T) {
	past := time.Now().Add(-48 * time.Hour)
	got := daysUntilExpiry(past)
	if got >= 0 {
		t.Errorf("expected negative days for past time, got %d", got)
	}
}

func TestDaysUntilExpiry_Future(t *testing.T) {
	future := time.Now().Add(72 * time.Hour)
	got := daysUntilExpiry(future)
	if got <= 0 {
		t.Errorf("expected positive days for future time, got %d", got)
	}
	if got != 3 {
		t.Errorf("expected 3 days (72h), got %d", got)
	}
}

func TestKeyUsageToStrings(t *testing.T) {
	usage := x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	got := keyUsageToStrings(usage)
	want := []string{"DigitalSignature", "KeyEncipherment"}
	if !equalStringSlices(got, want) {
		t.Errorf("keyUsageToStrings = %v, want %v", got, want)
	}
}

func TestExtKeyUsageToStrings(t *testing.T) {
	in := []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	got := extKeyUsageToStrings(in)
	want := []string{"ServerAuth", "ClientAuth"}
	if !equalStringSlices(got, want) {
		t.Errorf("extKeyUsageToStrings = %v, want %v", got, want)
	}
}

// contains and equalStringSlices are tiny local helpers (no need for strings pkg just for these).
func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/httpclient/ -run 'TestTLSVersionString|TestCipherSuiteName|TestFormatFingerprint|TestPositionLabel|TestCertToPEM|TestDaysUntilExpiry|TestKeyUsageToStrings|TestExtKeyUsageToStrings' -v`
Expected: FAIL — undefined: `tlsVersionString`, `cipherSuiteName`, etc.

- [ ] **Step 3: Implement helpers in tls.go**

Create `internal/httpclient/tls.go`:

```go
package httpclient

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"crypto/tls"
)

// tlsVersionString maps a TLS protocol version constant to its string label.
func tlsVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("0x%04x", version)
	}
}

// cipherSuiteName returns the standard cipher suite name for known IDs,
// or a hex string for unknown ones.
func cipherSuiteName(id uint16) string {
	for _, c := range tls.CipherSuites() {
		if c.ID == id {
			return c.Name
		}
	}
	return fmt.Sprintf("0x%04x", id)
}

// formatFingerprint returns the hex of der as uppercase, separated by sep.
func formatFingerprint(der []byte, sep string) string {
	hex := hex.EncodeToString(der)
	hex = strings.ToUpper(hex)
	var parts []string
	for i := 0; i < len(hex); i += 2 {
		parts = append(parts, hex[i:i+2])
	}
	return strings.Join(parts, sep)
}

// positionLabel returns "leaf" / "intermediate" / "root" based on chain position.
func positionLabel(idx, total int) string {
	if idx == 0 {
		return "leaf"
	}
	if idx == total-1 {
		return "root"
	}
	return "intermediate"
}

// certToPEM wraps DER bytes in a PEM CERTIFICATE block.
func certToPEM(der []byte) string {
	block := &pem.Block{Type: "CERTIFICATE", Bytes: der}
	return string(pem.EncodeToMemory(block))
}

// daysUntilExpiry returns the number of full days from now to notAfter.
// Negative when the certificate has already expired.
func daysUntilExpiry(notAfter time.Time) int {
	return int(notAfter.Sub(time.Now()) / (24 * time.Hour))
}

// keyUsageToStrings expands a KeyUsage bitmask into human names.
func keyUsageToStrings(usage x509.KeyUsage) []string {
	var out []string
	if usage&x509.KeyUsageDigitalSignature != 0 {
		out = append(out, "DigitalSignature")
	}
	if usage&x509.KeyUsageNonRepudiation != 0 {
		out = append(out, "NonRepudiation")
	}
	if usage&x509.KeyUsageKeyEncipherment != 0 {
		out = append(out, "KeyEncipherment")
	}
	if usage&x509.KeyUsageDataEncipherment != 0 {
		out = append(out, "DataEncipherment")
	}
	if usage&x509.KeyUsageKeyAgreement != 0 {
		out = append(out, "KeyAgreement")
	}
	if usage&x509.KeyUsageCertSign != 0 {
		out = append(out, "CertSign")
	}
	if usage&x509.KeyUsageCRLSign != 0 {
		out = append(out, "CRLSign")
	}
	if usage&x509.KeyUsageEncipherOnly != 0 {
		out = append(out, "EncipherOnly")
	}
	if usage&x509.KeyUsageDecipherOnly != 0 {
		out = append(out, "DecipherOnly")
	}
	return out
}

// extKeyUsageToStrings expands ExtendedKeyUsage values into names.
func extKeyUsageToStrings(usage []x509.ExtKeyUsage) []string {
	names := map[x509.ExtKeyUsage]string{
		x509.ExtKeyUsageServerAuth:      "ServerAuth",
		x509.ExtKeyUsageClientAuth:      "ClientAuth",
		x509.ExtKeyUsageCodeSigning:     "CodeSigning",
		x509.ExtKeyUsageEmailProtection: "EmailProtection",
		x509.ExtKeyUsageTimeStamping:    "TimeStamping",
		x509.ExtKeyUsageOCSPSigning:     "OCSPSigning",
	}
	out := make([]string, 0, len(usage))
	for _, u := range usage {
		if n, ok := names[u]; ok {
			out = append(out, n)
		} else {
			out = append(out, fmt.Sprintf("Unknown(%d)", u))
		}
	}
	return out
}

// base64StdEncode is used by later tasks; declare here so the import is used.
var _ = base64.StdEncoding.EncodeToString
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/httpclient/ -run 'TestTLSVersionString|TestCipherSuiteName|TestFormatFingerprint|TestPositionLabel|TestCertToPEM|TestDaysUntilExpiry|TestKeyUsageToStrings|TestExtKeyUsageToStrings' -v`
Expected: PASS for all 8.

- [ ] **Step 5: Commit**

```bash
git add internal/httpclient/tls.go internal/httpclient/tls_test.go
git commit -m "feat(httpclient): add TLS info primitive formatters"
```

---

