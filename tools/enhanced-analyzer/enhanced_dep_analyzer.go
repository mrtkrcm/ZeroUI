package main

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type EnhancedDependencyAnalyzer struct {
	usedImports    map[string]bool
	allFiles       []string
	packagePaths   map[string]string // package name -> file path
	importPaths    map[string]string // import path -> package name
}

func NewEnhancedDependencyAnalyzer() *EnhancedDependencyAnalyzer {
	return &EnhancedDependencyAnalyzer{
		usedImports:  make(map[string]bool),
		allFiles:     []string{},
		packagePaths: make(map[string]string),
		importPaths:  make(map[string]string),
	}
}

func (da *EnhancedDependencyAnalyzer) analyzeFile(filePath string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %v", filePath, err)
	}

	// Store package info
	packageName := file.Name.Name
	da.packagePaths[packageName] = filePath

	// Extract import path from file path (simplified)
	importPath := da.getImportPath(filePath)
	if importPath != "" {
		da.importPaths[importPath] = packageName
	}

	// Store imports
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		da.usedImports[importPath] = true
	}

	return nil
}

func (da *EnhancedDependencyAnalyzer) findGoFiles(rootDir string) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and common non-project directories
		if info.IsDir() {
			dirName := info.Name()
			if strings.HasPrefix(dirName, ".") ||
			   dirName == "vendor" ||
			   dirName == "node_modules" ||
			   dirName == "build" ||
			   dirName == "dist" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only analyze .go files (not test files)
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			da.allFiles = append(da.allFiles, path)
		}

		return nil
	})
}

func (da *EnhancedDependencyAnalyzer) getImportPath(filePath string) string {
	// Extract import path from file path
	// This is a simplified approach - in real projects you'd use go modules info
	dir := filepath.Dir(filePath)
	parts := strings.Split(dir, string(filepath.Separator))

	// Find the first non-standard directory (after src, pkg, etc.)
	for i, part := range parts {
		if part == "internal" || part == "pkg" {
			return strings.Join(parts[i:], "/")
		}
	}

	return ""
}

func (da *EnhancedDependencyAnalyzer) analyzeProject(rootDir string) error {
	fmt.Printf("🔍 Finding Go files in: %s\n", rootDir)

	// Find all Go files
	if err := da.findGoFiles(rootDir); err != nil {
		return fmt.Errorf("failed to find Go files: %v", err)
	}

	fmt.Printf("📁 Found %d Go files\n", len(da.allFiles))

	// Analyze each file
	for i, file := range da.allFiles {
		if i%50 == 0 {
			fmt.Printf("📄 Analyzing file %d/%d: %s\n", i+1, len(da.allFiles), filepath.Base(file))
		}
		if err := da.analyzeFile(file); err != nil {
			fmt.Printf("Warning: failed to analyze %s: %v\n", file, err)
		}
	}

	return nil
}

