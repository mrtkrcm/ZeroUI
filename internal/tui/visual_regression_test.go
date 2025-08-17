package tui

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

const (
	visualRegressionDir = "testdata/visual_regression"
	baselineImagesDir   = "testdata/baseline_images"
	diffImagesDir       = "testdata/diff_images"
)

// VisualRegressionTester handles visual regression testing for TUI
type VisualRegressionTester struct {
	engine            *toggle.Engine
	baselineThreshold float64
	pixelTolerance    int
	generateImages    bool
}

// TextToImageConverter converts TUI text output to images for visual comparison
type TextToImageConverter struct {
	charWidth  int
	charHeight int
	fontSize   int
	fontColor  color.Color
	bgColor    color.Color
	colorMap   map[string]color.Color
}

// VisualDiff represents differences between two visual outputs
type VisualDiff struct {
	TotalPixels     int
	DifferentPixels int
	Similarity      float64
	DiffRegions     []DiffRegion
	Summary         string
}

// DiffRegion represents a region where visuals differ
type DiffRegion struct {
	StartX, StartY int
	EndX, EndY     int
	Severity       DiffSeverity
	Description    string
}

type DiffSeverity int

const (
	MinorDiff DiffSeverity = iota
	ModerateDiff
	MajorDiff
	CriticalDiff
)

// NewVisualRegressionTester creates a visual regression testing system
func NewVisualRegressionTester(engine *toggle.Engine) *VisualRegressionTester {
	return &VisualRegressionTester{
		engine:            engine,
		baselineThreshold: 0.98, // 98% similarity required
		pixelTolerance:    5,    // Allow 5-pixel tolerance for anti-aliasing
		generateImages:    os.Getenv("GENERATE_TUI_IMAGES") == "true",
	}
}

// TestTUIVisualRegression runs comprehensive visual regression tests
func TestTUIVisualRegression(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	vrt := NewVisualRegressionTester(engine)

	// Create necessary directories
	dirs := []string{visualRegressionDir, baselineImagesDir, diffImagesDir}
	for _, dir := range dirs {
		require.NoError(t, os.MkdirAll(dir, 0755))
	}

	// Define visual test scenarios
	scenarios := []VisualTestScenario{
		{
			Name:        "MainGrid_Standard",
			Description: "Main application grid in standard terminal size",
			Width:       120,
			Height:      40,
			Setup: func(m *Model) error {
				// Default main grid setup
				return nil
			},
			CriticalElements: []string{"ZEROUI", "applications", "Ghostty", "VS Code"},
			ToleranceLevel:   ModerateDiff,
		},
		{
			Name:        "MainGrid_Small",
			Description: "Main application grid in small terminal",
			Width:       80,
			Height:      24,
			Setup: func(m *Model) error {
				return nil
			},
			CriticalElements: []string{"ZEROUI", "applications"},
			ToleranceLevel:   ModerateDiff,
		},
		{
			Name:        "HelpOverlay_Standard",
			Description: "Help overlay display",
			Width:       120,
			Height:      40,
			Setup: func(m *Model) error {
				m.showingHelp = true
				return nil
			},
			CriticalElements: []string{"Help", "Navigation", "quit"},
			ToleranceLevel:   MinorDiff,
		},
		{
			Name:        "ConfigEditor_View",
			Description: "Configuration editor interface",
			Width:       120,
			Height:      40,
			Setup: func(m *Model) error {
				m.state = ConfigEditView
				m.currentApp = "ghostty"
				return nil
			},
			CriticalElements: []string{"Config", "ghostty"},
			ToleranceLevel:   ModerateDiff,
		},
		{
			Name:        "ErrorDisplay_Standard",
			Description: "Error message display",
			Width:       120,
			Height:      40,
			Setup: func(m *Model) error {
				m.err = fmt.Errorf("test error message for visual regression")
				return nil
			},
			CriticalElements: []string{"Error", "test error message"},
			ToleranceLevel:   MinorDiff,
		},
		{
			Name:        "ResponsiveLarge_160x50",
			Description: "Large terminal responsive layout",
			Width:       160,
			Height:      50,
			Setup: func(m *Model) error {
				return nil
			},
			CriticalElements: []string{"ZEROUI", "applications", "4 columns"},
			ToleranceLevel:   ModerateDiff,
		},
		{
			Name:        "ResponsiveNarrow_60x20",
			Description: "Very narrow terminal layout",
			Width:       60,
			Height:      20,
			Setup: func(m *Model) error {
				return nil
			},
			CriticalElements: []string{"ZEROUI"},
			ToleranceLevel:   MajorDiff, // Allow more differences for extreme sizes
		},
	}

	// Run visual regression tests
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			vrt.runVisualRegressionTest(t, scenario)
		})
	}

	// Generate summary report
	vrt.generateVisualRegressionReport(t, scenarios)
}

