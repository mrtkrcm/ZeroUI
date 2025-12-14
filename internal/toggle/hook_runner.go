package toggle

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mrtkrcm/ZeroUI/internal/config"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/spf13/viper"
)

// HookRunner handles hook execution with security validation
type HookRunner struct {
	logger *logger.Logger
}

// NewHookRunner creates a new hook runner
func NewHookRunner(log *logger.Logger) *HookRunner {
	return &HookRunner{
		logger: log,
	}
}

// RunHooks executes hooks for a given type (pre-toggle, post-toggle, etc.)
func (hr *HookRunner) RunHooks(appConfig *config.AppConfig, hookType string) error {
	hookCmd, exists := appConfig.Hooks[hookType]
	if !exists {
		return nil // No hook defined, not an error
	}

	log := hr.logger.WithApp(appConfig.Name).WithContext(map[string]interface{}{
		"hook_type": hookType,
		"command":   hookCmd,
	})

	if viper.GetBool("verbose") {
		log.Debug("Running hook")
	}

	// Set environment variables safely
	if err := hr.setEnvironmentVariables(appConfig.Env); err != nil {
		return fmt.Errorf("failed to set environment variables: %w", err)
	}

	// Execute the hook command with security validation
	return hr.executeHookCommand(hookCmd, hookType, log)
}

// executeHookCommand executes a hook command with security checks
func (hr *HookRunner) executeHookCommand(hookCmd, hookType string, log *logger.Logger) error {
	// Parse command into parts
	parts := strings.Fields(hookCmd)
	if len(parts) == 0 {
		return nil // Empty command, not an error
	}

	// Security validation: Check if command is allowed
	if err := hr.validateHookCommand(hookCmd); err != nil {
		log.Error("Hook command validation failed", err)
		return fmt.Errorf("hook validation failed: %w", err)
	}

	// Security: Use only the command name, not any path
	commandName := filepath.Base(parts[0])

	// Resolve command from system PATH only (prevents path traversal)
	commandPath, err := exec.LookPath(commandName)
	if err != nil {
		return fmt.Errorf("command '%s' not found in system PATH: %w", commandName, err)
	}

	// Create command with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Reduced timeout for safety
	defer cancel()

	cmd := exec.CommandContext(ctx, commandPath, parts[1:]...)

	// Security: Restrict command capabilities
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil // Prevent interactive input

	// Set working directory to a safe, temporary location
	tempDir := os.TempDir()
	cmd.Dir = tempDir

	// Clear environment to prevent environment variable attacks
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.TempDir(),
		"TMPDIR=" + tempDir,
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hook %s failed: %w", hookType, err)
	}

	return nil
}

// validateHookCommand validates that the hook command is safe to execute
func (hr *HookRunner) validateHookCommand(hookCmd string) error {
	// Trim whitespace and check for empty command
	hookCmd = strings.TrimSpace(hookCmd)
	if hookCmd == "" {
		return fmt.Errorf("empty hook command")
	}

	// Check for dangerous characters and patterns - expanded list
	dangerousPatterns := []string{
		"|", "&&", "||", ";", "`", "$", "$(", "${", // Shell operators and command substitution
		"rm -rf", "rm -f", ">/dev/null", "2>&1", // Dangerous operations
		"curl", "wget", "nc", "telnet", "ssh", "scp", // Network operations
		"sudo", "su -", "chmod +x", "chown", "setuid", // Privilege escalation
		"../", "./", "~", "/etc/", "/usr/", "/var/", // Path traversal attempts
		"eval", "exec", "source", "bash -c", "sh -c", // Code execution
		">&", "<&", ">>", "<<", // Redirection operators
		"*", "?", "[", "]", // Glob patterns that could be dangerous
	}

	lowerCmd := strings.ToLower(hookCmd)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerCmd, pattern) {
			return fmt.Errorf("hook command contains dangerous pattern: %s", pattern)
		}
	}

	// Additional validation: no control characters
	for _, r := range hookCmd {
		if r < 32 && r != 9 && r != 10 && r != 13 { // Allow tab, newline, carriage return
			return fmt.Errorf("hook command contains control character")
		}
	}

	// Allow-list approach: only allow certain safe commands
	parts := strings.Fields(hookCmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty hook command")
	}

	// Strict allow-list - only essential safe commands
	allowedCommands := []string{
		"echo", "printf", "cat", "head", "tail", "wc",
		"grep", "sed", "awk", "sort", "uniq",
		"touch", "mkdir", "ls", "pwd",
		"notify-send", "osascript", // Notification commands
		"date", "sleep", // Time-related safe commands
	}

	// Extract just the command name, handle absolute paths
	command := filepath.Base(parts[0])
	// Also check if it's trying to use a path
	if strings.Contains(parts[0], "/") && parts[0] != command {
		return fmt.Errorf("hook command cannot use absolute or relative paths")
	}

	for _, allowed := range allowedCommands {
		if command == allowed {
			return nil
		}
	}

	return fmt.Errorf("hook command '%s' is not in the allowed list", command)
}

// setEnvironmentVariables safely sets environment variables
func (hr *HookRunner) setEnvironmentVariables(envVars map[string]string) error {
	// Prevent setting dangerous environment variables
	dangerousVars := []string{
		"PATH", "LD_LIBRARY_PATH", "DYLD_LIBRARY_PATH", // Path variables
		"SHELL", "USER", "HOME", // System variables
		"SUDO_USER", "SUDO_COMMAND", // Privilege escalation
	}

	for key, value := range envVars {
		// Check if variable name is dangerous
		upperKey := strings.ToUpper(key)
		for _, dangerous := range dangerousVars {
			if upperKey == dangerous {
				return fmt.Errorf("cannot set dangerous environment variable: %s", key)
			}
		}

		// Set the environment variable
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", key, err)
		}
	}

	return nil
}
