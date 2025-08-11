package extractor

import (
	"container/list"
	"sync"
	"time"
)

// Cache provides thread-safe LRU caching with TTL
type Cache interface {
	Get(key string) (*Config, bool)
	Set(key string, config *Config)
	Clear()
}

// LRUCache implements an LRU cache with TTL support
type LRUCache struct {
	mu       sync.RWMutex
	capacity int
	ttl      time.Duration
	items    map[string]*cacheItem
	order    *list.List
}

type cacheItem struct {
	key       string
	config    *Config
	timestamp time.Time
	element   *list.Element
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[string]*cacheItem),
		order:    list.New(),
	}
}

// Get retrieves a config from cache
func (c *LRUCache) Get(key string) (*Config, bool) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Check TTL
	if time.Since(item.timestamp) > c.ttl {
		c.mu.Lock()
		c.removeItem(item)
		c.mu.Unlock()
		return nil, false
	}

	// Move to front (most recently used)
	c.mu.Lock()
	c.order.MoveToFront(item.element)
	c.mu.Unlock()

	return item.config, true
}

// Set adds or updates a config in cache
func (c *LRUCache) Set(key string, config *Config) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing item
	if item, exists := c.items[key]; exists {
		item.config = config
		item.timestamp = time.Now()
		c.order.MoveToFront(item.element)
		return
	}

	// Add new item
	if len(c.items) >= c.capacity {
		// Evict least recently used
		c.evictOldest()
	}

	item := &cacheItem{
		key:       key,
		config:    config,
		timestamp: time.Now(),
	}
	item.element = c.order.PushFront(item)
	c.items[key] = item
}

// Clear removes all items from cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
	c.order.Init()
}

func (c *LRUCache) removeItem(item *cacheItem) {
	delete(c.items, item.key)
	c.order.Remove(item.element)
}

func (c *LRUCache) evictOldest() {
	oldest := c.order.Back()
	if oldest != nil {
		item := oldest.Value.(*cacheItem)
		c.removeItem(item)
	}
}

// NoOpCache provides a cache that doesn't cache anything (for testing)
type NoOpCache struct{}

func (n *NoOpCache) Get(key string) (*Config, bool) { return nil, false }
func (n *NoOpCache) Set(key string, config *Config)  {}
func (n *NoOpCache) Clear()                          {}