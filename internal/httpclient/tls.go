package httpclient

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"otester/internal/model"
)

// tlsVersionString maps a TLS protocol version constant to its string label.
func tlsVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "unknown"
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

// daysUntilExpiry returns the number of days from now to notAfter, rounded to
// the nearest whole day. Negative when the certificate has already expired.
func daysUntilExpiry(notAfter time.Time) int {
	return int(math.Round(notAfter.Sub(time.Now()).Hours() / 24))
}

// keyUsageToStrings expands a KeyUsage bitmask into human names.
func keyUsageToStrings(usage x509.KeyUsage) []string {
	var out []string
	if usage&x509.KeyUsageDigitalSignature != 0 {
		out = append(out, "DigitalSignature")
	}
	if usage&x509.KeyUsageContentCommitment != 0 {
		out = append(out, "ContentCommitment")
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

// sanToStrings renders every Subject Alternative Name on cert with a type prefix.
func sanToStrings(cert *x509.Certificate) []string {
	out := make([]string, 0, len(cert.DNSNames)+len(cert.IPAddresses)+len(cert.EmailAddresses)+len(cert.URIs))
	for _, d := range cert.DNSNames {
		out = append(out, "DNS:"+d)
	}
	for _, ip := range cert.IPAddresses {
		out = append(out, "IP:"+ip.String())
	}
	for _, e := range cert.EmailAddresses {
		out = append(out, "email:"+e)
	}
	for _, u := range cert.URIs {
		out = append(out, "URI:"+u.String())
	}
	return out
}

// publicKeyPEM marshals any supported public key as PKIX and PEM-encodes it.
func publicKeyPEM(pub any) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: der}
	return string(pem.EncodeToMemory(block)), nil
}

// extractKeyAlgorithm returns "RSA", "ECDSA", or "Ed25519" for known key types.
func extractKeyAlgorithm(pub any) string {
	switch pub.(type) {
	case *rsa.PublicKey:
		return "RSA"
	case *ecdsa.PublicKey:
		return "ECDSA"
	case ed25519.PublicKey:
		return "Ed25519"
	default:
		return "Unknown"
	}
}

// extractKeySize returns the bit size of the public key, or 0 for keys
// where size is not meaningful (Ed25519).
func extractKeySize(pub any) int {
	switch k := pub.(type) {
	case *rsa.PublicKey:
		return k.N.BitLen()
	case *ecdsa.PublicKey:
		return k.Curve.Params().BitSize
	case ed25519.PublicKey:
		return 0
	default:
		return 0
	}
}

// buildConnectionView returns a TLSConnectionView populated from cs, or nil if cs is nil.
func buildConnectionView(cs *tls.ConnectionState) *model.TLSConnectionView {
	if cs == nil {
		return nil
	}
	v := &model.TLSConnectionView{
		Version:            tlsVersionString(cs.Version),
		CipherSuite:        fmt.Sprintf("0x%04x", cs.CipherSuite),
		CipherSuiteName:    cipherSuiteName(cs.CipherSuite),
		NegotiatedProtocol: cs.NegotiatedProtocol,
		ServerName:         cs.ServerName,
		Resumed:            cs.DidResume,
		OCSPStapled:        cs.OCSPResponse != nil,
		PeerCertificates:   len(cs.PeerCertificates),
	}
	if len(cs.SignedCertificateTimestamps) > 0 {
		scts := make([]string, 0, len(cs.SignedCertificateTimestamps))
		for _, sct := range cs.SignedCertificateTimestamps {
			scts = append(scts, base64.StdEncoding.EncodeToString(sct))
		}
		v.SCTs = scts
	}
	return v
}

// buildCertificateViews converts a server-returned cert chain into CertificateView slices,
// labeling position by chain order (leaf / intermediate / root).
func buildCertificateViews(certs []*x509.Certificate) []model.CertificateView {
	if len(certs) == 0 {
		return []model.CertificateView{}
	}
	out := make([]model.CertificateView, 0, len(certs))
	for i, c := range certs {
		out = append(out, buildCertificateView(c, positionLabel(i, len(certs))))
	}
	return out
}

