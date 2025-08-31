package help

import (
	"strings"
	"time"
)

// ContextualHelp provides intelligent, context-aware help
type ContextualHelp struct {
	currentContext string
	lastActivity   time.Time
	helpHistory    []HelpItem
	suggestions    []Suggestion
}

// HelpItem represents a help entry
type HelpItem struct {
	Topic     string
	Content   string
	Category  string
	Priority  int
	LastShown time.Time
	ShowCount int
}

// Suggestion represents a contextual suggestion
type Suggestion struct {
	Text       string
	Action     string
	Confidence float64
	Category   string
	Trigger    string
}

// NewContextualHelp creates a new contextual help system
func NewContextualHelp() *ContextualHelp {
	return &ContextualHelp{
		helpHistory:  []HelpItem{},
		suggestions:  []Suggestion{},
		lastActivity: time.Now(),
	}
}

// UpdateContext updates the current context and generates suggestions
func (ch *ContextualHelp) UpdateContext(context string, userAction string) {
	ch.currentContext = context
	ch.lastActivity = time.Now()

	// Generate contextual suggestions
	ch.generateSuggestions(context, userAction)

	// Update help history
	ch.updateHelpHistory(context)
}

// GetSuggestions returns current contextual suggestions
func (ch *ContextualHelp) GetSuggestions() []Suggestion {
	// Filter suggestions by confidence and recency
	var active []Suggestion
	for _, suggestion := range ch.suggestions {
		if suggestion.Confidence > 0.3 {
			active = append(active, suggestion)
		}
	}
	return active
}

// GetHelp returns context-appropriate help
func (ch *ContextualHelp) GetHelp() string {
	switch ch.currentContext {
	case "editing":
		return ch.getEditingHelp()
	case "searching":
		return ch.getSearchHelp()
	case "navigation":
		return ch.getNavigationHelp()
	case "first-time":
		return ch.getFirstTimeHelp()
	default:
		return ch.getGeneralHelp()
	}
}

// getEditingHelp returns help for editing context
func (ch *ContextualHelp) getEditingHelp() string {
	return "✏️ Editing Mode:\n• Enter: Save changes\n• Esc: Cancel editing\n• Tab: Auto-complete"
}

// getSearchHelp returns help for search context
func (ch *ContextualHelp) getSearchHelp() string {
	return "🔍 Search Mode:\n• Enter: Apply search\n• Esc: Clear search\n• Type to filter results"
}

// getNavigationHelp returns help for navigation context
func (ch *ContextualHelp) getNavigationHelp() string {
	return "🎯 Navigation:\n• ↑↓: Move up/down\n• Enter: Edit item\n• ?: Help\n• q: Quit"
}

// getFirstTimeHelp returns help for first-time users
func (ch *ContextualHelp) getFirstTimeHelp() string {
	return "👋 Welcome!\n• Use ↑↓ to navigate\n• Press Enter to edit\n• Press ? for more help"
}

// getGeneralHelp returns general help
func (ch *ContextualHelp) getGeneralHelp() string {
	return "⚙️ Configuration Editor:\n• ↑↓: Navigate\n• Enter: Edit\n• ?: Help\n• q: Quit"
}

// generateSuggestions creates intelligent suggestions based on context
func (ch *ContextualHelp) generateSuggestions(context, action string) {
	ch.suggestions = []Suggestion{} // Reset suggestions

	switch context {
	case "editing":
		ch.addEditingSuggestions(action)
	case "searching":
		ch.addSearchSuggestions(action)
	case "navigation":
		ch.addNavigationSuggestions(action)
	case "configuration":
		ch.addConfigurationSuggestions(action)
	}

	// Add general suggestions
	ch.addGeneralSuggestions()
}