// VisualTestScenario defines a visual regression test scenario
type VisualTestScenario struct {
	Name             string
	Description      string
	Width            int
	Height           int
	Setup            func(*Model) error
	CriticalElements []string
	ToleranceLevel   DiffSeverity
}

// runVisualRegressionTest executes a single visual regression test
func (vrt *VisualRegressionTester) runVisualRegressionTest(t *testing.T, scenario VisualTestScenario) {
	// Create model
	model, err := NewModel(vrt.engine, "")
	require.NoError(t, err)

	// Configure model
	model.width = scenario.Width
	model.height = scenario.Height
	model.updateComponentSizes()

	// Run setup
	if scenario.Setup != nil {
		err = scenario.Setup(model)
		require.NoError(t, err)
	}

	// Render view
	currentView := model.View()

	// Save current snapshot
	currentSnapshotPath := filepath.Join(visualRegressionDir, fmt.Sprintf("%s_current.txt", scenario.Name))
	err = os.WriteFile(currentSnapshotPath, []byte(currentView), 0644)
	require.NoError(t, err)

	// Convert to image if enabled
	if vrt.generateImages {
		converter := NewTextToImageConverter()
		currentImg := converter.ConvertTextToImage(currentView, scenario.Width, scenario.Height)

		currentImgPath := filepath.Join(visualRegressionDir, fmt.Sprintf("%s_current.png", scenario.Name))
		vrt.saveImage(currentImg, currentImgPath)
	}

	// Load baseline
	baselinePath := filepath.Join(baselineImagesDir, fmt.Sprintf("%s_baseline.txt", scenario.Name))
	baselineExists := true
	baselineView := ""

	if baselineData, err := os.ReadFile(baselinePath); err == nil {
		baselineView = string(baselineData)
	} else {
		baselineExists = false
		t.Logf("No baseline found for %s, creating new baseline", scenario.Name)
	}

	if !baselineExists {
		// Create new baseline
		err = os.WriteFile(baselinePath, []byte(currentView), 0644)
		require.NoError(t, err)

		if vrt.generateImages {
			converter := NewTextToImageConverter()
			baselineImg := converter.ConvertTextToImage(currentView, scenario.Width, scenario.Height)
			baselineImgPath := filepath.Join(baselineImagesDir, fmt.Sprintf("%s_baseline.png", scenario.Name))
			vrt.saveImage(baselineImg, baselineImgPath)
		}

		t.Logf("Created new baseline for %s", scenario.Name)
		return
	}

	// Compare with baseline
	diff := vrt.compareVisualOutputs(baselineView, currentView, scenario)

	// Validate critical elements
	vrt.validateCriticalElements(t, scenario.CriticalElements, currentView)

	// Check if differences are acceptable
	if diff.Similarity < vrt.baselineThreshold {
		// Generate diff visualization
		diffPath := filepath.Join(diffImagesDir, fmt.Sprintf("%s_diff.txt", scenario.Name))
		vrt.generateDiffVisualization(baselineView, currentView, diffPath)

		// Determine if this is acceptable based on tolerance level
		acceptable := vrt.isDifferenceAcceptable(diff, scenario.ToleranceLevel)

		if !acceptable {
			t.Errorf("Visual regression detected in %s:\n"+
				"Similarity: %.2f%% (threshold: %.2f%%)\n"+
				"Different pixels: %d/%d\n"+
				"Summary: %s\n"+
				"Diff saved to: %s",
				scenario.Name, diff.Similarity*100, vrt.baselineThreshold*100,
				diff.DifferentPixels, diff.TotalPixels, diff.Summary, diffPath)
		} else {
			t.Logf("Visual difference detected in %s but within tolerance level %v",
				scenario.Name, scenario.ToleranceLevel)
		}
	}
}

