package db

import (
	"sync"
	"time"
)

// HTMLCache represents a simple cache for HTML fragments
type HTMLCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
}

// cacheItem represents a cached item with expiration
type cacheItem struct {
	value      string
	expiration time.Time
}

// NewHTMLCache creates a new HTML cache
func NewHTMLCache() *HTMLCache {
	return &HTMLCache{
		items: make(map[string]*cacheItem),
	}
}

// Set adds an item to the cache with expiration
func (c *HTMLCache) Set(key, value string, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &cacheItem{
		value:      value,
		expiration: time.Now().Add(expiration),
	}
}

// Get retrieves an item from the cache
func (c *HTMLCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return "", false
	}

	// Check if the item has expired
	if time.Now().After(item.expiration) {
		return "", false
	}

	return item.value, true
}

// Cleanup removes expired items from the cache
func (c *HTMLCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiration) {
			delete(c.items, key)
		}
	}
}
