package keybindings

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/stretchr/testify/assert"
)

func TestNewAppKeyMap(t *testing.T) {
	keyMap := NewAppKeyMap()
	
	// Test that all key bindings are properly initialized
	assert.NotNil(t, keyMap.Up)
	assert.NotNil(t, keyMap.Down)
	assert.NotNil(t, keyMap.Left)
	assert.NotNil(t, keyMap.Right)
	assert.NotNil(t, keyMap.Enter)
	assert.NotNil(t, keyMap.Select)
	assert.NotNil(t, keyMap.Back)
	assert.NotNil(t, keyMap.Home)
	assert.NotNil(t, keyMap.End)
	assert.NotNil(t, keyMap.Refresh)
	assert.NotNil(t, keyMap.Edit)
	assert.NotNil(t, keyMap.Save)
	assert.NotNil(t, keyMap.Cancel)
	assert.NotNil(t, keyMap.Reset)
	assert.NotNil(t, keyMap.Search)
	assert.NotNil(t, keyMap.Filter)
	assert.NotNil(t, keyMap.Help)
	assert.NotNil(t, keyMap.Quit)
	assert.NotNil(t, keyMap.ForceQuit)
	assert.NotNil(t, keyMap.ToggleMode)
	assert.NotNil(t, keyMap.TogglePreview)
	assert.NotNil(t, keyMap.ToggleHelp)
	assert.NotNil(t, keyMap.NextField)
	assert.NotNil(t, keyMap.PrevField)
	assert.NotNil(t, keyMap.SubmitForm)
	assert.NotNil(t, keyMap.CancelForm)
	assert.NotNil(t, keyMap.Debug)
	assert.NotNil(t, keyMap.Settings)
}

func TestAppKeyMap_ShortHelp(t *testing.T) {
	keyMap := NewAppKeyMap()
	shortHelp := keyMap.ShortHelp()
	
	// Should return 4 key bindings for short help
	assert.Len(t, shortHelp, 4)
	
	// Check that help keys are included
	expectedKeys := []key.Binding{keyMap.Help, keyMap.Select, keyMap.Back, keyMap.Quit}
	assert.Equal(t, expectedKeys, shortHelp)
}

func TestAppKeyMap_FullHelp(t *testing.T) {
	keyMap := NewAppKeyMap()
	fullHelp := keyMap.FullHelp()
	
	// Should return multiple rows of help
	assert.Len(t, fullHelp, 6)
	
	// Each row should have key bindings
	for i, row := range fullHelp {
		assert.NotEmpty(t, row, "Row %d should not be empty", i)
		
		// Each key binding should be valid
		for j, binding := range row {
			assert.NotNil(t, binding, "Binding at row %d, col %d should not be nil", i, j)
		}
	}
}

func TestFormKeyMap(t *testing.T) {
	formKeyMap := NewFormKeyMap()
	
	// Test that all form key bindings are initialized
	assert.NotNil(t, formKeyMap.NextField)
	assert.NotNil(t, formKeyMap.PrevField)
	assert.NotNil(t, formKeyMap.Submit)
	assert.NotNil(t, formKeyMap.Cancel)
	assert.NotNil(t, formKeyMap.Reset)
	assert.NotNil(t, formKeyMap.Clear)
}

func TestListKeyMap(t *testing.T) {
	listKeyMap := NewListKeyMap()
	
	// Test that all list key bindings are initialized
	assert.NotNil(t, listKeyMap.Up)
	assert.NotNil(t, listKeyMap.Down)
	assert.NotNil(t, listKeyMap.PageUp)
	assert.NotNil(t, listKeyMap.PageDown)
	assert.NotNil(t, listKeyMap.Home)
	assert.NotNil(t, listKeyMap.End)
	assert.NotNil(t, listKeyMap.Select)
	assert.NotNil(t, listKeyMap.Filter)
	assert.NotNil(t, listKeyMap.ClearFilter)
	assert.NotNil(t, listKeyMap.Refresh)
}

func TestHelpKeyMap(t *testing.T) {
	helpKeyMap := NewHelpKeyMap()
	
	// Test that all help key bindings are initialized
	assert.NotNil(t, helpKeyMap.Close)
	assert.NotNil(t, helpKeyMap.ScrollUp)
	assert.NotNil(t, helpKeyMap.ScrollDown)
	assert.NotNil(t, helpKeyMap.NextPage)
	assert.NotNil(t, helpKeyMap.PrevPage)
}

func TestKeyBindingKeys(t *testing.T) {
	keyMap := NewAppKeyMap()
	
	// Test that key bindings have the expected keys
	upKeys := keyMap.Up.Keys()
	assert.Contains(t, upKeys, "up")
	assert.Contains(t, upKeys, "k")
	
	downKeys := keyMap.Down.Keys()
	assert.Contains(t, downKeys, "down")
	assert.Contains(t, downKeys, "j")
	
	quitKeys := keyMap.Quit.Keys()
	assert.Contains(t, quitKeys, "q")
	assert.Contains(t, quitKeys, "ctrl+c")
	
	enterKeys := keyMap.Enter.Keys()
	assert.Contains(t, enterKeys, "enter")
	
	searchKeys := keyMap.Search.Keys()
	assert.Contains(t, searchKeys, "/")
}

