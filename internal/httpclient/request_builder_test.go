package httpclient

import (
	"context"
	"io"
	"testing"

	"otester/internal/model"
)

func TestBuildRequest_GET(t *testing.T) {
	input := &model.RequestInput{
		RequestID:      "test-1",
		Method:         "GET",
		URL:            "https://httpbin.org/get?foo=bar",
		TimeoutSeconds: 30,
	}

	req, err := BuildRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}
	if req.Method != "GET" {
		t.Errorf("expected GET, got %s", req.Method)
	}
	if req.URL.Query().Get("foo") != "bar" {
		t.Errorf("expected query foo=bar, got %v", req.URL.Query())
	}
}

func TestBuildRequest_POST_JSON(t *testing.T) {
	input := &model.RequestInput{
		RequestID:      "test-2",
		Method:         "POST",
		URL:            "https://httpbin.org/post",
		BodyType:       "json",
		Body:           `{"name":"test"}`,
		TimeoutSeconds: 30,
	}

	req, err := BuildRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}
	if req.Method != "POST" {
		t.Errorf("expected POST, got %s", req.Method)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", req.Header.Get("Content-Type"))
	}
}

func TestBuildRequest_QueryParams(t *testing.T) {
	input := &model.RequestInput{
		RequestID: "test-3",
		Method:    "GET",
		URL:       "https://httpbin.org/get",
		QueryParams: []model.KeyValue{
			{Key: "key1", Value: "val1", Enabled: true},
			{Key: "key2", Value: "val2", Enabled: false},
			{Key: "key3", Value: "val3", Enabled: true},
		},
		TimeoutSeconds: 30,
	}

	req, err := BuildRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if req.URL.Query().Get("key1") != "val1" {
		t.Errorf("expected key1=val1, got %v", req.URL.Query().Get("key1"))
	}
	if req.URL.Query().Get("key2") != "" {
		t.Errorf("expected key2 to be empty (disabled), got %v", req.URL.Query().Get("key2"))
	}
	if req.URL.Query().Get("key3") != "val3" {
		t.Errorf("expected key3=val3, got %v", req.URL.Query().Get("key3"))
	}
}

func TestBuildRequest_Headers(t *testing.T) {
	input := &model.RequestInput{
		RequestID: "test-4",
		Method:    "GET",
		URL:       "https://httpbin.org/get",
		Headers: []model.KeyValue{
			{Key: "X-Custom-Header", Value: "custom-value", Enabled: true},
			{Key: "X-Disabled-Header", Value: "disabled-value", Enabled: false},
		},
		TimeoutSeconds: 30,
	}

	req, err := BuildRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if req.Header.Get("X-Custom-Header") != "custom-value" {
		t.Errorf("expected X-Custom-Header=custom-value, got %s", req.Header.Get("X-Custom-Header"))
	}
	if req.Header.Get("X-Disabled-Header") != "" {
		t.Errorf("expected X-Disabled-Header to be empty, got %s", req.Header.Get("X-Disabled-Header"))
	}
}

func TestBuildRequest_InvalidURL(t *testing.T) {
	input := &model.RequestInput{
		RequestID:      "test-5",
		Method:         "GET",
		URL:            "://invalid-url",
		TimeoutSeconds: 30,
	}

	_, err := BuildRequest(context.Background(), input)
	if err == nil {
		t.Errorf("expected error for invalid URL, got nil")
	}
}

func TestBuildRequest_TextBodyType(t *testing.T) {
	input := &model.RequestInput{
		RequestID:      "test-6",
		Method:         "POST",
		URL:            "https://httpbin.org/post",
		BodyType:       "text",
		Body:           "plain text body",
		TimeoutSeconds: 30,
	}

	req, err := BuildRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}
	if req.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("expected Content-Type text/plain, got %s", req.Header.Get("Content-Type"))
	}
}

func TestBuildRequest_FormURLEncodedBodyType(t *testing.T) {
	input := &model.RequestInput{
		RequestID:      "test-7",
		Method:         "POST",
		URL:            "https://httpbin.org/post",
		BodyType:       "x-www-form-urlencoded",
		Body:           "name=test&value=123",
		TimeoutSeconds: 30,
	}

	req, err := BuildRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}
	if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		t.Errorf("expected Content-Type application/x-www-form-urlencoded, got %s", req.Header.Get("Content-Type"))
	}
}

func TestBuildRequest_XMLBodyType(t *testing.T) {
	input := &model.RequestInput{
		RequestID:      "test-xml",
		Method:         "POST",
		URL:            "https://httpbin.org/post",
		BodyType:       "xml",
		Body:           `<?xml version="1.0"?><root><item/></root>`,
		TimeoutSeconds: 30,
	}

	req, err := BuildRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}
	if req.Header.Get("Content-Type") != "application/xml" {
		t.Errorf("expected Content-Type application/xml, got %s", req.Header.Get("Content-Type"))
	}
}

// User-supplied Content-Type must win over the bodyType default. SOAP and
// many XML APIs expect text/xml (or application/soap+xml); unconditionally
// overwriting with application/xml causes the server to reject the request.
func TestBuildRequest_UserContentTypeWinsOverBodyType(t *testing.T) {
	input := &model.RequestInput{
		RequestID: "test-user-ct",
		Method:    "POST",
		URL:       "https://httpbin.org/post",
		BodyType:  "xml",
		Body:      `<root/>`,
		Headers: []model.KeyValue{
			{Key: "Content-Type", Value: "text/xml; charset=utf-8", Enabled: true},
		},
		TimeoutSeconds: 30,
	}

	req, err := BuildRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}
	got := req.Header.Get("Content-Type")
	if got != "text/xml; charset=utf-8" {
		t.Errorf("user Content-Type must be preserved, got %q", got)
	}
}

func TestBuildRequest_XMLBodyIsTransmitted(t *testing.T) {
	body := `<?xml version="1.0"?><root><item id="1"/></root>`
	input := &model.RequestInput{
		RequestID:      "test-xml-body",
		Method:         "POST",
		URL:            "https://httpbin.org/post",
		BodyType:       "xml",
		Body:           body,
		TimeoutSeconds: 30,
	}

	req, err := BuildRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}
	if req.ContentLength != int64(len(body)) {
		t.Errorf("ContentLength want %d, got %d", len(body), req.ContentLength)
	}
	got, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("ReadAll body: %v", err)
	}
	if string(got) != body {
		t.Errorf("body want %q, got %q", body, string(got))
	}
	if req.GetBody == nil {
		t.Error("GetBody must be set so TLS-retry can re-send the body")
	}
}
