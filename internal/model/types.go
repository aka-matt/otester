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
	TLS           *TLSInfo          `json:"tls,omitempty"`
	ErrorCode     string            `json:"errorCode"`
	ErrorMessage  string            `json:"errorMessage"`
}

type TokenStatus struct {
	ProfileID string    `json:"profileId"`
	HasToken  bool      `json:"hasToken"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	FromCache bool      `json:"fromCache"`
}

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