// compareVisualOutputs compares two visual outputs and returns differences
func (vrt *VisualRegressionTester) compareVisualOutputs(baseline, current string, scenario VisualTestScenario) *VisualDiff {
	baselineLines := strings.Split(stripAnsiCodes(baseline), "\n")
	currentLines := strings.Split(stripAnsiCodes(current), "\n")

	totalChars := 0
	differentChars := 0
	diffRegions := []DiffRegion{}

	maxLines := len(baselineLines)
	if len(currentLines) > maxLines {
		maxLines = len(currentLines)
	}

	currentRegion := (*DiffRegion)(nil)

	for i := 0; i < maxLines; i++ {
		var baselineLine, currentLine string

		if i < len(baselineLines) {
			baselineLine = baselineLines[i]
		}
		if i < len(currentLines) {
			currentLine = currentLines[i]
		}

		maxChars := len(baselineLine)
		if len(currentLine) > maxChars {
			maxChars = len(currentLine)
		}

		lineDifferences := 0

		for j := 0; j < maxChars; j++ {
			var baselineChar, currentChar rune

			if j < len(baselineLine) {
				baselineChar = rune(baselineLine[j])
			}
			if j < len(currentLine) {
				currentChar = rune(currentLine[j])
			}

			totalChars++

			if baselineChar != currentChar {
				differentChars++
				lineDifferences++

				// Track diff regions
				if currentRegion == nil {
					currentRegion = &DiffRegion{
						StartX: j,
						StartY: i,
						EndX:   j,
						EndY:   i,
					}
				} else {
					currentRegion.EndX = j
					currentRegion.EndY = i
				}
			} else if currentRegion != nil {
				// End of diff region
				currentRegion.Severity = vrt.calculateDiffSeverity(currentRegion, totalChars)
				currentRegion.Description = fmt.Sprintf("Diff region at (%d,%d) to (%d,%d)",
					currentRegion.StartX, currentRegion.StartY, currentRegion.EndX, currentRegion.EndY)
				diffRegions = append(diffRegions, *currentRegion)
				currentRegion = nil
			}
		}
	}

	// Close final diff region if exists
	if currentRegion != nil {
		currentRegion.Severity = vrt.calculateDiffSeverity(currentRegion, totalChars)
		diffRegions = append(diffRegions, *currentRegion)
	}

	similarity := 1.0
	if totalChars > 0 {
		similarity = float64(totalChars-differentChars) / float64(totalChars)
	}

	summary := fmt.Sprintf("%d differences across %d regions", differentChars, len(diffRegions))

	return &VisualDiff{
		TotalPixels:     totalChars,
		DifferentPixels: differentChars,
		Similarity:      similarity,
		DiffRegions:     diffRegions,
		Summary:         summary,
	}
}

// calculateDiffSeverity determines the severity of a diff region
func (vrt *VisualRegressionTester) calculateDiffSeverity(region *DiffRegion, totalPixels int) DiffSeverity {
	regionSize := (region.EndX - region.StartX + 1) * (region.EndY - region.StartY + 1)
	percentage := float64(regionSize) / float64(totalPixels)

	switch {
	case percentage < 0.01: // Less than 1%
		return MinorDiff
	case percentage < 0.05: // Less than 5%
		return ModerateDiff
	case percentage < 0.15: // Less than 15%
		return MajorDiff
	default:
		return CriticalDiff
	}
}