func TestKeyBindingHelp(t *testing.T) {
	keyMap := NewAppKeyMap()
	
	// Test that key bindings have help text
	assert.NotEmpty(t, keyMap.Up.Help().Key)
	assert.NotEmpty(t, keyMap.Up.Help().Desc)
	assert.Equal(t, "↑/k", keyMap.Up.Help().Key)
	assert.Equal(t, "move up", keyMap.Up.Help().Desc)
	
	assert.NotEmpty(t, keyMap.Down.Help().Key)
	assert.NotEmpty(t, keyMap.Down.Help().Desc)
	assert.Equal(t, "↓/j", keyMap.Down.Help().Key)
	assert.Equal(t, "move down", keyMap.Down.Help().Desc)
	
	assert.NotEmpty(t, keyMap.Quit.Help().Key)
	assert.NotEmpty(t, keyMap.Quit.Help().Desc)
	assert.Equal(t, "q/ctrl+c", keyMap.Quit.Help().Key)
	assert.Equal(t, "quit", keyMap.Quit.Help().Desc)
}

func TestSpecialKeyBindings(t *testing.T) {
	keyMap := NewAppKeyMap()
	
	// Test debug key binding
	debugKeys := keyMap.Debug.Keys()
	assert.Contains(t, debugKeys, "ctrl+shift+d")
	
	// Test settings key binding
	settingsKeys := keyMap.Settings.Keys()
	assert.Contains(t, settingsKeys, "ctrl+,")
	
	// Test force quit
	forceQuitKeys := keyMap.ForceQuit.Keys()
	assert.Contains(t, forceQuitKeys, "ctrl+d")
	
	// Test toggle mode
	toggleModeKeys := keyMap.ToggleMode.Keys()
	assert.Contains(t, toggleModeKeys, "ctrl+m")
}

func TestFormSpecificKeys(t *testing.T) {
	formKeyMap := NewFormKeyMap()
	
	// Test tab navigation
	nextKeys := formKeyMap.NextField.Keys()
	assert.Contains(t, nextKeys, "tab")
	assert.Contains(t, nextKeys, "down")
	
	prevKeys := formKeyMap.PrevField.Keys()
	assert.Contains(t, prevKeys, "shift+tab")
	assert.Contains(t, prevKeys, "up")
	
	// Test form submission
	submitKeys := formKeyMap.Submit.Keys()
	assert.Contains(t, submitKeys, "enter")
	assert.Contains(t, submitKeys, "ctrl+s")
	
	// Test form cancellation
	cancelKeys := formKeyMap.Cancel.Keys()
	assert.Contains(t, cancelKeys, "esc")
	
	// Test form reset
	resetKeys := formKeyMap.Reset.Keys()
	assert.Contains(t, resetKeys, "ctrl+r")
	
	// Test field clear
	clearKeys := formKeyMap.Clear.Keys()
	assert.Contains(t, clearKeys, "ctrl+l")
}

func TestListSpecificKeys(t *testing.T) {
	listKeyMap := NewListKeyMap()
	
	// Test navigation keys
	upKeys := listKeyMap.Up.Keys()
	assert.Contains(t, upKeys, "up")
	assert.Contains(t, upKeys, "k")
	
	downKeys := listKeyMap.Down.Keys()
	assert.Contains(t, downKeys, "down")
	assert.Contains(t, downKeys, "j")
	
	// Test page navigation
	pageUpKeys := listKeyMap.PageUp.Keys()
	assert.Contains(t, pageUpKeys, "pgup")
	assert.Contains(t, pageUpKeys, "b")
	
	pageDownKeys := listKeyMap.PageDown.Keys()
	assert.Contains(t, pageDownKeys, "pgdown")
	assert.Contains(t, pageDownKeys, "f")
	
	// Test home/end navigation
	homeKeys := listKeyMap.Home.Keys()
	assert.Contains(t, homeKeys, "home")
	assert.Contains(t, homeKeys, "g")
	
	endKeys := listKeyMap.End.Keys()
	assert.Contains(t, endKeys, "end")
	assert.Contains(t, endKeys, "G")
	
	// Test selection
	selectKeys := listKeyMap.Select.Keys()
	assert.Contains(t, selectKeys, "enter")
	assert.Contains(t, selectKeys, " ")
	
	// Test filtering
	filterKeys := listKeyMap.Filter.Keys()
	assert.Contains(t, filterKeys, "/")
	
	clearFilterKeys := listKeyMap.ClearFilter.Keys()
	assert.Contains(t, clearFilterKeys, "ctrl+l")
	assert.Contains(t, clearFilterKeys, "esc")
	
	// Test refresh
	refreshKeys := listKeyMap.Refresh.Keys()
	assert.Contains(t, refreshKeys, "r")
	assert.Contains(t, refreshKeys, "F5")
}

func TestHelpSpecificKeys(t *testing.T) {
	helpKeyMap := NewHelpKeyMap()
	
	// Test close help
	closeKeys := helpKeyMap.Close.Keys()
	assert.Contains(t, closeKeys, "q")
	assert.Contains(t, closeKeys, "esc")
	
	// Test scrolling
	scrollUpKeys := helpKeyMap.ScrollUp.Keys()
	assert.Contains(t, scrollUpKeys, "up")
	assert.Contains(t, scrollUpKeys, "k")
	
	scrollDownKeys := helpKeyMap.ScrollDown.Keys()
	assert.Contains(t, scrollDownKeys, "down")
	assert.Contains(t, scrollDownKeys, "j")
	
	// Test page navigation
	nextPageKeys := helpKeyMap.NextPage.Keys()
	assert.Contains(t, nextPageKeys, "right")
	assert.Contains(t, nextPageKeys, "l")
	assert.Contains(t, nextPageKeys, "pgdown")
	
	prevPageKeys := helpKeyMap.PrevPage.Keys()
	assert.Contains(t, prevPageKeys, "left")
	assert.Contains(t, prevPageKeys, "h")
	assert.Contains(t, prevPageKeys, "pgup")
}