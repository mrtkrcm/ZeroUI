package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

// AppKeyMap defines key bindings for the application
type AppKeyMap struct {
	// Navigation
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding

	// Selection
	Enter key.Binding
	Space key.Binding
	Tab   key.Binding

	// Actions
	Toggle    key.Binding
	Cycle     key.Binding
	Presets   key.Binding
	Back      key.Binding
	Help      key.Binding
	Quit      key.Binding
	ForceQuit key.Binding

	// Advanced
	Search key.Binding
	Reset  key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() AppKeyMap {
	return AppKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left/previous value"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right/next value"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select/confirm"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle/cycle value"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch focus"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle value"),
		),
		Cycle: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "cycle through values"),
		),
		Presets: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "open presets"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "go back"),
		),
		Help: key.NewBinding(
			key.WithKeys("?", "h"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "force quit"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset to default"),
		),
	}
}

// ShortHelp returns key bindings to be shown in the mini help view
func (k AppKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k AppKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Navigation
		{k.Up, k.Down, k.Left, k.Right},
		// Selection
		{k.Enter, k.Space, k.Tab},
		// Actions
		{k.Toggle, k.Cycle, k.Presets, k.Reset},
		// System
		{k.Help, k.Back, k.Quit, k.Search},
	}
}
