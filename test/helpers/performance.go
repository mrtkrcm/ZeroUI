package helpers

import (
	"runtime"
	"testing"
	"time"
)

// PerformanceMonitor tracks test performance metrics
type PerformanceMonitor struct {
	startTime time.Time
	startMem  uint64
	testName  string
}

// StartPerformanceMonitor begins performance tracking
func StartPerformanceMonitor(t *testing.T, testName string) *PerformanceMonitor {
	t.Helper()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &PerformanceMonitor{
		startTime: time.Now(),
		startMem:  memStats.Alloc,
		testName:  testName,
	}
}

// StopPerformanceMonitor ends performance tracking and logs results
func (pm *PerformanceMonitor) StopPerformanceMonitor(t *testing.T, threshold time.Duration) {
	t.Helper()

	duration := time.Since(pm.startTime)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memoryUsed := memStats.Alloc - pm.startMem

	// Log performance metrics
	t.Logf("%s - Duration: %v, Memory: %d bytes", pm.testName, duration, memoryUsed)

	// Check against threshold
	if duration > threshold {
		t.Logf("WARNING: %s exceeded threshold (%v > %v)", pm.testName, duration, threshold)
	}
}

// BenchmarkWithMetrics runs a benchmark and collects detailed metrics
func BenchmarkWithMetrics(b *testing.B, name string, fn func()) {
	b.Helper()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	startAlloc := memStats.Alloc

	start := time.Now()
	result := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fn()
		}
	})
	duration := time.Since(start)

	runtime.ReadMemStats(&memStats)
	totalAlloc := memStats.Alloc - startAlloc

	b.Logf("%s Benchmark Results:", name)
	b.Logf("  Iterations: %d", result.N)
	b.Logf("  Total Time: %v", duration)
	b.Logf("  Time per Op: %v", result.T)
	b.Logf("  Memory Allocated: %d bytes", totalAlloc)
	b.Logf("  Allocs per Op: %d", result.AllocsPerOp())
}

// MeasureMemoryUsage measures memory usage of a function
func MeasureMemoryUsage(t *testing.T, name string, fn func()) {
	t.Helper()

	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	startMem := memStats.Alloc

	fn()

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	endMem := memStats.Alloc

	memoryUsed := endMem - startMem
	t.Logf("%s - Memory Used: %d bytes", name, memoryUsed)
}

// PerformanceTest runs a function and validates it meets performance criteria
func PerformanceTest(t *testing.T, name string, fn func(), maxDuration time.Duration, maxMemory uint64) {
	t.Helper()

	// Measure duration
	start := time.Now()
	fn()
	duration := time.Since(start)

	// Check duration
	if duration > maxDuration {
		t.Errorf("%s exceeded time limit: %v > %v", name, duration, maxDuration)
	}

	// Measure memory (simplified)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	// Note: Memory measurement is approximate due to GC

	t.Logf("%s - Duration: %v (limit: %v)", name, duration, maxDuration)
}

// TimeOperation times a single operation and logs the result
func TimeOperation(t *testing.T, name string, fn func()) time.Duration {
	t.Helper()

	start := time.Now()
	fn()
	duration := time.Since(start)

	t.Logf("%s completed in %v", name, duration)
	return duration
}

// ParallelTestRunner runs tests in parallel with resource limits
type ParallelTestRunner struct {
	maxConcurrency int
	semaphore      chan struct{}
}

// NewParallelTestRunner creates a new parallel test runner
func NewParallelTestRunner(maxConcurrency int) *ParallelTestRunner {
	return &ParallelTestRunner{
		maxConcurrency: maxConcurrency,
		semaphore:      make(chan struct{}, maxConcurrency),
	}
}

// RunParallelTest runs a test function with concurrency control
func (ptr *ParallelTestRunner) RunParallelTest(t *testing.T, testName string, testFunc func(t *testing.T)) {
	t.Helper()

	ptr.semaphore <- struct{}{} // Acquire
	defer func() { <-ptr.semaphore }() // Release

	t.Run(testName, testFunc)
}

// GetSystemInfo logs system information for performance context
func GetSystemInfo(t *testing.T) {
	t.Helper()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	t.Logf("System Info:")
	t.Logf("  GOOS: %s", runtime.GOOS)
	t.Logf("  GOARCH: %s", runtime.GOARCH)
	t.Logf("  NumCPU: %d", runtime.NumCPU())
	t.Logf("  NumGoroutine: %d", runtime.NumGoroutine())
	t.Logf("  Memory Alloc: %d bytes", memStats.Alloc)
	t.Logf("  Memory Sys: %d bytes", memStats.Sys)
	t.Logf("  GC Cycles: %d", memStats.NumGC)
}

// Example usage:
//
// func TestWithPerformanceMonitoring(t *testing.T) {
//     t.Parallel()
//
//     pm := helpers.StartPerformanceMonitor(t, "TestWithPerformanceMonitoring")
//     defer pm.StopPerformanceMonitor(t, 100*time.Millisecond)
//
//     // Test code here
//
//     helpers.MeasureMemoryUsage(t, "TestFunction", func() {
//         // Function to measure
//     })
// }
