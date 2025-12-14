package cmd

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/mrtkrcm/ZeroUI/test/helpers"
)

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

	// Run the command (it will fail because no subcommands, but shouldn't panic)
	defer func() {
		os.Stdout = old
		w.Close()
	}()

	// This should not panic
	ExecuteWithContext(ctx)

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
