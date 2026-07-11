package httpclient

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"otester/internal/model"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("expected non-nil Client")
	}
	if client.client == nil {
		t.Fatal("expected non-nil http.Client")
	}
	if client.cancelFuncs == nil {
		t.Fatal("expected non-nil cancelFuncs map")
	}
	if client.client.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", client.client.Timeout)
	}
}

func TestClient_CancelRequest(t *testing.T) {
	client := NewClient()

	// Cancel a non-existent request should not error
	err := client.CancelRequest("non-existent-id")
	if err != nil {
		t.Errorf("expected no error for non-existent request, got %v", err)
	}
}

func TestClient_DoRequest_InvalidURL(t *testing.T) {
	client := NewClient()

	input := &model.RequestInput{
		RequestID:      "test-invalid-url",
		Method:         "GET",
		URL:            "://invalid-url",
		TimeoutSeconds: 30,
	}

	output, err := client.DoRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("DoRequest returned error: %v", err)
	}

	if output.ErrorCode != string(model.ErrInvalidURL) {
		t.Errorf("expected ErrorCode %s, got %s", model.ErrInvalidURL, output.ErrorCode)
	}
}

func TestClient_DoRequest_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient()
	client.client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	input := &model.RequestInput{
		RequestID:      "test-success",
		Method:         "GET",
		URL:            server.URL,
		TimeoutSeconds: 30,
	}

	output, err := client.DoRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("DoRequest returned error: %v", err)
	}

	if output.ErrorCode != "" {
		t.Errorf("expected no ErrorCode, got %s", output.ErrorCode)
	}
	if output.StatusCode != http.StatusOK {
		t.Errorf("expected StatusCode 200, got %d", output.StatusCode)
	}
	if output.TLS == nil {
		t.Fatal("expected TLS info to be populated for HTTPS server")
	}
	if output.TLS.Status != "ok" {
		t.Errorf("expected TLS.Status=ok, got %s", output.TLS.Status)
	}
	if len(output.TLS.Certificates) < 1 {
		t.Errorf("expected >= 1 certificate, got %d", len(output.TLS.Certificates))
	}
}

func TestClient_DoRequest_Timeout(t *testing.T) {
	// Server that delays response longer than the timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()

	input := &model.RequestInput{
		RequestID:      "test-timeout",
		Method:         "GET",
		URL:            server.URL,
		TimeoutSeconds: 1, // 1 second timeout
	}

	output, err := client.DoRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("DoRequest returned error: %v", err)
	}

	if output.ErrorCode != string(model.ErrRequestTimeout) {
		t.Errorf("expected ErrorCode %s, got %s", model.ErrRequestTimeout, output.ErrorCode)
	}
	if output.TLS == nil {
		t.Fatal("expected TLS info to be populated even on timeout")
	}
	if output.TLS.Status != "no_tls_attempted" {
		t.Errorf("expected TLS.Status=no_tls_attempted for timeout, got %s", output.TLS.Status)
	}
}

func TestClient_DoRequest_Cancellation(t *testing.T) {
	// Server that responds slowly - we'll cancel before it responds
	done := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block until the connection is closed or context is cancelled
		select {
		case <-r.Context().Done():
		case <-done:
		}
	}))
	defer func() { close(done); server.Close() }()

	client := NewClient()

	input := &model.RequestInput{
		RequestID:      "test-cancel",
		Method:         "GET",
		URL:            server.URL,
		TimeoutSeconds: 30,
	}

	// Start request in goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	var output *model.ResponseOutput
	var err error
	go func() {
		output, err = client.DoRequest(context.Background(), input)
		wg.Done()
	}()

	// Give the request a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the request
	client.CancelRequest(input.RequestID)

	// Wait for the request to complete
	wg.Wait()

	if err != nil {
		t.Fatalf("DoRequest returned error: %v", err)
	}

	if output.ErrorCode != string(model.ErrRequestCancelled) {
		t.Errorf("expected ErrorCode %s, got %s (status: %d)", model.ErrRequestCancelled, output.ErrorCode, output.StatusCode)
	}
}

