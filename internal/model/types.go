package model

import "time"

type KeyValue struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

type AppInfo struct {
	Version string `json:"version"`
	Name    string `json:"name"`
}

type ErrorCode string

const (
	ErrConfigFileNotFound        ErrorCode = "CONFIG_FILE_NOT_FOUND"
	ErrConfigReadFailed          ErrorCode = "CONFIG_READ_FAILED"
	ErrConfigParseFailed         ErrorCode = "CONFIG_PARSE_FAILED"
	ErrConfigValidationFailed    ErrorCode = "CONFIG_VALIDATION_FAILED"
	ErrEndpointDisabled          ErrorCode = "ENDPOINT_DISABLED"
	ErrInvalidURL                ErrorCode = "INVALID_URL"
	ErrInvalidMethod             ErrorCode = "INVALID_METHOD"
	ErrInvalidBody               ErrorCode = "INVALID_BODY"
	ErrRequestTimeout            ErrorCode = "REQUEST_TIMEOUT"
	ErrRequestCancelled          ErrorCode = "REQUEST_CANCELLED"
	ErrDNSLookupFailed           ErrorCode = "DNS_LOOKUP_FAILED"
	ErrTLSHandshakeFailed        ErrorCode = "TLS_HANDSHAKE_FAILED"
	ErrConnectionFailed          ErrorCode = "CONNECTION_FAILED"
	ErrResponseTooLarge          ErrorCode = "RESPONSE_TOO_LARGE"
	ErrOAuthProfileNotFound      ErrorCode = "OAUTH_PROFILE_NOT_FOUND"
	ErrOAuthTokenRequestFailed   ErrorCode = "OAUTH_TOKEN_REQUEST_FAILED"
	ErrOAuthTokenResponseInvalid ErrorCode = "OAUTH_TOKEN_RESPONSE_INVALID"
	ErrOAuthAccessTokenMissing   ErrorCode = "OAUTH_ACCESS_TOKEN_MISSING"
	ErrAPIRequestFailed          ErrorCode = "API_REQUEST_FAILED"
)

type RequestInput struct {
	RequestID      string     `json:"requestId"`
	EndpointID     string     `json:"endpointId"`
	Method         string     `json:"method"`
	URL            string     `json:"url"`
	Headers        []KeyValue `json:"headers"`
	QueryParams    []KeyValue `json:"queryParams"`
	BodyType       string     `json:"bodyType"`
	Body           string     `json:"body"`
	TimeoutSeconds int        `json:"timeoutSeconds"`
	UseOAuth       bool       `json:"useOAuth"`
	OAuthProfileID string     `json:"oauthProfileId"`
}

type ResponseOutput struct {
	RequestID     string            `json:"requestId"`
	StatusCode    int               `json:"statusCode"`
	Status        string            `json:"status"`
	Headers       map[string][]string `json:"headers"`
	Body          string            `json:"body"`
	BodyTruncated bool              `json:"bodyTruncated"`
	DurationMs    int64             `json:"durationMs"`
	SizeBytes     int64             `json:"sizeBytes"`
	ContentType   string            `json:"contentType"`
	UsedOAuth     bool              `json:"usedOAuth"`
	TokenFromCache bool             `json:"tokenFromCache"`
	ErrorCode     string            `json:"errorCode"`
	ErrorMessage  string            `json:"errorMessage"`
}

type TokenStatus struct {
	ProfileID string    `json:"profileId"`
	HasToken  bool      `json:"hasToken"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	FromCache bool      `json:"fromCache"`
}
