package httpclient

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

func TestBuildCertificateViews_Labels(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	// Chain: [leaf (self-signed as a stand-in), leaf]  → positions: leaf, root
	cert1 := buildSelfSignedCert(t, &priv.PublicKey, priv, "a.example")
	cert2 := buildSelfSignedCert(t, &priv.PublicKey, priv, "b.example")
	got := buildCertificateViews([]*x509.Certificate{cert1, cert2})
	if len(got) != 2 {
		t.Fatalf("expected 2 certs, got %d", len(got))
	}
	if got[0].Position != "leaf" {
		t.Errorf("first position = %q, want leaf", got[0].Position)
	}
	if got[1].Position != "root" {
		t.Errorf("last position = %q, want root", got[1].Position)
	}
}

func TestBuildCertificateView_Fields(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	cert := buildSelfSignedCert(t, &priv.PublicKey, priv, "leaf.example")
	got := buildCertificateView(cert, "leaf")

	if got.Subject == "" {
		t.Error("Subject empty")
	}
	if got.Issuer == "" {
		t.Error("Issuer empty")
	}
	if got.SerialNumber == "" {
		t.Error("SerialNumber empty")
	}
	if got.SignatureAlgorithm == "" {
		t.Error("SignatureAlgorithm empty")
	}
	if got.NotBefore == "" || got.NotAfter == "" {
		t.Errorf("NotBefore/NotAfter empty: %q / %q", got.NotBefore, got.NotAfter)
	}
	if got.IsExpired {
		t.Error("freshly built cert should not be expired")
	}
	if got.IsNotYetValid {
		t.Error("freshly built cert should be currently valid")
	}
	if got.KeyAlgorithm != "RSA" {
		t.Errorf("KeyAlgorithm = %q, want RSA", got.KeyAlgorithm)
	}
	if got.KeySize != 2048 {
		t.Errorf("KeySize = %d, want 2048", got.KeySize)
	}
	if got.PublicKeyPEM == "" || !contains(got.PublicKeyPEM, "BEGIN PUBLIC KEY") {
		t.Errorf("PublicKeyPEM invalid: %q", got.PublicKeyPEM)
	}
	if got.FingerprintSHA1 == "" || got.FingerprintSHA256 == "" {
		t.Errorf("fingerprints empty: SHA1=%q SHA256=%q", got.FingerprintSHA1, got.FingerprintSHA256)
	}
	if got.PEM == "" || !contains(got.PEM, "BEGIN CERTIFICATE") {
		t.Errorf("PEM invalid: %q", got.PEM)
	}
	if got.RawDER == "" {
		t.Error("RawDER empty")
	}
	if got.SignatureBytes == "" {
		t.Error("SignatureBytes empty")
	}
	if !sliceContains(got.KeyUsage, "DigitalSignature") {
		t.Errorf("expected DigitalSignature in key usage: %v", got.KeyUsage)
	}
	if !sliceContains(got.ExtendedKeyUsage, "ServerAuth") {
		t.Errorf("expected ServerAuth in extended key usage: %v", got.ExtendedKeyUsage)
	}
}

func TestBuildConnectionView_Nil(t *testing.T) {
	got := buildConnectionView(nil)
	if got != nil {
		t.Errorf("expected nil for nil ConnectionState, got %+v", got)
	}
}

func TestBuildConnectionView_TLS13(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	resp, err := server.Client().Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	got := buildConnectionView(resp.TLS)
	if got == nil {
		t.Fatal("expected non-nil connection view")
	}
	if got.Version != "TLS 1.2" && got.Version != "TLS 1.3" {
		t.Errorf("Version = %q, expected TLS 1.2 or 1.3", got.Version)
	}
	if got.PeerCertificates < 1 {
		t.Errorf("PeerCertificates = %d, expected >= 1", got.PeerCertificates)
	}
}

func TestBuildTLSInfo_NonHTTPS(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader("")),
		TLS:        nil,
	}
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	got := BuildTLSInfo(resp, req)
	if got == nil {
		t.Fatal("expected non-nil TLSInfo")
	}
	if got.Status != "no_tls_attempted" {
		t.Errorf("Status = %q, want no_tls_attempted", got.Status)
	}
	if got.TargetHost != "example.com" {
		t.Errorf("TargetHost = %q, want example.com", got.TargetHost)
	}
	if got.Connection != nil {
		t.Errorf("Connection should be nil for non-HTTPS, got %+v", got.Connection)
	}
	if len(got.Certificates) != 0 {
		t.Errorf("Certificates should be empty for non-HTTPS, got %d", len(got.Certificates))
	}
}

func TestBuildTLSInfo_HTTPS(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL+"/foo", nil)
	resp, err := server.Client().Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	got := BuildTLSInfo(resp, req)
	if got == nil || got.Status != "ok" {
		t.Fatalf("expected Status=ok, got %+v", got)
	}
	if got.Connection == nil {
		t.Fatal("expected non-nil Connection")
	}
	if len(got.Certificates) < 1 {
		t.Errorf("expected >= 1 cert, got %d", len(got.Certificates))
	}
	if got.Certificates[0].Position != "leaf" {
		t.Errorf("first cert position = %q, want leaf", got.Certificates[0].Position)
	}
}

func TestBuildTLSInfoFromError_TLS(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com/foo", nil)
	err := errors.New("tls: handshake failure")
	got := BuildTLSInfoFromError(req, err)
	if got.Status != "handshake_failed" {
		t.Errorf("Status = %q, want handshake_failed", got.Status)
	}
	if got.Error != "tls: handshake failure" {
		t.Errorf("Error = %q, want %q", got.Error, "tls: handshake failure")
	}
	if got.AttemptedServerName != "example.com" {
		t.Errorf("AttemptedServerName = %q, want example.com", got.AttemptedServerName)
	}
}

func TestBuildTLSInfoFromError_X509(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com/foo", nil)
	err := errors.New("x509: certificate is valid for other.example, not example.com")
	got := BuildTLSInfoFromError(req, err)
	if got.Status != "handshake_failed" {
		t.Errorf("Status = %q, want handshake_failed", got.Status)
	}
}

func TestBuildTLSInfoFromError_ConnectionRefused(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com/foo", nil)
	err := errors.New("dial tcp 127.0.0.1:1: connect: connection refused")
	got := BuildTLSInfoFromError(req, err)
	if got.Status != "no_tls_attempted" {
		t.Errorf("Status = %q, want no_tls_attempted", got.Status)
	}
}

func TestBuildTLSInfoFromError_DNS(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://nonexistent.invalid/foo", nil)
	err := errors.New("no such host")
	got := BuildTLSInfoFromError(req, err)
	if got.Status != "no_tls_attempted" {
		t.Errorf("Status = %q, want no_tls_attempted", got.Status)
	}
}

func TestBuildTLSInfoFromError_Timeout(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com/foo", nil)
	got := BuildTLSInfoFromError(req, context.DeadlineExceeded)
	if got.Status != "no_tls_attempted" {
		t.Errorf("Status = %q, want no_tls_attempted", got.Status)
	}
	if !strings.Contains(got.Error, "deadline") && !strings.Contains(got.Error, "timeout") {
		t.Errorf("Error = %q, expected deadline/timeout wording", got.Error)
	}
}