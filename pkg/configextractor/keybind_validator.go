package configextractor

import (
	"strings"
)

// KeybindValidationResult represents the result of validating a keybind
type KeybindValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
}

// KeybindValidator validates keybind configurations
type KeybindValidator struct {
	validModifiers   map[string]bool
	validSpecialKeys map[string]bool
	validActions     map[string]bool
}

// NewKeybindValidator creates a new keybind validator
func NewKeybindValidator() *KeybindValidator {
	return &KeybindValidator{
		validModifiers: map[string]bool{
			"ctrl": true, "shift": true, "alt": true, "super": true,
			"meta": true, "cmd": true, "command": true, "opt": true, "option": true,
		},
		validSpecialKeys: map[string]bool{
			"escape": true, "enter": true, "return": true, "tab": true,
			"space": true, "backspace": true, "delete": true, "insert": true,
			"home": true, "end": true, "pageup": true, "pagedown": true,
			"up": true, "down": true, "left": true, "right": true,
			"f1": true, "f2": true, "f3": true, "f4": true, "f5": true,
			"f6": true, "f7": true, "f8": true, "f9": true, "f10": true,
			"f11": true, "f12": true, "f13": true, "f14": true, "f15": true,
			"f16": true, "f17": true, "f18": true, "f19": true, "f20": true,
		},
		validActions: map[string]bool{
			"copy": true, "paste": true, "cut": true, "select_all": true,
			"undo": true, "redo": true, "quit": true, "close": true,
			"new_window": true, "new_tab": true, "close_tab": true,
			"next_tab": true, "prev_tab": true, "goto_tab": true,
			"split_horizontal": true, "split_vertical": true, "close_split": true,
			"resize_split_left": true, "resize_split_right": true,
			"resize_split_up": true, "resize_split_down": true,
			"focus_left": true, "focus_right": true, "focus_up": true, "focus_down": true,
			"reload_config": true, "toggle_fullscreen": true,
			"scroll_up": true, "scroll_down": true, "scroll_page_up": true,
			"scroll_page_down": true, "scroll_home": true, "scroll_end": true,
			"clear": true, "reset": true,
		},
	}
}

// ValidateKeybind validates a complete keybind string
func (v *KeybindValidator) ValidateKeybind(keybind string) KeybindValidationResult {
	result := KeybindValidationResult{Valid: true}

	if !strings.Contains(keybind, "=") {
		result.Valid = false
		result.Errors = append(result.Errors, "keybind must contain '=' separator")
		return result
	}

	parts := strings.SplitN(keybind, "=", 2)
	keys := strings.TrimSpace(parts[0])
	action := strings.TrimSpace(parts[1])

	if keys == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "keybind keys cannot be empty")
		return result
	}

	if action == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "keybind action cannot be empty")
		return result
	}

	keyResult := v.validateKeyCombination(keys)
	if !keyResult.Valid {
		result.Valid = false
		result.Errors = append(result.Errors, keyResult.Errors...)
	}
	result.Warnings = append(result.Warnings, keyResult.Warnings...)

	actionResult := v.validateAction(action)
	if !actionResult.Valid {
		result.Valid = false
		result.Errors = append(result.Errors, actionResult.Errors...)
	}
	result.Warnings = append(result.Warnings, actionResult.Warnings...)

	return result
}

func (v *KeybindValidator) validateKeyCombination(keys string) KeybindValidationResult {
	result := KeybindValidationResult{Valid: true}

	if keys == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "keybind must contain at least one valid key")
		return result
	}

	components := strings.Split(strings.ToLower(keys), "+")

	for _, comp := range components {
		comp = strings.TrimSpace(comp)
		if comp == "" {
			continue
		}

		if v.validModifiers[comp] || v.validSpecialKeys[comp] {
			continue
		}

		if len(comp) == 1 && ((comp[0] >= 'a' && comp[0] <= 'z') || (comp[0] >= '0' && comp[0] <= '9')) {
			continue
		}

		result.Valid = false
		result.Errors = append(result.Errors, "invalid key component: '"+comp+"'")
	}

	return result
}

func (v *KeybindValidator) validateAction(action string) KeybindValidationResult {
	result := KeybindValidationResult{Valid: true}

	baseAction := action
	if idx := strings.Index(action, ":"); idx != -1 {
		baseAction = action[:idx]
	}

	if !v.validActions[baseAction] {
		result.Warnings = append(result.Warnings, "unknown action: '"+baseAction+"'")
	}

	return result
}
