package cache

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// FileEntry represents a cached file with its content and metadata
type FileEntry struct {
	Content   []byte
	ModTime   time.Time
	Size      int64
	CacheTime time.Time
}

// FileCache provides caching for file contents with mtime validation
type FileCache struct {
	cache sync.Map // path -> *FileEntry
	mu    sync.RWMutex
	stats CacheStats
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
}

// NewFileCache creates a new file cache
func NewFileCache() *FileCache {
	return &FileCache{}
}

// ReadFile reads a file from cache or disk, caching the result
func (fc *FileCache) ReadFile(path string) ([]byte, error) {
	// Try to get file info first
	stat, err := os.Stat(path)
	if err != nil {
		fc.incrementMisses()
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Check cache
	if cached, ok := fc.cache.Load(path); ok {
		entry := cached.(*FileEntry)
		// Validate cache entry with mtime
		if entry.ModTime.Equal(stat.ModTime()) && entry.Size == stat.Size() {
			fc.incrementHits()
			return entry.Content, nil
		}
		// Cache is stale, need to re-read
	}

	fc.incrementMisses()

	// Read file from disk
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Cache the content
	entry := &FileEntry{
		Content:   content,
		ModTime:   stat.ModTime(),
		Size:      stat.Size(),
		CacheTime: time.Now(),
	}
	fc.cache.Store(path, entry)

	return content, nil
}

// InvalidatePath removes a specific path from the cache
func (fc *FileCache) InvalidatePath(path string) {
	fc.cache.Delete(path)
	fc.mu.Lock()
	fc.stats.Evictions++
	fc.mu.Unlock()
}

// InvalidateAll clears the entire cache
func (fc *FileCache) InvalidateAll() {
	fc.cache.Range(func(key, value interface{}) bool {
		fc.cache.Delete(key)
		return true
	})
	fc.mu.Lock()
	fc.stats.Evictions++
	fc.mu.Unlock()
}

// GetStats returns cache statistics
func (fc *FileCache) GetStats() CacheStats {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.stats
}

// GetHitRate returns the cache hit rate as a percentage
func (fc *FileCache) GetHitRate() float64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	total := fc.stats.Hits + fc.stats.Misses
	if total == 0 {
		return 0
	}
	return float64(fc.stats.Hits) / float64(total) * 100
}

// incrementHits safely increments hit counter
func (fc *FileCache) incrementHits() {
	fc.mu.Lock()
	fc.stats.Hits++
	fc.mu.Unlock()
}

// incrementMisses safely increments miss counter
func (fc *FileCache) incrementMisses() {
	fc.mu.Lock()
	fc.stats.Misses++
	fc.mu.Unlock()
}

// CleanupOldEntries removes cache entries older than the specified duration
func (fc *FileCache) CleanupOldEntries(maxAge time.Duration) int {
	now := time.Now()
	removed := 0

	fc.cache.Range(func(key, value interface{}) bool {
		entry := value.(*FileEntry)
		if now.Sub(entry.CacheTime) > maxAge {
			fc.cache.Delete(key)
			removed++
		}
		return true
	})

	if removed > 0 {
		fc.mu.Lock()
		fc.stats.Evictions += uint64(removed)
		fc.mu.Unlock()
	}

	return removed
}

// StartCleanupWorker starts a background goroutine that periodically cleans old entries
func (fc *FileCache) StartCleanupWorker(interval, maxAge time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fc.CleanupOldEntries(maxAge)
			case <-stop:
				return
			}
		}
	}()

	return stop
}
