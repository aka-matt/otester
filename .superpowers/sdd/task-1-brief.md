### Task 1: Backend types — `TLSInfo`, `TLSConnectionView`, `CertificateView`

**Files:**
- Modify: `internal/model/types.go` (append new types, add field to `ResponseOutput`)
- Modify: `internal/model/types_test.go` (add JSON round-trip test)

**Interfaces:**
- Produces: `model.TLSInfo`, `model.TLSConnectionView`, `model.CertificateView` structs with JSON tags
- Produces: `model.ResponseOutput.TLS *TLSInfo` field with `json:"tls,omitempty"`

- [ ] **Step 1: Write the failing test**

Append to `internal/model/types_test.go`:

```go
func TestTLSInfoJSON(t *testing.T) {
	info := TLSInfo{
		Status:                "ok",
		TargetHost:            "example.com:443",
		AttemptedServerName:   "example.com",
		Connection: &TLSConnectionView{
			Version:          "TLS 1.3",
			CipherSuite:      "TLS_AES_256_GCM_SHA384",
			CipherSuiteName:  "AES-256-GCM",
			NegotiatedProtocol: "h2",
			ServerName:       "example.com",
			Resumed:          false,
			SCTs:             []string{"AQID"},
			OCSPStapled:      true,
			PeerCertificates: 2,
		},
		Certificates: []CertificateView{
			{
				Position: "leaf",
				Subject:  "CN=example.com",
				Issuer:   "CN=R3, O=Let's Encrypt",
			},
		},
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal TLSInfo: %v", err)
	}
	if !strings.Contains(string(data), `"status":"ok"`) {
		t.Errorf("expected status field in JSON, got %s", data)
	}
	if !strings.Contains(string(data), `"targetHost":"example.com:443"`) {
		t.Errorf("expected targetHost field in JSON, got %s", data)
	}
	var back TLSInfo
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal TLSInfo: %v", err)
	}
	if back.Status != "ok" || back.TargetHost != "example.com:443" {
		t.Errorf("roundtrip mismatch: %+v", back)
	}
}

func TestResponseOutput_TLS_Omitempty(t *testing.T) {
	out := ResponseOutput{RequestID: "r1"}
	data, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(data), `"tls"`) {
		t.Errorf("expected no tls field when TLS is nil, got %s", data)
	}
}
```

Add `"strings"` to the imports block at the top of `internal/model/types_test.go` (alongside `"encoding/json"` and `"testing"`).

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/model/ -run 'TestTLSInfoJSON|TestResponseOutput_TLS_Omitempty' -v`
Expected: FAIL with compile errors referencing `TLSInfo`, `TLSConnectionView`, `CertificateView`.

- [ ] **Step 3: Implement the types**

In `internal/model/types.go`, append at the end:

```go
type TLSInfo struct {
	Status              string               `json:"status"`
	Error               string               `json:"error,omitempty"`
	TargetHost          string               `json:"targetHost"`
	AttemptedServerName string               `json:"attemptedServerName,omitempty"`
	Connection          *TLSConnectionView   `json:"connection,omitempty"`
	Certificates        []CertificateView    `json:"certificates"`
}

type TLSConnectionView struct {
	Version            string   `json:"version"`
	CipherSuite        string   `json:"cipherSuite"`
	CipherSuiteName    string   `json:"cipherSuiteName"`
	NegotiatedProtocol string   `json:"negotiatedProtocol,omitempty"`
	ServerName         string   `json:"serverName"`
	Resumed            bool     `json:"resumed"`
	SCTs               []string `json:"scts,omitempty"`
	OCSPStapled        bool     `json:"ocspStapled"`
	PeerCertificates   int      `json:"peerCertificates"`
}

type CertificateView struct {
	Position              string   `json:"position"`
	Subject               string   `json:"subject"`
	Issuer                string   `json:"issuer"`
	SerialNumber          string   `json:"serialNumber"`
	Version               int      `json:"version"`
	SignatureAlgorithm    string   `json:"signatureAlgorithm"`
	NotBefore             string   `json:"notBefore"`
	NotAfter              string   `json:"notAfter"`
	IsExpired             bool     `json:"isExpired"`
	IsNotYetValid         bool     `json:"isNotYetValid"`
	DaysToExpiry          int      `json:"daysToExpiry"`
	SubjectKeyId          string   `json:"subjectKeyId"`
	AuthorityKeyId        string   `json:"authorityKeyId"`
	SANs                  []string `json:"sans"`
	KeyAlgorithm          string   `json:"keyAlgorithm"`
	KeySize               int      `json:"keySize"`
	PublicKeyPEM          string   `json:"publicKeyPem"`
	FingerprintSHA1       string   `json:"fingerprintSha1"`
	FingerprintSHA256     string   `json:"fingerprintSha256"`
	IsCA                  bool     `json:"isCa"`
	MaxPathLength         int      `json:"maxPathLength"`
	KeyUsage              []string `json:"keyUsage"`
	ExtendedKeyUsage      []string `json:"extendedKeyUsage"`
	CRLDistributionPoints []string `json:"crlDistributionPoints,omitempty"`
	Policies              []string `json:"policies,omitempty"`
	RawDER                string   `json:"rawDer"`
	PEM                   string   `json:"pem"`
	SignatureBytes        string   `json:"signatureBytes"`
}
```

In the `ResponseOutput` struct, add this field (place it just before `ErrorCode`):

```go
	TLS *TLSInfo `json:"tls,omitempty"`
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/model/ -run 'TestTLSInfoJSON|TestResponseOutput_TLS_Omitempty' -v`
Expected: PASS for both.

- [ ] **Step 5: Run full model package tests**

Run: `go test ./internal/model/ -v`
Expected: All tests pass.

- [ ] **Step 6: Commit**

```bash
git add internal/model/types.go internal/model/types_test.go
git commit -m "feat(model): add TLSInfo, TLSConnectionView, CertificateView types"
```

---

