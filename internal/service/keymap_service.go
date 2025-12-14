package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mrtkrcm/ZeroUI/pkg/configextractor"
)

// KeymapService handles keymap operations across different applications
type KeymapService struct {
	configService *ConfigService
}

// NewKeymapService creates a new keymap service
func NewKeymapService(configService *ConfigService) *KeymapService {
	return &KeymapService{
		configService: configService,
	}
}

// KeymapInfo represents a parsed keymap
type KeymapInfo struct {
	Keys   string
	Action string
	App    string
}

// KeymapPreset represents a collection of keymaps for a specific theme
type KeymapPreset struct {
	Name        string
	Description string
	Keymaps     []string
}

// GetKeymapsForApp retrieves all keymaps for a specific application
func (s *KeymapService) GetKeymapsForApp(app string) ([]KeymapInfo, error) {
	values, err := s.configService.GetCurrentValues(app)
	if err != nil {
		return nil, err
	}

	var keymaps []KeymapInfo
	for key, value := range values {
		if strings.HasPrefix(key, "keybind") {
			if strVal, ok := value.(string); ok && strVal != "" {
				if info := s.parseKeymapString(strVal, app); info != nil {
					keymaps = append(keymaps, *info)
				}
			} else if sliceVal, ok := value.([]interface{}); ok {
				for _, item := range sliceVal {
					if strItem, ok := item.(string); ok && strItem != "" {
						if info := s.parseKeymapString(strItem, app); info != nil {
							keymaps = append(keymaps, *info)
						}
					}
				}
			}
		}
	}

	// Sort by keys for consistent output
	sort.Slice(keymaps, func(i, j int) bool {
		return keymaps[i].Keys < keymaps[j].Keys
	})

	return keymaps, nil
}

// parseKeymapString parses a keymap string into KeymapInfo
func (s *KeymapService) parseKeymapString(keymapStr, app string) *KeymapInfo {
	if !strings.Contains(keymapStr, "=") {
		return nil
	}

	parts := strings.SplitN(keymapStr, "=", 2)
	if len(parts) != 2 {
		return nil
	}

	return &KeymapInfo{
		Keys:   strings.TrimSpace(parts[0]),
		Action: strings.TrimSpace(parts[1]),
		App:    app,
	}
}

// ValidateKeymap validates a single keymap
func (s *KeymapService) ValidateKeymap(keymap string) (*configextractor.KeybindValidationResult, error) {
	validator := configextractor.NewKeybindValidator()
	result := validator.ValidateKeybind(keymap)
	return &result, nil
}

// GetKeymapPresets returns available keymap presets for an application
func (s *KeymapService) GetKeymapPresets(app string) map[string]*KeymapPreset {
	presets := make(map[string]*KeymapPreset)

	// Define presets based on application type
	switch app {
	case "ghostty":
		presets["vim-like"] = &KeymapPreset{
			Name:        "vim-like",
			Description: "Vim-style navigation and control",
			Keymaps: []string{
				"ctrl+h=previous_tab",
				"ctrl+l=next_tab",
				"ctrl+j=scroll_page_down",
				"ctrl+k=scroll_page_up",
				"ctrl+w=close_surface",
				"shift+insert=paste_from_clipboard",
			},
		}

		presets["tmux-like"] = &KeymapPreset{
			Name:        "tmux-like",
			Description: "Tmux-style prefix-based shortcuts",
			Keymaps: []string{
				"ctrl+b+c=new_tab",
				"ctrl+b+n=next_tab",
				"ctrl+b+p=previous_tab",
				"ctrl+b+x=close_surface",
				"ctrl+b+%=split_horizontal",
				"ctrl+b+\"=split_vertical",
				"ctrl+b+left=goto_split:left",
				"ctrl+b+right=goto_split:right",
			},
		}

		presets["modern"] = &KeymapPreset{
			Name:        "modern",
			Description: "Modern shortcuts with super key",
			Keymaps: []string{
				"super+t=new_tab",
				"super+w=close_surface",
				"super+c=copy_to_clipboard",
				"super+v=paste_from_clipboard",
				"super+plus=increase_font_size:1",
				"super+minus=decrease_font_size:1",
				"super+0=reset_font_size",
				"ctrl+shift+t=new_tab",
			},
		}

	case "vscode", "zed":
		presets["vim-like"] = &KeymapPreset{
			Name:        "vim-like",
			Description: "Vim-style navigation for code editors",
			Keymaps: []string{
				"ctrl+h=workbench.action.previousEditor",
				"ctrl+l=workbench.action.nextEditor",
				"ctrl+j=workbench.action.quickOpen",
				"ctrl+k=workbench.action.showCommands",
			},
		}

	default:
		presets["basic"] = &KeymapPreset{
			Name:        "basic",
			Description: "Basic shortcuts for any application",
			Keymaps: []string{
				"ctrl+c=copy",
				"ctrl+v=paste",
				"ctrl+x=cut",
				"ctrl+z=undo",
				"ctrl+y=redo",
				"ctrl+a=select_all",
			},
		}
	}

	return presets
}

