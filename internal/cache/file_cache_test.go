package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileCache(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "filecache_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatal(err)
	}

	cache := NewFileCache()

	// Test first read (cache miss)
	content1, err := cache.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content1) != string(testContent) {
		t.Errorf("Content mismatch: got %q, want %q", content1, testContent)
	}

	stats := cache.GetStats()
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	if stats.Hits != 0 {
		t.Errorf("Expected 0 hits, got %d", stats.Hits)
	}

	// Test second read (cache hit)
	content2, err := cache.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read cached file: %v", err)
	}
	if string(content2) != string(testContent) {
		t.Errorf("Cached content mismatch: got %q, want %q", content2, testContent)
	}

	stats = cache.GetStats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
}

func TestFileCacheInvalidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filecache_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	cache := NewFileCache()

	// Read file to cache it
	content1, _ := cache.ReadFile(testFile)
	if string(content1) != "original" {
		t.Errorf("Expected 'original', got %q", content1)
	}

	// Sleep to ensure different mtime
	time.Sleep(10 * time.Millisecond)

	// Modify the file
	if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
		t.Fatal(err)
	}

	// Read again - should detect change and reload
	content2, _ := cache.ReadFile(testFile)
	if string(content2) != "modified" {
		t.Errorf("Expected 'modified', got %q", content2)
	}

	stats := cache.GetStats()
	if stats.Misses != 2 {
		t.Errorf("Expected 2 misses (initial + after modification), got %d", stats.Misses)
	}
}

func TestFileCacheNonExistent(t *testing.T) {
	cache := NewFileCache()

	_, err := cache.ReadFile("/non/existent/file")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	stats := cache.GetStats()
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss for non-existent file, got %d", stats.Misses)
	}
}

func TestFileCacheHitRate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filecache_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	cache := NewFileCache()

	// First read - miss
	cache.ReadFile(testFile)

	// Next 9 reads - hits
	for i := 0; i < 9; i++ {
		cache.ReadFile(testFile)
	}

	hitRate := cache.GetHitRate()
	expectedRate := 90.0 // 9 hits out of 10 total
	if hitRate != expectedRate {
		t.Errorf("Expected hit rate of %.1f%%, got %.1f%%", expectedRate, hitRate)
	}
}

func TestFileCacheCleanup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filecache_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple test files
	files := make([]string, 3)
	for i := range files {
		files[i] = filepath.Join(tmpDir, string(rune('a'+i))+".txt")
		if err := os.WriteFile(files[i], []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cache := NewFileCache()

	// Read all files to cache them
	for _, f := range files {
		cache.ReadFile(f)
	}

	// Clean up entries older than 0 seconds (all of them)
	removed := cache.CleanupOldEntries(0)
	if removed != 3 {
		t.Errorf("Expected to remove 3 entries, removed %d", removed)
	}

	stats := cache.GetStats()
	if stats.Evictions != 3 {
		t.Errorf("Expected 3 evictions, got %d", stats.Evictions)
	}
}

func BenchmarkFileCache(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "filecache_bench")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "bench.txt")
	content := make([]byte, 1024) // 1KB file
	for i := range content {
		content[i] = byte(i % 256)
	}
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		b.Fatal(err)
	}

	b.Run("WithCache", func(b *testing.B) {
		cache := NewFileCache()
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := cache.ReadFile(testFile)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("WithoutCache", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := os.ReadFile(testFile)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