// addEditingSuggestions provides editing-specific suggestions
func (ch *ContextualHelp) addEditingSuggestions(action string) {
	suggestions := []Suggestion{
		{
			Text:       "💡 Press Tab for auto-complete suggestions",
			Action:     "show-autocomplete",
			Confidence: 0.8,
			Category:   "editing",
			Trigger:    "editing-started",
		},
		{
			Text:       "💡 Press Enter to save, Esc to cancel",
			Action:     "show-save-cancel",
			Confidence: 0.9,
			Category:   "editing",
			Trigger:    "editing-started",
		},
		{
			Text:       "💡 Use Ctrl+Space for context help",
			Action:     "show-context-help",
			Confidence: 0.6,
			Category:   "editing",
			Trigger:    "editing-active",
		},
	}

	ch.suggestions = append(ch.suggestions, suggestions...)
}

// addSearchSuggestions provides search-specific suggestions
func (ch *ContextualHelp) addSearchSuggestions(action string) {
	suggestions := []Suggestion{
		{
			Text:       "🔍 Try searching for 'font', 'color', or 'theme'",
			Action:     "show-search-examples",
			Confidence: 0.7,
			Category:   "searching",
			Trigger:    "search-started",
		},
		{
			Text:       "🔍 Use quotes for exact matches: \"font-family\"",
			Action:     "show-search-tips",
			Confidence: 0.6,
			Category:   "searching",
			Trigger:    "search-active",
		},
	}

	ch.suggestions = append(ch.suggestions, suggestions...)
}

// addNavigationSuggestions provides navigation-specific suggestions
func (ch *ContextualHelp) addNavigationSuggestions(action string) {
	suggestions := []Suggestion{
		{
			Text:       "🎯 Click items to select, double-click to edit",
			Action:     "show-mouse-navigation",
			Confidence: 0.7,
			Category:   "navigation",
			Trigger:    "navigation-active",
		},
		{
			Text:       "🎯 Use number keys (1-9) for quick navigation",
			Action:     "show-quick-navigation",
			Confidence: 0.5,
			Category:   "navigation",
			Trigger:    "navigation-active",
		},
	}

	ch.suggestions = append(ch.suggestions, suggestions...)
}

// addConfigurationSuggestions provides configuration-specific suggestions
func (ch *ContextualHelp) addConfigurationSuggestions(action string) {
	suggestions := []Suggestion{
		{
			Text:       "⚙️ Modified settings are marked with ✨",
			Action:     "show-changed-indicator",
			Confidence: 0.8,
			Category:   "configuration",
			Trigger:    "configuration-loaded",
		},
		{
			Text:       "⚙️ Press Ctrl+S to save your changes",
			Action:     "show-save-shortcut",
			Confidence: 0.9,
			Category:   "configuration",
			Trigger:    "configuration-changed",
		},
	}

	ch.suggestions = append(ch.suggestions, suggestions...)
}

// addGeneralSuggestions provides general helpful suggestions
func (ch *ContextualHelp) addGeneralSuggestions() {
	suggestions := []Suggestion{
		{
			Text:       "❓ Press ? anytime for detailed help",
			Action:     "show-help-reminder",
			Confidence: 0.4,
			Category:   "general",
			Trigger:    "app-started",
		},
		{
			Text:       "🚀 Pro tip: Use mouse wheel for smooth scrolling",
			Action:     "show-mouse-tip",
			Confidence: 0.3,
			Category:   "general",
			Trigger:    "app-active",
		},
	}

	ch.suggestions = append(ch.suggestions, suggestions...)
}

// updateHelpHistory tracks help usage patterns
func (ch *ContextualHelp) updateHelpHistory(context string) {
	// Track help topics for personalization
	helpItem := HelpItem{
		Topic:     context,
		LastShown: time.Now(),
		ShowCount: 1,
	}

	// Update existing or add new
	found := false
	for i := range ch.helpHistory {
		if ch.helpHistory[i].Topic == context {
			ch.helpHistory[i].LastShown = time.Now()
			ch.helpHistory[i].ShowCount++
			found = true
			break
		}
	}

	if !found {
		ch.helpHistory = append(ch.helpHistory, helpItem)
	}
}

// GetQuickHelp returns a concise help message
func (ch *ContextualHelp) GetQuickHelp() string {
	switch ch.currentContext {
	case "editing":
		return "Enter: Save • Esc: Cancel • Tab: Auto-complete"
	case "searching":
		return "Enter: Search • Esc: Cancel"
	case "navigation":
		return "↑↓: Navigate • Enter: Edit • ?: Help"
	default:
		return "↑↓/jk: Navigate • Enter: Edit • /: Search • ?: Help • q: Quit"
	}
}

