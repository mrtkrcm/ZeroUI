package feedback

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// LoadingSystem manages loading states and progress indicators
type LoadingSystem struct {
	activeLoaders map[string]*Loader
	spinners      map[string]spinner.Model
	theme         LoadingTheme
}

// Loader represents an active loading operation
type Loader struct {
	ID          string
	Message     string
	StartTime   time.Time
	Spinner     spinner.Model
	Progress    *Progress
	Steps       []LoadingStep
	CurrentStep int
}

// LoadingStep represents a step in a multi-step operation
type LoadingStep struct {
	Name        string
	Description string
	Duration    time.Duration
	Completed   bool
}

// Progress represents progress information
type Progress struct {
	Current int
	Total   int
	Message string
}

// LoadingTheme defines the visual theme for loading indicators
type LoadingTheme struct {
	PrimaryColor    string
	SecondaryColor  string
	SuccessColor    string
	ErrorColor      string
	TextColor       string
	BackgroundColor string
}

// DefaultLoadingTheme provides a beautiful default theme
var DefaultLoadingTheme = LoadingTheme{
	PrimaryColor:    "#bd93f9",
	SecondaryColor:  "#6272a4",
	SuccessColor:    "#50fa7b",
	ErrorColor:      "#ff5555",
	TextColor:       "#f8f8f2",
	BackgroundColor: "#1e1e2e",
}

// NewLoadingSystem creates a new loading system
func NewLoadingSystem() *LoadingSystem {
	return &LoadingSystem{
		activeLoaders: make(map[string]*Loader),
		spinners:      make(map[string]spinner.Model),
		theme:         DefaultLoadingTheme,
	}
}

// StartLoading starts a loading operation
func (ls *LoadingSystem) StartLoading(id, message string) {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(ls.theme.PrimaryColor))

	loader := &Loader{
		ID:        id,
		Message:   message,
		StartTime: time.Now(),
		Spinner:   s,
	}

	ls.activeLoaders[id] = loader
	ls.spinners[id] = s
}

// StartStepLoading starts a multi-step loading operation
func (ls *LoadingSystem) StartStepLoading(id, message string, steps []string) {
	ls.StartLoading(id, message)

	loader := ls.activeLoaders[id]
	loader.Steps = make([]LoadingStep, len(steps))
	for i, step := range steps {
		loader.Steps[i] = LoadingStep{
			Name:        step,
			Description: step,
		}
	}
}

// UpdateStep updates the current step of a multi-step operation
func (ls *LoadingSystem) UpdateStep(id string, stepIndex int) {
	if loader, exists := ls.activeLoaders[id]; exists {
		loader.CurrentStep = stepIndex
		if stepIndex < len(loader.Steps) {
			loader.Steps[stepIndex].Completed = true
			loader.Message = loader.Steps[stepIndex].Description
		}
	}
}

// UpdateProgress updates the progress of a loading operation
func (ls *LoadingSystem) UpdateProgress(id string, current, total int, message string) {
	if loader, exists := ls.activeLoaders[id]; exists {
		if loader.Progress == nil {
			loader.Progress = &Progress{}
		}
		loader.Progress.Current = current
		loader.Progress.Total = total
		loader.Progress.Message = message
		if message != "" {
			loader.Message = message
		}
	}
}

// CompleteLoading completes a loading operation successfully
func (ls *LoadingSystem) CompleteLoading(id string, successMessage string) {
	if loader, exists := ls.activeLoaders[id]; exists {
		loader.Message = successMessage
		loader.Spinner = spinner.New()
		loader.Spinner.Spinner = spinner.Dot
		loader.Spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(ls.theme.SuccessColor))

		// Auto-remove after a delay
		time.AfterFunc(2*time.Second, func() {
			delete(ls.activeLoaders, id)
			delete(ls.spinners, id)
		})
	}
}

// FailLoading marks a loading operation as failed
func (ls *LoadingSystem) FailLoading(id string, errorMessage string) {
	if loader, exists := ls.activeLoaders[id]; exists {
		loader.Message = errorMessage
		loader.Spinner = spinner.New()
		loader.Spinner.Spinner = spinner.Dot
		loader.Spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(ls.theme.ErrorColor))

		// Keep failed operations visible longer
		time.AfterFunc(5*time.Second, func() {
			delete(ls.activeLoaders, id)
			delete(ls.spinners, id)
		})
	}
}

// CancelLoading cancels a loading operation
func (ls *LoadingSystem) CancelLoading(id string) {
	delete(ls.activeLoaders, id)
	delete(ls.spinners, id)
}

// IsLoading checks if an operation is currently loading
func (ls *LoadingSystem) IsLoading(id string) bool {
	_, exists := ls.activeLoaders[id]
	return exists
}

// GetActiveLoaders returns all active loading operations
func (ls *LoadingSystem) GetActiveLoaders() map[string]*Loader {
	return ls.activeLoaders
}

