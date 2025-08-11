package tui

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

// AutomationFramework provides comprehensive TUI testing automation
type AutomationFramework struct {
	engine         *toggle.Engine
	testSuites     map[string]*TestSuite
	results        *TestResults
	config         *AutomationConfig
	logger         *log.Logger
	watchMode      bool
	reportPath     string
	mu             sync.RWMutex
}

// AutomationConfig configures the testing framework
type AutomationConfig struct {
	TestTimeout       time.Duration
	SnapshotThreshold float64 // Similarity threshold for visual comparisons
	MaxConcurrent     int
	RetryAttempts     int
	EnableProfiling   bool
	WatchFiles        []string
	OutputFormats     []OutputFormat
}

type OutputFormat string

const (
	OutputJSON OutputFormat = "json"
	OutputHTML OutputFormat = "html"
	OutputText OutputFormat = "text"
)

// TestScenario defines a complete UI test scenario
type TestScenario struct {
	Name        string
	Description string
	Width       int
	Height      int
	Setup       func(*Model) error
	Interactions []Interaction
	Validations []Validation
	Tags        []string // for filtering tests
}

// Interaction represents a user interaction
type Interaction struct {
	Type        InteractionType
	Key         tea.KeyType
	Runes       []rune
	Delay       time.Duration
	Description string
}

type InteractionType int

const (
	KeyPress InteractionType = iota
	WindowResize
	MouseClick
	Wait
)

// Validation represents a rendering validation
type Validation struct {
	Type        ValidationType
	Pattern     string
	Count       int
	Position    Position
	Description string
}

type ValidationType int

const (
	Contains ValidationType = iota
	NotContains
	LineCount
	WidthCheck
	HeightCheck
	RegexMatch
	VisualStructure
)

type Position struct {
	Line   int
	Column int
}

// TestSuite contains multiple related test cases
type TestSuite struct {
	Name        string
	Description string
	Tests       []*AutomationTest
	Setup       func() error
	Teardown    func() error
	Tags        []string
	Parallel    bool
}

// AutomationTest represents a single automated test
type AutomationTest struct {
	ID          string
	Name        string
	Description string
	Scenario    TestScenario
	Expected    *ExpectedResult
	Timeout     time.Duration
	Retries     int
	Critical    bool // If true, suite fails when this test fails
}

// ExpectedResult defines what we expect from a test
type ExpectedResult struct {
	VisualHash      string
	ContainsText    []string
	NotContainsText []string
	MinRenderTime   time.Duration
	MaxRenderTime   time.Duration
	MemoryUsage     int64 // Max memory usage in bytes
	StructureRules  []StructureRule
}

// StructureRule validates UI structure
type StructureRule struct {
	Type        StructureType
	Pattern     string
	MinOccurs   int
	MaxOccurs   int
	LineRange   LineRange
	Description string
}

type StructureType string

const (
	BoxCharacters   StructureType = "box_chars"
	TextAlignment   StructureType = "text_align"
	ColorCodes      StructureType = "color_codes"
	WidthConsistent StructureType = "width_consistent"
	HeightConsistent StructureType = "height_consistent"
)

type LineRange struct {
	Start int
	End   int
}

// TestResults tracks all test execution results
type TestResults struct {
	StartTime      time.Time
	EndTime        time.Time
	TotalTests     int
	PassedTests    int
	FailedTests    int
	SkippedTests   int
	SuiteResults   map[string]*SuiteResult
	OverallHealth  float64 // 0-100 score
	mu             sync.RWMutex
}

// SuiteResult tracks results for a test suite
type SuiteResult struct {
	Suite        *TestSuite
	StartTime    time.Time
	EndTime      time.Time
	TestResults  []*TestResult
	Passed       bool
	ErrorMessage string
}

// TestResult tracks individual test results
type TestResult struct {
	Test           *AutomationTest
	Passed         bool
	ExecutionTime  time.Duration
	MemoryUsed     int64
	ErrorMessage   string
	VisualSnapshot string
	DiffOutput     string
	Metrics        *TestMetrics
}

// TestMetrics provides detailed test metrics
type TestMetrics struct {
	RenderTime    time.Duration
	UpdateCalls   int
	ViewCalls     int
	MemoryPeak    int64
	CPUUsage      float64
	GCPauses      time.Duration
}

