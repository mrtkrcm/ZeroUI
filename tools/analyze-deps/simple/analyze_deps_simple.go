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

func main() {
	rootDir := "."

	fmt.Printf("🔍 Analyzing dependencies in: %s\n", rootDir)

	// Find all Go files
	var goFiles []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and tools
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") ||
			   info.Name() == "tools" ||
			   info.Name() == "vendor" ||
			   info.Name() == "node_modules" ||
			   info.Name() == "build" ||
			   info.Name() == "dist" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only analyze .go files (not test files)
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			goFiles = append(goFiles, path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error finding Go files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📁 Found %d Go files\n", len(goFiles))

	// Analyze imports
	usedImports := make(map[string]bool)

	for i, file := range goFiles {
		if i%50 == 0 {
			fmt.Printf("📄 Analyzing file %d/%d: %s\n", i+1, len(goFiles), filepath.Base(file))
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			usedImports[importPath] = true
		}
	}

	fmt.Printf("📦 Found %d unique imports\n", len(usedImports))

	// Read go.mod dependencies
	goModPath := "go.mod"
	deps, err := readGoModDeps(goModPath)
	if err != nil {
		fmt.Printf("Error reading go.mod: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📋 Total dependencies in go.mod: %d\n", len(deps))

	// Find unused dependencies
	var unused []string
	for _, dep := range deps {
		if !usedImports[dep] {
			// Skip standard library and self
			if !strings.Contains(dep, ".") || strings.Contains(dep, "github.com/mrtkrcm/ZeroUI") {
				continue
			}
			unused = append(unused, dep)
		}
	}

	sort.Strings(unused)

	fmt.Printf("⚠️  Potentially unused dependencies: %d\n\n", len(unused))

	// Categorize
	categories := categorizeDeps(unused)

	fmt.Printf("📊 Categorized Results:\n")
	fmt.Printf("   High Risk (core): %d\n", len(categories["high_risk"]))
	fmt.Printf("   Medium Risk: %d\n", len(categories["medium_risk"]))
	fmt.Printf("   Low Risk: %d\n", len(categories["low_risk"]))
	fmt.Printf("   Development: %d\n", len(categories["development"]))
	fmt.Printf("   Unknown: %d\n\n", len(categories["unknown"]))

	// Show low risk ones (safe to remove)
	if len(categories["low_risk"]) > 0 {
		fmt.Printf("✅ SAFE TO REMOVE (Low Risk):\n")
		for i, dep := range categories["low_risk"] {
			if i >= 20 {
				fmt.Printf("   ... and %d more\n", len(categories["low_risk"])-20)
				break
			}
			fmt.Printf("   - %s\n", dep)
		}
		fmt.Printf("\n")
	}

	// Show development ones
	if len(categories["development"]) > 0 {
		fmt.Printf("🛠️  DEVELOPMENT TOOLS (Check CI):\n")
		for i, dep := range categories["development"] {
			if i >= 10 {
				fmt.Printf("   ... and %d more\n", len(categories["development"])-10)
				break
			}
			fmt.Printf("   - %s\n", dep)
		}
		fmt.Printf("\n")
	}

	fmt.Printf("✅ Analysis Complete!\n")
}

func readGoModDeps(goModPath string) ([]string, error) {
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
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				deps = append(deps, parts[0])
			}
		}
	}

	return deps, scanner.Err()
}

func categorizeDeps(deps []string) map[string][]string {
	categories := map[string][]string{
		"high_risk":   {},
		"medium_risk": {},
		"low_risk":    {},
		"development": {},
		"unknown":     {},
	}

	highRisk := []string{
		"github.com/charmbracelet/bubbletea",
		"github.com/charmbracelet/lipgloss",
		"github.com/spf13/cobra",
		"github.com/spf13/viper",
		"go.uber.org",
		"golang.org/x",
		"google.golang.org",
	}

	mediumRisk := []string{
		"github.com/prometheus",
		"github.com/stretchr/testify",
		"github.com/rs/zerolog",
		"github.com/sirupsen/logrus",
	}

	devTools := []string{
		"github.com/air-verse/air",
		"github.com/golangci",
		"honnef.co/go/tools",
		"github.com/4meepo/tagalign",
		"github.com/Abirdcfly/dupword",
	}

	for _, dep := range deps {
		categorized := false

		// Check high risk
		for _, pattern := range highRisk {
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
		for _, pattern := range mediumRisk {
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
		for _, pattern := range devTools {
			if strings.Contains(dep, pattern) {
				categories["development"] = append(categories["development"], dep)
				categorized = true
				break
			}
		}
		if categorized {
			continue
		}

		// Low risk
		categories["low_risk"] = append(categories["low_risk"], dep)
	}

	return categories
}