// DetectConflicts finds conflicting keymaps
func (s *KeymapService) DetectConflicts(app string) ([]string, error) {
	keymaps, err := s.GetKeymapsForApp(app)
	if err != nil {
		return nil, err
	}

	keyToActions := make(map[string][]string)
	var conflicts []string

	for _, km := range keymaps {
		if actions, exists := keyToActions[km.Keys]; exists {
			// Check if action is already mapped to these keys
			found := false
			for _, action := range actions {
				if action == km.Action {
					found = true
					break
				}
			}
			if !found {
				conflicts = append(conflicts, fmt.Sprintf("Keys '%s' mapped to multiple actions: %s",
					km.Keys, strings.Join(append(actions, km.Action), ", ")))
			}
		} else {
			keyToActions[km.Keys] = []string{km.Action}
		}
	}

	return conflicts, nil
}

// FormatKeymapForDisplay formats a keymap for display
func (s *KeymapService) FormatKeymapForDisplay(info KeymapInfo) string {
	return fmt.Sprintf("%s → %s", info.Keys, info.Action)
}

// SuggestAlternatives suggests alternative keymaps for common conflicts
func (s *KeymapService) SuggestAlternatives(conflictingKeys string) []string {
	suggestions := map[string][]string{
		"ctrl+c": {"ctrl+shift+c", "super+c", "alt+c"},
		"ctrl+v": {"ctrl+shift+v", "super+v", "alt+v"},
		"ctrl+w": {"ctrl+shift+w", "super+w", "alt+w"},
		"ctrl+t": {"ctrl+shift+t", "super+t", "alt+t"},
	}

	if alternatives, exists := suggestions[conflictingKeys]; exists {
		return alternatives
	}

	// Generate generic alternatives
	return []string{
		conflictingKeys + "+shift",
		"super+" + strings.TrimPrefix(conflictingKeys, "ctrl+"),
		"alt+" + strings.TrimPrefix(conflictingKeys, "ctrl+"),
	}
}

// AddKeymap adds a new keymap to an application
func (s *KeymapService) AddKeymap(app, keymap string) error {
	// Validate before adding
	validator := configextractor.NewKeybindValidator()
	result := validator.ValidateKeybind(keymap)
	if !result.Valid {
		return fmt.Errorf("invalid keymap format: %v", result.Errors)
	}

	return s.configService.AppendConfiguration(app, "keybind", keymap)
}

// RemoveKeymap removes a keymap by its keys (e.g., "ctrl+c")
func (s *KeymapService) RemoveKeymap(app, keys string) error {
	// We need to find the full keymap string to remove it consistently
	// Use fuzzy matching logic similar to engine but here we can be smarter
	return s.configService.RemoveConfiguration(app, "keybind", keys)
}

// ValidateAllKeymaps validates all keymaps for an application
func (s *KeymapService) ValidateAllKeymaps(app string) (int, int, []string, error) {
	keymaps, err := s.GetKeymapsForApp(app)
	if err != nil {
		return 0, 0, nil, err
	}

	var validCount, invalidCount int
	var errorMessages []string
	validator := configextractor.NewKeybindValidator()

	for _, km := range keymaps {
		// Reconstruct keymap string or validate parts?
		// Better to validate the reconstruction "keys=action"
		keymapStr := fmt.Sprintf("%s=%s", km.Keys, km.Action)
		result := validator.ValidateKeybind(keymapStr)
		if result.Valid {
			validCount++
		} else {
			invalidCount++
			errorMessages = append(errorMessages, fmt.Sprintf("Invalid keymap '%s': %v", keymapStr, result.Errors))
		}
	}

	return validCount, invalidCount, errorMessages, nil
}
