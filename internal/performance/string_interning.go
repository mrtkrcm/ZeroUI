package performance

import (
	"sync"
)

// StringInterner provides memory-efficient string deduplication
type StringInterner struct {
	mu      sync.RWMutex
	strings map[string]string
	stats   *InternStats
}

// InternStats tracks interning performance metrics
type InternStats struct {
	Requests    int64
	Hits        int64
	TotalSaved  int64 // Bytes saved through interning
	UniqueCount int64
}

var (
	globalInterner *StringInterner
	internerOnce   sync.Once
)

// GlobalInterner returns the singleton string interner
func GlobalInterner() *StringInterner {
	internerOnce.Do(func() {
		globalInterner = NewStringInterner()
		
		// Pre-populate with common configuration strings
		commonStrings := []string{
			"true", "false", "enabled", "disabled", "default", "auto",
			"left", "right", "center", "top", "bottom", "none",
			"small", "medium", "large", "normal", "bold", "italic",
			"json", "yaml", "toml", "xml", "ini", "conf",
			"string", "number", "boolean", "array", "object",
			"required", "optional", "deprecated", "experimental",
		}
		
		for _, s := range commonStrings {
			globalInterner.intern(s)
		}
	})
	return globalInterner
}

// NewStringInterner creates a new string interner
func NewStringInterner() *StringInterner {
	return &StringInterner{
		strings: make(map[string]string, 256),
		stats:   &InternStats{},
	}
}

// Intern returns a canonical representation of the string
func (si *StringInterner) Intern(s string) string {
	si.stats.Requests++
	
	// Fast path: read lock only
	si.mu.RLock()
	if canonical, exists := si.strings[s]; exists {
		si.mu.RUnlock()
		si.stats.Hits++
		return canonical
	}
	si.mu.RUnlock()
	
	// Slow path: write lock for new string
	return si.intern(s)
}

func (si *StringInterner) intern(s string) string {
	si.mu.Lock()
	defer si.mu.Unlock()
	
	// Double-check after acquiring write lock
	if canonical, exists := si.strings[s]; exists {
		si.stats.Hits++
		return canonical
	}
	
	// Create canonical copy and store it
	canonical := string([]byte(s)) // Force allocation to ensure it's not sharing underlying array
	si.strings[s] = canonical
	si.stats.UniqueCount++
	si.stats.TotalSaved += int64(len(s))
	
	// Prevent unbounded growth
	if len(si.strings) > 10000 {
		si.evictOldEntries()
	}
	
	return canonical
}

// evictOldEntries removes some entries to prevent memory bloat
func (si *StringInterner) evictOldEntries() {
	// Simple eviction: remove every 4th entry
	// In production, you might use LRU or frequency-based eviction
	count := 0
	for k := range si.strings {
		if count%4 == 0 {
			delete(si.strings, k)
		}
		count++
		if count > 2500 { // Remove ~25% of entries
			break
		}
	}
}

// Stats returns current interning statistics
func (si *StringInterner) Stats() InternStats {
	si.mu.RLock()
	defer si.mu.RUnlock()
	stats := *si.stats
	stats.UniqueCount = int64(len(si.strings))
	return stats
}

// HitRate returns the cache hit rate as a percentage
func (si *StringInterner) HitRate() float64 {
	stats := si.Stats()
	if stats.Requests == 0 {
		return 0
	}
	return float64(stats.Hits) / float64(stats.Requests) * 100
}

// Clear removes all interned strings
func (si *StringInterner) Clear() {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.strings = make(map[string]string, 256)
	si.stats = &InternStats{}
}

// InternCommonConfigValue interns strings commonly found in config values
func InternCommonConfigValue(value string) string {
	// Only intern strings that are likely to be repeated
	if len(value) < 2 || len(value) > 64 {
		return value
	}
	
	// Check if it looks like a common config value
	if isCommonConfigValue(value) {
		return GlobalInterner().Intern(value)
	}
	
	return value
}

// isCommonConfigValue heuristically determines if a string is worth interning
func isCommonConfigValue(s string) bool {
	// Intern boolean-like values
	switch s {
	case "true", "false", "yes", "no", "on", "off", "enabled", "disabled":
		return true
	}
	
	// Intern common enum values
	if len(s) <= 16 && (containsOnlyAlphaNum(s) || containsCommonSeparators(s)) {
		return true
	}
	
	return false
}

func containsOnlyAlphaNum(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

func containsCommonSeparators(s string) bool {
	for _, r := range s {
		switch r {
		case '-', '_', '.', '/':
			continue
		default:
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
				return false
			}
		}
	}
	return true
}