// buildCertificateView converts one x509.Certificate into the JSON view model.
func buildCertificateView(cert *x509.Certificate, position string) model.CertificateView {
	now := time.Now()

	maxPathLen := -1
	if cert.MaxPathLen > 0 || cert.MaxPathLenZero {
		maxPathLen = cert.MaxPathLen
	}

	policies := make([]string, 0, len(cert.PolicyIdentifiers))
	for _, p := range cert.PolicyIdentifiers {
		policies = append(policies, p.String())
	}

	pemStr := certToPEM(cert.Raw)
	pubPEM, _ := publicKeyPEM(cert.PublicKey)

	var (
		subjKeyID string
		authKeyID string
	)
	if len(cert.SubjectKeyId) > 0 {
		subjKeyID = formatFingerprint(cert.SubjectKeyId, ":")
	}
	if len(cert.AuthorityKeyId) > 0 {
		authKeyID = formatFingerprint(cert.AuthorityKeyId, ":")
	}

	return model.CertificateView{
		Position:              position,
		Subject:               cert.Subject.String(),
		Issuer:                cert.Issuer.String(),
		SerialNumber:          cert.SerialNumber.Text(16),
		Version:               cert.Version,
		SignatureAlgorithm:    cert.SignatureAlgorithm.String(),
		NotBefore:             cert.NotBefore.UTC().Format(time.RFC3339),
		NotAfter:              cert.NotAfter.UTC().Format(time.RFC3339),
		IsExpired:             now.After(cert.NotAfter),
		IsNotYetValid:         now.Before(cert.NotBefore),
		DaysToExpiry:          daysUntilExpiry(cert.NotAfter),
		SubjectKeyId:          subjKeyID,
		AuthorityKeyId:        authKeyID,
		SANs:                  sanToStrings(cert),
		KeyAlgorithm:          extractKeyAlgorithm(cert.PublicKey),
		KeySize:               extractKeySize(cert.PublicKey),
		PublicKeyPEM:          pubPEM,
		FingerprintSHA1:       formatFingerprint(sha1Of(cert.Raw), ":"),
		FingerprintSHA256:     formatFingerprint(sha256Of(cert.Raw), ":"),
		IsCA:                  cert.BasicConstraintsValid && cert.IsCA,
		MaxPathLength:         maxPathLen,
		KeyUsage:              keyUsageToStrings(cert.KeyUsage),
		ExtendedKeyUsage:      extKeyUsageToStrings(cert.ExtKeyUsage),
		CRLDistributionPoints: cert.CRLDistributionPoints,
		Policies:              policies,
		RawDER:                base64.StdEncoding.EncodeToString(cert.Raw),
		PEM:                   pemStr,
		SignatureBytes:        base64.StdEncoding.EncodeToString(cert.Signature),
	}
}

func sha1Of(b []byte) []byte {
	h := sha1.Sum(b)
	return h[:]
}

func sha256Of(b []byte) []byte {
	h := sha256.Sum256(b)
	return h[:]
}

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

// InsecureTLSClient returns a fresh *http.Client whose Transport skips
// certificate validation. It is intended for one-shot retry attempts
// after a TLS handshake failure; the returned client must not be shared
// across goroutines because each call constructs a new Transport. The
// timeout is enforced by the shared client's policy so a hung insecure
// connection cannot outlive the request's deadline.
func InsecureTLSClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

// BuildTLSInfoFromError classifies a transport-level error and produces a TLSInfo
// reflecting whether TLS was attempted and whether it failed.
func BuildTLSInfoFromError(req *http.Request, err error) *model.TLSInfo {
	host := ""
	sni := ""
	if req != nil && req.URL != nil {
		host = req.URL.Host
		// For https URLs, the "host" portion is what would be sent as SNI.
		if req.URL.Scheme == "https" {
			sni = req.URL.Hostname()
		}
	}

	status := "no_tls_attempted"
	msg := ""
	if err != nil {
		msg = err.Error()
	}

	// context cancellations and timeouts are never TLS handshake failures.
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		status = "no_tls_attempted"
	} else if msg != "" && (strings.Contains(msg, "tls:") || strings.Contains(msg, "x509:")) {
		status = "handshake_failed"
	}

	info := &model.TLSInfo{
		Status:              status,
		Error:               msg,
		TargetHost:          host,
		AttemptedServerName: sni,
		Certificates:        []model.CertificateView{},
	}
	if status == "ok" {
		info.Error = ""
	}
	return info
}