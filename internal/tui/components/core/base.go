package core

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// BaseComponent provides common functionality for all components
type BaseComponent struct {
	id       string
	width    int
	height   int
	x        int
	y        int
	focused  bool
	visible  bool
	children []Component
}

// NewBaseComponent creates a new base component
func NewBaseComponent(id string) *BaseComponent {
	return &BaseComponent{
		id:      id,
		visible: true,
	}
}

// ID returns the component ID
func (b *BaseComponent) ID() string {
	return b.id
}

// SetSize sets the component size
func (b *BaseComponent) SetSize(width, height int) tea.Cmd {
	b.width = width
	b.height = height
	return nil
}

// GetSize returns the component size
func (b *BaseComponent) GetSize() (int, int) {
	return b.width, b.height
}

// SetPosition sets the component position
func (b *BaseComponent) SetPosition(x, y int) tea.Cmd {
	b.x = x
	b.y = y
	return nil
}

// GetPosition returns the component position
func (b *BaseComponent) GetPosition() (int, int) {
	return b.x, b.y
}

// Focus sets the component as focused
func (b *BaseComponent) Focus() tea.Cmd {
	b.focused = true
	return nil
}

// Blur removes focus from the component
func (b *BaseComponent) Blur() tea.Cmd {
	b.focused = false
	return nil
}

// IsFocused returns whether the component is focused
func (b *BaseComponent) IsFocused() bool {
	return b.focused
}

// SetVisible sets the component visibility
func (b *BaseComponent) SetVisible(visible bool) tea.Cmd {
	b.visible = visible
	return nil
}

// IsVisible returns whether the component is visible
func (b *BaseComponent) IsVisible() bool {
	return b.visible
}

// AddChild adds a child component
func (b *BaseComponent) AddChild(child Component) {
	b.children = append(b.children, child)
}

// RemoveChild removes a child component
func (b *BaseComponent) RemoveChild(child Component) {
	for i, c := range b.children {
		if c == child {
			b.children = append(b.children[:i], b.children[i+1:]...)
			break
		}
	}
}

// Children returns all child components
func (b *BaseComponent) Children() []Component {
	return b.children
}

// Width returns the component width
func (b *BaseComponent) Width() int {
	return b.width
}

// Height returns the component height
func (b *BaseComponent) Height() int {
	return b.height
}

// X returns the component X position
func (b *BaseComponent) X() int {
	return b.x
}

// Y returns the component Y position
func (b *BaseComponent) Y() int {
	return b.y
}

// Default implementations for tea.Model interface
func (b *BaseComponent) Init() tea.Cmd {
	return nil
}

func (b *BaseComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return b, nil
}

func (b *BaseComponent) View() string {
	return ""
}

// KeyMap represents key bindings for a component
type KeyMap struct {
	bindings []key.Binding
}

// NewKeyMap creates a new key map
func NewKeyMap() *KeyMap {
	return &KeyMap{
		bindings: make([]key.Binding, 0),
	}
}

// Add adds a key binding
func (k *KeyMap) Add(binding key.Binding) {
	k.bindings = append(k.bindings, binding)
}

// Remove removes a key binding by key
func (k *KeyMap) Remove(keys string) {
	for i, binding := range k.bindings {
		if binding.Keys()[0] == keys {
			k.bindings = append(k.bindings[:i], k.bindings[i+1:]...)
			break
		}
	}
}

// Bindings returns all key bindings
func (k *KeyMap) Bindings() []key.Binding {
	return k.bindings
}

// Clear removes all key bindings
func (k *KeyMap) Clear() {
	k.bindings = make([]key.Binding, 0)
}

// DefaultKeyBindings returns common key bindings
func DefaultKeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("ctrl+c/q", "quit"),
		),
		key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
		key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev"),
		),
	}
}

// NavigationKeyBindings returns navigation key bindings
func NavigationKeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select"),
		),
	}
}