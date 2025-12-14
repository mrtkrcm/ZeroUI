// zeroui/pkg/configextractor/config_diff.go
package configextractor

import (
	"fmt"
	"strings"
)

// ConfigDiff represents the differences between two configurations
type ConfigDiff struct {
	Added     map[string]interface{} `json:"added"`
	Modified  map[string]ValueDiff   `json:"modified"`
	Removed   map[string]interface{} `json:"removed"`
	Unchanged map[string]interface{} `json:"unchanged"`
}

// ValueDiff represents a change in a single value
type ValueDiff struct {
	Old interface{} `json:"old"`
	New interface{} `json:"new"`
}

// ConfigDiffer provides configuration diffing functionality
type ConfigDiffer struct{}

// NewConfigDiffer creates a new configuration differ
func NewConfigDiffer() *ConfigDiffer {
	return &ConfigDiffer{}
}

// DiffConfigurations compares two configuration maps and returns the differences
func (d *ConfigDiffer) DiffConfigurations(old, new map[string]interface{}) ConfigDiff {
	diff := ConfigDiff{
		Added:     make(map[string]interface{}),
		Modified:  make(map[string]ValueDiff),
		Removed:   make(map[string]interface{}),
		Unchanged: make(map[string]interface{}),
	}

	// Check for added and modified keys
	for key, newValue := range new {
		if oldValue, exists := old[key]; exists {
			if fmt.Sprintf("%v", oldValue) != fmt.Sprintf("%v", newValue) {
				diff.Modified[key] = ValueDiff{Old: oldValue, New: newValue}
			} else {
				diff.Unchanged[key] = newValue
			}
		} else {
			diff.Added[key] = newValue
		}
	}

	// Check for removed keys
	for key, oldValue := range old {
		if _, exists := new[key]; !exists {
			diff.Removed[key] = oldValue
		}
	}

	return diff
}

// HasChanges returns true if the diff contains any changes
func (d ConfigDiff) HasChanges() bool {
	return len(d.Added) > 0 || len(d.Modified) > 0 || len(d.Removed) > 0
}

// Summary returns a human-readable summary of the changes
func (d ConfigDiff) Summary() string {
	if !d.HasChanges() {
		return "No changes"
	}

	var parts []string
	if len(d.Added) > 0 {
		parts = append(parts, fmt.Sprintf("+%d added", len(d.Added)))
	}
	if len(d.Modified) > 0 {
		parts = append(parts, fmt.Sprintf("~%d modified", len(d.Modified)))
	}
	if len(d.Removed) > 0 {
		parts = append(parts, fmt.Sprintf("-%d removed", len(d.Removed)))
	}
	if len(d.Unchanged) > 0 {
		parts = append(parts, fmt.Sprintf("=%d unchanged", len(d.Unchanged)))
	}

	return strings.Join(parts, ", ")
}

// FormatDiff returns a formatted string representation of the diff
func (d ConfigDiff) FormatDiff() string {
	var output strings.Builder

	if len(d.Added) > 0 {
		output.WriteString("Added:\n")
		for key, value := range d.Added {
			output.WriteString(fmt.Sprintf("  + %s = %v\n", key, value))
		}
	}

	if len(d.Modified) > 0 {
		output.WriteString("Modified:\n")
		for key, diff := range d.Modified {
			output.WriteString(fmt.Sprintf("  ~ %s: %v → %v\n", key, diff.Old, diff.New))
		}
	}

	if len(d.Removed) > 0 {
		output.WriteString("Removed:\n")
		for key, value := range d.Removed {
			output.WriteString(fmt.Sprintf("  - %s = %v\n", key, value))
		}
	}

	if len(d.Unchanged) > 0 {
		output.WriteString("Unchanged:\n")
		for key, value := range d.Unchanged {
			output.WriteString(fmt.Sprintf("  = %s = %v\n", key, value))
		}
	}

	return strings.TrimSpace(output.String())
}
