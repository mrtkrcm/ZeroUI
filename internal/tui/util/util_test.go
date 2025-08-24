package util

import (
	"testing"
)

func TestInfoTypeString(t *testing.T) {
	tests := []struct {
		infoType InfoType
		expected string
	}{
		{InfoTypeInfo, "info"},
		{InfoTypeSuccess, "success"},
		{InfoTypeWarn, "warn"},
		{InfoTypeError, "error"},
		{InfoType(999), "unknown"}, // Test unknown type
	}

	for _, test := range tests {
		result := test.infoType.String()
		if result != test.expected {
			t.Errorf("InfoType.String() = %v, want %v", result, test.expected)
		}
	}
}

func TestInfoMsg(t *testing.T) {
	// Test InfoMsg creation
	msg := InfoMsg{
		Type: InfoTypeSuccess,
		Msg:  "Operation completed successfully",
	}

	if msg.Type != InfoTypeSuccess {
		t.Errorf("Expected InfoTypeSuccess, got %v", msg.Type)
	}

	if msg.Msg != "Operation completed successfully" {
		t.Errorf("Expected 'Operation completed successfully', got '%s'", msg.Msg)
	}
}

func TestClearStatusMsg(t *testing.T) {
	// Test ClearStatusMsg creation
	msg := ClearStatusMsg{}

	// ClearStatusMsg is a marker struct, just test that it can be created
	_ = msg // Use the variable to avoid unused variable error
}

// TestErrorMsg tests the ErrorMsg type
func TestErrorMsg(t *testing.T) {
	err := ErrorMsg{
		Err: &testError{msg: "test error"},
	}

	if err.Err.Error() != "test error" {
		t.Errorf("Expected 'test error', got '%s'", err.Err.Error())
	}
}

// TestSuccessMsg tests the SuccessMsg type
func TestSuccessMsg(t *testing.T) {
	msg := SuccessMsg{
		Title: "Success",
		Body:  "Operation completed",
	}

	if msg.Title != "Success" {
		t.Errorf("Expected 'Success', got '%s'", msg.Title)
	}

	if msg.Body != "Operation completed" {
		t.Errorf("Expected 'Operation completed', got '%s'", msg.Body)
	}
}

// TestShowInfoMsg tests the ShowInfoMsg type
func TestShowInfoMsg(t *testing.T) {
	msg := ShowInfoMsg("This is an info message")

	if string(msg) != "This is an info message" {
		t.Errorf("Expected 'This is an info message', got '%s'", string(msg))
	}
}

// Helper type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