// NewAutomationFramework creates a new TUI testing framework
func NewAutomationFramework(engine *toggle.Engine) *AutomationFramework {
	return &AutomationFramework{
		engine:     engine,
		testSuites: make(map[string]*TestSuite),
		results: &TestResults{
			SuiteResults: make(map[string]*SuiteResult),
		},
		config: &AutomationConfig{
			TestTimeout:       30 * time.Second,
			SnapshotThreshold: 0.95,
			MaxConcurrent:     4,
			RetryAttempts:     3,
			EnableProfiling:   true,
			OutputFormats:     []OutputFormat{OutputJSON, OutputHTML},
		},
		logger: log.New(os.Stdout, "[TUI-Automation] ", log.LstdFlags|log.Lshortfile),
	}
}

// AddTestSuite registers a new test suite
func (af *AutomationFramework) AddTestSuite(suite *TestSuite) {
	af.mu.Lock()
	defer af.mu.Unlock()
	
	af.testSuites[suite.Name] = suite
	af.logger.Printf("Registered test suite: %s with %d tests", suite.Name, len(suite.Tests))
}

// RunAllTests executes all registered test suites
func (af *AutomationFramework) RunAllTests(ctx context.Context) error {
	af.results.StartTime = time.Now()
	
	af.logger.Println("Starting comprehensive TUI testing...")
	
	// Run suites
	for suiteName, suite := range af.testSuites {
		af.logger.Printf("Running suite: %s", suiteName)
		
		result := af.runSuite(ctx, suite)
		af.results.SuiteResults[suiteName] = result
		
		if result.Passed {
			af.results.PassedTests += len(result.TestResults)
		} else {
			af.results.FailedTests += len(result.TestResults)
		}
		af.results.TotalTests += len(result.TestResults)
	}
	
	af.results.EndTime = time.Now()
	af.calculateOverallHealth()
	
	// Generate reports
	return af.generateReports()
}

// runSuite executes a single test suite
func (af *AutomationFramework) runSuite(ctx context.Context, suite *TestSuite) *SuiteResult {
	result := &SuiteResult{
		Suite:       suite,
		StartTime:   time.Now(),
		TestResults: make([]*TestResult, 0, len(suite.Tests)),
		Passed:      true,
	}
	
	// Run setup
	if suite.Setup != nil {
		if err := suite.Setup(); err != nil {
			result.Passed = false
			result.ErrorMessage = fmt.Sprintf("Setup failed: %v", err)
			return result
		}
	}
	
	// Run tests
	if suite.Parallel && af.config.MaxConcurrent > 1 {
		result.TestResults = af.runTestsParallel(ctx, suite.Tests)
	} else {
		result.TestResults = af.runTestsSequential(ctx, suite.Tests)
	}
	
	// Check if any critical test failed
	for _, testResult := range result.TestResults {
		if testResult.Test.Critical && !testResult.Passed {
			result.Passed = false
			if result.ErrorMessage == "" {
				result.ErrorMessage = fmt.Sprintf("Critical test failed: %s", testResult.Test.Name)
			}
		}
	}
	
	// Run teardown
	if suite.Teardown != nil {
		if err := suite.Teardown(); err != nil {
			af.logger.Printf("Teardown warning for suite %s: %v", suite.Name, err)
		}
	}
	
	result.EndTime = time.Now()
	return result
}

// runTestsSequential runs tests one by one
func (af *AutomationFramework) runTestsSequential(ctx context.Context, tests []*AutomationTest) []*TestResult {
	results := make([]*TestResult, 0, len(tests))
	
	for _, test := range tests {
		result := af.runSingleTest(ctx, test)
		results = append(results, result)
		
		// Log progress
		status := "PASS"
		if !result.Passed {
			status = "FAIL"
		}
		af.logger.Printf("Test %s: %s (%v)", test.Name, status, result.ExecutionTime)
	}
	
	return results
}

