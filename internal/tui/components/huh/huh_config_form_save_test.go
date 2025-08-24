package components

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test that getFormValues only returns fields that changed vs original/default
func TestHuhForm_ChangeOnlyValues(t *testing.T) {
	form := NewHuhConfigForm("ghostty")

	fields := []ConfigField{
		// Already set in file; unchanged should not be returned
		{Key: "font-family", Type: FieldTypeString, Value: "SF Mono", IsSet: true, Description: "Font"},
		// Not set; default provided; unchanged should not be returned
		{Key: "theme", Type: FieldTypeSelect, Value: "Dracula", Options: []string{"Dracula", "Gruvbox"}, IsSet: false},
		// Boolean set to true originally; flip to false should return
		{Key: "bold-text", Type: FieldTypeBool, Value: true, IsSet: true},
		// Int not set, no default; provide non-empty value should return
		{Key: "font-size", Type: FieldTypeInt, IsSet: false},
	}

	form.SetFields(fields)

	// Simulate user edits via bindings
	// 1) font-family unchanged -> should not appear
	if ptr := form.stringBindings["font-family"]; ptr != nil {
		*ptr = "SF Mono"
	}
	// 2) theme unchanged vs default -> should not appear
	if ptr := form.stringBindings["theme"]; ptr != nil {
		*ptr = "Dracula"
	}
	// 3) bold-text changed true -> false -> should appear
	if bptr := form.boolBindings["bold-text"]; bptr != nil {
		*bptr = false
	}
	// 4) font-size new non-empty value -> should appear
	if ptr := form.stringBindings["font-size"]; ptr != nil {
		*ptr = "14"
	}

	vals := form.getFormValues()

	// Only bold-text and font-size should be included
	assert.NotContains(t, vals, "font-family")
	assert.NotContains(t, vals, "theme")

	if v, ok := vals["bold-text"]; assert.True(t, ok) {
		assert.Equal(t, "false", v)
	}
	if v, ok := vals["font-size"]; assert.True(t, ok) {
		assert.Equal(t, "14", v)
	}

	// Ensure map contains exactly 2 entries
	assert.Len(t, vals, 2)

	_ = os.Setenv("_TEST_OK", "1")
}
