package performance

import (
	"runtime"
	"testing"
)

// TestMemoryImprovements validates memory usage improvements
func TestMemoryImprovements(t *testing.T) {
	// Force GC before measurements
	runtime.GC()
	runtime.GC()

	var m runtime.MemStats

	t.Run("MapPoolMemory", func(t *testing.T) {
		// Measure without pooling
		runtime.GC()
		runtime.ReadMemStats(&m)
		allocBefore := m.Alloc

		for i := 0; i < 1000; i++ {
			_ = make(map[string]bool, 16)
		}

		runtime.ReadMemStats(&m)
		allocWithout := m.Alloc - allocBefore

		// Measure with pooling
		runtime.GC()
		runtime.ReadMemStats(&m)
		allocBefore = m.Alloc

		for i := 0; i < 1000; i++ {
			m := GetStringBoolMap()
			PutStringBoolMap(m)
		}

		runtime.ReadMemStats(&m)
		allocWith := m.Alloc - allocBefore

		improvement := float64(allocWithout-allocWith) / float64(allocWithout) * 100
		t.Logf("Memory improvement with map pooling: %.2f%% (without: %d bytes, with: %d bytes)",
			improvement, allocWithout, allocWith)

		if allocWith >= allocWithout {
			t.Error("Map pooling did not improve memory usage")
		}
	})

	t.Run("SlicePreallocationMemory", func(t *testing.T) {
		const items = 100
		const iterations = 100

		// Without preallocation
		runtime.GC()
		runtime.ReadMemStats(&m)
		totalAllocsBefore := m.TotalAlloc

		for i := 0; i < iterations; i++ {
			var slice []string
			for j := 0; j < items; j++ {
				slice = append(slice, "item")
			}
			_ = slice
		}

		runtime.ReadMemStats(&m)
		totalAllocsWithout := m.TotalAlloc - totalAllocsBefore

		// With preallocation
		runtime.GC()
		runtime.ReadMemStats(&m)
		totalAllocsBefore = m.TotalAlloc

		for i := 0; i < iterations; i++ {
			slice := make([]string, 0, items)
			for j := 0; j < items; j++ {
				slice = append(slice, "item")
			}
			_ = slice
		}

		runtime.ReadMemStats(&m)
		totalAllocsWith := m.TotalAlloc - totalAllocsBefore

		improvement := float64(totalAllocsWithout-totalAllocsWith) / float64(totalAllocsWithout) * 100
		t.Logf("Memory improvement with slice preallocation: %.2f%% (without: %d bytes, with: %d bytes)",
			improvement, totalAllocsWithout, totalAllocsWith)

		if totalAllocsWith >= totalAllocsWithout {
			t.Error("Slice preallocation did not improve memory usage")
		}
	})
}

// TestGCPressure tests the reduction in GC pressure
func TestGCPressure(t *testing.T) {
	runtime.GC()
	var m runtime.MemStats

	t.Run("WithoutPooling", func(t *testing.T) {
		runtime.ReadMemStats(&m)
		gcsBefore := m.NumGC

		// Create many temporary allocations
		for i := 0; i < 10000; i++ {
			m := make(map[string]bool)
			m["test"] = true
			_ = m
		}

		runtime.GC()
		runtime.ReadMemStats(&m)
		gcsAfter := m.NumGC

		t.Logf("GC runs without pooling: %d", gcsAfter-gcsBefore)
	})

	t.Run("WithPooling", func(t *testing.T) {
		runtime.ReadMemStats(&m)
		gcsBefore := m.NumGC

		// Use pooled allocations
		for i := 0; i < 10000; i++ {
			m := GetStringBoolMap()
			m["test"] = true
			PutStringBoolMap(m)
		}

		runtime.GC()
		runtime.ReadMemStats(&m)
		gcsAfter := m.NumGC

		t.Logf("GC runs with pooling: %d", gcsAfter-gcsBefore)
	})
}
