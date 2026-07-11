package security

import (
	"strings"
	"testing"
)

func TestRedactBearerToken(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4ifQ.foobar"
	redacted := RedactBearerToken(token)
	if redacted == token {
		t.Error("token should be redacted")
	}
	// token[:6] = eyJhbG, token[len-4:] = obar
	if redacted != "eyJhbG...obar" {
		t.Errorf("unexpected redaction format: %s", redacted)
	}
}

func TestRedactAuthorizationHeader(t *testing.T) {
	header := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.foobar"
	redacted := RedactAuthorizationHeader(header)
	if strings.Contains(redacted, "eyJhbG") && !strings.Contains(redacted, "eyJh") {
		t.Error("token should be redacted in header")
	}
}

func TestRedactURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "client secret redacted",
			input:    "https://api.example.com/oauth?client_secret=mysecret&code=authcode",
			expected: "https://api.example.com/oauth?client_secret=******&code=authcode",
		},
		{
			name:     "access token redacted",
			input:    "https://api.example.com/token?access_token=mytoken&scope=read",
			expected: "https://api.example.com/token?access_token=******&scope=read",
		},
		{
			name:     "no secrets",
			input:    "https://api.example.com/users/123",
			expected: "https://api.example.com/users/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactURL(tt.input)
			if result != tt.expected {
				t.Errorf("got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestRedactHeaders(t *testing.T) {
	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.foobar"},
	}

	redacted := RedactHeaders(headers)

	if redacted["Content-Type"][0] != "application/json" {
		t.Error("Content-Type should not be modified")
	}

	auth := redacted["Authorization"][0]
	if strings.Contains(auth, "eyJhbG") && !strings.Contains(auth, "eyJh") {
		t.Error("Authorization header token should be redacted")
	}
}

func TestRedactJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "client secret redacted",
			input:    `{"client_secret":"mysecret", "grant_type": "password"}`,
			expected: `{"client_secret":"******", "grant_type": "password"}`,
		},
		{
			name:     "access token redacted",
			input:    `{"access_token":"mytoken", "token_type": "Bearer"}`,
			expected: `{"access_token":"******", "token_type": "Bearer"}`,
		},
		{
			name:     "no secrets",
			input:    `{"username": "user123", "email": "test@example.com"}`,
			expected: `{"username": "user123", "email": "test@example.com"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactJSON(tt.input)
			if result != tt.expected {
				t.Errorf("got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestRedactBearerTokenShort(t *testing.T) {
	short := "short"
	redacted := RedactBearerToken(short)
	if redacted != "******" {
		t.Errorf("short token should be fully redacted, got %s", redacted)
	}
}
