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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Execute CLI commands with context support
	if err := cli.ExecuteWithContext(ctx); err != nil {
		// CLI handles its own error output
		os.Exit(1)
	}
}
