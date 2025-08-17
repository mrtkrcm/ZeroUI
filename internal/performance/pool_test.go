package performance

import (
	"sync"
	"testing"
)

// TestStringBoolMapPool tests the string bool map pooling
func TestStringBoolMapPool(t *testing.T) {
	// Test basic get and put
	m1 := GetStringBoolMap()
	if m1 == nil {
		t.Fatal("GetStringBoolMap returned nil")
	}

	// Add some data
	m1["test1"] = true
	m1["test2"] = false

	// Return to pool
	PutStringBoolMap(m1)

	// Get another map - should be cleared
	m2 := GetStringBoolMap()
	if len(m2) != 0 {
		t.Errorf("Expected empty map, got %d items", len(m2))
	}

	// Test that it's the same map (pooling works)
	m2["new"] = true
	PutStringBoolMap(m2)

	m3 := GetStringBoolMap()
	if len(m3) != 0 {
		t.Error("Map not properly cleared before reuse")
	}
}

// TestStringBoolMapPoolConcurrency tests concurrent access to the pool
func TestStringBoolMapPoolConcurrency(t *testing.T) {
	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				m := GetStringBoolMap()
				m["key1"] = true
				m["key2"] = false
				if len(m) != 2 {
					t.Errorf("Goroutine %d: unexpected map size %d", id, len(m))
				}
				PutStringBoolMap(m)
			}
		}(i)
	}

	wg.Wait()
}

// TestStringBoolMapPoolLargeMap tests that large maps are not returned to pool
func TestStringBoolMapPoolLargeMap(t *testing.T) {
	m := GetStringBoolMap()

	// Add more than 1024 items (pool limit)
	for i := 0; i < 1025; i++ {
		m[string(rune(i))] = true
	}

	PutStringBoolMap(m) // Should not be pooled

	// Get a new map - should not be the large one
	m2 := GetStringBoolMap()
	if len(m2) > 0 {
		t.Error("Large map was incorrectly returned to pool")
	}
}

// TestStringBuilderPool tests string builder pooling
func TestStringBuilderPool(t *testing.T) {
	sb1 := GetStringBuilder()
	if sb1 == nil {
		t.Fatal("GetStringBuilder returned nil")
	}

	sb1.WriteString("test")
	result := sb1.String()
	if result != "test" {
		t.Errorf("Expected 'test', got '%s'", result)
	}

	PutStringBuilder(sb1)

	// Get another builder - should be reset
	sb2 := GetStringBuilder()
	if sb2.Len() != 0 {
		t.Error("String builder not properly reset")
	}
}

// TestGetSpacer tests the spacer function
func TestGetSpacer(t *testing.T) {
	tests := []struct {
		n        int
		expected string
	}{
		{0, ""},
		{-1, ""},
		{1, " "},
		{5, "     "},
		{10, "          "},
	}

	for _, tt := range tests {
		result := GetSpacer(tt.n)
		if result != tt.expected {
			t.Errorf("GetSpacer(%d) = %q, want %q", tt.n, result, tt.expected)
		}
		if len(result) != max(0, tt.n) {
			t.Errorf("GetSpacer(%d) length = %d, want %d", tt.n, len(result), max(0, tt.n))
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
