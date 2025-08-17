package components

// Message types for component communication

// PresetAppliedMsg is sent when a preset is applied
type PresetAppliedMsg struct {
	PresetName string
	Values     map[string]interface{}
}
