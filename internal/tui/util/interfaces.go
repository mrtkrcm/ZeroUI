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

// ClearStatusMsg clears status messages
type ClearStatusMsg struct{}