// Render renders all active loading indicators
func (ls *LoadingSystem) Render(width int) string {
	if len(ls.activeLoaders) == 0 {
		return ""
	}

	var rendered []string
	for _, loader := range ls.activeLoaders {
		rendered = append(rendered, ls.renderLoader(loader, width))
	}

	return strings.Join(rendered, "\n")
}

// Update updates all loading spinners
func (ls *LoadingSystem) Update() {
	for id := range ls.activeLoaders {
		if s, exists := ls.spinners[id]; exists {
			var cmd interface{} // This would be tea.Cmd in real implementation
			ls.spinners[id], _ = s.Update(cmd)
		}
	}
}

// Private methods
func (ls *LoadingSystem) renderLoader(loader *Loader, width int) string {
	var content strings.Builder

	// Add spinner
	spinnerView := loader.Spinner.View()
	content.WriteString(spinnerView)
	content.WriteString(" ")

	// Add message
	content.WriteString(loader.Message)

	// Add progress if available
	if loader.Progress != nil {
		progress := loader.Progress
		if progress.Total > 0 {
			percentage := float64(progress.Current) / float64(progress.Total) * 100
			content.WriteString(fmt.Sprintf(" (%.1f%%)", percentage))

			// Add progress bar
			content.WriteString("\n")
			content.WriteString(ls.renderProgressBar(progress, width-4))
		}
	}

	// Add step information for multi-step operations
	if len(loader.Steps) > 0 {
		content.WriteString("\n")
		for i, step := range loader.Steps {
			var status string
			if step.Completed {
				status = "✅"
			} else if i == loader.CurrentStep {
				status = "⏳"
			} else {
				status = "⏸️"
			}
			content.WriteString(fmt.Sprintf("%s %s\n", status, step.Name))
		}
	}

	// Style the content
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ls.theme.TextColor)).
		Background(lipgloss.Color(ls.theme.BackgroundColor)).
		Padding(0, 1).
		Width(width)

	return style.Render(content.String())
}

func (ls *LoadingSystem) renderProgressBar(progress *Progress, width int) string {
	if progress.Total == 0 {
		return ""
	}

	percentage := float64(progress.Current) / float64(progress.Total)
	filled := int(float64(width) * percentage)

	var bar strings.Builder
	bar.WriteString("╭")
	for i := 0; i < width; i++ {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}
	bar.WriteString("╮")

	return bar.String()
}

// Preset loading operations
func (ls *LoadingSystem) StartConfigSave() {
	steps := []string{
		"Validating configuration",
		"Applying changes",
		"Saving to file",
		"Refreshing display",
	}
	ls.StartStepLoading("config-save", "Saving configuration...", steps)
}

func (ls *LoadingSystem) StartFileLoad() {
	ls.StartLoading("file-load", "Loading configuration file...")
}

func (ls *LoadingSystem) StartValidation() {
	ls.StartLoading("validation", "Validating configuration...")
}

func (ls *LoadingSystem) StartBackup() {
	ls.StartLoading("backup", "Creating backup...")
}

// Utility methods
func (ls *LoadingSystem) GetElapsedTime(id string) time.Duration {
	if loader, exists := ls.activeLoaders[id]; exists {
		return time.Since(loader.StartTime)
	}
	return 0
}

func (ls *LoadingSystem) GetProgress(id string) (current, total int, message string) {
	if loader, exists := ls.activeLoaders[id]; exists && loader.Progress != nil {
		return loader.Progress.Current, loader.Progress.Total, loader.Progress.Message
	}
	return 0, 0, ""
}

// Animation and effects
func (ls *LoadingSystem) AddPulseEffect(id string) {
	if loader, exists := ls.activeLoaders[id]; exists {
		// Add pulsing effect to the spinner
		loader.Spinner.Style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ls.theme.PrimaryColor)).
			Blink(true)
	}
}

func (ls *LoadingSystem) AddColorTransition(id string, fromColor, toColor string) {
	if loader, exists := ls.activeLoaders[id]; exists {
		// Transition spinner color
		loader.Spinner.Style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(toColor))
	}
}

// Accessibility features
func (ls *LoadingSystem) EnableScreenReaderAnnouncements() {
	// Add screen reader support for loading states
	for _, loader := range ls.activeLoaders {
		// In a real implementation, this would announce to screen readers
		_ = loader
	}
}

func (ls *LoadingSystem) GetLoadingAnnouncements() []string {
	var announcements []string
	for _, loader := range ls.activeLoaders {
		message := fmt.Sprintf("Loading: %s", loader.Message)
		if loader.Progress != nil {
			percentage := float64(loader.Progress.Current) / float64(loader.Progress.Total) * 100
			message += fmt.Sprintf(" (%.1f%% complete)", percentage)
		}
		announcements = append(announcements, message)
	}
	return announcements
}

// Performance monitoring
func (ls *LoadingSystem) GetPerformanceStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["active_loaders"] = len(ls.activeLoaders)
	stats["total_operations"] = len(ls.activeLoaders) // This would track historical data

	// Calculate average completion time (simplified)
	totalTime := 0.0
	count := 0
	for _, loader := range ls.activeLoaders {
		if !loader.StartTime.IsZero() {
			totalTime += time.Since(loader.StartTime).Seconds()
			count++
		}
	}

	if count > 0 {
		stats["average_duration"] = totalTime / float64(count)
	}

	return stats
}

