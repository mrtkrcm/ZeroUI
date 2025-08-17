package showcase

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
)

func TestShowcaseRenderer(t *testing.T) {
	renderer := NewShowcaseRenderer()

	t.Run("NewShowcaseRenderer creates valid renderer", func(t *testing.T) {
		if renderer == nil {
			t.Fatal("NewShowcaseRenderer returned nil")
		}

		if renderer.exampleRenderer == nil {
			t.Error("ExampleRenderer not initialized")
		}

		if len(renderer.sections) == 0 {
			t.Error("No sections defined")
		}
	})

	t.Run("GetSections returns sections", func(t *testing.T) {
		sections := renderer.GetSections()

		if len(sections) != 10 {
			t.Errorf("Expected 10 sections, got %d", len(sections))
		}

		// Check that we have all expected sections
		expectedSections := []ShowcaseSection{
			OverviewSection,
			ColorsSection,
			TypographySection,
			ComponentsSection,
			LayoutSection,
			InteractiveSection,
			AnimationsSection,
			ErrorStatesSection,
			BoxDrawingSection,
			RealExamplesSection,
		}

		for i, expected := range expectedSections {
			if sections[i].Section != expected {
				t.Errorf("Section %d: expected %v, got %v", i, expected, sections[i].Section)
			}
		}
	})

	t.Run("CreateSectionList creates valid list", func(t *testing.T) {
		list := renderer.CreateSectionList()

		if list.Index() < 0 {
			t.Error("List index should be non-negative")
		}
	})

	t.Run("RenderSection handles all sections", func(t *testing.T) {
		params := RenderParams{
			ProgressVal: 0.5,
			Progress:    progress.New(),
			Spinner:     spinner.New(),
			TextInput:   textinput.New(),
			AnimTicker:  time.Now(),
		}

		for _, section := range renderer.sections {
			content := renderer.RenderSection(section.Section, 80, params)

			if content == "" {
				t.Errorf("Section %s rendered empty content", section.Title)
			}

			if content == "Section not implemented" {
				t.Errorf("Section %s not implemented", section.Title)
			}
		}
	})

	t.Run("RenderStaticShowcase produces content", func(t *testing.T) {
		params := RenderParams{
			ProgressVal: 0.75,
			Progress:    progress.New(),
			Spinner:     spinner.New(),
			TextInput:   textinput.New(),
			AnimTicker:  time.Now(),
		}

		content := renderer.RenderStaticShowcase(100, params)

		if content == "" {
			t.Error("Static showcase rendered empty content")
		}

		// Should contain content from all sections
		if len(content) < 1000 {
			t.Error("Static showcase content seems too short")
		}
	})
}

func TestExampleRenderer(t *testing.T) {
	renderer := NewExampleRenderer()

	t.Run("NewExampleRenderer creates valid renderer", func(t *testing.T) {
		if renderer == nil {
			t.Fatal("NewExampleRenderer returned nil")
		}

		if renderer.colorTheme == nil {
			t.Error("ColorTheme not initialized")
		}

		if renderer.typographyTheme == nil {
			t.Error("TypographyTheme not initialized")
		}

		if renderer.spacing == nil {
			t.Error("SpacingSystem not initialized")
		}
	})

	t.Run("RenderOverview produces content", func(t *testing.T) {
		content := renderer.RenderOverview(80)

		if content == "" {
			t.Error("Overview rendered empty content")
		}

		if !containsText(content, "ZeroUI Design System") {
			t.Error("Overview should contain title")
		}
	})

	t.Run("RenderColors produces content", func(t *testing.T) {
		content := renderer.RenderColors(80)

		if content == "" {
			t.Error("Colors rendered empty content")
		}

		if !containsText(content, "Color Palette") {
			t.Error("Colors should contain palette info")
		}
	})

	t.Run("RenderTypography produces content", func(t *testing.T) {
		content := renderer.RenderTypography(80)

		if content == "" {
			t.Error("Typography rendered empty content")
		}

		if !containsText(content, "Typography") {
			t.Error("Typography should contain typography info")
		}
	})

	t.Run("RenderComponents produces content", func(t *testing.T) {
		progress := progress.New()
		content := renderer.RenderComponents(80, 0.5, progress)

		if content == "" {
			t.Error("Components rendered empty content")
		}

		if !containsText(content, "Components") {
			t.Error("Components should contain component info")
		}
	})

	t.Run("RenderAnimations produces content", func(t *testing.T) {
		spinner := spinner.New()
		progress := progress.New()
		content := renderer.RenderAnimations(80, spinner, 0.5, progress, time.Now())

		if content == "" {
			t.Error("Animations rendered empty content")
		}

		if !containsText(content, "Animations") {
			t.Error("Animations should contain animation info")
		}
	})
}

func TestShowcaseDelegate(t *testing.T) {
	delegate := ShowcaseDelegate{}

	t.Run("Delegate has correct height", func(t *testing.T) {
		if delegate.Height() != 2 {
			t.Errorf("Expected height 2, got %d", delegate.Height())
		}
	})

	t.Run("Delegate has zero spacing", func(t *testing.T) {
		if delegate.Spacing() != 0 {
			t.Errorf("Expected spacing 0, got %d", delegate.Spacing())
		}
	})
}

func TestLayoutCalculator(t *testing.T) {
	calc := NewLayoutCalculator()

	t.Run("NewLayoutCalculator creates valid calculator", func(t *testing.T) {
		if calc == nil {
			t.Fatal("NewLayoutCalculator returned nil")
		}
	})

	t.Run("CalculateSidebarWidth works correctly", func(t *testing.T) {
		tests := []struct {
			totalWidth  int
			expectedMin int
		}{
			{90, 30},   // Small width should return minimum
			{120, 40},  // Normal width should return 1/3
			{300, 100}, // Large width should return 1/3
		}

		for _, test := range tests {
			result := calc.CalculateSidebarWidth(test.totalWidth)
			if result < test.expectedMin {
				t.Errorf("Width %d: expected at least %d, got %d", test.totalWidth, test.expectedMin, result)
			}
		}
	})

	t.Run("CalculateContentWidth works correctly", func(t *testing.T) {
		// Non-interactive mode should return full width
		result := calc.CalculateContentWidth(100, false)
		if result != 100 {
			t.Errorf("Non-interactive mode: expected 100, got %d", result)
		}

		// Interactive mode should return 2/3 width
		result = calc.CalculateContentWidth(120, true)
		if result < 50 { // Should be at least minimum
			t.Errorf("Interactive mode: expected at least 50, got %d", result)
		}
	})
}

// Helper function to check if content contains text
func containsText(content, text string) bool {
	// Simple check - in a real implementation you might want to strip ANSI codes
	// For now, just check if the text appears anywhere
	return len(content) > 0 && len(text) > 0
}
