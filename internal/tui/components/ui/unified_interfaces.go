package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrtkrcm/ZeroUI/internal/tui/components/core"
)

// UnifiedComponent represents a fully integrated TUI component
type UnifiedComponent interface {
	core.Component
	core.Sizeable
	core.Focusable
	core.KeyHandler

	// Component identification
	ID() string
	Title() string
	Description() string

	// State management
	IsActive() bool
	SetActive(bool) tea.Cmd
	GetState() interface{}
	SetState(interface{}) tea.Cmd

	// Data management
	GetData() interface{}
	SetData(interface{}) tea.Cmd

	// Validation
	IsValid() bool
	Validate() []ValidationError

	// Lifecycle
	OnMount() tea.Cmd
	OnUnmount() tea.Cmd
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Type    ValidationErrorType
}

type ValidationErrorType string

const (
	ErrorTypeRequired ValidationErrorType = "required"
	ErrorTypeInvalid  ValidationErrorType = "invalid"
	ErrorTypeFormat   ValidationErrorType = "format"
	ErrorTypeCustom   ValidationErrorType = "custom"
)

// ComponentManager manages unified components
type ComponentManager struct {
	components map[string]UnifiedComponent
	active     string
	layout     *LayoutManager
	styles     *StylingManager
}

// NewComponentManager creates a new component manager
func NewComponentManager() *ComponentManager {
	return &ComponentManager{
		components: make(map[string]UnifiedComponent),
		layout:     NewLayoutManager(),
		styles:     NewStylingManager(),
	}
}

// RegisterComponent registers a unified component
func (cm *ComponentManager) RegisterComponent(comp UnifiedComponent) {
	cm.components[comp.ID()] = comp
}

// GetComponent gets a component by ID
func (cm *ComponentManager) GetComponent(id string) UnifiedComponent {
	return cm.components[id]
}

// SetActive sets the active component
func (cm *ComponentManager) SetActive(id string) tea.Cmd {
	if cm.active != "" {
		if comp := cm.components[cm.active]; comp != nil {
			comp.SetActive(false)
		}
	}

	cm.active = id
	if comp := cm.components[id]; comp != nil {
		return comp.SetActive(true)
	}
	return nil
}

// GetActive returns the active component
func (cm *ComponentManager) GetActive() UnifiedComponent {
	if cm.active == "" {
		return nil
	}
	return cm.components[cm.active]
}

// GetAllComponents returns all registered components
func (cm *ComponentManager) GetAllComponents() map[string]UnifiedComponent {
	return cm.components
}

// Update updates all components
func (cm *ComponentManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	for _, comp := range cm.components {
		_, cmd := comp.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return cm, tea.Batch(cmds...)
}

// View renders the component manager
func (cm *ComponentManager) View() string {
	if cm.active == "" {
		return "No active component"
	}

	comp := cm.components[cm.active]
	if comp == nil {
		return "Component not found"
	}

	return comp.View()
}

// Init initializes all components
func (cm *ComponentManager) Init() tea.Cmd {
	var cmds []tea.Cmd

	for _, comp := range cm.components {
		if cmd := comp.OnMount(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

// GetKeyBindings returns all key bindings from active component
func (cm *ComponentManager) GetKeyBindings() []key.Binding {
	if comp := cm.GetActive(); comp != nil {
		return comp.KeyBindings()
	}
	return []key.Binding{}
}

// HandleKey handles key events for active component
func (cm *ComponentManager) HandleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if comp := cm.GetActive(); comp != nil {
		return comp.HandleKey(msg)
	}
	return cm, nil
}

// LayoutManager manages component layout
type LayoutManager struct {
	regions map[string]LayoutRegion
}

// LayoutRegion represents a layout region
type LayoutRegion struct {
	X      int
	Y      int
	Width  int
	Height int
}

// NewLayoutManager creates a new layout manager
func NewLayoutManager() *LayoutManager {
	return &LayoutManager{
		regions: make(map[string]LayoutRegion),
	}
}

// SetRegion sets a layout region
func (lm *LayoutManager) SetRegion(name string, region LayoutRegion) {
	lm.regions[name] = region
}

// GetRegion gets a layout region
func (lm *LayoutManager) GetRegion(name string) (LayoutRegion, bool) {
	region, exists := lm.regions[name]
	return region, exists
}

// StylingManager manages component styling
type StylingManager struct {
	theme interface{} // Will be connected to the actual theme system
}

// NewStylingManager creates a new styling manager
func NewStylingManager() *StylingManager {
	return &StylingManager{}
}
