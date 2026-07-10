package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestKeyValueJSON(t *testing.T) {
	kv := KeyValue{Key: "Content-Type", Value: "application/json", Enabled: true}
	data, err := json.Marshal(kv)
	if err != nil {
		t.Fatalf("failed to marshal KeyValue: %v", err)
	}
	var result KeyValue
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal KeyValue: %v", err)
	}
	if result.Key != kv.Key || result.Value != kv.Value || result.Enabled != kv.Enabled {
		t.Errorf("KeyValue mismatch: got %+v, want %+v", result, kv)
	}
}

func TestAppInfoJSON(t *testing.T) {
	info := AppInfo{Version: "1.0.0", Name: "otester"}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal AppInfo: %v", err)
	}
	var result AppInfo
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal AppInfo: %v", err)
	}
	if result != info {
		t.Errorf("AppInfo mismatch: got %+v, want %+v", result, info)
	}
}

func TestErrorCodeConstants(t *testing.T) {
	codes := []ErrorCode{
		ErrConfigFileNotFound,
		ErrConfigReadFailed,
		ErrConfigParseFailed,
		ErrConfigValidationFailed,
		ErrEndpointDisabled,
		ErrInvalidURL,
		ErrInvalidMethod,
		ErrInvalidBody,
		ErrRequestTimeout,
		ErrRequestCancelled,
		ErrDNSLookupFailed,
		ErrTLSHandshakeFailed,
		ErrConnectionFailed,
		ErrResponseTooLarge,
		ErrOAuthProfileNotFound,
		ErrOAuthTokenRequestFailed,
		ErrOAuthTokenResponseInvalid,
		ErrOAuthAccessTokenMissing,
		ErrAPIRequestFailed,
	}
	for _, c := range codes {
		if c == "" {
			t.Errorf("ErrorCode constant should not be empty")
		}
	}
}

func TestRequestInputJSON(t *testing.T) {
	input := RequestInput{
		RequestID:      "req-123",
		EndpointID:     "ep-456",
		Method:         "POST",
		URL:            "https://api.example.com/users",
		Headers:        []KeyValue{{Key: "Authorization", Value: "Bearer token", Enabled: true}},
		QueryParams:    []KeyValue{{Key: "page", Value: "1", Enabled: true}},
		BodyType:       "json",
		Body:           `{"name":"test"}`,
		TimeoutSeconds: 30,
		UseOAuth:       true,
		OAuthProfileID: "oauth-1",
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal RequestInput: %v", err)
	}
	var result RequestInput
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal RequestInput: %v", err)
	}
	if result.RequestID != input.RequestID || result.Method != input.Method || result.UseOAuth != input.UseOAuth {
		t.Errorf("RequestInput mismatch: got %+v, want %+v", result, input)
	}
}

func TestResponseOutputJSON(t *testing.T) {
	output := ResponseOutput{
		RequestID:     "req-123",
		StatusCode:    200,
		Status:        "OK",
		Headers:       map[string][]string{"Content-Type": {"application/json"}},
		Body:          `{"success":true}`,
		BodyTruncated: false,
		DurationMs:    150,
		SizeBytes:     17,
		ContentType:   "application/json",
		UsedOAuth:     false,
		TokenFromCache: false,
		ErrorCode:     "",
		ErrorMessage:  "",
	}
	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal ResponseOutput: %v", err)
	}
	var result ResponseOutput
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal ResponseOutput: %v", err)
	}
	if result.StatusCode != output.StatusCode || result.DurationMs != output.DurationMs {
		t.Errorf("ResponseOutput mismatch: got %+v, want %+v", result, output)
	}
}

func TestTokenStatusJSON(t *testing.T) {
	now := time.Now()
	status := TokenStatus{
		ProfileID: "oauth-1",
		HasToken:  true,
		ExpiresAt: &now,
		FromCache: false,
	}
	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("failed to marshal TokenStatus: %v", err)
	}
	var result TokenStatus
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal TokenStatus: %v", err)
	}
	if result.ProfileID != status.ProfileID || result.HasToken != status.HasToken || result.FromCache != status.FromCache {
		t.Errorf("TokenStatus mismatch: got %+v, want %+v", result, status)
	}
}

func TestTokenStatusOptionalExpiresAt(t *testing.T) {
	status := TokenStatus{
		ProfileID: "oauth-1",
		HasToken:  false,
		ExpiresAt: nil,
		FromCache: false,
	}
	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("failed to marshal TokenStatus: %v", err)
	}
	var result TokenStatus
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal TokenStatus: %v", err)
	}
	if result.ExpiresAt != nil {
		t.Errorf("TokenStatus ExpiresAt should be nil when not set")
	}
}
