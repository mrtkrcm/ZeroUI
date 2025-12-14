package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
)

// RunOptions defines options for executing the CLI runtime
type RunOptions struct {
	Args       []string
	SignalChan chan os.Signal
	BaseCtx    context.Context
}

// Run starts the CLI runtime using default options
func Run() int {
	return RunWithOptions(RunOptions{})
}

// RunWithOptions centralizes context creation, signal handling, and command execution
func RunWithOptions(opts RunOptions) int {
	baseCtx := opts.BaseCtx
	if baseCtx == nil {
		baseCtx = context.Background()
	}

	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	sigChan := opts.SignalChan
	if sigChan == nil {
		sigChan = make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigChan)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- ExecuteContext(ctx, opts.Args)
	}()

	for {
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				cancel()
			}
		case err := <-errCh:
			if err == nil {
				return 0
			}
			if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
				return 0
			}
			return 1
		case <-ctx.Done():
			// Wait for the command to acknowledge cancellation
		}
	}
}
