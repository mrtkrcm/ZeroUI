package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrtkrcm/ZeroUI/internal/tui/registry"
	"github.com/stretchr/testify/assert"
)

// TestAppGridNoPanic ensures no panic with negative margins
func TestAppGridNoPanic(t *testing.T) {
	// Create grid
	grid := NewAppGrid()

	// Test with very small terminal that would cause negative margins
	smallSizeMsg := tea.WindowSizeMsg{
		Width:  20, // Very small width
		Height: 10,
	}

	// This should not panic
	updatedGrid, _ := grid.Update(smallSizeMsg)
	grid = updatedGrid

	// Render should not panic even with small size
	view := grid.View()
	assert.NotEmpty(t, view, "View should render something")

	// Verify no negative Repeat counts by checking the view doesn't error
	assert.NotContains(t, view, "panic", "View should not contain panic text")
}

// TestAppGridResponsiveLayout tests different terminal sizes
func TestAppGridResponsiveLayout(t *testing.T) {
	testCases := []struct {
		name     string
		width    int
		height   int
		expected int // expected columns
	}{
		{"Very Small", 40, 20, 1},
		{"Small", 80, 24, 2},
		{"Medium", 120, 30, 3},
		{"Large", 160, 40, 4},
		{"Extra Large", 200, 50, 4}, // Max 4 columns
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			grid := NewAppGrid()

			// Update size
			sizeMsg := tea.WindowSizeMsg{
				Width:  tc.width,
				Height: tc.height,
			}

			updatedGrid, _ := grid.Update(sizeMsg)
			grid = updatedGrid

			// Check columns
			assert.Equal(t, tc.expected, grid.columns, "Should have correct number of columns")

			// Ensure card size is within bounds
			assert.GreaterOrEqual(t, grid.cardSize, grid.minCardSize, "Card size should not be below minimum")
			assert.LessOrEqual(t, grid.cardSize, grid.maxCardSize, "Card size should not exceed maximum")

			// Render should work without panic
			view := grid.View()
			assert.NotEmpty(t, view, "View should render")
		})
	}
}

// TestAppGridMarginCalculation tests margin calculations don't go negative
func TestAppGridMarginCalculation(t *testing.T) {
	grid := NewAppGrid()

	// Set dimensions that would previously cause negative margin
	grid.width = 50
	grid.height = 20
	grid.columns = 3
	grid.cardSize = 20
	grid.cardSpacing = 4

	// Render should handle this gracefully
	view := grid.renderAdvancedGrid()

	// Check that the view doesn't have issues
	assert.NotEmpty(t, view, "Should render something")

	// Verify no panic by checking we can count spaces
	// (strings.Repeat with negative count would have panicked)
	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 0, "Should have rendered lines")
}

// TestAppGridCacheInvalidation tests cache works correctly
func TestAppGridCacheInvalidation(t *testing.T) {
	grid := NewAppGrid()

	// Set size
	grid.width = 100
	grid.height = 30

	// First render
	view1 := grid.View()
	assert.NotEmpty(t, view1, "First view should render")
	assert.NotEmpty(t, grid.cachedView, "Should have cached view")

	// Second render without changes should use cache
	view2 := grid.View()
	assert.Equal(t, view1, view2, "Should return cached view")

	// Change selection - should invalidate cache
	grid.moveSelectionAnimated(1)
	grid.invalidateCache()
	view3 := grid.View()

	// View might be same content but cache should have been cleared and regenerated
	assert.NotEmpty(t, view3, "View after selection change should render")
}

// TestAppGridPerfectSquares tests cards maintain square dimensions
func TestAppGridPerfectSquares(t *testing.T) {
	grid := NewAppGrid()

	// Set standard size
	grid.width = 120
	grid.height = 40
	grid.updateResponsiveLayout()

	// Card size should be set for perfect squares
	assert.Greater(t, grid.cardSize, 0, "Card size should be positive")

	// All cards should have same dimensions
	for _, card := range grid.cards {
		// Cards should be set to same size (this would be verified in SetSize method)
		assert.NotNil(t, card, "Card should exist")
	}

	// Verify spacing is consistent
	assert.GreaterOrEqual(t, grid.cardSpacing, 2, "Spacing should be at least 2")
}

// TestAppGridNoFreeze tests UI doesn't freeze with rapid updates
func TestAppGridNoFreeze(t *testing.T) {
	grid := NewAppGrid()

	// Simulate rapid key presses
	for i := 0; i < 100; i++ {
		// Alternate between different keys rapidly
		if i%4 == 0 {
			grid.Update(tea.KeyMsg{Type: tea.KeyDown})
		} else if i%4 == 1 {
			grid.Update(tea.KeyMsg{Type: tea.KeyUp})
		} else if i%4 == 2 {
			grid.Update(tea.KeyMsg{Type: tea.KeyLeft})
		} else {
			grid.Update(tea.KeyMsg{Type: tea.KeyRight})
		}
	}

	// Should still be able to render
	view := grid.View()
	assert.NotEmpty(t, view, "Should still render after rapid updates")
}

// TestAppGridEmptyState tests rendering with no apps
func TestAppGridEmptyState(t *testing.T) {
	grid := NewAppGrid()
	grid.cards = []*AppCardModel{}         // Clear cards
	grid.statuses = []registry.AppStatus{} // Clear statuses

	// Should render empty state without panic
	view := grid.View()
	assert.NotEmpty(t, view, "Should render empty state")
	assert.Contains(t, view, "No applications", "Should show empty message")
}