// runTestsParallel runs tests concurrently
func (af *AutomationFramework) runTestsParallel(ctx context.Context, tests []*AutomationTest) []*TestResult {
	resultsChan := make(chan *TestResult, len(tests))
	semaphore := make(chan struct{}, af.config.MaxConcurrent)
	
	var wg sync.WaitGroup
	
	// Start tests
	for _, test := range tests {
		wg.Add(1)
		go func(t *AutomationTest) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			result := af.runSingleTest(ctx, t)
			resultsChan <- result
		}(test)
	}
	
	// Wait for completion
	go func() {
		wg.Wait()
		close(resultsChan)
	}()
	
	// Collect results
	var results []*TestResult
	for result := range resultsChan {
		results = append(results, result)
		
		status := "PASS"
		if !result.Passed {
			status = "FAIL"
		}
		af.logger.Printf("Test %s: %s (%v)", result.Test.Name, status, result.ExecutionTime)
	}
	
	return results
}

// runSingleTest executes an individual test with retries
func (af *AutomationFramework) runSingleTest(ctx context.Context, test *AutomationTest) *TestResult {
	var result *TestResult
	
	for attempt := 0; attempt <= test.Retries; attempt++ {
		if attempt > 0 {
			af.logger.Printf("Retrying test %s (attempt %d/%d)", test.Name, attempt+1, test.Retries+1)
		}
		
		result = af.executeTest(ctx, test)
		
		if result.Passed {
			break
		}
	}
	
	return result
}

// executeTest runs the actual test
func (af *AutomationFramework) executeTest(ctx context.Context, test *AutomationTest) *TestResult {
	startTime := time.Now()
	
	result := &TestResult{
		Test:    test,
		Passed:  false,
		Metrics: &TestMetrics{},
	}
	
	// Create timeout context
	testCtx, cancel := context.WithTimeout(ctx, test.Timeout)
	defer cancel()
	
	// Create model
	model, err := NewModel(af.engine, "")
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to create model: %v", err)
		result.ExecutionTime = time.Since(startTime)
		return result
	}
	
	// Set up scenario
	model.width = test.Scenario.Width
	model.height = test.Scenario.Height
	model.updateComponentSizes()
	
	if test.Scenario.Setup != nil {
		if err := test.Scenario.Setup(model); err != nil {
			result.ErrorMessage = fmt.Sprintf("Setup failed: %v", err)
			result.ExecutionTime = time.Since(startTime)
			return result
		}
	}
	
	// Execute interactions
	renderStartTime := time.Now()
	for _, interaction := range test.Scenario.Interactions {
		select {
		case <-testCtx.Done():
			result.ErrorMessage = "Test timeout"
			result.ExecutionTime = time.Since(startTime)
			return result
		default:
		}
		
		switch interaction.Type {
		case KeyPress:
			var keyMsg tea.KeyMsg
			if interaction.Key != 0 {
				keyMsg = tea.KeyMsg{Type: interaction.Key}
			} else if len(interaction.Runes) > 0 {
				keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: interaction.Runes}
			}
			
			updatedModel, _ := model.Update(keyMsg)
			model = updatedModel.(*Model)
			result.Metrics.UpdateCalls++
			
		case WindowResize:
			resizeMsg := tea.WindowSizeMsg{Width: test.Scenario.Width, Height: test.Scenario.Height}
			updatedModel, _ := model.Update(resizeMsg)
			model = updatedModel.(*Model)
			result.Metrics.UpdateCalls++
			
		case Wait:
			time.Sleep(interaction.Delay)
		}
	}
	
	// Capture final view
	finalView := model.View()
	result.VisualSnapshot = finalView
	result.Metrics.RenderTime = time.Since(renderStartTime)
	result.Metrics.ViewCalls++
	
	// Run validations
	if af.validateExpectedResult(test.Expected, finalView, result) {
		result.Passed = true
	}
	
	result.ExecutionTime = time.Since(startTime)
	return result
}

// validateExpectedResult checks if the test output matches expectations
func (af *AutomationFramework) validateExpectedResult(expected *ExpectedResult, actualView string, result *TestResult) bool {
	if expected == nil {
		return true // No expectations means success
	}
	
	var errors []string
	
	// Check visual hash
	if expected.VisualHash != "" {
		actualHash := af.hashSnapshot(actualView)
		if actualHash != expected.VisualHash {
			errors = append(errors, "Visual hash mismatch")
		}
	}
	
	// Check text contains
	for _, text := range expected.ContainsText {
		if !strings.Contains(actualView, text) {
			errors = append(errors, fmt.Sprintf("Missing expected text: %s", text))
		}
	}
	
	// Check text not contains
	for _, text := range expected.NotContainsText {
		if strings.Contains(actualView, text) {
			errors = append(errors, fmt.Sprintf("Found unexpected text: %s", text))
		}
	}
	
	// Check render time
	if expected.MinRenderTime > 0 && result.Metrics.RenderTime < expected.MinRenderTime {
		errors = append(errors, "Render time too fast")
	}
	if expected.MaxRenderTime > 0 && result.Metrics.RenderTime > expected.MaxRenderTime {
		errors = append(errors, "Render time too slow")
	}
	
	// Check structure rules
	for _, rule := range expected.StructureRules {
		if !af.validateStructureRule(rule, actualView) {
			errors = append(errors, fmt.Sprintf("Structure rule failed: %s", rule.Description))
		}
	}
	
	if len(errors) > 0 {
		result.ErrorMessage = fmt.Sprintf("Validation errors: %v", errors)
		return false
	}
	
	return true
}

