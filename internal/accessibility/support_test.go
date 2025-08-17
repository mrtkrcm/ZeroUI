package accessibility

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectAccessibilityNeeds(t *testing.T) {
	// Clean environment for each test
	originalVars := map[string]string{
		"ACCESSIBLE":           os.Getenv("ACCESSIBLE"),
		"SCREEN_READER":        os.Getenv("SCREEN_READER"),
		"HIGH_CONTRAST":        os.Getenv("HIGH_CONTRAST"),
		"NO_COLOR":            os.Getenv("NO_COLOR"),
		"REDUCED_MOTION":      os.Getenv("REDUCED_MOTION"),
		"VERBOSE_DESCRIPTIONS": os.Getenv("VERBOSE_DESCRIPTIONS"),
		"SIMPLE_UI":           os.Getenv("SIMPLE_UI"),
	}
	
	// Clean up after test
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("Default no accessibility", func(t *testing.T) {
		clearEnvVars()
		opts := DetectAccessibilityNeeds()
		
		assert.False(t, opts.Enabled)
		assert.False(t, opts.ScreenReaderMode)
		assert.False(t, opts.HighContrast)
		assert.False(t, opts.ReducedMotion)
		assert.False(t, opts.NoColor)
		assert.False(t, opts.VerboseDescriptions)
		assert.False(t, opts.SimplifiedUI)
	})

	t.Run("ACCESSIBLE environment variable", func(t *testing.T) {
		clearEnvVars()
		os.Setenv("ACCESSIBLE", "1")
		
		opts := DetectAccessibilityNeeds()
		assert.True(t, opts.Enabled)
	})

	t.Run("Screen reader detection", func(t *testing.T) {
		clearEnvVars()
		os.Setenv("SCREEN_READER", "1")
		
		opts := DetectAccessibilityNeeds()
		assert.True(t, opts.Enabled) // Auto-enabled
		assert.True(t, opts.ScreenReaderMode)
	})

	t.Run("High contrast mode", func(t *testing.T) {
		clearEnvVars()
		os.Setenv("HIGH_CONTRAST", "true")
		
		opts := DetectAccessibilityNeeds()
		assert.True(t, opts.Enabled) // Auto-enabled
		assert.True(t, opts.HighContrast)
	})

	t.Run("No color mode", func(t *testing.T) {
		clearEnvVars()
		os.Setenv("NO_COLOR", "1")
		
		opts := DetectAccessibilityNeeds()
		assert.True(t, opts.NoColor)
	})

	t.Run("Reduced motion", func(t *testing.T) {
		clearEnvVars()
		os.Setenv("REDUCED_MOTION", "true")
		
		opts := DetectAccessibilityNeeds()
		assert.True(t, opts.Enabled) // Auto-enabled
		assert.True(t, opts.ReducedMotion)
	})

	t.Run("Multiple accessibility features", func(t *testing.T) {
		clearEnvVars()
		os.Setenv("SCREEN_READER", "1")
		os.Setenv("HIGH_CONTRAST", "true")
		os.Setenv("VERBOSE_DESCRIPTIONS", "yes")
		
		opts := DetectAccessibilityNeeds()
		assert.True(t, opts.Enabled)
		assert.True(t, opts.ScreenReaderMode)
		assert.True(t, opts.HighContrast)
		assert.True(t, opts.VerboseDescriptions)
	})
}

func TestAccessibilityOptions_Methods(t *testing.T) {
	t.Run("GetHuhAccessibilityMode", func(t *testing.T) {
		opts := AccessibilityOptions{Enabled: true}
		assert.True(t, opts.GetHuhAccessibilityMode())
		
		opts = AccessibilityOptions{ScreenReaderMode: true}
		assert.True(t, opts.GetHuhAccessibilityMode())
		
		opts = AccessibilityOptions{}
		assert.False(t, opts.GetHuhAccessibilityMode())
	})

	t.Run("GetColorMode", func(t *testing.T) {
		opts := AccessibilityOptions{NoColor: true}
		assert.Equal(t, ColorModeNone, opts.GetColorMode())
		
		opts = AccessibilityOptions{HighContrast: true}
		assert.Equal(t, ColorModeHighContrast, opts.GetColorMode())
		
		opts = AccessibilityOptions{}
		assert.Equal(t, ColorModeNormal, opts.GetColorMode())
	})

	t.Run("GetAnimationMode", func(t *testing.T) {
		opts := AccessibilityOptions{ReducedMotion: true}
		assert.Equal(t, AnimationModeNone, opts.GetAnimationMode())
		
		opts = AccessibilityOptions{}
		assert.Equal(t, AnimationModeNormal, opts.GetAnimationMode())
	})

	t.Run("GetDescriptionLevel", func(t *testing.T) {
		opts := AccessibilityOptions{VerboseDescriptions: true}
		assert.Equal(t, DescriptionLevelVerbose, opts.GetDescriptionLevel())
		
		opts = AccessibilityOptions{ScreenReaderMode: true}
		assert.Equal(t, DescriptionLevelVerbose, opts.GetDescriptionLevel())
		
		opts = AccessibilityOptions{Enabled: true}
		assert.Equal(t, DescriptionLevelDetailed, opts.GetDescriptionLevel())
		
		opts = AccessibilityOptions{}
		assert.Equal(t, DescriptionLevelBasic, opts.GetDescriptionLevel())
	})
}

