package rpc

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestConcurrentPluginLoadSimulation simulates concurrent plugin loading
func TestConcurrentPluginLoadSimulation(t *testing.T) {
	// Simulate loading 10 plugins
	numPlugins := 10
	loadDelay := 10 * time.Millisecond

	t.Run("Sequential", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < numPlugins; i++ {
			time.Sleep(loadDelay) // Simulate load time
		}
		duration := time.Since(start)
		t.Logf("Sequential loading of %d plugins took %v", numPlugins, duration)

		expectedMin := time.Duration(numPlugins) * loadDelay
		if duration < expectedMin {
			t.Errorf("Sequential load too fast: %v < %v", duration, expectedMin)
		}
	})

	t.Run("Concurrent", func(t *testing.T) {
		start := time.Now()
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, 4) // Limit concurrency like in our implementation

		for i := 0; i < numPlugins; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				time.Sleep(loadDelay) // Simulate load time
			}()
		}

		wg.Wait()
		duration := time.Since(start)
		t.Logf("Concurrent loading of %d plugins took %v", numPlugins, duration)

		// Should be significantly faster than sequential
		expectedMax := time.Duration(numPlugins) * loadDelay
		if duration >= expectedMax {
			t.Errorf("Concurrent load not faster than sequential: %v >= %v", duration, expectedMax)
		}

		// With semaphore of 4, minimum time should be (numPlugins/4) * loadDelay
		expectedMin := time.Duration((numPlugins+3)/4) * loadDelay
		if duration < expectedMin-5*time.Millisecond { // Allow some timing variance
			t.Errorf("Concurrent load too fast (impossible): %v < %v", duration, expectedMin)
		}
	})
}

// TestSemaphorePattern tests the semaphore pattern used in LoadPluginsConcurrently
func TestSemaphorePattern(t *testing.T) {
	maxConcurrent := 4
	totalTasks := 20
	var currentConcurrent int32
	var maxObserved int32

	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i := 0; i < totalTasks; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Track concurrent execution
			current := atomic.AddInt32(&currentConcurrent, 1)

			// Update max if needed
			for {
				max := atomic.LoadInt32(&maxObserved)
				if current <= max || atomic.CompareAndSwapInt32(&maxObserved, max, current) {
					break
				}
			}

			// Simulate work
			time.Sleep(5 * time.Millisecond)

			atomic.AddInt32(&currentConcurrent, -1)
		}(i)
	}

	wg.Wait()

	t.Logf("Max concurrent execution: %d (limit was %d)", maxObserved, maxConcurrent)

	if maxObserved > int32(maxConcurrent) {
		t.Errorf("Semaphore failed: max concurrent %d exceeded limit %d", maxObserved, maxConcurrent)
	}

	if maxObserved == 0 {
		t.Error("No concurrent execution detected")
	}
}

// BenchmarkConcurrentVsSequential compares concurrent vs sequential performance
func BenchmarkConcurrentVsSequential(b *testing.B) {
	numItems := 10
	workDuration := time.Millisecond

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < numItems; j++ {
				time.Sleep(workDuration)
			}
		}
	})

	b.Run("Concurrent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup
			semaphore := make(chan struct{}, 4)

			for j := 0; j < numItems; j++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					semaphore <- struct{}{}
					defer func() { <-semaphore }()
					time.Sleep(workDuration)
				}()
			}

			wg.Wait()
		}
	})
}
