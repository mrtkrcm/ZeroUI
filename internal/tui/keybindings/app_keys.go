package keybindings

import (
	"github.com/charmbracelet/bubbles/key"
)

// AppKeyMap defines the key bindings for the application
type AppKeyMap struct {
	// Navigation
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Enter  key.Binding
	Select key.Binding
	Back   key.Binding
	Home   key.Binding
	End    key.Binding

	// Application actions
	Refresh key.Binding
	Edit    key.Binding
	Save    key.Binding
	Cancel  key.Binding
	Reset   key.Binding

	// UI controls
	Search    key.Binding
	Filter    key.Binding
	Help      key.Binding
	Quit      key.Binding
	ForceQuit key.Binding

	// View switching
	ToggleMode    key.Binding
	TogglePreview key.Binding
	ToggleHelp    key.Binding

	// Form navigation
	NextField  key.Binding
	PrevField  key.Binding
	SubmitForm key.Binding
	CancelForm key.Binding

	// Advanced
	Debug    key.Binding
	Settings key.Binding
}

// NewAppKeyMap creates a new key map with default bindings
func NewAppKeyMap() AppKeyMap {
	return AppKeyMap{
		// Navigation
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
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select/confirm"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select item"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "go back"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "go to bottom"),
		),

		// Application actions
		Refresh: key.NewBinding(
			key.WithKeys("r", "F5"),
			key.WithHelp("r/F5", "refresh"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e", "F2"),
			key.WithHelp("e/F2", "edit"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("ctrl+c/esc", "cancel"),
		),
		Reset: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "reset"),
		),

		// UI controls
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter"),
		),
		Help: key.NewBinding(
			key.WithKeys("?", "h"),
			key.WithHelp("?/h", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "force quit"),
		),

		// View switching
		ToggleMode: key.NewBinding(
			key.WithKeys("ctrl+m"),
			key.WithHelp("ctrl+m", "toggle mode"),
		),
		TogglePreview: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "toggle preview"),
		),
		ToggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),

		// Form navigation
		NextField: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		PrevField: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous field"),
		),
		SubmitForm: key.NewBinding(
			key.WithKeys("enter", "ctrl+s"),
			key.WithHelp("enter/ctrl+s", "submit form"),
		),
		CancelForm: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc/ctrl+c", "cancel form"),
		),

		// Advanced
		Debug: key.NewBinding(
			key.WithKeys("ctrl+shift+d"),
			key.WithHelp("ctrl+shift+d", "debug info"),
		),
		Settings: key.NewBinding(
			key.WithKeys("ctrl+,"),
			key.WithHelp("ctrl+,", "settings"),
		),
	}
}

// ShortHelp returns key bindings to be shown in the mini help view
func (k AppKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Help, k.Select, k.Back, k.Quit,
	}
}

// FullHelp returns key bindings to be shown in the full help view
func (k AppKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Navigation
		{k.Up, k.Down, k.Left, k.Right},
		{k.Home, k.End, k.Enter, k.Back},

		// Actions
		{k.Select, k.Edit, k.Refresh, k.Save},
		{k.Search, k.Filter, k.Reset, k.Cancel},

		// UI Controls
		{k.Help, k.ToggleMode, k.TogglePreview, k.Settings},
		{k.Quit, k.ForceQuit, k.Debug},
	}
}

// FormKeyMap provides key bindings specific to forms
type FormKeyMap struct {
	NextField key.Binding
	PrevField key.Binding
	Submit    key.Binding
	Cancel    key.Binding
	Reset     key.Binding
	Clear     key.Binding
}

// NewFormKeyMap creates form-specific key bindings
func NewFormKeyMap() FormKeyMap {
	return FormKeyMap{
		NextField: key.NewBinding(
			key.WithKeys("tab", "down"),
			key.WithHelp("tab/↓", "next field"),
		),
		PrevField: key.NewBinding(
			key.WithKeys("shift+tab", "up"),
			key.WithHelp("shift+tab/↑", "previous field"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter", "ctrl+s"),
			key.WithHelp("enter/ctrl+s", "submit"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Reset: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "reset form"),
		),
		Clear: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "clear field"),
		),
	}
}

// ListKeyMap provides key bindings for list navigation
type ListKeyMap struct {
	Up          key.Binding
	Down        key.Binding
	PageUp      key.Binding
	PageDown    key.Binding
	Home        key.Binding
	End         key.Binding
	Select      key.Binding
	Filter      key.Binding
	ClearFilter key.Binding
	Refresh     key.Binding
}

// NewListKeyMap creates list-specific key bindings
func NewListKeyMap() ListKeyMap {
	return ListKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("pgup/b", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "f"),
			key.WithHelp("pgdown/f", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "bottom"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys("ctrl+l", "esc"),
			key.WithHelp("ctrl+l/esc", "clear filter"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r", "F5"),
			key.WithHelp("r/F5", "refresh"),
		),
	}
}

// HelpKeyMap provides help-specific bindings
type HelpKeyMap struct {
	Close      key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
	NextPage   key.Binding
	PrevPage   key.Binding
}

// NewHelpKeyMap creates help-specific key bindings
func NewHelpKeyMap() HelpKeyMap {
	return HelpKeyMap{
		Close: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "close help"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "scroll up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "scroll down"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("right", "l", "pgdown"),
			key.WithHelp("→/l/pgdn", "next page"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("left", "h", "pgup"),
			key.WithHelp("←/h/pgup", "previous page"),
		),
	}
}
