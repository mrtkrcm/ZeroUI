package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mrtkrcm/ZeroUI/internal/cli"
)

func main() {
	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Handle signals in a goroutine
	go func() {
		for {
			select {
			case sig := <-sigChan:
				switch sig {
				case syscall.SIGINT:
					cancel()
					return
				case syscall.SIGTERM:
					// Allow some time for graceful shutdown
					go func() {
						cancel()
					}()
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Execute CLI commands with context support
	if err := cli.ExecuteWithContext(ctx); err != nil {
		// CLI handles its own error output
		os.Exit(1)
	}
}