func TestClient_DoRequest_QueryParams(t *testing.T) {
	var receivedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"query":"` + r.URL.RawQuery + `"}`))
	}))
	defer server.Close()

	client := NewClient()

	input := &model.RequestInput{
		RequestID: "test-query-params",
		Method:    "GET",
		URL:       server.URL,
		QueryParams: []model.KeyValue{
			{Key: "foo", Value: "bar", Enabled: true},
			{Key: "baz", Value: "qux", Enabled: true},
		},
		TimeoutSeconds: 30,
	}

	output, err := client.DoRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("DoRequest returned error: %v", err)
	}

	if output.ErrorCode != "" {
		t.Errorf("expected no ErrorCode, got %s", output.ErrorCode)
	}
	// Query param order may vary, so check both params are present
	if !strings.Contains(receivedQuery, "foo=bar") || !strings.Contains(receivedQuery, "baz=qux") {
		t.Errorf("expected query containing 'foo=bar' and 'baz=qux', got %q", receivedQuery)
	}
}

func TestClient_DoRequest_Headers(t *testing.T) {
	var receivedHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Custom-Header")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"header":"` + receivedHeader + `"}`))
	}))
	defer server.Close()

	client := NewClient()

	input := &model.RequestInput{
		RequestID: "test-headers",
		Method:    "GET",
		URL:       server.URL,
		Headers: []model.KeyValue{
			{Key: "X-Custom-Header", Value: "my-value", Enabled: true},
		},
		TimeoutSeconds: 30,
	}

	output, err := client.DoRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("DoRequest returned error: %v", err)
	}

	if output.ErrorCode != "" {
		t.Errorf("expected no ErrorCode, got %s", output.ErrorCode)
	}
	if receivedHeader != "my-value" {
		t.Errorf("expected header 'my-value', got %q", receivedHeader)
	}
}

func TestClient_DoRequest_TLSHandshakeFailure(t *testing.T) {
	// httptest server uses "example.com" as its cert's DNS name; we connect
	// via 127.0.0.1 with hostname verification on, which causes x509 hostname mismatch.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	// Use the default transport (does NOT skip verify) so the handshake fails.

	// Replace the host portion of server.URL with 127.0.0.1 but keep the port.
	u, _ := url.Parse(server.URL)
	host := "127.0.0.1"
	inputURL := strings.Replace(server.URL, u.Hostname(), host, 1)

	input := &model.RequestInput{
		RequestID:      "test-tls-fail",
		Method:         "GET",
		URL:            inputURL,
		TimeoutSeconds: 5,
	}

	output, err := client.DoRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("DoRequest returned error: %v", err)
	}

	if output.TLS == nil {
		t.Fatal("expected TLS info on handshake failure")
	}
	if output.TLS.Status != "handshake_failed" {
		t.Errorf("expected TLS.Status=handshake_failed, got %s", output.TLS.Status)
	}
	if !strings.Contains(output.TLS.Error, "x509:") && !strings.Contains(output.TLS.Error, "tls:") {
		t.Errorf("expected tls: or x509: in error, got %q", output.TLS.Error)
	}
}

func TestClient_DoRequest_PostWithBody(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"received":"` + receivedBody + `"}`))
	}))
	defer server.Close()

	client := NewClient()

	input := &model.RequestInput{
		RequestID:      "test-post-body",
		Method:         "POST",
		URL:            server.URL,
		BodyType:       "json",
		Body:           `{"name":"test"}`,
		TimeoutSeconds: 30,
	}

	output, err := client.DoRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("DoRequest returned error: %v", err)
	}

	if output.ErrorCode != "" {
		t.Errorf("expected no ErrorCode, got %s", output.ErrorCode)
	}
	if receivedBody != `{"name":"test"}` {
		t.Errorf("expected body %q, got %q", `{"name":"test"}`, receivedBody)
	}
}
