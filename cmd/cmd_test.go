package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/mrtkrcm/ZeroUI/test/helpers"
	"github.com/spf13/cobra"
)

func executeCommand(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	cmd := rootCmd
	var stdout, stderr bytes.Buffer

	oldOut := cmd.OutOrStdout()
	oldErr := cmd.ErrOrStderr()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	err := cmd.Execute()
	cmd.SetOut(oldOut)
	cmd.SetErr(oldErr)
	cmd.SetArgs(nil)

	code := 0
	if err != nil {
		code = 1
	}

	return code, stdout.String(), stderr.String()
}

func TestMain(m *testing.M) {
	helpers.RunTestMainWithCleanup(m, "cmd", "zeroui-cmd-test-home-", nil)
}

func TestRootCmd(t *testing.T) {
	// Test that root command exists and can be executed
	// The root command is initialized as a global variable
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	// Test command basic properties
	if rootCmd.Use != "zeroui" {
		t.Errorf("Expected command use to be 'zeroui', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Root command should have a short description")
	}
}

func TestExecuteWithContext(t *testing.T) {
	// Test that ExecuteWithContext doesn't panic with a basic context
	ctx := context.Background()

	// Capture stdout to avoid printing during test
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	// Replace stdin with a pipe to avoid triggering TUI launches
	oldStdin := os.Stdin
	stdinR, stdinW, _ := os.Pipe()
	os.Stdin = stdinR

	// Run the command (it will fail because no subcommands, but shouldn't panic)
	defer func() {
		os.Stdout = old
		w.Close()
		stdinW.Close()
		os.Stdin = oldStdin
	}()

	// This should not panic
	if err := ExecuteWithContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("ExecuteWithContext returned error: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = old
}

func TestCommandStructure(t *testing.T) {
	// Test that the root command has the expected structure
	if rootCmd.Use != "zeroui" {
		t.Errorf("Expected root command use to be 'zeroui', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Root command should have a short description")
	}

	// Test that root command has subcommands
	if len(rootCmd.Commands()) == 0 {
		t.Error("Root command should have subcommands")
	}

	// Test that we have some expected subcommands
	subCommandNames := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		subCommandNames[cmd.Use] = true
	}

	// Check for some key subcommands
	expectedCommands := []string{"toggle", "list", "ui", "preset"}
	for _, expected := range expectedCommands {
		found := false
		for cmdName := range subCommandNames {
			if strings.Contains(cmdName, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find subcommand containing '%s', but didn't", expected)
		}
	}
}

func TestContainerInitialization(t *testing.T) {
	// Test that the container is properly initialized
	// This tests the init() function that sets up the container
	if appContainer == nil {
		t.Error("appContainer should be initialized by init() function")
	}
}

func TestUnknownCommand(t *testing.T) {
	code, _, stderr := executeCommand(t, "unknown")

	if code != 1 {
		t.Fatalf("expected exit code 1 for unknown command, got %d", code)
	}

	if !strings.Contains(stderr, "unknown command \"unknown\"") {
		t.Fatalf("expected unknown command message, got %q", stderr)
	}
}

func TestUnknownFlag(t *testing.T) {
	code, _, stderr := executeCommand(t, "--does-not-exist")

	if code != 1 {
		t.Fatalf("expected exit code 1 for unknown flag, got %d", code)
	}

	if !strings.Contains(stderr, "unknown flag") {
		t.Fatalf("expected unknown flag message, got %q", stderr)
	}
}

func TestMissingArgsValidation(t *testing.T) {
	code, _, stderr := executeCommand(t, "toggle", "ghostty")

	if code != 1 {
		t.Fatalf("expected exit code 1 for missing args, got %d", code)
	}

	if !strings.Contains(stderr, "accepts 3 arg(s)") {
		t.Fatalf("expected argument validation message, got %q", stderr)
	}
}

func TestRunWithSIGINTTriggersCleanup(t *testing.T) {
	signalChan := make(chan os.Signal, 1)

	cleanupCalled := make(chan struct{}, 1)
	RegisterCleanupHook(func() { cleanupCalled <- struct{}{} })
	t.Cleanup(func() { cleanupHooks = nil })

	started := make(chan struct{}, 1)
	blockCmd := &cobra.Command{
		Use:   "test-block",
		Short: "test command that waits for cancellation",
		RunE: func(cmd *cobra.Command, args []string) error {
			started <- struct{}{}
			<-cmd.Context().Done()
			return cmd.Context().Err()
		},
	}
	rootCmd.AddCommand(blockCmd)
	t.Cleanup(func() { rootCmd.RemoveCommand(blockCmd) })

	exitCodeCh := make(chan int, 1)
	go func() {
		exitCodeCh <- RunWithOptions(RunOptions{Args: []string{"test-block"}, SignalChan: signalChan})
	}()

	helpers.WaitForCondition(t, func() bool {
		select {
		case <-started:
			return true
		default:
			return false
		}
	}, time.Second, "command should start")

	signalChan <- syscall.SIGINT

	select {
	case code := <-exitCodeCh:
		if code != 0 {
			t.Fatalf("expected exit code 0 for SIGINT, got %d", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run did not complete after SIGINT")
	}

	helpers.WaitForCondition(t, func() bool {
		select {
		case <-cleanupCalled:
			return true
		default:
			return false
		}
	}, time.Second, "cleanup hook should be called after SIGINT")
}

func TestRunWithSIGTERMTriggersCleanup(t *testing.T) {
	signalChan := make(chan os.Signal, 1)

	cleanupCalled := make(chan struct{}, 1)
	RegisterCleanupHook(func() { cleanupCalled <- struct{}{} })
	t.Cleanup(func() { cleanupHooks = nil })

	started := make(chan struct{}, 1)
	blockCmd := &cobra.Command{
		Use:   "test-block",
		Short: "test command that waits for cancellation",
		RunE: func(cmd *cobra.Command, args []string) error {
			started <- struct{}{}
			<-cmd.Context().Done()
			return cmd.Context().Err()
		},
	}
	rootCmd.AddCommand(blockCmd)
	t.Cleanup(func() { rootCmd.RemoveCommand(blockCmd) })

	exitCodeCh := make(chan int, 1)
	go func() {
		exitCodeCh <- RunWithOptions(RunOptions{Args: []string{"test-block"}, SignalChan: signalChan})
	}()

	helpers.WaitForCondition(t, func() bool {
		select {
		case <-started:
			return true
		default:
			return false
		}
	}, time.Second, "command should start")

	signalChan <- syscall.SIGTERM

	select {
	case code := <-exitCodeCh:
		if code != 0 {
			t.Fatalf("expected exit code 0 for SIGTERM, got %d", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run did not complete after SIGTERM")
	}

	helpers.WaitForCondition(t, func() bool {
		select {
		case <-cleanupCalled:
			return true
		default:
			return false
		}
	}, time.Second, "cleanup hook should be called after SIGTERM")
}
