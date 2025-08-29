package tui

import (
	"fmt"
	"strings"

	ui "github.com/mrtkrcm/ZeroUI/internal/tui/components/ui"
)

// UIImplementation represents different UI implementation options
type UIImplementation string

const (
	UIImplementationStandard    UIImplementation = "standard"
	UIImplementationEnhanced    UIImplementation = "enhanced"
	UIImplementationDelightful  UIImplementation = "delightful"
	UIImplementationMinimal     UIImplementation = "minimal"
)

// UISelector manages UI implementation selection and switching
type UISelector struct {
	currentImplementation UIImplementation
	availableImplementations []UIImplementation
	implementationConfigs   map[UIImplementation]*UIConfig
}

// UIConfig holds configuration for a specific UI implementation
type UIConfig struct {
	Name         string
	Description  string
	Features     []string
	Complexity   string // "low", "medium", "high"
	Performance  string // "fast", "balanced", "rich"
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
	// Initialize enhanced UI manager
	uiManager := ui.NewUIIntegrationManager()
	uiManager.Initialize(120, 40)

	// Create enhanced components
	enhancedList := ui.NewEnhancedApplicationList()

	// For now, create a standard model and enhance it
	model, err := NewModel(nil, initialApp, nil)
	if err != nil {
		return nil, err
	}

	// Add enhanced components to the model
	model.uiManager = uiManager
	model.enhancedAppList = enhancedList

	return model, nil
}

// createDelightfulModel creates a delightful model with animations and effects
func (uis *UISelector) createDelightfulModel(initialApp string) (*Model, error) {
	// Create a standard model for now
	// TODO: Implement delightful UI components
	model, err := NewModel(nil, initialApp, nil)
	return model, err
}

// createMinimalModel creates a minimal model with essential features only
func (uis *UISelector) createMinimalModel(initialApp string) (*Model, error) {
	// Create a standard model for now
	// TODO: Implement minimal UI components
	model, err := NewModel(nil, initialApp, nil)
	return model, err
}

// GetRecommendedImplementation returns the recommended implementation based on context
func (uis *UISelector) GetRecommendedImplementation() UIImplementation {
	// For now, recommend enhanced UI as the best balance
	return UIImplementationEnhanced
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
