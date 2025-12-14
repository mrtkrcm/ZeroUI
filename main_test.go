package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/mrtkrcm/ZeroUI/test/helpers"
)

func TestMain(m *testing.M) {
	helpers.RunTestMainWithCleanup(m, "main", "zeroui-main-test-home-", nil)
}

func TestSignalHandling(t *testing.T) {
	// Test that signal handling works correctly
	// We'll test the signal handling logic without actually running main()

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling like in main()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Test SIGINT handling
	go func() {
		sigChan <- syscall.SIGINT
	}()

	// Start a goroutine that simulates the main loop
	done := make(chan bool)
	go func() {
		defer close(done)
		for {
			select {
			case sig := <-sigChan:
				switch sig {
				case syscall.SIGINT:
					cancel()
					return
				case syscall.SIGTERM:
					go func() {
						time.Sleep(100 * time.Millisecond) // Shorter timeout for test
						cancel()
					}()
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for the signal handling to complete
	select {
	case <-done:
		// Signal was handled correctly
	case <-time.After(1 * time.Second):
		t.Fatal("Signal handling timed out")
	}

	// Verify context was cancelled
	select {
	case <-ctx.Done():
		// Context was cancelled as expected
	default:
		t.Error("Context should have been cancelled after signal")
	}
}

func TestContextCancellation(t *testing.T) {
	// Test that context cancellation works properly
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	// Verify context is cancelled
	select {
	case <-ctx.Done():
		// Context is cancelled as expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled immediately")
	}
}

func TestGracefulShutdown(t *testing.T) {
	// Test the graceful shutdown logic
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Send SIGTERM
	go func() {
		sigChan <- syscall.SIGTERM
	}()

	// Simulate the main loop with graceful shutdown
	shutdownComplete := make(chan bool)
	go func() {
		defer close(shutdownComplete)
		for {
			select {
			case sig := <-sigChan:
				if sig == syscall.SIGTERM {
					// Start graceful shutdown timeout
					go func() {
						time.Sleep(200 * time.Millisecond) // Short timeout for test
						cancel()
					}()
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for shutdown to complete
	select {
	case <-shutdownComplete:
		// Shutdown completed successfully
	case <-time.After(1 * time.Second):
		t.Fatal("Graceful shutdown timed out")
	}
}
