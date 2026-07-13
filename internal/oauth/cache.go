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

// GetStatus returns the token entry for a key. If refreshBefore > 0 and the
// cached entry is within that window of expiry, ok is false so callers treat
// the entry as missing (forcing a re-fetch). The returned bool mirrors Get:
// true only when the entry exists AND is still safely usable.
func (c *TokenCache) GetStatus(key string, refreshBefore time.Duration) (token *CachedToken, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, found := c.tokens[key]
	if !found {
		return nil, false
	}
	if refreshBefore > 0 && time.Until(t.ExpiresAt) < refreshBefore {
		return nil, false
	}
	return t, true
}

func BuildCacheKey(tokenURL, clientID, scope string) string {
	return tokenURL + "|" + clientID + "|" + scope
}