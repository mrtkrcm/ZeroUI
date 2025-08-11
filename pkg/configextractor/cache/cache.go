package cache

import (
	"sync"
	"time"
)

// LRUCache implements an LRU cache with TTL for configuration data
type LRUCache struct {
	mu         sync.RWMutex
	entries    map[string]*entry
	order      *list         // Doubly linked list for LRU ordering
	maxSize    int           // Maximum number of entries
	defaultTTL time.Duration // Default TTL for entries
}

// entry represents a cache entry
type entry struct {
	key       string
	config    *Config
	timestamp time.Time
	ttl       time.Duration
	listNode  *node // Reference to position in LRU list
}

// Doubly linked list for efficient LRU operations
type list struct {
	head, tail *node
	size       int
}

type node struct {
	key        string
	prev, next *node
}

// NewLRU creates a new LRU cache with TTL
func NewLRU(maxSize int, defaultTTL time.Duration) *LRUCache {
	return &LRUCache{
		entries:    make(map[string]*entry, maxSize),
		order:      newList(),
		maxSize:    maxSize,
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a cached configuration
func (c *LRUCache) Get(key string) (*Config, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check TTL expiration
	if time.Since(entry.timestamp) > entry.ttl {
		c.removeEntry(key, entry)
		return nil, false
	}

	// Move to front (most recently used)
	c.order.moveToFront(entry.listNode)

	return entry.config, true
}

// Set stores a configuration in cache with specified TTL
func (c *LRUCache) Set(key string, config *Config, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if entry already exists
	if existingEntry, exists := c.entries[key]; exists {
		// Update existing entry
		existingEntry.config = config
		existingEntry.timestamp = time.Now()
		existingEntry.ttl = ttl
		c.order.moveToFront(existingEntry.listNode)
		return
	}

	// Evict LRU entry if cache is full
	if len(c.entries) >= c.maxSize {
		c.evictLRU()
	}

	// Add new entry
	listNode := c.order.addToFront(key)
	entry := &entry{
		key:       key,
		config:    config,
		timestamp: time.Now(),
		ttl:       ttl,
		listNode:  listNode,
	}

	c.entries[key] = entry
}

// Clear removes all expired entries
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	var toRemove []string

	for key, entry := range c.entries {
		if now.Sub(entry.timestamp) > entry.ttl {
			toRemove = append(toRemove, key)
		}
	}

	for _, key := range toRemove {
		if entry, exists := c.entries[key]; exists {
			c.removeEntry(key, entry)
		}
	}
}

// Stats returns cache statistics
func (c *LRUCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	expired := 0
	now := time.Now()

	for _, entry := range c.entries {
		if now.Sub(entry.timestamp) > entry.ttl {
			expired++
		}
	}

	return CacheStats{
		Size:     len(c.entries),
		MaxSize:  c.maxSize,
		Expired:  expired,
		HitRatio: 0, // Would need hit/miss tracking for accurate ratio
	}
}

// removeEntry removes an entry from cache and LRU list
func (c *LRUCache) removeEntry(key string, entry *entry) {
	delete(c.entries, key)
	c.order.remove(entry.listNode)
}

// evictLRU removes the least recently used entry
func (c *LRUCache) evictLRU() {
	if c.order.tail != nil {
		key := c.order.tail.key
		if entry, exists := c.entries[key]; exists {
			c.removeEntry(key, entry)
		}
	}
}

// Doubly linked list implementation for LRU ordering

// newList creates a new doubly linked list
func newList() *list {
	return &list{}
}

// addToFront adds a new node to the front of the list
func (l *list) addToFront(key string) *node {
	node := &node{key: key}

	if l.head == nil {
		// First node
		l.head = node
		l.tail = node
	} else {
		// Add to front
		node.next = l.head
		l.head.prev = node
		l.head = node
	}

	l.size++
	return node
}

// remove removes a node from the list
func (l *list) remove(node *node) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		l.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		l.tail = node.prev
	}

	l.size--
}

// moveToFront moves an existing node to the front
func (l *list) moveToFront(node *node) {
	if l.head == node {
		return // Already at front
	}

	// Remove from current position
	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		l.tail = node.prev
	}

	// Add to front
	node.prev = nil
	node.next = l.head
	if l.head != nil {
		l.head.prev = node
	}
	l.head = node

	// Update tail if this was the only node
	if l.tail == nil {
		l.tail = node
	}
}

// CacheStats provides cache statistics
type CacheStats struct {
	Size     int     `json:"size"`
	MaxSize  int     `json:"max_size"`
	Expired  int     `json:"expired"`
	HitRatio float64 `json:"hit_ratio"`
}

// InMemoryCache is a simple in-memory cache implementation
type InMemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*simpleEntry
}

type simpleEntry struct {
	config    *Config
	timestamp time.Time
	ttl       time.Duration
}

// NewInMemory creates a simple in-memory cache
func NewInMemory() *InMemoryCache {
	return &InMemoryCache{
		entries: make(map[string]*simpleEntry),
	}
}

// Get retrieves a cached configuration
func (c *InMemoryCache) Get(key string) (*Config, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check TTL
	if time.Since(entry.timestamp) > entry.ttl {
		return nil, false
	}

	return entry.config, true
}

// Set stores a configuration in cache
func (c *InMemoryCache) Set(key string, config *Config, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &simpleEntry{
		config:    config,
		timestamp: time.Now(),
		ttl:       ttl,
	}
}

// Clear removes expired entries
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.Sub(entry.timestamp) > entry.ttl {
			delete(c.entries, key)
		}
	}
}

// NoOpCache is a cache implementation that doesn't cache anything
type NoOpCache struct{}

// NewNoOp creates a no-op cache
func NewNoOp() *NoOpCache {
	return &NoOpCache{}
}

// Get always returns cache miss
func (c *NoOpCache) Get(key string) (*Config, bool) {
	return nil, false
}

// Set does nothing
func (c *NoOpCache) Set(key string, config *Config, ttl time.Duration) {
	// No-op
}

// Clear does nothing
func (c *NoOpCache) Clear() {
	// No-op
}