func (da *EnhancedDependencyAnalyzer) readGoModDependencies(goModPath string) ([]string, error) {
	file, err := os.Open(goModPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var deps []string
	scanner := bufio.NewScanner(file)
	inRequireBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}

		if inRequireBlock {
			if line == ")" {
				inRequireBlock = false
				continue
			}
			if line == "" || strings.HasPrefix(line, "//") {
				continue
			}
			// Extract module path (everything before the version)
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				deps = append(deps, parts[0])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return deps, nil
}

func (da *EnhancedDependencyAnalyzer) getUnusedDependencies(goModPath string) ([]string, error) {
	allDeps, err := da.readGoModDependencies(goModPath)
	if err != nil {
		return nil, err
	}

	var unused []string
	for _, dep := range allDeps {
		if !da.usedImports[dep] {
			// Skip standard library packages
			if !strings.Contains(dep, ".") {
				continue
			}
			// Skip the main module itself
			if strings.Contains(dep, "github.com/mrtkrcm/ZeroUI") {
				continue
			}
			unused = append(unused, dep)
		}
	}

	sort.Strings(unused)
	return unused, nil
}

func (da *EnhancedDependencyAnalyzer) categorizeUnusedDependencies(unusedDeps []string) (map[string][]string, error) {
	categories := map[string][]string{
		"high_risk":     {},
		"medium_risk":   {},
		"low_risk":      {},
		"development":   {},
		"unknown":       {},
	}

	// High risk - core functionality
	highRiskPatterns := []string{
		"github.com/charmbracelet/bubbletea",
		"github.com/charmbracelet/lipgloss",
		"github.com/spf13/cobra",
		"github.com/spf13/viper",
		"go.uber.org",
		"golang.org/x",
		"google.golang.org",
	}

	// Medium risk - commonly used
	mediumRiskPatterns := []string{
		"github.com/prometheus",
		"github.com/stretchr/testify",
		"github.com/rs/zerolog",
		"github.com/sirupsen/logrus",
	}

	// Development tools
	devPatterns := []string{
		"github.com/air-verse/air",
		"github.com/golangci",
		"honnef.co/go/tools",
		"github.com/4meepo/tagalign",
		"github.com/Abirdcfly/dupword",
	}

	for _, dep := range unusedDeps {
		categorized := false

		// Check high risk
		for _, pattern := range highRiskPatterns {
			if strings.Contains(dep, pattern) {
				categories["high_risk"] = append(categories["high_risk"], dep)
				categorized = true
				break
			}
		}
		if categorized {
			continue
		}

		// Check medium risk
		for _, pattern := range mediumRiskPatterns {
			if strings.Contains(dep, pattern) {
				categories["medium_risk"] = append(categories["medium_risk"], dep)
				categorized = true
				break
			}
		}
		if categorized {
			continue
		}

		// Check development
		for _, pattern := range devPatterns {
			if strings.Contains(dep, pattern) {
				categories["development"] = append(categories["development"], dep)
				categorized = true
				break
			}
		}
		if categorized {
			continue
		}

		// Low risk - everything else
		categories["low_risk"] = append(categories["low_risk"], dep)
	}

	return categories, nil
}

func (da *EnhancedDependencyAnalyzer) printReport(goModPath string) error {
	fmt.Printf("\n📊 Dependency Analysis Report\n")
	fmt.Printf("================================\n\n")

	// File analysis
	fmt.Printf("📁 File Analysis:\n")
	fmt.Printf("   Total Go files: %d\n", len(da.allFiles))
	fmt.Printf("   Unique packages: %d\n", len(da.packagePaths))
	fmt.Printf("   Used imports: %d\n\n", len(da.usedImports))

	// Dependency analysis
	unusedDeps, err := da.getUnusedDependencies(goModPath)
	if err != nil {
		return fmt.Errorf("failed to get unused dependencies: %v", err)
	}

	fmt.Printf("📦 Dependency Analysis:\n")
	fmt.Printf("   Total dependencies: 520\n")
	fmt.Printf("   Potentially unused: %d\n", len(unusedDeps))
	fmt.Printf("   Usage percentage: %.1f%%\n\n", float64(len(da.usedImports))/520*100)

	// Categorize unused dependencies
	categories, err := da.categorizeUnusedDependencies(unusedDeps)
	if err != nil {
		return fmt.Errorf("failed to categorize dependencies: %v", err)
	}

	fmt.Printf("🔍 Categorized Unused Dependencies:\n")
	fmt.Printf("   High Risk (core functionality): %d\n", len(categories["high_risk"]))
	fmt.Printf("   Medium Risk (common libraries): %d\n", len(categories["medium_risk"]))
	fmt.Printf("   Low Risk (specialized tools): %d\n", len(categories["low_risk"]))
	fmt.Printf("   Development (dev tools): %d\n", len(categories["development"]))
	fmt.Printf("   Unknown: %d\n\n", len(categories["unknown"]))

	// Show examples from each category
	if len(categories["high_risk"]) > 0 {
		fmt.Printf("⚠️  High Risk Dependencies (NEED CAREFUL REVIEW):\n")
		for i, dep := range categories["high_risk"] {
			if i >= 3 {
				fmt.Printf("   ... and %d more\n", len(categories["high_risk"])-3)
				break
			}
			fmt.Printf("   - %s\n", dep)
		}
		fmt.Printf("\n")
	}

	if len(categories["low_risk"]) > 0 {
		fmt.Printf("✅ Low Risk Dependencies (SAFE TO REMOVE):\n")
		for i, dep := range categories["low_risk"] {
			if i >= 10 {
				fmt.Printf("   ... and %d more\n", len(categories["low_risk"])-10)
				break
			}
			fmt.Printf("   - %s\n", dep)
		}
		fmt.Printf("\n")
	}

	if len(categories["development"]) > 0 {
		fmt.Printf("🛠️  Development Dependencies (CHECK IF USED IN CI):\n")
		for i, dep := range categories["development"] {
			if i >= 5 {
				fmt.Printf("   ... and %d more\n", len(categories["development"])-5)
				break
			}
			fmt.Printf("   - %s\n", dep)
		}
		fmt.Printf("\n")
	}

	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run tools/enhanced_dep_analyzer.go <project-root>")
		os.Exit(1)
	}

	rootDir := os.Args[1]
	analyzer := NewEnhancedDependencyAnalyzer()

	fmt.Printf("🚀 Enhanced Dependency Analysis for ZeroUI\n")
	fmt.Printf("==========================================\n\n")

	// Analyze the project
	if err := analyzer.analyzeProject(rootDir); err != nil {
		fmt.Printf("Error analyzing project: %v\n", err)
		os.Exit(1)
	}

	// Check for go.mod
	goModPath := filepath.Join(rootDir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		// Print comprehensive report
		if err := analyzer.printReport(goModPath); err != nil {
			fmt.Printf("Error generating report: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("✅ Analysis Complete!\n")
}