// isDifferenceAcceptable checks if the difference is within tolerance
func (vrt *VisualRegressionTester) isDifferenceAcceptable(diff *VisualDiff, toleranceLevel DiffSeverity) bool {
	// Check if any diff region exceeds the tolerance level
	for _, region := range diff.DiffRegions {
		if region.Severity > toleranceLevel {
			return false
		}
	}

	// Additional similarity threshold based on tolerance level
	var minSimilarity float64
	switch toleranceLevel {
	case MinorDiff:
		minSimilarity = 0.99
	case ModerateDiff:
		minSimilarity = 0.95
	case MajorDiff:
		minSimilarity = 0.90
	case CriticalDiff:
		minSimilarity = 0.80
	}

	return diff.Similarity >= minSimilarity
}

// validateCriticalElements ensures critical UI elements are present
func (vrt *VisualRegressionTester) validateCriticalElements(t *testing.T, elements []string, view string) {
	cleanView := stripAnsiCodes(view)

	for _, element := range elements {
		assert.Contains(t, cleanView, element,
			"Critical UI element missing: %s", element)
	}
}

// generateDiffVisualization creates a visual diff representation
func (vrt *VisualRegressionTester) generateDiffVisualization(baseline, current, outputPath string) error {
	baselineLines := strings.Split(stripAnsiCodes(baseline), "\n")
	currentLines := strings.Split(stripAnsiCodes(current), "\n")

	var diff strings.Builder
	diff.WriteString("VISUAL DIFF ANALYSIS\n")
	diff.WriteString("===================\n\n")

	maxLines := len(baselineLines)
	if len(currentLines) > maxLines {
		maxLines = len(currentLines)
	}

	for i := 0; i < maxLines; i++ {
		var baselineLine, currentLine string

		if i < len(baselineLines) {
			baselineLine = baselineLines[i]
		}
		if i < len(currentLines) {
			currentLine = currentLines[i]
		}

		if baselineLine != currentLine {
			diff.WriteString(fmt.Sprintf("Line %d:\n", i+1))
			diff.WriteString(fmt.Sprintf("  BASELINE: %s\n", baselineLine))
			diff.WriteString(fmt.Sprintf("  CURRENT:  %s\n", currentLine))
			diff.WriteString("\n")
		}
	}

	return os.WriteFile(outputPath, []byte(diff.String()), 0644)
}

// generateVisualRegressionReport creates a summary report
func (vrt *VisualRegressionTester) generateVisualRegressionReport(t *testing.T, scenarios []VisualTestScenario) {
	reportPath := filepath.Join(visualRegressionDir, "regression_report.md")

	var report strings.Builder
	report.WriteString("# Visual Regression Test Report\n\n")
	report.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))

	report.WriteString("## Test Summary\n\n")
	report.WriteString(fmt.Sprintf("Total scenarios tested: %d\n\n", len(scenarios)))

	report.WriteString("## Scenarios\n\n")
	for _, scenario := range scenarios {
		report.WriteString(fmt.Sprintf("### %s\n", scenario.Name))
		report.WriteString(fmt.Sprintf("- **Description**: %s\n", scenario.Description))
		report.WriteString(fmt.Sprintf("- **Dimensions**: %dx%d\n", scenario.Width, scenario.Height))
		report.WriteString(fmt.Sprintf("- **Tolerance**: %v\n", scenario.ToleranceLevel))
		report.WriteString(fmt.Sprintf("- **Critical Elements**: %v\n", scenario.CriticalElements))
		report.WriteString("\n")
	}

	report.WriteString("## Files Generated\n\n")
	report.WriteString("- Current snapshots: `testdata/visual_regression/`\n")
	report.WriteString("- Baseline images: `testdata/baseline_images/`\n")
	report.WriteString("- Diff visualizations: `testdata/diff_images/`\n\n")

	err := os.WriteFile(reportPath, []byte(report.String()), 0644)
	if err != nil {
		t.Logf("Failed to write regression report: %v", err)
	} else {
		t.Logf("Visual regression report saved to: %s", reportPath)
	}
}

