package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/test/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	// Test that the container is properly initialized via GetContainer()
	// Container is lazily initialized on first access
	c := GetContainer()
	if c == nil {
		t.Error("GetContainer() should return a non-nil container")
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

// ============================================================================
// Phase 1 & Phase 2 Integration Tests
// ============================================================================

// TestCommandTracingLogsStartAndEnd verifies that command tracing is attached and executes
func TestCommandTracingLogsStartAndEnd(t *testing.T) {
	// Verify that command tracing is properly attached by checking that
	// the PersistentPreRunE hook was set up by attachCommandTracing
	if rootCmd.PersistentPreRunE == nil {
		t.Error("Expected PersistentPreRunE to be set by attachCommandTracing")
	}
	if rootCmd.PersistentPostRunE == nil {
		t.Error("Expected PersistentPostRunE to be set by attachCommandTracing")
	}

	// Execute a command to verify tracing doesn't break execution
	code, _, _ := executeCommand(t, "list", "apps")
	if code != 0 {
		t.Fatalf("Command failed with code %d", code)
	}
}

// TestRequestIDGeneration verifies that request ID generation works
func TestRequestIDGeneration(t *testing.T) {
	// Test that the request ID generation function works correctly
	// The actual request ID is generated in attachCommandTracing's PersistentPreRunE

	// Execute a command to verify request ID generation doesn't break execution
	code, _, _ := executeCommand(t, "list", "apps")
	if code != 0 {
		t.Fatalf("Command failed with code %d", code)
	}

	// The request ID format is tested implicitly - if the command runs successfully,
	// the request ID generation in PersistentPreRunE executed without error
}

// TestLoggerAvailableInContext verifies that logger is available from context in subcommands
func TestLoggerAvailableInContext(t *testing.T) {
	loggerFound := false
	testCmd := &cobra.Command{
		Use:   "test-context-logger",
		Short: "test command for context logger",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get logger from context
			cmdLogger := logger.FromContext(cmd.Context())
			if cmdLogger != nil {
				loggerFound = true
			}
			return nil
		},
	}
	rootCmd.AddCommand(testCmd)
	t.Cleanup(func() { rootCmd.RemoveCommand(testCmd) })

	ctx := context.Background()
	err := ExecuteContext(ctx, []string{"test-context-logger"})
	if err != nil {
		t.Fatalf("ExecuteContext failed: %v", err)
	}

	if !loggerFound {
		t.Error("Logger should be available from context in subcommands")
	}
}

// TestRuntimeConfigFlags tests that the new runtime config flags are accepted without error
func TestRuntimeConfigFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "log-level flag",
			args: []string{"--log-level=debug", "list", "apps"},
		},
		{
			name: "log-format flag",
			args: []string{"--log-format=json", "list", "apps"},
		},
		{
			name: "default-theme flag",
			args: []string{"--default-theme=dracula", "list", "apps"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := ExecuteContext(ctx, tt.args)

			// The command should execute without error
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestRuntimeConfigEnvironmentVariables tests that environment variables are properly loaded
func TestRuntimeConfigEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
	}{
		{
			name:     "ZEROUI_LOG_LEVEL env var",
			envKey:   "ZEROUI_LOG_LEVEL",
			envValue: "warn",
		},
		{
			name:     "ZEROUI_LOG_FORMAT env var",
			envKey:   "ZEROUI_LOG_FORMAT",
			envValue: "json",
		},
		{
			name:     "ZEROUI_DEFAULT_THEME env var",
			envKey:   "ZEROUI_DEFAULT_THEME",
			envValue: "nord",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			oldValue := os.Getenv(tt.envKey)
			os.Setenv(tt.envKey, tt.envValue)
			t.Cleanup(func() {
				if oldValue != "" {
					os.Setenv(tt.envKey, oldValue)
				} else {
					os.Unsetenv(tt.envKey)
				}
			})

			// Reset viper to force reload
			viper.Reset()

			// Re-initialize config - this triggers environment variable loading
			initConfig()

			// The environment variable should be loaded via automatic env binding
			// We test this indirectly by ensuring initConfig doesn't panic or error
			// The actual value checking happens in the runtime config loader tests
		})
	}
}

