package graph

import (
	"sync"
	"time"

	"github.com/auriora/onemount/internal/logging"
)

// CacheEntry represents a cached API response with expiration
type CacheEntry struct {
	Data      []byte
	ExpiresAt time.Time
}

// ResponseCache implements a TTL-based cache for API responses
type ResponseCache struct {
	cache      map[string]CacheEntry
	mutex      sync.RWMutex
	defaultTTL time.Duration
}

// NewResponseCache creates a new response cache with the specified default TTL
func NewResponseCache(defaultTTL time.Duration) *ResponseCache {
	cache := &ResponseCache{
		cache:      make(map[string]CacheEntry),
		defaultTTL: defaultTTL,
	}

	// Start a background goroutine to clean up expired entries
	go cache.cleanupLoop()

	return cache
}

// Get retrieves a cached response if it exists and is not expired
func (c *ResponseCache) Get(key string) ([]byte, bool) {
	c.mutex.RLock()
	entry, exists := c.cache[key]
	c.mutex.RUnlock()

	if !exists {
		return nil, false
	}

	// Check if the entry has expired
	if time.Now().After(entry.ExpiresAt) {
		// Remove expired entry
		c.mutex.Lock()
		delete(c.cache, key)
		c.mutex.Unlock()
		return nil, false
	}

	return entry.Data, true
}

// Set adds or updates a cached response with the default TTL
func (c *ResponseCache) Set(key string, data []byte) {
	c.SetWithTTL(key, data, c.defaultTTL)
}

// SetWithTTL adds or updates a cached response with a specific TTL
func (c *ResponseCache) SetWithTTL(key string, data []byte, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Invalidate removes a specific entry from the cache
func (c *ResponseCache) Invalidate(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.cache, key)
}

// InvalidatePrefix removes all entries with keys starting with the given prefix
func (c *ResponseCache) InvalidatePrefix(prefix string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key := range c.cache {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.cache, key)
		}
	}
}

// Clear removes all entries from the cache
func (c *ResponseCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]CacheEntry)
}

// Size returns the number of entries in the cache
func (c *ResponseCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.cache)
}

// cleanupLoop periodically removes expired entries from the cache
func (c *ResponseCache) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpired()
	}
}

// cleanupExpired removes all expired entries from the cache
func (c *ResponseCache) cleanupExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for key, entry := range c.cache {
		if now.After(entry.ExpiresAt) {
			delete(c.cache, key)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		logging.Debug().Int("count", expiredCount).Msg("Removed expired entries from response cache")
	}
}
