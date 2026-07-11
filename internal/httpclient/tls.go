package httpclient

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math"
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

// base64StdEncode is used by later tasks; declare here so the import is used.
var _ = base64.StdEncoding.EncodeToString