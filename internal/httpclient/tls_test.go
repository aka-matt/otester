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