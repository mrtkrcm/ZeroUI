package accessibility

import (
	"os"
	"strconv"
	"strings"
)

// AccessibilityOptions holds accessibility configuration
type AccessibilityOptions struct {
	Enabled             bool
	ScreenReaderMode    bool
	HighContrast        bool
	ReducedMotion       bool
	NoColor             bool
	VerboseDescriptions bool
	SimplifiedUI        bool
}

// DetectAccessibilityNeeds detects accessibility requirements from environment
func DetectAccessibilityNeeds() AccessibilityOptions {
	opts := AccessibilityOptions{}

	// Check environment variables for accessibility preferences
	opts.Enabled = os.Getenv("ACCESSIBLE") != "" ||
		os.Getenv("ACCESSIBILITY") != "" ||
		os.Getenv("A11Y") != ""

	opts.ScreenReaderMode = os.Getenv("SCREEN_READER") != "" ||
		os.Getenv("NVDA") != "" ||
		os.Getenv("JAWS") != "" ||
		os.Getenv("ORCA") != ""

	opts.HighContrast = os.Getenv("HIGH_CONTRAST") != "" ||
		strings.ToLower(os.Getenv("CONTRAST")) == "high"

	opts.ReducedMotion = os.Getenv("REDUCED_MOTION") != "" ||
		strings.ToLower(os.Getenv("MOTION")) == "reduced"

	opts.NoColor = os.Getenv("NO_COLOR") != "" ||
		os.Getenv("MONOCHROME") != ""

	opts.VerboseDescriptions = os.Getenv("VERBOSE_DESCRIPTIONS") != "" ||
		parseEnvBool("ACCESSIBILITY_VERBOSE", false)

	opts.SimplifiedUI = os.Getenv("SIMPLE_UI") != "" ||
		parseEnvBool("ACCESSIBILITY_SIMPLE", false)

	// Auto-enable accessibility if any specific feature is requested
	if opts.ScreenReaderMode || opts.HighContrast || opts.ReducedMotion {
		opts.Enabled = true
	}

	return opts
}

// GetHuhAccessibilityMode returns the accessibility setting for Huh forms
func (opts AccessibilityOptions) GetHuhAccessibilityMode() bool {
	return opts.Enabled || opts.ScreenReaderMode
}

// GetColorMode returns whether colors should be used
func (opts AccessibilityOptions) GetColorMode() ColorMode {
	if opts.NoColor {
		return ColorModeNone
	}
	if opts.HighContrast {
		return ColorModeHighContrast
	}
	return ColorModeNormal
}

// GetAnimationMode returns whether animations should be used
func (opts AccessibilityOptions) GetAnimationMode() AnimationMode {
	if opts.ReducedMotion {
		return AnimationModeNone
	}
	return AnimationModeNormal
}

// ColorMode represents color display preferences
type ColorMode int

const (
	ColorModeNormal ColorMode = iota
	ColorModeHighContrast
	ColorModeNone
)

// AnimationMode represents animation preferences
type AnimationMode int

const (
	AnimationModeNormal AnimationMode = iota
	AnimationModeNone
)

// parseEnvBool parses an environment variable as a boolean
func parseEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		// Try common string representations
		switch strings.ToLower(value) {
		case "yes", "y", "on", "1", "true", "enable", "enabled":
			return true
		case "no", "n", "off", "0", "false", "disable", "disabled":
			return false
		default:
			return defaultValue
		}
	}

	return parsed
}

// GetDescriptionLevel returns appropriate description verbosity
func (opts AccessibilityOptions) GetDescriptionLevel() DescriptionLevel {
	if opts.VerboseDescriptions || opts.ScreenReaderMode {
		return DescriptionLevelVerbose
	}
	if opts.Enabled {
		return DescriptionLevelDetailed
	}
	return DescriptionLevelBasic
}

// DescriptionLevel represents how verbose descriptions should be
type DescriptionLevel int

const (
	DescriptionLevelBasic DescriptionLevel = iota
	DescriptionLevelDetailed
	DescriptionLevelVerbose
)

// GetAccessibleTitle creates an accessible title with context
func (opts AccessibilityOptions) GetAccessibleTitle(title, context string) string {
	if !opts.Enabled {
		return title
	}

	if context != "" && opts.VerboseDescriptions {
		return title + " - " + context
	}

	return title
}

// GetAccessibleHelp provides context-appropriate help text
func (opts AccessibilityOptions) GetAccessibleHelp(basicHelp, detailedHelp string) string {
	switch opts.GetDescriptionLevel() {
	case DescriptionLevelVerbose:
		if detailedHelp != "" {
			return detailedHelp
		}
		return basicHelp + " (Use arrow keys to navigate, Enter to select, Esc to go back)"
	case DescriptionLevelDetailed:
		if detailedHelp != "" {
			return detailedHelp
		}
		return basicHelp + " (Arrow keys: navigate, Enter: select)"
	default:
		return basicHelp
	}
}

// GetStatusDescription provides accessible status descriptions
func (opts AccessibilityOptions) GetStatusDescription(status string) string {
	if !opts.Enabled {
		return status
	}

	statusDescriptions := map[string]string{
		"configured":   "Application is configured and ready",
		"needs_config": "Application needs configuration",
		"error":        "Application has configuration errors",
		"unknown":      "Application status is unknown",
		"loading":      "Loading application information",
		"success":      "Operation completed successfully",
		"warning":      "Warning: attention required",
	}

	if description, exists := statusDescriptions[status]; exists {
		return description
	}

	return status
}

// ShouldUseProgressBar determines if progress bars should be shown
func (opts AccessibilityOptions) ShouldUseProgressBar() bool {
	return !opts.ReducedMotion && !opts.SimplifiedUI
}

// ShouldUseSpinner determines if spinners should be shown
func (opts AccessibilityOptions) ShouldUseSpinner() bool {
	return !opts.ReducedMotion && !opts.SimplifiedUI
}

// GetFocusIndicator returns appropriate focus indication
func (opts AccessibilityOptions) GetFocusIndicator() string {
	if opts.ScreenReaderMode {
		return "► " // Clear focus indicator for screen readers
	}
	if opts.HighContrast {
		return ">> " // High contrast focus indicator
	}
	return "• " // Standard focus indicator
}

// GetKeyboardHelp returns context-appropriate keyboard help
func (opts AccessibilityOptions) GetKeyboardHelp() string {
	if opts.GetDescriptionLevel() == DescriptionLevelVerbose {
		return "Navigation: Arrow keys or hjkl, Select: Enter or Space, Back: Escape, Help: Question mark, Quit: q or Ctrl+C"
	}
	if opts.GetDescriptionLevel() == DescriptionLevelDetailed {
		return "↑↓←→/hjkl: navigate, Enter/Space: select, Esc: back, ?: help, q: quit"
	}
	return "arrows: move, enter: select, ?: help, q: quit"
}

// ValidateAccessibilitySettings validates and adjusts accessibility settings
func ValidateAccessibilitySettings(opts *AccessibilityOptions) {
	// Ensure consistent settings
	if opts.ScreenReaderMode {
		opts.Enabled = true
		opts.VerboseDescriptions = true
	}

	if opts.HighContrast {
		opts.Enabled = true
	}

	if opts.SimplifiedUI {
		opts.ReducedMotion = true
	}
}

// GetCurrentAccessibilityOptions returns the current accessibility configuration
func GetCurrentAccessibilityOptions() AccessibilityOptions {
	opts := DetectAccessibilityNeeds()
	ValidateAccessibilitySettings(&opts)
	return opts
}
