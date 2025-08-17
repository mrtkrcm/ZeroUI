package util

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents a generic TUI model
type Model interface {
	tea.Model
}

// InfoMsg represents informational messages
type InfoMsg struct {
	Type InfoType
	Msg  string
}

type InfoType int

const (
	InfoTypeInfo InfoType = iota
	InfoTypeSuccess
	InfoTypeWarn
	InfoTypeError
)

// String returns the string representation of InfoType
func (it InfoType) String() string {
	switch it {
	case InfoTypeInfo:
		return "info"
	case InfoTypeSuccess:
		return "success"
	case InfoTypeWarn:
		return "warn"
	case InfoTypeError:
		return "error"
	default:
		return "unknown"
	}
}

// ClearStatusMsg clears status messages
type ClearStatusMsg struct{}
