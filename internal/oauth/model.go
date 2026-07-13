package oauth

import "time"

type TokenResponse struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
}

type CachedToken struct {
	AccessToken string
	ExpiresAt   time.Time
}

type TokenStatus struct {
	ProfileID string     `json:"profileId"`
	HasToken  bool       `json:"hasToken"`
	FromCache bool       `json:"fromCache"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}