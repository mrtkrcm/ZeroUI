package core

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Component represents a basic TUI component
type Component interface {
	tea.Model
}

// Focusable represents a component that can receive focus
type Focusable interface {
	Component
	Focus() tea.Cmd
	Blur() tea.Cmd
	IsFocused() bool
}

// Sizeable represents a component that can be resized
type Sizeable interface {
	Component
	SetSize(width, height int) tea.Cmd
	GetSize() (int, int)
}

// Positionable represents a component that can be positioned
type Positionable interface {
	Component
	SetPosition(x, y int) tea.Cmd
	GetPosition() (int, int)
}

// KeyHandler represents a component that handles key bindings
type KeyHandler interface {
	Component
	KeyBindings() []key.Binding
	HandleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd)
}

// Layout represents a component that manages layout
type Layout interface {
	Sizeable
	AddChild(child Component)
	RemoveChild(child Component)
	Children() []Component
}

// Container represents a component that can contain other components
type Container interface {
	Layout
	Focusable
}

// Page represents a full-screen page component
type Page interface {
	Container
	ID() string
	Title() string
	Description() string
}

// Dialog represents a modal dialog component
type Dialog interface {
	Focusable
	Sizeable
	ID() string
	IsModal() bool
	OnClose() tea.Cmd
}

// List represents a list component
type List interface {
	Focusable
	Sizeable
	SetItems(items []ListItem) tea.Cmd
	GetItems() []ListItem
	SelectedIndex() int
	SelectedItem() ListItem
	SetSelected(index int) tea.Cmd
	MoveUp() tea.Cmd
	MoveDown() tea.Cmd
}

// ListItem represents an item in a list
type ListItem interface {
	ID() string
	Title() string
	Description() string
	FilterValue() string
}

// Form represents a form component with type-safe configuration
type Form interface {
	Focusable
	Sizeable
	AddField(field FormField) tea.Cmd
	RemoveField(id string) tea.Cmd
	GetField(id string) FormField
	GetFields() []FormField
	IsValid() bool
	GetData() ConfigData
	GetLegacyData() map[string]interface{} // For backward compatibility
	Reset() tea.Cmd
	Validate() []ValidationError
}

// FormField represents a field in a form with type safety
type FormField interface {
	Focusable
	ID() string
	Label() string
	Value() ConfigValue
	SetValue(value ConfigValue) tea.Cmd
	IsValid() bool
	ValidationError() string
	Type() ValueType
}

// StatusBar represents a status bar component
type StatusBar interface {
	Sizeable
	SetStatus(status string) tea.Cmd
	SetMessage(message string, duration int) tea.Cmd
	SetProgress(current, total int) tea.Cmd
	AddKeyBinding(binding key.Binding) tea.Cmd
	RemoveKeyBinding(keys string) tea.Cmd
}

// HelpSystem represents a help system component
type HelpSystem interface {
	Sizeable
	SetKeyBindings(bindings []key.Binding) tea.Cmd
	ShowFullHelp() tea.Cmd
	ShowShortHelp() tea.Cmd
	ToggleHelp() tea.Cmd
	IsShowingFullHelp() bool
}