// SkeletonLoader provides placeholder loading states for content
type SkeletonLoader struct {
	lines     int
	width     int
	animFrame int
	style     SkeletonStyle
}

// SkeletonStyle defines the visual appearance of skeleton loaders
type SkeletonStyle struct {
	BaseColor      string
	HighlightColor string
	BorderColor    string
	Rounded        bool
}

// DefaultSkeletonStyle provides a default skeleton loading style
var DefaultSkeletonStyle = SkeletonStyle{
	BaseColor:      "#3c3c4c",
	HighlightColor: "#4c4c5c",
	BorderColor:    "#5c5c6c",
	Rounded:        true,
}

// NewSkeletonLoader creates a new skeleton loader
func NewSkeletonLoader(lines, width int) *SkeletonLoader {
	return &SkeletonLoader{
		lines:     lines,
		width:     width,
		animFrame: 0,
		style:     DefaultSkeletonStyle,
	}
}

// NewSkeletonLoaderWithStyle creates a skeleton loader with custom style
func NewSkeletonLoaderWithStyle(lines, width int, style SkeletonStyle) *SkeletonLoader {
	return &SkeletonLoader{
		lines:     lines,
		width:     width,
		animFrame: 0,
		style:     style,
	}
}

// Update advances the shimmer animation
func (sl *SkeletonLoader) Update() {
	sl.animFrame = (sl.animFrame + 1) % 6
}

// View renders the skeleton loader with shimmer effect
func (sl *SkeletonLoader) View() string {
	shimmerChars := []string{"░", "▒", "▓", "▓", "▒", "░"}

	var lines []string
	for i := 0; i < sl.lines; i++ {
		// Vary line widths to simulate different content lengths
		lineWidth := sl.width
		switch i % 4 {
		case 0:
			lineWidth = sl.width
		case 1:
			lineWidth = int(float64(sl.width) * 0.7)
		case 2:
			lineWidth = int(float64(sl.width) * 0.85)
		case 3:
			lineWidth = int(float64(sl.width) * 0.6)
		}

		// Create shimmer effect by offsetting animation frame per line
		offset := (sl.animFrame + i) % len(shimmerChars)
		shimmerChar := shimmerChars[offset]

		line := strings.Repeat(shimmerChar, lineWidth)
		lines = append(lines, line)
	}

	// Apply styling
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(sl.style.HighlightColor)).
		Background(lipgloss.Color(sl.style.BaseColor))

	if sl.style.Rounded {
		style = style.
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(sl.style.BorderColor))
	}

	return style.Render(strings.Join(lines, "\n"))
}

// SetLines updates the number of lines
func (sl *SkeletonLoader) SetLines(lines int) {
	sl.lines = lines
}

// SetWidth updates the width
func (sl *SkeletonLoader) SetWidth(width int) {
	sl.width = width
}

// SetStyle updates the skeleton style
func (sl *SkeletonLoader) SetStyle(style SkeletonStyle) {
	sl.style = style
}

// SkeletonCard renders a card-shaped skeleton placeholder
func SkeletonCard(width, height int) string {
	sl := NewSkeletonLoader(height, width)
	return sl.View()
}

// SkeletonList renders a list of skeleton items
func SkeletonList(itemCount, itemWidth int) string {
	var items []string
	for i := 0; i < itemCount; i++ {
		sl := NewSkeletonLoader(2, itemWidth)
		sl.animFrame = i % 6 // Stagger animation
		items = append(items, sl.View())
	}
	return strings.Join(items, "\n\n")
}

// SkeletonForm renders a form-shaped skeleton placeholder
func SkeletonForm(fieldCount, width int) string {
	var fields []string
	for i := 0; i < fieldCount; i++ {
		// Label skeleton (shorter)
		labelSl := NewSkeletonLoader(1, int(float64(width)*0.3))
		labelSl.animFrame = i % 6

		// Input skeleton (full width)
		inputSl := NewSkeletonLoader(1, width)
		inputSl.animFrame = (i + 2) % 6

		fields = append(fields, labelSl.View()+"\n"+inputSl.View())
	}
	return strings.Join(fields, "\n\n")
}

// SkeletonGrid renders a grid of skeleton items
func SkeletonGrid(rows, cols, itemWidth, itemHeight int) string {
	var gridRows []string
	for r := 0; r < rows; r++ {
		var rowItems []string
		for c := 0; c < cols; c++ {
			sl := NewSkeletonLoader(itemHeight, itemWidth)
			sl.animFrame = (r*cols + c) % 6
			rowItems = append(rowItems, sl.View())
		}
		gridRows = append(gridRows, lipgloss.JoinHorizontal(lipgloss.Top, rowItems...))
	}
	return strings.Join(gridRows, "\n")
}