// GetDetailedHelp returns comprehensive help
func (ch *ContextualHelp) GetDetailedHelp() string {
	help := []string{
		"╭─ ZeroUI Help ──────────────────────────────────────────────────╮",
		"│                                                                │",
		"│ 🎯 Navigation                                                  │",
		"│   ↑/↓, j/k        Move up/down                                │",
		"│   Mouse wheel      Smooth scrolling                           │",
		"│   Click            Select item                                │",
		"│   Double-click     Quick edit                                  │",
		"│   Number keys      Quick navigation (1-9)                     │",
		"│                                                                │",
		"│ ✏️  Editing                                                     │",
		"│   Enter/Space      Start editing                              │",
		"│   Enter            Save changes                               │",
		"│   Esc              Cancel editing                             │",
		"│   Tab              Auto-complete                              │",
		"│   Ctrl+Space       Context help                               │",
		"│                                                                │",
		"│ 🔍 Search                                                      │",
		"│   /                Start search                               │",
		"│   Enter            Apply search                               │",
		"│   Esc              Clear search                               │",
		"│                                                                │",
		"│ 💾 Actions                                                     │",
		"│   Ctrl+S          Save configuration                          │",
		"│   u               Undo last change                            │",
		"│   Ctrl+Z          Undo                                        │",
		"│   Ctrl+Y          Redo                                        │",
		"│                                                                │",
		"│ ❓ Help & Info                                                  │",
		"│   ?                Toggle this help                           │",
		"│   Ctrl+H          Context help                                │",
		"│   F1              Keyboard shortcuts                          │",
		"│                                                                │",
		"│ 🚪 Exit                                                        │",
		"│   q, Ctrl+C       Quit application                            │",
		"│   Ctrl+Q          Force quit                                  │",
		"│                                                                │",
		"╰────────────────────────────────────────────────────────────────╯",
	}

	return strings.Join(help, "\n")
}

// GetFieldHelp returns field-specific help
func (ch *ContextualHelp) GetFieldHelp(fieldName, fieldType string) string {
	// Check field name patterns (more specific first)
	fieldNameLower := strings.ToLower(fieldName)

	switch {
	case strings.Contains(fieldNameLower, "font.size"):
		return "📏 Font size values are typically in pixels or points (12pt, 14px, etc.)"
	case strings.Contains(fieldNameLower, "font"):
		return "🎨 Font settings affect text appearance. Try 'JetBrains Mono' or 'Fira Code' for coding."
	case strings.Contains(fieldNameLower, "color"):
		return "🎨 Colors use hex format (#RRGGBB) or named colors (red, blue, etc.)"
	case strings.Contains(fieldNameLower, "size"):
		return "📏 Size values are typically in pixels or points (12pt, 14px, etc.)"
	case strings.Contains(fieldNameLower, "path"):
		return "📁 Paths can be absolute (/usr/bin) or relative (./config)"
	case fieldType == "boolean":
		return "✓ Boolean values: true/false, yes/no, 1/0, on/off"
	case fieldType == "number":
		return "🔢 Numeric values support integers and decimals"
	default:
		return "💡 This field accepts text input. Press Tab for suggestions."
	}
}

// AnalyzeUserBehavior analyzes user patterns to provide better suggestions
func (ch *ContextualHelp) AnalyzeUserBehavior(action string) {
	// Simple pattern analysis for better suggestions
	switch {
	case strings.Contains(action, "search"):
		ch.improveSearchSuggestions()
	case strings.Contains(action, "edit"):
		ch.improveEditingSuggestions()
	case strings.Contains(action, "navigate"):
		ch.improveNavigationSuggestions()
	}
}

// improveSearchSuggestions enhances search suggestions based on usage
func (ch *ContextualHelp) improveSearchSuggestions() {
	// Could analyze search patterns and provide better examples
}

// improveEditingSuggestions enhances editing suggestions
func (ch *ContextualHelp) improveEditingSuggestions() {
	// Could track editing patterns and provide more relevant tips
}

