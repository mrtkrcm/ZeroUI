package core

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Application-level messages
type (
	// WindowResizeMsg is sent when the terminal window is resized
	WindowResizeMsg struct {
		Width  int
		Height int
	}

	// PageChangeMsg is sent when switching between pages
	PageChangeMsg struct {
		From string
		To   string
	}

	// ThemeChangeMsg is sent when the theme changes
	ThemeChangeMsg struct {
		ThemeName string
	}

	// StatusUpdateMsg is sent to update the status bar
	StatusUpdateMsg struct {
		Message string
	}

	// ErrorMsg is sent when an error occurs
	ErrorMsg struct {
		Error error
		Title string
	}

	// SuccessMsg is sent when an operation succeeds
	SuccessMsg struct {
		Message string
	}

	// LoadingMsg is sent to show/hide loading state
	LoadingMsg struct {
		Loading bool
		Message string
	}
)

// Component-level messages
type (
	// FocusMsg is sent to focus a component
	FocusMsg struct {
		ComponentID string
	}

	// BlurMsg is sent to blur a component
	BlurMsg struct {
		ComponentID string
	}

	// UpdateMsg is sent to update component data
	UpdateMsg struct {
		ComponentID string
		Data        ConfigData
	}

	// RefreshMsg is sent to refresh component data
	RefreshMsg struct {
		ComponentID string
	}
)

// List-specific messages
type (
	// ListItemSelectedMsg is sent when a list item is selected
	ListItemSelectedMsg struct {
		ListID string
		Index  int
		Item   ListItem
	}

	// ListFilterMsg is sent to filter list items
	ListFilterMsg struct {
		ListID string
		Filter string
	}

	// ListSortMsg is sent to sort list items
	ListSortMsg struct {
		ListID    string
		SortBy    string
		Ascending bool
	}
)

// Form-specific messages
type (
	// FormFieldChangedMsg is sent when a form field value changes
	FormFieldChangedMsg struct {
		FormID  string
		FieldID string
		Value   interface{}
	}

	// FormSubmitMsg is sent when a form is submitted
	FormSubmitMsg struct {
		FormID string
		Data   map[string]interface{}
	}

	// FormValidationMsg is sent to trigger form validation
	FormValidationMsg struct {
		FormID string
	}
)

// Dialog-specific messages
type (
	// DialogOpenMsg is sent to open a dialog
	DialogOpenMsg struct {
		DialogID string
		Data     interface{}
	}

	// DialogCloseMsg is sent to close a dialog
	DialogCloseMsg struct {
		DialogID string
		Data     interface{}
	}

	// DialogResultMsg is sent when a dialog returns a result
	DialogResultMsg struct {
		DialogID string
		Result   interface{}
		Action   string // "ok", "cancel", "close", etc.
	}
)

// Configuration-specific messages
type (
	// ConfigLoadedMsg is sent when configuration is loaded
	ConfigLoadedMsg struct {
		AppName string
		Config  interface{}
	}

	// ConfigChangedMsg is sent when configuration changes
	ConfigChangedMsg struct {
		AppName string
		Key     string
		Value   interface{}
	}

	// ConfigSavedMsg is sent when configuration is saved
	ConfigSavedMsg struct {
		AppName string
		Success bool
		Error   error
	}

	// PresetAppliedMsg is sent when a preset is applied
	PresetAppliedMsg struct {
		AppName    string
		PresetName string
		Success    bool
		Error      error
	}
)

// Utility functions for creating commands from messages
func WindowResize(width, height int) tea.Cmd {
	return func() tea.Msg {
		return WindowResizeMsg{Width: width, Height: height}
	}
}

func PageChange(from, to string) tea.Cmd {
	return func() tea.Msg {
		return PageChangeMsg{From: from, To: to}
	}
}

func ShowError(err error, title string) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{Error: err, Title: title}
	}
}

func ShowSuccess(message string) tea.Cmd {
	return func() tea.Msg {
		return SuccessMsg{Message: message}
	}
}

func ShowLoading(loading bool, message string) tea.Cmd {
	return func() tea.Msg {
		return LoadingMsg{Loading: loading, Message: message}
	}
}

func UpdateStatus(message string) tea.Cmd {
	return func() tea.Msg {
		return StatusUpdateMsg{Message: message}
	}
}

func SelectListItem(listID string, index int, item ListItem) tea.Cmd {
	return func() tea.Msg {
		return ListItemSelectedMsg{ListID: listID, Index: index, Item: item}
	}
}

func SubmitForm(formID string, data map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		return FormSubmitMsg{FormID: formID, Data: data}
	}
}

func OpenDialog(dialogID string, data interface{}) tea.Cmd {
	return func() tea.Msg {
		return DialogOpenMsg{DialogID: dialogID, Data: data}
	}
}

func CloseDialog(dialogID string, data interface{}) tea.Cmd {
	return func() tea.Msg {
		return DialogCloseMsg{DialogID: dialogID, Data: data}
	}
}
