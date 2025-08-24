package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormFooterHintsIncludePresetsAndChangedOnly(t *testing.T) {
	form := NewHuhConfigForm("ghostty")
	// Provide one simple field so the form builds
	form.SetFields([]ConfigField{{Key: "theme", Type: FieldTypeString, Value: "dark", IsSet: true}})

	view := form.View()
	assert.Contains(t, view, "C changed-only", "Footer should mention changed-only toggle")
	assert.Contains(t, view, "p presets", "Footer should mention presets shortcut")
}