func TestAccessibilityOptions_Helpers(t *testing.T) {
	t.Run("GetAccessibleTitle", func(t *testing.T) {
		opts := AccessibilityOptions{Enabled: false}
		result := opts.GetAccessibleTitle("Test", "Context")
		assert.Equal(t, "Test", result)
		
		opts = AccessibilityOptions{Enabled: true}
		result = opts.GetAccessibleTitle("Test", "")
		assert.Equal(t, "Test", result)
		
		opts = AccessibilityOptions{Enabled: true, VerboseDescriptions: true}
		result = opts.GetAccessibleTitle("Test", "Context")
		assert.Equal(t, "Test - Context", result)
	})

	t.Run("GetAccessibleHelp", func(t *testing.T) {
		opts := AccessibilityOptions{}
		result := opts.GetAccessibleHelp("Basic help", "Detailed help")
		assert.Equal(t, "Basic help", result)
		
		opts = AccessibilityOptions{Enabled: true}
		result = opts.GetAccessibleHelp("Basic", "Detailed")
		assert.Equal(t, "Detailed", result)
		
		opts = AccessibilityOptions{VerboseDescriptions: true}
		result = opts.GetAccessibleHelp("Basic", "")
		assert.Contains(t, result, "Use arrow keys to navigate")
	})

	t.Run("GetStatusDescription", func(t *testing.T) {
		opts := AccessibilityOptions{Enabled: false}
		result := opts.GetStatusDescription("configured")
		assert.Equal(t, "configured", result)
		
		opts = AccessibilityOptions{Enabled: true}
		result = opts.GetStatusDescription("configured")
		assert.Equal(t, "Application is configured and ready", result)
		
		result = opts.GetStatusDescription("unknown_status")
		assert.Equal(t, "unknown_status", result)
	})

	t.Run("ShouldUseProgressBar", func(t *testing.T) {
		opts := AccessibilityOptions{}
		assert.True(t, opts.ShouldUseProgressBar())
		
		opts = AccessibilityOptions{ReducedMotion: true}
		assert.False(t, opts.ShouldUseProgressBar())
		
		opts = AccessibilityOptions{SimplifiedUI: true}
		assert.False(t, opts.ShouldUseProgressBar())
	})

	t.Run("GetFocusIndicator", func(t *testing.T) {
		opts := AccessibilityOptions{}
		assert.Equal(t, "• ", opts.GetFocusIndicator())
		
		opts = AccessibilityOptions{ScreenReaderMode: true}
		assert.Equal(t, "► ", opts.GetFocusIndicator())
		
		opts = AccessibilityOptions{HighContrast: true}
		assert.Equal(t, ">> ", opts.GetFocusIndicator())
	})

	t.Run("GetKeyboardHelp", func(t *testing.T) {
		opts := AccessibilityOptions{}
		result := opts.GetKeyboardHelp()
		assert.Contains(t, result, "arrows")
		
		opts = AccessibilityOptions{VerboseDescriptions: true}
		result = opts.GetKeyboardHelp()
		assert.Contains(t, result, "Navigation:")
		
		opts = AccessibilityOptions{Enabled: true}
		result = opts.GetKeyboardHelp()
		assert.Contains(t, result, "navigate")
	})
}

func TestValidateAccessibilitySettings(t *testing.T) {
	t.Run("Screen reader enables accessibility", func(t *testing.T) {
		opts := &AccessibilityOptions{ScreenReaderMode: true}
		ValidateAccessibilitySettings(opts)
		
		assert.True(t, opts.Enabled)
		assert.True(t, opts.VerboseDescriptions)
	})

	t.Run("High contrast enables accessibility", func(t *testing.T) {
		opts := &AccessibilityOptions{HighContrast: true}
		ValidateAccessibilitySettings(opts)
		
		assert.True(t, opts.Enabled)
	})

	t.Run("Simplified UI enables reduced motion", func(t *testing.T) {
		opts := &AccessibilityOptions{SimplifiedUI: true}
		ValidateAccessibilitySettings(opts)
		
		assert.True(t, opts.ReducedMotion)
	})
}

func TestParseEnvBool(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
		def      bool
	}{
		{"empty uses default", "", true, true},
		{"empty uses default false", "", false, false},
		{"true", "true", true, false},
		{"false", "false", false, true},
		{"yes", "yes", true, false},
		{"no", "no", false, true},
		{"1", "1", true, false},
		{"0", "0", false, true},
		{"on", "on", true, false},
		{"off", "off", false, true},
		{"enable", "enable", true, false},
		{"disable", "disable", false, true},
		{"invalid uses default", "invalid", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue == "" {
				os.Unsetenv("TEST_ENV")
			} else {
				os.Setenv("TEST_ENV", tt.envValue)
			}
			
			result := parseEnvBool("TEST_ENV", tt.def)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCurrentAccessibilityOptions(t *testing.T) {
	// This is an integration test for the full flow
	clearEnvVars()
	os.Setenv("SCREEN_READER", "1")
	os.Setenv("HIGH_CONTRAST", "true")
	
	opts := GetCurrentAccessibilityOptions()
	
	// Should be validated automatically
	assert.True(t, opts.Enabled)
	assert.True(t, opts.ScreenReaderMode)
	assert.True(t, opts.HighContrast)
	assert.True(t, opts.VerboseDescriptions) // Set by validation
}

// Helper function to clear environment variables
func clearEnvVars() {
	envVars := []string{
		"ACCESSIBLE", "ACCESSIBILITY", "A11Y",
		"SCREEN_READER", "NVDA", "JAWS", "ORCA",
		"HIGH_CONTRAST", "CONTRAST",
		"REDUCED_MOTION", "MOTION",
		"NO_COLOR", "MONOCHROME",
		"VERBOSE_DESCRIPTIONS", "ACCESSIBILITY_VERBOSE",
		"SIMPLE_UI", "ACCESSIBILITY_SIMPLE",
	}
	
	for _, env := range envVars {
		os.Unsetenv(env)
	}
}