package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestMain configures a deterministic environment for the package tests:
//   - Prepends repo-local testdata/bin (if present) to PATH so stub binaries
//     such as testdata/bin/ghostty are preferred.
//   - Creates an isolated temporary HOME directory for the test run.
//   - Creates test configuration files for applications to enable detection.
//
// The original environment is restored after tests complete.
func TestMain(m *testing.M) {
	// Create temporary HOME directory
	tmpHome, err := os.MkdirTemp("", "zeroui-internal-tui-test-home-")
	if err != nil {
		panic(err)
	}

	// Set HOME environment variable
	origHOME, hadHOME := os.LookupEnv("HOME")
	if err := os.Setenv("HOME", tmpHome); err != nil {
		panic(err)
	}

	// Create test application configurations
	setupTestConfigurations(tmpHome)

	// Run tests
	exitCode := m.Run()

	// Cleanup
	if hadHOME {
		os.Setenv("HOME", origHOME)
	} else {
		os.Unsetenv("HOME")
	}
	os.RemoveAll(tmpHome)

	os.Exit(exitCode)
}

// setupTestConfigurations creates test configuration files and executables for applications
// so that the TUI can detect and display them during testing
func setupTestConfigurations(homeDir string) {
	// Create .config directory
	configDir := filepath.Join(homeDir, ".config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return
	}

	// Create bin directory for test executables
	binDir := filepath.Join(homeDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err == nil {
		// Add to PATH so executables are found
		os.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	}

	// Create test executables (simple shell scripts)
	testExecutables := []string{"ghostty", "zed", "nvim", "code", "alacritty", "wezterm"}
	for _, exec := range testExecutables {
		execPath := filepath.Join(binDir, exec)
		script := fmt.Sprintf("#!/bin/bash\necho 'Test executable for %s'\n", exec)
		os.WriteFile(execPath, []byte(script), 0755)
	}

	// Create Ghostty configuration
	ghosttyDir := filepath.Join(configDir, "ghostty")
	if err := os.MkdirAll(ghosttyDir, 0755); err == nil {
		ghosttyConfig := `font-family = "JetBrains Mono"
font-size = 14
theme = "dark"
window-padding-x = 8
window-padding-y = 8
`
		os.WriteFile(filepath.Join(ghosttyDir, "config"), []byte(ghosttyConfig), 0644)
	}

	// Create Alacritty configuration
	alacrittyDir := filepath.Join(configDir, "alacritty")
	if err := os.MkdirAll(alacrittyDir, 0755); err == nil {
		alacrittyConfig := `font:
  normal:
    family: "JetBrains Mono"
  size: 14.0
`
		os.WriteFile(filepath.Join(alacrittyDir, "alacritty.yml"), []byte(alacrittyConfig), 0644)
	}

	// Create WezTerm configuration
	weztermDir := filepath.Join(configDir, "wezterm")
	if err := os.MkdirAll(weztermDir, 0755); err == nil {
		weztermConfig := `local wezterm = require 'wezterm'
return {
  font = wezterm.font("JetBrains Mono"),
  font_size = 14.0,
  color_scheme = "OneDark",
}
`
		os.WriteFile(filepath.Join(weztermDir, "wezterm.lua"), []byte(weztermConfig), 0644)
	}

	// Create Zed configuration
	zedDir := filepath.Join(configDir, "zed")
	if err := os.MkdirAll(zedDir, 0755); err == nil {
		zedSettings := `{
  "theme": "One Dark",
  "buffer_font_family": "JetBrains Mono",
  "buffer_font_size": 14,
  "vim_mode": true,
  "format_on_save": "on"
}`
		os.WriteFile(filepath.Join(zedDir, "settings.json"), []byte(zedSettings), 0644)
	}

	// Create Neovim configuration
	nvimDir := filepath.Join(configDir, "nvim")
	if err := os.MkdirAll(nvimDir, 0755); err == nil {
		nvimConfig := `-- Basic Neovim configuration
vim.opt.number = true
vim.opt.relativenumber = true
vim.opt.tabstop = 4
vim.opt.shiftwidth = 4
vim.opt.expandtab = true
`
		os.WriteFile(filepath.Join(nvimDir, "init.lua"), []byte(nvimConfig), 0644)
	}

	// Create VS Code configuration
	vscodeDir := filepath.Join(configDir, "Code", "User")
	if err := os.MkdirAll(vscodeDir, 0755); err == nil {
		vscodeSettings := `{
    "editor.fontFamily": "JetBrains Mono",
    "editor.fontSize": 14,
    "editor.theme": "Default Dark+",
    "editor.formatOnSave": true,
    "editor.wordWrap": "on"
}`
		os.WriteFile(filepath.Join(vscodeDir, "settings.json"), []byte(vscodeSettings), 0644)
	}

	// Create Tmux configuration
	tmuxConfig := `set -g default-terminal "screen-256color"
set -g status-style 'bg=#333333,fg=#5eacd3'
set -g window-status-current-style 'bg=#5eacd3,fg=#1e1e1e'
`
	os.WriteFile(filepath.Join(homeDir, ".tmux.conf"), []byte(tmuxConfig), 0644)
}
