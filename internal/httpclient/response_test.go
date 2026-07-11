package httpclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"otester/internal/model"
)

func TestReadBody_LargeBody(t *testing.T) {
	// Create a body much larger than maxBodySize to ensure truncation
	largeBody := strings.Repeat("x", maxBodySize*2)

	_, truncated, err := readBody(strings.NewReader(largeBody))
	if err != nil {
		t.Fatalf("readBody failed: %v", err)
	}

	if !truncated {
		t.Errorf("expected truncated=true, got truncated=%v", truncated)
	}
}

func TestReadBody_SmallBody(t *testing.T) {
	smallBody := "small body"

	data, truncated, err := readBody(strings.NewReader(smallBody))
	if err != nil {
		t.Fatalf("readBody failed: %v", err)
	}

	if truncated {
		t.Errorf("expected truncated=false, got truncated=%v", truncated)
	}
	if string(data) != smallBody {
		t.Errorf("expected data %q, got %q", smallBody, string(data))
	}
}

func TestParseResponse_Success(t *testing.T) {
	body := `{"message":"success"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
	defer server.Close()

	resp, err := server.Client().Get(server.URL)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	duration := 100 * time.Millisecond
	output, err := ParseResponse(resp, nil, "test-request", duration)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}

	if output.RequestID != "test-request" {
		t.Errorf("expected RequestID test-request, got %s", output.RequestID)
	}
	if output.StatusCode != http.StatusOK {
		t.Errorf("expected StatusCode 200, got %d", output.StatusCode)
	}
	if output.Body != body {
		t.Errorf("expected Body %q, got %q", body, output.Body)
	}
	if output.BodyTruncated {
		t.Errorf("expected BodyTruncated false, got true")
	}
	if output.DurationMs != duration.Milliseconds() {
		t.Errorf("expected DurationMs %d, got %d", duration.Milliseconds(), output.DurationMs)
	}
	if output.TLS == nil || output.TLS.Status != "no_tls_attempted" {
		t.Errorf("expected TLS.Status=no_tls_attempted, got %+v", output.TLS)
	}
}

func TestParseResponse_TruncatedBody(t *testing.T) {
	// Create a body larger than maxBodySize
	largeBody := strings.Repeat("x", maxBodySize+100)
	bodyReader := io.LimitReader(strings.NewReader(largeBody), maxBodySize+1)

	// Use a test server that returns a large body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		io.Copy(w, bodyReader)
	}))
	defer server.Close()

	resp, err := server.Client().Get(server.URL)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	output, err := ParseResponse(resp, nil, "test-truncated", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}

	if !output.BodyTruncated {
		t.Errorf("expected BodyTruncated true, got false")
	}
	if int64(len(output.Body)) != maxBodySize {
		t.Errorf("expected Body length %d, got %d", maxBodySize, len(output.Body))
	}
}

func TestHandleRequestError_Timeout(t *testing.T) {
	output, err := handleRequestError("test-timeout", context.DeadlineExceeded, nil, 30*time.Second, nil)
	if err != nil {
		t.Fatalf("handleRequestError returned error: %v", err)
	}

	if output.ErrorCode != string(model.ErrRequestTimeout) {
		t.Errorf("expected ErrorCode %s, got %s", model.ErrRequestTimeout, output.ErrorCode)
	}
	if output.ErrorMessage != "request timed out" {
		t.Errorf("expected ErrorMessage 'request timed out', got %s", output.ErrorMessage)
	}
}

func TestHandleRequestError_Cancelled(t *testing.T) {
	output, err := handleRequestError("test-cancelled", context.Canceled, nil, 100*time.Millisecond, nil)
	if err != nil {
		t.Fatalf("handleRequestError returned error: %v", err)
	}

	if output.ErrorCode != string(model.ErrRequestCancelled) {
		t.Errorf("expected ErrorCode %s, got %s", model.ErrRequestCancelled, output.ErrorCode)
	}
	if output.ErrorMessage != "request was cancelled" {
		t.Errorf("expected ErrorMessage 'request was cancelled', got %s", output.ErrorMessage)
	}
}

func TestHandleRequestError_ConnectionFailed(t *testing.T) {
	err := errors.New("connection refused")
	output, err2 := handleRequestError("test-conn", nil, err, 0, nil)
	if err2 != nil {
		t.Fatalf("handleRequestError returned error: %v", err2)
	}

	if output.ErrorCode != string(model.ErrConnectionFailed) {
		t.Errorf("expected ErrorCode %s, got %s", model.ErrConnectionFailed, output.ErrorCode)
	}
	if output.ErrorMessage != "connection refused" {
		t.Errorf("expected ErrorMessage 'connection refused', got %s", output.ErrorMessage)
	}
}
