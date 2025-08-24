package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	fmt.Printf("🔍 Analyzing Actual Dependency Usage\n")
	fmt.Printf("===================================\n\n")

	// Dependencies to check (medium risk + development)
	depsToCheck := []string{
		// Medium risk
		"github.com/prometheus",
		"github.com/stretchr/testify",
		"github.com/rs/zerolog",
		"github.com/sirupsen/logrus",

		// Development tools
		"github.com/air-verse/air",
		"github.com/golangci",
		"honnef.co/go/tools",
		"github.com/4meepo/tagalign",
		"github.com/Abirdcfly/dupword",
	}

	// Find all Go files
	var goFiles []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

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

		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			goFiles = append(goFiles, path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error finding Go files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📁 Found %d Go files\n\n", len(goFiles))

	// Analyze imports
	usedImports := make(map[string]bool)
	for _, file := range goFiles {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			usedImports[importPath] = true
		}
	}

	fmt.Printf("📊 Dependency Usage Analysis:\n")
	fmt.Printf("   Total unique imports: %d\n\n", len(usedImports))

	// Check each dependency
	actuallyUsed := []string{}
	notUsed := []string{}

	for _, dep := range depsToCheck {
		found := false
		for usedImport := range usedImports {
			if strings.Contains(usedImport, dep) {
				found = true
				actuallyUsed = append(actuallyUsed, dep)
				break
			}
		}
		if !found {
			notUsed = append(notUsed, dep)
		}
	}

	sort.Strings(actuallyUsed)
	sort.Strings(notUsed)

	fmt.Printf("✅ ACTUALLY USED DEPENDENCIES:\n")
	for _, dep := range actuallyUsed {
		fmt.Printf("   - %s\n", dep)
	}

	fmt.Printf("\n❌ NOT USED (SAFE TO REMOVE):\n")
	for _, dep := range notUsed {
		fmt.Printf("   - %s\n", dep)
	}

	fmt.Printf("\n📋 Summary:\n")
	fmt.Printf("   Used: %d dependencies\n", len(actuallyUsed))
	fmt.Printf("   Not used: %d dependencies\n", len(notUsed))
	fmt.Printf("   Usage rate: %.1f%%\n", float64(len(actuallyUsed))/float64(len(depsToCheck))*100)

	// Check for CI usage
	fmt.Printf("\n🔍 Checking for CI/Development Tool Usage:\n")

	// Look for Makefile, .github, etc.
	ciFiles := []string{}
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			name := info.Name()
			if name == "Makefile" ||
			   name == ".github" ||
			   name == "go.mod" ||
			   name == "go.sum" ||
			   strings.Contains(path, ".github/") {
				ciFiles = append(ciFiles, path)
			}
		}
		return nil
	})

	// Check CI files for development tool usage
	ciToolsFound := []string{}
	for _, file := range ciFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		contentStr := string(content)
		for _, tool := range []string{"air", "golangci", "gofumpt", "golint"} {
			if strings.Contains(contentStr, tool) {
				ciToolsFound = append(ciToolsFound, tool)
			}
		}
	}

	if len(ciToolsFound) > 0 {
		fmt.Printf("⚠️  CI/Development tools found in use:\n")
		for _, tool := range ciToolsFound {
			fmt.Printf("   - %s\n", tool)
		}
	} else {
		fmt.Printf("✅ No CI tool usage found in project files\n")
	}

	fmt.Printf("\n📝 Recommendations:\n")
	if len(actuallyUsed) > 0 {
		fmt.Printf("   🔒 Keep these dependencies (actually used):\n")
		for _, dep := range actuallyUsed {
			fmt.Printf("     - %s\n", dep)
		}
	}

	if len(notUsed) > 0 {
		fmt.Printf("   🗑️  Safe to remove these dependencies:\n")
		for _, dep := range notUsed {
			fmt.Printf("     - %s\n", dep)
		}
	}

	fmt.Printf("\n✅ Analysis Complete!\n")
}
