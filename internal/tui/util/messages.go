package util

// Message types for TUI communication

// ErrorMsg represents an error message
type ErrorMsg struct {
	Err error
}

// Error returns the error
func (e ErrorMsg) Error() string {
	return e.Err.Error()
}

// SuccessMsg represents a success message
type SuccessMsg struct {
	Title string
	Body  string
}

// ShowInfoMsg represents an informational message
type ShowInfoMsg string