// validateStructureRule checks a single structure rule
func (af *AutomationFramework) validateStructureRule(rule StructureRule, view string) bool {
	lines := strings.Split(stripAnsiCodes(view), "\n")
	
	startLine := 0
	endLine := len(lines) - 1
	
	if rule.LineRange.Start > 0 {
		startLine = rule.LineRange.Start - 1
	}
	if rule.LineRange.End > 0 && rule.LineRange.End < len(lines) {
		endLine = rule.LineRange.End - 1
	}
	
	occurrences := 0
	
	for i := startLine; i <= endLine && i < len(lines); i++ {
		line := lines[i]
		
		switch rule.Type {
		case BoxCharacters:
			if strings.ContainsAny(line, "╭╮╯╰│─║═") {
				occurrences++
			}
		case TextAlignment:
			// Check if text matches alignment pattern
			if strings.Contains(line, rule.Pattern) {
				occurrences++
			}
		case ColorCodes:
			if strings.Contains(view, "\x1b[") { // Check original view with ANSI codes
				occurrences++
			}
		case WidthConsistent:
			// All lines should be within expected width
			if len(line) <= af.parseIntFromPattern(rule.Pattern) {
				occurrences++
			}
		case HeightConsistent:
			// Total lines should match expectation
			if len(lines) <= af.parseIntFromPattern(rule.Pattern) {
				occurrences++
			}
		}
	}
	
	return occurrences >= rule.MinOccurs && 
		   (rule.MaxOccurs == 0 || occurrences <= rule.MaxOccurs)
}

// parseIntFromPattern extracts integer from pattern string
func (af *AutomationFramework) parseIntFromPattern(pattern string) int {
	// Simple implementation - extend as needed
	var result int
	fmt.Sscanf(pattern, "%d", &result)
	return result
}

// hashSnapshot creates hash for visual comparison
func (af *AutomationFramework) hashSnapshot(snapshot string) string {
	// Implement same logic as in automated_rendering_test.go
	// This is a placeholder - use the actual implementation
	return fmt.Sprintf("%x", len(snapshot)) // Simplified
}

// calculateOverallHealth computes the overall test health score
func (af *AutomationFramework) calculateOverallHealth() {
	if af.results.TotalTests == 0 {
		af.results.OverallHealth = 0
		return
	}
	
	af.results.OverallHealth = float64(af.results.PassedTests) / float64(af.results.TotalTests) * 100
}

// generateReports creates test reports in various formats
func (af *AutomationFramework) generateReports() error {
	timestamp := time.Now().Format("20060102-150405")
	reportDir := fmt.Sprintf("testdata/reports/%s", timestamp)
	
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return err
	}
	
	for _, format := range af.config.OutputFormats {
		var filename string
		var err error
		
		switch format {
		case OutputJSON:
			filename = filepath.Join(reportDir, "report.json")
			err = af.generateJSONReport(filename)
		case OutputHTML:
			filename = filepath.Join(reportDir, "report.html")
			err = af.generateHTMLReport(filename)
		case OutputText:
			filename = filepath.Join(reportDir, "report.txt")
			err = af.generateTextReport(filename)
		}
		
		if err != nil {
			af.logger.Printf("Failed to generate %s report: %v", format, err)
		} else {
			af.logger.Printf("Generated %s report: %s", format, filename)
		}
	}
	
	return nil
}

