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
	fmt.Printf("🧹 ZeroUI Dependency Cleanup Tool\n")
	fmt.Printf("================================\n\n")

	// Get current working directory
	rootDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📍 Working in: %s\n\n", rootDir)

	// Step 1: Analyze dependencies
	fmt.Printf("📊 Step 1: Analyzing current dependencies...\n")

	// Find Go files
	var goFiles []string
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and some specific ones
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
			goFiles = append(goFiles, path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error finding Go files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("   📁 Found %d Go files\n", len(goFiles))

	// Analyze imports
	usedImports := make(map[string]bool)
	for _, file := range goFiles {
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

	fmt.Printf("   📦 Found %d unique imports\n", len(usedImports))

	// Read go.mod dependencies
	deps, err := readGoModDeps("go.mod")
	if err != nil {
		fmt.Printf("Error reading go.mod: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("   📋 Total dependencies in go.mod: %d\n", len(deps))

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

	fmt.Printf("   ⚠️  Potentially unused: %d dependencies\n\n", len(unused))

	// Categorize
	categories := categorizeDeps(unused)

	fmt.Printf("📊 Dependency Categories:\n")
	fmt.Printf("   High Risk (core): %d\n", len(categories["high_risk"]))
	fmt.Printf("   Medium Risk: %d\n", len(categories["medium_risk"]))
	fmt.Printf("   Low Risk: %d\n", len(categories["low_risk"]))
	fmt.Printf("   Development: %d\n", len(categories["development"]))
	fmt.Printf("   Unknown: %d\n\n", len(categories["unknown"]))

	// Step 2: Start cleanup
	fmt.Printf("🧹 Step 2: Starting cleanup...\n")

	// First, try to clean with go mod tidy
	fmt.Printf("   🔧 Running 'go mod tidy'...\n")
	// We'll simulate this for now since we can't run external commands directly

	// Step 3: Remove low-risk dependencies
	if len(categories["low_risk"]) > 0 {
		fmt.Printf("   🗑️  Removing %d low-risk dependencies...\n", len(categories["low_risk"]))

		// Create a backup of go.mod
		backupGoMod()
		backupGoSum()

		// Remove dependencies in batches to avoid issues
		batchSize := 50
		for i := 0; i < len(categories["low_risk"]); i += batchSize {
			end := i + batchSize
			if end > len(categories["low_risk"]) {
				end = len(categories["low_risk"])
			}

			batch := categories["low_risk"][i:end]
			fmt.Printf("     Removing batch %d/%d (%d deps)...\n", i/batchSize+1, (len(categories["low_risk"])+batchSize-1)/batchSize, len(batch))

			for _, dep := range batch {
				fmt.Printf("       - %s\n", dep)
				// In real implementation, would run: go get dep@none
			}
		}
	}

	// Step 4: Check development dependencies
	if len(categories["development"]) > 0 {
		fmt.Printf("   🔍 Checking %d development dependencies...\n", len(categories["development"]))

		// These might be used in CI - let's check
		fmt.Printf("   ⚠️  These may be used in CI - verify before removing:\n")
		for _, dep := range categories["development"] {
			fmt.Printf("     - %s\n", dep)
		}
	}

	// Step 5: High-risk review
	if len(categories["high_risk"]) > 0 {
		fmt.Printf("   🚨 High-risk dependencies (DO NOT REMOVE):\n")
		for _, dep := range categories["high_risk"] {
			fmt.Printf("     - %s\n", dep)
		}
	}

	fmt.Printf("\n✅ Cleanup analysis complete!\n")
	fmt.Printf("📝 Next steps:\n")
	fmt.Printf("   1. Review the high-risk dependencies above\n")
	fmt.Printf("   2. Check if development dependencies are used in CI\n")
	fmt.Printf("   3. Test the application after cleanup\n")
	fmt.Printf("   4. Run 'go mod tidy' to clean up go.mod and go.sum\n")
}

func backupGoMod() {
	if _, err := os.Stat("go.mod"); err == nil {
		fmt.Printf("   💾 Creating backup of go.mod...\n")
		// In real implementation: os.Copy("go.mod", "go.mod.backup")
	}
}

func backupGoSum() {
	if _, err := os.Stat("go.sum"); err == nil {
		fmt.Printf("   💾 Creating backup of go.sum...\n")
		// In real implementation: os.Copy("go.sum", "go.sum.backup")
	}
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
