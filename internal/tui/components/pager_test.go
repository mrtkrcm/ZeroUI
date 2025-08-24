package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfigPager(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		content     string
		wantInView  string
		wantMissing string
	}{
		{
			name:     "displays file path",
			filePath: "/home/user/.config/app/config.toml",
			content:  "test = true",
			wantInView: "Viewing: /home/user/.config/app/config.toml",
		},
		{
			name:     "shows line numbers",
			filePath: "test.conf",
			content:  "line1\nline2\nline3",
			wantInView: "   1",
		},
		{
			name:     "handles empty content",
			filePath: "empty.conf",
			content:  "",
			wantInView: "No configuration file content available",
		},
		{
			name:     "syntax highlights comments",
			filePath: "config.toml",
			content:  "# This is a comment\nkey = value",
			wantInView: "# This is a comment",
		},
		{
			name:     "syntax highlights sections",
			filePath: "config.ini",
			content:  "[section]\nkey = value",
			wantInView: "[section]",
		},
		{
			name:     "shows scroll percentage",
			filePath: "test.conf",
			content:  strings.Repeat("line\n", 100),
			wantInView: "0%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create pager
			pager := NewConfigPager()
			pager.SetContent(tt.filePath, tt.content)
			
			// Initialize with window size
			pager.Update(tea.WindowSizeMsg{
				Width:  80,
				Height: 24,
			})
			
			// Get the view
			view := pager.View()
			
			// Check expected content
			if tt.wantInView != "" && !strings.Contains(view, tt.wantInView) {
				t.Errorf("View should contain %q, got:\n%s", tt.wantInView, view)
			}
			
			// Check missing content
			if tt.wantMissing != "" && strings.Contains(view, tt.wantMissing) {
				t.Errorf("View should not contain %q, got:\n%s", tt.wantMissing, view)
			}
		})
	}
}

func TestPagerNavigation(t *testing.T) {
	// Create pager with long content
	pager := NewConfigPager()
	content := strings.Repeat("line\n", 100)
	pager.SetContent("test.conf", content)
	
	// Initialize
	pager.Update(tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	})
	
	// Test scrolling down
	pager.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	
	// Test jumping to bottom
	pager.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	view := pager.View()
	if !strings.Contains(view, "100%") && !strings.Contains(view, "99%") {
		t.Error("Should show near 100% after jumping to bottom")
	}
	
	// Test jumping to top
	pager.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	view = pager.View()
	if !strings.Contains(view, "0%") && !strings.Contains(view, "1%") {
		t.Error("Should show near 0% after jumping to top")
	}
}

func TestPagerSyntaxHighlighting(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantStyle bool
	}{
		{
			name:      "comment with #",
			line:      "# This is a comment",
			wantStyle: true,
		},
		{
			name:      "comment with //",
			line:      "// This is also a comment",
			wantStyle: true,
		},
		{
			name:      "section header",
			line:      "[database]",
			wantStyle: true,
		},
		{
			name:      "key-value with =",
			line:      "port = 5432",
			wantStyle: true,
		},
		{
			name:      "key-value with :",
			line:      "host: localhost",
			wantStyle: true,
		},
	}
	
	pager := NewConfigPager()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			highlighted := pager.highlightLine(tt.line)
			// Syntax highlighting should add ANSI codes or modify the line
			if tt.wantStyle && highlighted == tt.line {
				t.Errorf("Expected syntax highlighting for %q", tt.line)
			}
		})
	}
}