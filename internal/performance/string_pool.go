package performance

import (
	"strings"
	"sync"
)

// StringBuilderPool provides reusable string builders to reduce allocations
var StringBuilderPool = sync.Pool{
	New: func() interface{} {
		builder := &strings.Builder{}
		builder.Grow(1024) // Pre-allocate 1KB
		return builder
	},
}

// GetBuilder gets a string builder from the pool
func GetBuilder() *strings.Builder {
	return StringBuilderPool.Get().(*strings.Builder)
}

// PutBuilder returns a string builder to the pool
func PutBuilder(builder *strings.Builder) {
	if builder.Cap() < 64*1024 { // Don't pool builders that grew too large
		builder.Reset()
		StringBuilderPool.Put(builder)
	}
}

// SpacerCache caches common spacing strings to avoid repeated allocations
var spacerCache = make(map[int]string, 16)
var spacerMutex sync.RWMutex

// GetSpacer returns a cached spacer string of the given length
func GetSpacer(length int) string {
	if length <= 0 {
		return ""
	}
	
	spacerMutex.RLock()
	if spacer, exists := spacerCache[length]; exists {
		spacerMutex.RUnlock()
		return spacer
	}
	spacerMutex.RUnlock()
	
	// Create and cache the spacer
	spacer := strings.Repeat(" ", length)
	spacerMutex.Lock()
	if len(spacerCache) < 16 { // Limit cache size
		spacerCache[length] = spacer
	}
	spacerMutex.Unlock()
	
	return spacer
}