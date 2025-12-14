package runtimeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestLoad_MergesPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yaml := []byte(`
log_level: warn
log_format: json
default_theme: Dracula
config_dir: %s
`)
	if err := os.WriteFile(configPath, []byte(fmt.Sprintf(string(yaml), tmpDir)), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	t.Setenv("ZEROUi_LOG_LEVEL", "error")

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("log-level", "", "")
	flags.String("config-dir", tmpDir, "")
	flags.String("log-format", "", "")
	flags.String("default-theme", "", "")
	_ = flags.Set("log-level", "debug")
	_ = flags.Set("log-format", "console")

	loader := NewLoader(viper.New())

	cfg, err := loader.Load(configPath, flags)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.LogLevel != "debug" { // flag should win over env & file
		t.Errorf("expected log_level from flag to be 'debug', got %q", cfg.LogLevel)
	}

	if cfg.LogFormat != "console" { // flag wins over file
		t.Errorf("expected log_format to be 'console', got %q", cfg.LogFormat)
	}

	if cfg.DefaultTheme != "Dracula" { // config file value
		t.Errorf("expected default_theme from file to be 'Dracula', got %q", cfg.DefaultTheme)
	}

	if cfg.ConfigDir != filepath.Clean(tmpDir) {
		t.Errorf("expected config_dir to resolve to %q, got %q", tmpDir, cfg.ConfigDir)
	}
}

func TestValidateConfigErrors(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		ConfigFile:   filepath.Join(tmpDir, "missing.yaml"),
		ConfigDir:    filepath.Join(tmpDir, "nosuchdir"),
		LogLevel:     "loud",
		LogFormat:    "pretty",
		DefaultTheme: "Solarized",
	}

	err := validateConfig(cfg)
	if err == nil {
		t.Fatalf("expected validation errors, got none")
	}

	errorText := err.Error()
	expectedFragments := []string{
		"does not exist",
		"config_dir",
		"log_level",
		"log_format",
		"default_theme",
	}

	for _, fragment := range expectedFragments {
		if !strings.Contains(errorText, fragment) {
			t.Errorf("expected validation error to mention %q, got: %s", fragment, errorText)
		}
	}
}
