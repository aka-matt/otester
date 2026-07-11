package httpclient

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/url"
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

// buildSelfSignedCert creates a minimal self-signed cert with the given key
// and a single DNSName SAN. Used as a fixture for SAN / key-extraction tests.
func buildSelfSignedCert(t *testing.T, pub any, signer any, dnsName string) *x509.Certificate {
	t.Helper()
	certTemplate := &x509.Certificate{
		SerialNumber:          bigInt(t, 1),
		Subject:               pkixName("CN=test.example"),
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{dnsName},
		IPAddresses:           []net.IP{net.ParseIP("192.0.2.1")},
		EmailAddresses:        []string{"admin@test.example"},
		URIs:                  []*url.URL{{Scheme: "https", Host: "test.example", Path: "/.well-known/acme-challenge/abc"}},
	}
	der, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, pub, signer)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	return cert
}

func bigInt(t *testing.T, v int64) *big.Int {
	t.Helper()
	return big.NewInt(v)
}

func pkixName(cn string) pkix.Name {
	return pkix.Name{CommonName: cn}
}

func TestSANToStrings(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	cert := buildSelfSignedCert(t, &priv.PublicKey, priv, "test.example")

	got := sanToStrings(cert)
	wantSubstrings := []string{"DNS:test.example", "IP:192.0.2.1", "email:admin@test.example"}
	for _, w := range wantSubstrings {
		if !sliceContains(got, w) {
			t.Errorf("expected SAN %q in %v", w, got)
		}
	}
}

func TestPublicKeyPEM(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pemStr, err := publicKeyPEM(&priv.PublicKey)
	if err != nil {
		t.Fatalf("publicKeyPEM: %v", err)
	}
	if !contains(pemStr, "-----BEGIN PUBLIC KEY-----") {
		t.Errorf("expected BEGIN PUBLIC KEY header in %q", pemStr)
	}
}

func TestKeyAlgorithm_RSA(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	if got := extractKeyAlgorithm(&priv.PublicKey); got != "RSA" {
		t.Errorf("extractKeyAlgorithm RSA = %q, want RSA", got)
	}
	if got := extractKeySize(&priv.PublicKey); got != 2048 {
		t.Errorf("extractKeySize RSA = %d, want 2048", got)
	}
}

func TestKeyAlgorithm_ECDSA(t *testing.T) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if got := extractKeyAlgorithm(&priv.PublicKey); got != "ECDSA" {
		t.Errorf("extractKeyAlgorithm ECDSA = %q, want ECDSA", got)
	}
	if got := extractKeySize(&priv.PublicKey); got != 256 {
		t.Errorf("extractKeySize ECDSA P-256 = %d, want 256", got)
	}
}

func TestKeyAlgorithm_Ed25519(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	if got := extractKeyAlgorithm(pub); got != "Ed25519" {
		t.Errorf("extractKeyAlgorithm Ed25519 = %q, want Ed25519", got)
	}
	if got := extractKeySize(pub); got != 0 {
		t.Errorf("extractKeySize Ed25519 = %d, want 0", got)
	}
}

func sliceContains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}