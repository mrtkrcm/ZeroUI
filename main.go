package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrtkrcm/ZeroUI/cmd"
)

func main() {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle signals
	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGINT:
			// Ctrl+C - immediate but graceful shutdown
			cancel()
		case syscall.SIGTERM:
			// Termination request - graceful shutdown with timeout
			go func() {
				time.Sleep(5 * time.Second) // Give 5 seconds for graceful shutdown
				os.Exit(1)
			}()
			cancel()
		}
	}()

	// Execute the command with context
	cmd.ExecuteWithContext(ctx)
}
