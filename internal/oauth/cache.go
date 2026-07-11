package oauth

import (
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type TokenCache struct {
	tokens map[string]*CachedToken
	mu     sync.RWMutex
	singleflight.Group
}

func NewTokenCache() *TokenCache {
	return &TokenCache{
		tokens: make(map[string]*CachedToken),
	}
}

func (c *TokenCache) Get(key string, refreshBefore time.Duration) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	token, ok := c.tokens[key]
	if !ok {
		return "", false
	}

	// Refresh if within refresh window
	if time.Until(token.ExpiresAt) < refreshBefore {
		return "", false
	}

	return token.AccessToken, true
}

func (c *TokenCache) Set(key string, token *CachedToken) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokens[key] = token
}

func (c *TokenCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.tokens, key)
}

func (c *TokenCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokens = make(map[string]*CachedToken)
}

// GetStatus returns the token status for a given key without exposing internal mutex
func (c *TokenCache) GetStatus(key string) (token *CachedToken, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	token, ok = c.tokens[key]
	return token, ok
}

func BuildCacheKey(tokenURL, clientID, scope string) string {
	return tokenURL + "|" + clientID + "|" + scope
}