// NewTextToImageConverter creates a text-to-image converter for visual testing
func NewTextToImageConverter() *TextToImageConverter {
	return &TextToImageConverter{
		charWidth:  8,
		charHeight: 16,
		fontSize:   12,
		fontColor:  color.RGBA{255, 255, 255, 255}, // White text
		bgColor:    color.RGBA{0, 0, 0, 255},       // Black background
		colorMap: map[string]color.Color{
			"red":     color.RGBA{255, 0, 0, 255},
			"green":   color.RGBA{0, 255, 0, 255},
			"blue":    color.RGBA{0, 0, 255, 255},
			"yellow":  color.RGBA{255, 255, 0, 255},
			"magenta": color.RGBA{255, 0, 255, 255},
			"cyan":    color.RGBA{0, 255, 255, 255},
		},
	}
}

// ConvertTextToImage converts terminal text output to an image
func (tic *TextToImageConverter) ConvertTextToImage(text string, termWidth, termHeight int) image.Image {
	// Create image
	img := image.NewRGBA(image.Rect(0, 0, termWidth*tic.charWidth, termHeight*tic.charHeight))

	// Fill background
	for y := 0; y < termHeight*tic.charHeight; y++ {
		for x := 0; x < termWidth*tic.charWidth; x++ {
			img.Set(x, y, tic.bgColor)
		}
	}

	// Process text lines
	lines := strings.Split(stripAnsiCodes(text), "\n")
	for lineNum, line := range lines {
		if lineNum >= termHeight {
			break
		}

		for charNum, char := range line {
			if charNum >= termWidth {
				break
			}

			// Simple character rendering (would need actual font rendering for production)
			if char != ' ' {
				tic.drawCharacter(img, char, charNum*tic.charWidth, lineNum*tic.charHeight)
			}
		}
	}

	return img
}

// drawCharacter draws a simple representation of a character
func (tic *TextToImageConverter) drawCharacter(img *image.RGBA, char rune, x, y int) {
	// Simplified character rendering - draws a small rectangle for each character
	// In production, you would use actual font rendering
	for dy := 2; dy < tic.charHeight-2; dy++ {
		for dx := 1; dx < tic.charWidth-1; dx++ {
			// Simple pattern based on character value
			if (int(char)+dx+dy)%3 == 0 {
				img.Set(x+dx, y+dy, tic.fontColor)
			}
		}
	}
}

// saveImage saves an image to a file
func (vrt *VisualRegressionTester) saveImage(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	return png.Encode(file, img)
}

// TestContinuousIntegration tests TUI in CI environment
func TestContinuousIntegration(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("Skipping CI-specific tests")
	}

	// CI-specific visual tests with stricter requirements
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Test standard CI terminal size
	model.width = 80
	model.height = 24
	model.updateComponentSizes()

	view := model.View()

	// CI validations
	assert.NotEmpty(t, view, "CI: View should not be empty")
	assert.Contains(t, view, "ZEROUI", "CI: Should contain app title")

	// Ensure output fits in CI terminal
	lines := strings.Split(stripAnsiCodes(view), "\n")
	assert.LessOrEqual(t, len(lines), 24, "CI: Should fit in 24 lines")

	for i, line := range lines {
		assert.LessOrEqual(t, len(line), 80, "CI: Line %d should fit in 80 characters", i)
	}
}

// BenchmarkTUIRendering benchmarks TUI rendering performance
func BenchmarkTUIRendering(b *testing.B) {
	engine, _ := toggle.NewEngine()
	model, _ := NewTestModel(engine, "")

	model.width = 120
	model.height = 40
	model.updateComponentSizes()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

// BenchmarkTUIInteraction benchmarks TUI interaction handling
func BenchmarkTUIInteraction(b *testing.B) {
	engine, _ := toggle.NewEngine()
	model, _ := NewTestModel(engine, "")

	b.ResetTimer()
	b.ReportAllocs()

	keys := []tea.KeyType{tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight}

	for i := 0; i < b.N; i++ {
		keyMsg := tea.KeyMsg{Type: keys[i%len(keys)]}
		model.Update(keyMsg)
	}
}
