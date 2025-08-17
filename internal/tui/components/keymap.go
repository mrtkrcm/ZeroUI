package components

import (
	"github.com/charmbracelet/bubbles/key"
)

// EnhancedKeyMap provides enhanced key bindings with better UX
type EnhancedKeyMap struct {
	// Navigation
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding

	// Selection and actions
	Enter  key.Binding
	Space  key.Binding
	Tab    key.Binding
	Escape key.Binding

	// Application specific
	Toggle key.Binding
	Cycle  key.Binding
	Reset  key.Binding
	Save   key.Binding
	Reload key.Binding

	// Interface
	Search   key.Binding
	Filter   key.Binding
	Help     key.Binding
	Quit     key.Binding
	FullHelp key.Binding

	// Advanced
	Debug      key.Binding
	Screenshot key.Binding
	Export     key.Binding
}

// NewEnhancedKeyMap creates an enhanced key mapping with better UX
func NewEnhancedKeyMap() EnhancedKeyMap {
	return EnhancedKeyMap{
		// Navigation - follows vim/standard conventions
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
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("PgUp", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("PgDn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("Home/g", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("End/G", "go to bottom"),
		),

		// Selection and actions
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("Enter", "select/confirm"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("Space", "quick toggle"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("Tab", "next field"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("Esc", "back/cancel"),
		),

		// Application specific
		Toggle: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle value"),
		),
		Cycle: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "cycle values"),
		),
		Reset: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("Ctrl+R", "reset to default"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("Ctrl+S", "save changes"),
		),
		Reload: key.NewBinding(
			key.WithKeys("r", "f5"),
			key.WithHelp("r/F5", "refresh"),
		),

		// Interface
		Search: key.NewBinding(
			key.WithKeys("/", "ctrl+f"),
			key.WithHelp("/", "search"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter"),
		),
		Help: key.NewBinding(
			key.WithKeys("?", "h"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		FullHelp: key.NewBinding(
			key.WithKeys("F1"),
			key.WithHelp("F1", "full help"),
		),

		// Advanced
		Debug: key.NewBinding(
			key.WithKeys("ctrl+alt+d"),
			key.WithHelp("Ctrl+Alt+D", "debug mode"),
		),
		Screenshot: key.NewBinding(
			key.WithKeys("ctrl+alt+s"),
			key.WithHelp("Ctrl+Alt+S", "screenshot"),
		),
		Export: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("Ctrl+E", "export config"),
		),
	}
}

// GetContextualHelp returns help items for the current context
func (k EnhancedKeyMap) GetContextualHelp(context string) []HelpItem {
	switch context {
	case "app_grid":
		return []HelpItem{
			{"↑↓←→/hjkl", "Navigate apps"},
			{"Enter", "Open app config"},
			{"Space", "Quick toggle"},
			{"/", "Search apps"},
			{"f", "Filter by status"},
			{"r", "Refresh"},
			{"?", "Toggle help"},
			{"q", "Quit"},
		}
	case "config_edit":
		return []HelpItem{
			{"↑↓/jk", "Navigate fields"},
			{"Enter", "Edit field"},
			{"Space", "Quick toggle"},
			{"Tab", "Next field"},
			{"Ctrl+S", "Save config"},
			{"Ctrl+R", "Reset field"},
			{"Esc", "Back to grid"},
			{"?", "Toggle help"},
		}
	case "search":
		return []HelpItem{
			{"Type", "Search query"},
			{"↑↓/jk", "Navigate results"},
			{"Enter", "Select result"},
			{"Esc", "Cancel search"},
		}
	default:
		return []HelpItem{
			{"?", "Show help"},
			{"q", "Quit"},
		}
	}
}
