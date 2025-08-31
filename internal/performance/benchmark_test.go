package performance

import (
	"os"
	"strings"
	"testing"
)

// BenchmarkStringBoolMapPool tests the performance of pooled maps vs regular allocation
func BenchmarkStringBoolMapPool(b *testing.B) {
	b.Run("WithPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			m := GetStringBoolMap() // Now uses direct allocation for small maps
			m["key1"] = true
			m["key2"] = false
			m["key3"] = true
			// No need to call PutStringBoolMap for direct-allocated maps
		}
	})
	
	b.Run("WithPooled", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			m := GetStringBoolMapPooled() // Uses sync.Pool
			m["key1"] = true
			m["key2"] = false
			m["key3"] = true
			PutStringBoolMap(m) // Return to pool
		}
	})

	b.Run("WithoutPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			m := make(map[string]bool)
			m["key1"] = true
			m["key2"] = false
			m["key3"] = true
		}
	})
}

// BenchmarkStringBuilderPool tests the performance of pooled string builders
func BenchmarkStringBuilderPool(b *testing.B) {
	b.Run("WithPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sb := GetStringBuilder() // Now uses direct allocation for small strings
			sb.WriteString("hello")
			sb.WriteString(" ")
			sb.WriteString("world")
			_ = sb.String()
			// No need to call PutStringBuilder for direct-allocated builders
		}
	})
	
	b.Run("WithPooled", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sb := GetStringBuilderPooled() // Uses sync.Pool
			sb.WriteString("hello")
			sb.WriteString(" ")
			sb.WriteString("world")
			_ = sb.String()
			PutStringBuilder(sb) // Return to pool
		}
	})

	b.Run("WithoutPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			sb.WriteString("hello")
			sb.WriteString(" ")
			sb.WriteString("world")
			_ = sb.String()
		}
	})
}

// BenchmarkSlicePreallocation tests the performance of pre-allocated slices
func BenchmarkSlicePreallocation(b *testing.B) {
	items := 100

	b.Run("WithPreallocation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			slice := make([]string, 0, items)
			for j := 0; j < items; j++ {
				slice = append(slice, "item")
			}
		}
	})

	b.Run("WithoutPreallocation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var slice []string
			for j := 0; j < items; j++ {
				slice = append(slice, "item")
			}
		}
	})
}

// BenchmarkHomeCache tests the performance of cached home directory lookup
func BenchmarkHomeCache(b *testing.B) {
	b.Run("Cached", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = GetHomeDir()
		}
	})

	b.Run("Uncached", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = os.UserHomeDir()
		}
	})
}