// generateJSONReport creates JSON test report
func (af *AutomationFramework) generateJSONReport(filename string) error {
	// Implementation would serialize af.results to JSON
	// This is a placeholder
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	fmt.Fprintf(file, "{\n")
	fmt.Fprintf(file, "  \"start_time\": \"%s\",\n", af.results.StartTime.Format(time.RFC3339))
	fmt.Fprintf(file, "  \"end_time\": \"%s\",\n", af.results.EndTime.Format(time.RFC3339))
	fmt.Fprintf(file, "  \"total_tests\": %d,\n", af.results.TotalTests)
	fmt.Fprintf(file, "  \"passed_tests\": %d,\n", af.results.PassedTests)
	fmt.Fprintf(file, "  \"failed_tests\": %d,\n", af.results.FailedTests)
	fmt.Fprintf(file, "  \"overall_health\": %.2f\n", af.results.OverallHealth)
	fmt.Fprintf(file, "}\n")
	
	return nil
}

// generateHTMLReport creates HTML test report
func (af *AutomationFramework) generateHTMLReport(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	
	// Write HTML header
	fmt.Fprintln(writer, "<!DOCTYPE html>")
	fmt.Fprintln(writer, "<html><head><title>TUI Test Report</title>")
	fmt.Fprintln(writer, "<style>")
	fmt.Fprintln(writer, "body { font-family: monospace; margin: 20px; }")
	fmt.Fprintln(writer, ".passed { color: green; }")
	fmt.Fprintln(writer, ".failed { color: red; }")
	fmt.Fprintln(writer, ".suite { border: 1px solid #ccc; margin: 10px 0; padding: 10px; }")
	fmt.Fprintln(writer, "</style></head><body>")
	
	// Write summary
	fmt.Fprintf(writer, "<h1>TUI Test Report</h1>\n")
	fmt.Fprintf(writer, "<h2>Summary</h2>\n")
	fmt.Fprintf(writer, "<p>Total Tests: %d</p>\n", af.results.TotalTests)
	fmt.Fprintf(writer, "<p>Passed: <span class=\"passed\">%d</span></p>\n", af.results.PassedTests)
	fmt.Fprintf(writer, "<p>Failed: <span class=\"failed\">%d</span></p>\n", af.results.FailedTests)
	fmt.Fprintf(writer, "<p>Health Score: %.2f%%</p>\n", af.results.OverallHealth)
	
	// Write suite details
	fmt.Fprintln(writer, "<h2>Test Suites</h2>")
	for suiteName, suiteResult := range af.results.SuiteResults {
		statusClass := "passed"
		if !suiteResult.Passed {
			statusClass = "failed"
		}
		
		fmt.Fprintf(writer, "<div class=\"suite\">\n")
		fmt.Fprintf(writer, "<h3 class=\"%s\">%s</h3>\n", statusClass, suiteName)
		fmt.Fprintf(writer, "<p>Duration: %v</p>\n", suiteResult.EndTime.Sub(suiteResult.StartTime))
		fmt.Fprintf(writer, "</div>\n")
	}
	
	fmt.Fprintln(writer, "</body></html>")
	return nil
}

// generateTextReport creates plain text test report
func (af *AutomationFramework) generateTextReport(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	fmt.Fprintln(file, "TUI Test Report")
	fmt.Fprintln(file, "===============")
	fmt.Fprintf(file, "Total Tests: %d\n", af.results.TotalTests)
	fmt.Fprintf(file, "Passed: %d\n", af.results.PassedTests)
	fmt.Fprintf(file, "Failed: %d\n", af.results.FailedTests)
	fmt.Fprintf(file, "Health Score: %.2f%%\n", af.results.OverallHealth)
	fmt.Fprintf(file, "Duration: %v\n", af.results.EndTime.Sub(af.results.StartTime))
	
	return nil
}

// EnableWatchMode enables continuous testing when files change
func (af *AutomationFramework) EnableWatchMode(watchPaths []string) {
	af.watchMode = true
	af.config.WatchFiles = watchPaths
	// Implementation would use fsnotify to watch for file changes
	// and automatically re-run tests
}

// stripAnsiCodes removes ANSI escape sequences from text
func stripAnsiCodes(str string) string {
	var result strings.Builder
	inEscape := false
	
	for _, ch := range str {
		if ch == '\x1b' {
			inEscape = true
		} else if inEscape {
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inEscape = false
			}
		} else {
			result.WriteRune(ch)
		}
	}
	
	return result.String()
}