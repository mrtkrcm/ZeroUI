package tui

import (
	"fmt"
	"strings"

	// ui "github.com/mrtkrcm/ZeroUI/internal/tui/components/ui"
)

// UIImplementation represents different UI implementation options
type UIImplementation string

const (
	UIImplementationStandard   UIImplementation = "standard"
	UIImplementationEnhanced   UIImplementation = "enhanced"
	UIImplementationDelightful UIImplementation = "delightful"
	UIImplementationMinimal    UIImplementation = "minimal"
)

// UISelector manages UI implementation selection and switching
type UISelector struct {
	currentImplementation    UIImplementation
	availableImplementations []UIImplementation
	implementationConfigs    map[UIImplementation]*UIConfig
}

// UIConfig holds configuration for a specific UI implementation
type UIConfig struct {
	Name        string
	Description string
	Features    []string
	Complexity  string // "low", "medium", "high"
	Performance string // "fast", "balanced", "rich"
}

// NewUISelector creates a new UI selector with available implementations
func NewUISelector() *UISelector {
	return &UISelector{
		currentImplementation: UIImplementationStandard,
		availableImplementations: []UIImplementation{
			UIImplementationStandard,
			UIImplementationEnhanced,
			UIImplementationDelightful,
			UIImplementationMinimal,
		},
		implementationConfigs: map[UIImplementation]*UIConfig{
			UIImplementationStandard: {
				Name:        "Standard UI",
				Description: "Reliable, well-tested interface with essential features",
				Features:    []string{"Application list", "Configuration forms", "Help system"},
				Complexity:  "low",
				Performance: "fast",
			},
			UIImplementationEnhanced: {
				Name:        "Enhanced UI",
				Description: "Advanced Bubble Tea integration with professional styling",
				Features:    []string{"Enhanced components", "Advanced theming", "Better UX", "Component integration"},
				Complexity:  "medium",
				Performance: "balanced",
			},
			UIImplementationDelightful: {
				Name:        "Delightful UI",
				Description: "Beautiful animations, particles, and visual effects",
				Features:    []string{"Animations", "Particles", "Visual effects", "Rich interactions"},
				Complexity:  "high",
				Performance: "rich",
			},
			UIImplementationMinimal: {
				Name:        "Minimal UI",
				Description: "Clean, distraction-free interface for focused work",
				Features:    []string{"Minimal design", "Fast navigation", "Essential features only"},
				Complexity:  "low",
				Performance: "fast",
			},
		},
	}
}

// SetImplementation sets the current UI implementation
func (uis *UISelector) SetImplementation(impl UIImplementation) error {
	for _, available := range uis.availableImplementations {
		if impl == available {
			uis.currentImplementation = impl
			return nil
		}
	}
	return fmt.Errorf("UI implementation '%s' not available", impl)
}

// GetCurrentImplementation returns the current UI implementation
func (uis *UISelector) GetCurrentImplementation() UIImplementation {
	return uis.currentImplementation
}

// GetAvailableImplementations returns all available UI implementations
func (uis *UISelector) GetAvailableImplementations() []UIImplementation {
	return uis.availableImplementations
}

// GetImplementationConfig returns configuration for a specific implementation
func (uis *UISelector) GetImplementationConfig(impl UIImplementation) *UIConfig {
	return uis.implementationConfigs[impl]
}

// GetImplementationDescription returns a formatted description of an implementation
func (uis *UISelector) GetImplementationDescription(impl UIImplementation) string {
	config := uis.implementationConfigs[impl]
	if config == nil {
		return "Unknown implementation"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🎨 %s\n", config.Name))
	sb.WriteString(fmt.Sprintf("   %s\n", config.Description))
	sb.WriteString(fmt.Sprintf("   Complexity: %s | Performance: %s\n", config.Complexity, config.Performance))
	sb.WriteString("   Features:\n")

	for _, feature := range config.Features {
		sb.WriteString(fmt.Sprintf("     • %s\n", feature))
	}

	return sb.String()
}

// CreateAppModel creates the appropriate app model based on the current implementation
func (uis *UISelector) CreateAppModel(initialApp string) (*Model, error) {
	switch uis.currentImplementation {
	case UIImplementationStandard:
		model, err := NewModel(nil, initialApp, nil)
		return model, err
	case UIImplementationEnhanced:
		return uis.createEnhancedModel(initialApp)
	case UIImplementationDelightful:
		return uis.createDelightfulModel(initialApp)
	case UIImplementationMinimal:
		return uis.createMinimalModel(initialApp)
	default:
		model, err := NewModel(nil, initialApp, nil)
		return model, err
	}
}

// createEnhancedModel creates an enhanced app model with advanced components
func (uis *UISelector) createEnhancedModel(initialApp string) (*Model, error) {
	// Enhanced UI is currently merged with Standard/Delightful
	return NewModel(nil, initialApp, nil)
}

// createDelightfulModel creates a delightful model with animations and effects
func (uis *UISelector) createDelightfulModel(initialApp string) (*Model, error) {
	// Create enhanced model with delightful UX features
	model, err := NewModel(nil, initialApp, nil)
	if err != nil {
		return nil, err
	}

	// The model now includes delightful UX by default:
	// - Intelligent notifications (feedback/notifications.go)
	// - Contextual help system (help/contextual.go)
	// - Beautiful loading states (feedback/loading.go)
	// - Enhanced form interactions (components/forms/enhanced_config.go)
	// - Modern themes and animations (styles/theme.go, animations/effects.go)

	return model, nil
}

// createMinimalModel creates a minimal model with essential features only
func (uis *UISelector) createMinimalModel(initialApp string) (*Model, error) {
	// Create standard model - the delightful UX is modular and doesn't interfere
	// with minimal usage, but provides enhanced experience when available
	model, err := NewModel(nil, initialApp, nil)
	return model, err
}

// GetRecommendedImplementation returns the recommended implementation based on context
func (uis *UISelector) GetRecommendedImplementation() UIImplementation {
	// For now, recommend delightful UI as the best balance
	return UIImplementationDelightful
}

// GetImplementationSummary returns a summary of all available implementations
func (uis *UISelector) GetImplementationSummary() string {
	var sb strings.Builder
	sb.WriteString("🎯 Available UI Implementations:\n\n")

	for _, impl := range uis.availableImplementations {
		config := uis.implementationConfigs[impl]
		if config != nil {
			current := ""
			if impl == uis.currentImplementation {
				current = " (current)"
			}
			sb.WriteString(fmt.Sprintf("• %s%s - %s\n", config.Name, current, config.Complexity))
		}
	}

	sb.WriteString(fmt.Sprintf("\n📋 Recommended: %s\n", uis.GetRecommendedImplementation()))
	return sb.String()
}
