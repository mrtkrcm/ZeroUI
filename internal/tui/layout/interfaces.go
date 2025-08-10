package layout

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Sizeable represents components that can be resized
type Sizeable interface {
	SetSize(width, height int) tea.Cmd
	GetSize() (int, int)
}

// Focusable represents components that can receive focus
type Focusable interface {
	Focus() tea.Cmd
	Blur() tea.Cmd
	IsFocused() bool
}

// Help represents components that provide key bindings
type Help interface {
	Bindings() []key.Binding
}

// Positional represents components that can be positioned
type Positional interface {
	SetPosition(x, y int) tea.Cmd
}