// improveNavigationSuggestions enhances navigation suggestions
func (ch *ContextualHelp) improveNavigationSuggestions() {
	// Could adapt to user's preferred navigation method
}

// GetOnboardingTips provides tips for new users
func (ch *ContextualHelp) GetOnboardingTips() []string {
	return []string{
		"👋 Welcome to ZeroUI! Press ? for help anytime",
		"🎯 Start by pressing Enter on any setting to edit it",
		"🔍 Use / to search for specific settings",
		"💾 Don't forget to press Ctrl+S to save your changes",
		"⚙️ Modified settings are marked with ✨",
		"🚀 Pro tip: Use number keys for quick navigation",
	}
}

// GetKeyboardShortcuts returns all available keyboard shortcuts
func (ch *ContextualHelp) GetKeyboardShortcuts() map[string]string {
	return map[string]string{
		// Navigation
		"↑, k":      "Move up",
		"↓, j":      "Move down",
		"←, h":      "Move left",
		"→, l":      "Move right",
		"Tab":       "Next section",
		"Shift+Tab": "Previous section",

		// Actions
		"Enter":  "Edit/Save",
		"Space":  "Select/Edit",
		"Esc":    "Cancel/Exit",
		"Delete": "Remove item",

		// Search
		"/":      "Start search",
		"Ctrl+F": "Find in page",

		// Help
		"?":      "Toggle help",
		"Ctrl+H": "Context help",
		"F1":     "Keyboard shortcuts",

		// Edit
		"Ctrl+A": "Select all",
		"Ctrl+C": "Copy",
		"Ctrl+V": "Paste",
		"Ctrl+X": "Cut",
		"Ctrl+Z": "Undo",
		"Ctrl+Y": "Redo",

		// Application
		"Ctrl+S": "Save",
		"Ctrl+Q": "Quit",
		"F11":    "Toggle fullscreen",
	}
}

// GetEasterEggs returns fun hidden features
func (ch *ContextualHelp) GetEasterEggs() []string {
	return []string{
		"🎮 Try pressing Ctrl+Alt+Shift+P for a surprise!",
		"🌈 Long press Tab to see color animations",
		"🎵 Type 'music' in search to see themed suggestions",
		"🎨 Press Ctrl+Alt+C to cycle color themes",
		"🚀 Type 'konami' (↑↑↓↓←→←→BA) for bonus features",
	}
}

// GenerateMotivationalMessage creates encouraging feedback
func (ch *ContextualHelp) GenerateMotivationalMessage(action string) string {
	messages := map[string][]string{
		"saved": {
			"💾 Configuration saved successfully!",
			"✅ Your changes are now active!",
			"🎉 Settings updated and ready to go!",
		},
		"edited": {
			"✨ Great edit! Your configuration is looking good.",
			"🎯 Perfect! That setting is now optimized.",
			"🚀 Nice work! Your setup just got better.",
		},
		"searched": {
			"🔍 Found what you were looking for!",
			"🎯 Search complete! Check out these results.",
			"📋 Here are the settings matching your search.",
		},
	}

	if msgs, exists := messages[action]; exists {
		return msgs[time.Now().Unix()%int64(len(msgs))]
	}

	return "🎉 Great job!"
}

// TrackFeatureUsage tracks which features users use most
func (ch *ContextualHelp) TrackFeatureUsage(feature string) {
	// Could be used for analytics and improving UX based on usage patterns
}

// GetTipsAndTricks returns advanced usage tips
func (ch *ContextualHelp) GetTipsAndTricks() []string {
	return []string{
		"💡 Use regex in search: /^font.*/ for font-related settings",
		"💡 Press Ctrl+Click to multi-select items",
		"💡 Drag and drop items to reorder (in supported views)",
		"💡 Right-click for context menu with advanced options",
		"💡 Hold Shift while navigating for faster movement",
		"💡 Use Ctrl+Enter to save and continue editing next item",
		"💡 Type 'reset' in any field to restore default value",
		"💡 Use arrow keys in number fields for fine adjustments",
	}
}
