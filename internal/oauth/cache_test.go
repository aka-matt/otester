package oauth

import (
	"strings"
	"testing"
	"time"
)

func TestTokenCache_GetSet(t *testing.T) {
	cache := NewTokenCache()
	cache.Set("key1", &CachedToken{
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	})

	token, ok := cache.Get("key1", 30*time.Second)
	if !ok {
		t.Fatal("expected token to be found")
	}
	if token != "test-token" {
		t.Errorf("expected 'test-token', got '%s'", token)
	}
}

func TestTokenCache_GetNotFound(t *testing.T) {
	cache := NewTokenCache()

	token, ok := cache.Get("nonexistent", 30*time.Second)
	if ok {
		t.Error("expected token not to be found")
	}
	if token != "" {
		t.Errorf("expected empty token, got '%s'", token)
	}
}

func TestTokenCache_GetExpired(t *testing.T) {
	cache := NewTokenCache()
	cache.Set("key1", &CachedToken{
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(30 * time.Second), // Expires in 30 seconds
	})

	// Request with 60 second refreshBefore - token should be considered expired
	token, ok := cache.Get("key1", 60*time.Second)
	if ok {
		t.Error("expected token to be considered expired")
	}
	if token != "" {
		t.Errorf("expected empty token, got '%s'", token)
	}
}

func TestTokenCache_Delete(t *testing.T) {
	cache := NewTokenCache()
	cache.Set("key1", &CachedToken{
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	})

	cache.Delete("key1")

	token, ok := cache.Get("key1", 30*time.Second)
	if ok {
		t.Error("expected token not to be found after delete")
	}
	if token != "" {
		t.Errorf("expected empty token, got '%s'", token)
	}
}

func TestTokenCache_Clear(t *testing.T) {
	cache := NewTokenCache()
	cache.Set("key1", &CachedToken{
		AccessToken: "test-token-1",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	})
	cache.Set("key2", &CachedToken{
		AccessToken: "test-token-2",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	})

	cache.Clear()

	_, ok1 := cache.Get("key1", 30*time.Second)
	_, ok2 := cache.Get("key2", 30*time.Second)
	if ok1 || ok2 {
		t.Error("expected all tokens to be cleared")
	}
}

func TestBuildCacheKey(t *testing.T) {
	key := BuildCacheKey("https://login.microsoftonline.com/tenant/oauth2/v2.0/token",
		"client-id", "https://graph.microsoft.com/.default")
	if !strings.Contains(key, "client-id") {
		t.Error("cache key should contain client-id")
	}
	if !strings.Contains(key, "https://login.microsoftonline.com/tenant/oauth2/v2.0/token") {
		t.Error("cache key should contain token URL")
	}
	if !strings.Contains(key, "https://graph.microsoft.com/.default") {
		t.Error("cache key should contain scope")
	}
	if key != "https://login.microsoftonline.com/tenant/oauth2/v2.0/token|client-id|https://graph.microsoft.com/.default" {
		t.Errorf("cache key has unexpected format: %s", key)
	}
}