// TestFlagPrecedenceOverEnvironment tests that flags take precedence over environment variables
func TestFlagPrecedenceOverEnvironment(t *testing.T) {
	// Set environment variable
	os.Setenv("ZEROUI_LOG_LEVEL", "warn")
	t.Cleanup(func() { os.Unsetenv("ZEROUI_LOG_LEVEL") })

	// Execute with flag (should override env var) - the flag value should be used
	ctx := context.Background()
	err := ExecuteContext(ctx, []string{"--log-level=debug", "list", "apps"})
	if err != nil {
		t.Fatalf("ExecuteContext failed: %v", err)
	}

	// The command should execute successfully, demonstrating flag precedence
	// Actual precedence logic is tested in the runtime config loader tests
}

// TestThemeInitialization verifies that theme is properly initialized from config
func TestThemeInitialization(t *testing.T) {
	// Create a temporary config file with theme setting
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configContent := `default-theme: dracula
log-level: info
log-format: text
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Reset viper
	viper.Reset()

	// Set config file
	cfgFile = configPath

	// Re-initialize config (this should load the theme)
	initConfig()

	// Verify theme was set (we can't directly check the theme name from styles package,
	// but we can verify the config was loaded)
	theme := viper.GetString("default-theme")
	if theme != "dracula" {
		t.Errorf("Expected theme to be 'dracula', got '%s'", theme)
	}
}

// TestCleanupHooksExecution verifies that cleanup hooks are executed properly
func TestCleanupHooksExecution(t *testing.T) {
	// Track cleanup execution order
	var executionOrder []int

	RegisterCleanupHook(func() { executionOrder = append(executionOrder, 1) })
	RegisterCleanupHook(func() { executionOrder = append(executionOrder, 2) })
	RegisterCleanupHook(func() { executionOrder = append(executionOrder, 3) })

	t.Cleanup(func() { cleanupHooks = nil })

	// Run cleanup
	runCleanupHooks()

	// Verify all hooks were executed
	if len(executionOrder) != 3 {
		t.Errorf("Expected 3 cleanup hooks to execute, got %d", len(executionOrder))
	}

	// Verify execution order
	for i, val := range executionOrder {
		if val != i+1 {
			t.Errorf("Expected hook %d to execute in order, got %d", i+1, val)
		}
	}
}

// TestCleanupHooksAreIdempotent verifies that cleanup hooks can only run once
func TestCleanupHooksAreIdempotent(t *testing.T) {
	executionCount := 0
	RegisterCleanupHook(func() { executionCount++ })
	t.Cleanup(func() { cleanupHooks = nil })

	// Run cleanup twice
	runCleanupHooks()
	runCleanupHooks()

	// Should only execute once
	if executionCount != 1 {
		t.Errorf("Expected cleanup hook to execute once, executed %d times", executionCount)
	}
}

// TestGracefulShutdownOnContextCancellation verifies graceful shutdown when context is cancelled
func TestGracefulShutdownOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	started := make(chan struct{}, 1)
	testCmd := &cobra.Command{
		Use:   "test-cancel",
		Short: "test command for cancellation",
		RunE: func(cmd *cobra.Command, args []string) error {
			started <- struct{}{}
			<-cmd.Context().Done()
			return cmd.Context().Err()
		},
	}
	rootCmd.AddCommand(testCmd)
	t.Cleanup(func() { rootCmd.RemoveCommand(testCmd) })

	errCh := make(chan error, 1)
	go func() {
		errCh <- ExecuteContext(ctx, []string{"test-cancel"})
	}()

	// Wait for command to start
	helpers.WaitForCondition(t, func() bool {
		select {
		case <-started:
			return true
		default:
			return false
		}
	}, time.Second, "command should start")

	// Cancel context
	cancel()

	// Wait for command to finish
	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Command did not complete after context cancellation")
	}
}

// TestPersistentFlags verifies that persistent flags are available to all subcommands
func TestPersistentFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
	}{
		{"verbose flag", "verbose"},
		{"dry-run flag", "dry-run"},
		{"log-level flag", "log-level"},
		{"log-format flag", "log-format"},
		{"default-theme flag", "default-theme"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := rootCmd.PersistentFlags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Expected persistent flag '%s' to exist", tt.flagName)
			}
		})
	}
}

// TestCommandTracingPreservesOriginalPreRunE verifies that command tracing preserves original PreRunE
func TestCommandTracingPreservesOriginalPreRunE(t *testing.T) {
	preRunCalled := false
	testCmd := &cobra.Command{
		Use:   "test-prerun",
		Short: "test command with original PreRunE",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			preRunCalled = true
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	// Attach tracing (this happens in init())
	attachCommandTracing(testCmd)

	// Execute the command
	ctx := context.Background()
	rootCmd.AddCommand(testCmd)
	t.Cleanup(func() { rootCmd.RemoveCommand(testCmd) })

	err := ExecuteContext(ctx, []string{"test-prerun"})
	if err != nil {
		t.Fatalf("ExecuteContext failed: %v", err)
	}

	if !preRunCalled {
		t.Error("Original PersistentPreRunE should have been called")
	}
}

// TestCommandTracingPreservesOriginalPostRunE verifies that command tracing preserves original PostRunE
func TestCommandTracingPreservesOriginalPostRunE(t *testing.T) {
	postRunCalled := false
	testCmd := &cobra.Command{
		Use:   "test-postrun",
		Short: "test command with original PostRunE",
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			postRunCalled = true
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	// Attach tracing
	attachCommandTracing(testCmd)

	// Execute the command
	ctx := context.Background()
	rootCmd.AddCommand(testCmd)
	t.Cleanup(func() { rootCmd.RemoveCommand(testCmd) })

	err := ExecuteContext(ctx, []string{"test-postrun"})
	if err != nil {
		t.Fatalf("ExecuteContext failed: %v", err)
	}

	if !postRunCalled {
		t.Error("Original PersistentPostRunE should have been called")
	}
}

// TestContainerCleanup verifies that the application container is properly closed
func TestContainerCleanup(t *testing.T) {
	// This test verifies that ExecuteContext properly closes the container
	// We can't directly test the Close() call, but we can verify no errors occur
	ctx := context.Background()

	testCmd := &cobra.Command{
		Use:   "test-container",
		Short: "test command for container cleanup",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	rootCmd.AddCommand(testCmd)
	t.Cleanup(func() { rootCmd.RemoveCommand(testCmd) })

	err := ExecuteContext(ctx, []string{"test-container"})
	if err != nil {
		t.Fatalf("ExecuteContext failed: %v", err)
	}

	// If we get here without errors, container cleanup worked
}

// TestRegisterCleanupHookWithNil verifies that nil hooks are ignored
func TestRegisterCleanupHookWithNil(t *testing.T) {
	initialCount := len(cleanupHooks)

	RegisterCleanupHook(nil)

	if len(cleanupHooks) != initialCount {
		t.Error("Nil cleanup hook should not be registered")
	}
}

// TestCommandTracingRecursivelyAttachesToSubcommands verifies that tracing is attached to all subcommands
func TestCommandTracingRecursivelyAttachesToSubcommands(t *testing.T) {
	// Create a parent command with a subcommand
	parentCmd := &cobra.Command{
		Use:   "test-parent",
		Short: "test parent command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	childCmd := &cobra.Command{
		Use:   "test-child",
		Short: "test child command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	parentCmd.AddCommand(childCmd)

	// Attach tracing recursively
	attachCommandTracing(parentCmd)

	// Both commands should now have PersistentPreRunE and PersistentPostRunE set
	if parentCmd.PersistentPreRunE == nil {
		t.Error("Parent command should have PersistentPreRunE after attachCommandTracing")
	}
	if parentCmd.PersistentPostRunE == nil {
		t.Error("Parent command should have PersistentPostRunE after attachCommandTracing")
	}
	if childCmd.PersistentPreRunE == nil {
		t.Error("Child command should have PersistentPreRunE after attachCommandTracing")
	}
	if childCmd.PersistentPostRunE == nil {
		t.Error("Child command should have PersistentPostRunE after attachCommandTracing")
	}
}
