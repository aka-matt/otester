### Task 1: Extend `TLSInfo` with `ValidationSkipped` and `OriginalError`

**Files:**
- Modify: `internal/model/types.go` (append two fields to `TLSInfo`)
- Modify: `internal/model/types_test.go` (add JSON round-trip test)

**Interfaces (after this task):**
- `model.TLSInfo.ValidationSkipped bool` — `json:"validationSkipped"`
- `model.TLSInfo.OriginalError string` — `json:"originalError,omitempty"`

- [ ] **Step 1: Write the failing test**

Append to `internal/model/types_test.go`:

```go
func TestTLSInfo_ValidationSkipped_Roundtrip(t *testing.T) {
	info := TLSInfo{
		Status:            "ok",
		TargetHost:        "self-signed.example:443",
		ValidationSkipped: true,
		OriginalError:     "x509: certificate signed by unknown authority",
		Certificates:      []CertificateView{},
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal TLSInfo: %v", err)
	}
	if !strings.Contains(string(data), `"validationSkipped":true`) {
		t.Errorf("expected validationSkipped in JSON, got %s", data)
	}
	if !strings.Contains(string(data), `"originalError":"x509: certificate signed by unknown authority"`) {
		t.Errorf("expected originalError in JSON, got %s", data)
	}
	var back TLSInfo
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal TLSInfo: %v", err)
	}
	if !back.ValidationSkipped {
		t.Error("ValidationSkipped did not round-trip as true")
	}
	if back.OriginalError != "x509: certificate signed by unknown authority" {
		t.Errorf("OriginalError round-trip mismatch: %q", back.OriginalError)
	}
}

func TestTLSInfo_DefaultFields_Omit(t *testing.T) {
	info := TLSInfo{Status: "ok", TargetHost: "example.com"}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(data), `"validationSkipped"`) {
		t.Errorf("expected validationSkipped omitted when false, got %s", data)
	}
	if strings.Contains(string(data), `"originalError"`) {
		t.Errorf("expected originalError omitted when empty, got %s", data)
	}
}
```

(If `strings` is not already imported in `types_test.go`, add it to the imports block.)

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/model/ -run 'TestTLSInfo_ValidationSkipped_Roundtrip|TestTLSInfo_DefaultFields_Omit' -v`
Expected: FAIL with compile errors referencing `ValidationSkipped` / `OriginalError` on `TLSInfo`.

- [ ] **Step 3: Implement the new fields**

In `internal/model/types.go`, find the existing `TLSInfo` struct and add two fields. The struct currently looks like:

```go
type TLSInfo struct {
	Status              string               `json:"status"`
	Error               string               `json:"error,omitempty"`
	TargetHost          string               `json:"targetHost"`
	AttemptedServerName string               `json:"attemptedServerName,omitempty"`
	Connection          *TLSConnectionView   `json:"connection,omitempty"`
	Certificates        []CertificateView    `json:"certificates"`
}
```

Add the two new fields just after `Certificates`:

```go
type TLSInfo struct {
	Status              string             `json:"status"`
	Error               string             `json:"error,omitempty"`
	TargetHost          string             `json:"targetHost"`
	AttemptedServerName string             `json:"attemptedServerName,omitempty"`
	Connection          *TLSConnectionView `json:"connection,omitempty"`
	Certificates        []CertificateView  `json:"certificates"`
	ValidationSkipped   bool               `json:"validationSkipped"`
	OriginalError       string             `json:"originalError,omitempty"`
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/model/ -run 'TestTLSInfo_ValidationSkipped_Roundtrip|TestTLSInfo_DefaultFields_Omit' -v`
Expected: PASS for both.

- [ ] **Step 5: Run full model package tests**

Run: `go test ./internal/model/ -v`
Expected: All tests pass (including pre-existing `TestTLSInfoJSON` and `TestResponseOutput_TLS_Omitempty`).

- [ ] **Step 6: Commit**

```bash
git add internal/model/types.go internal/model/types_test.go
git commit -m "feat(model): add ValidationSkipped and OriginalError to TLSInfo"
```

---

