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

type DependencyAnalyzer struct {
	usedImports map[string]bool
	allFiles    []string
}

func NewDependencyAnalyzer() *DependencyAnalyzer {
	return &DependencyAnalyzer{
		usedImports: make(map[string]bool),
		allFiles:    []string{},
	}
}

func (da *DependencyAnalyzer) analyzeFile(filePath string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %v", filePath, err)
	}

	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		da.usedImports[importPath] = true
	}

	return nil
}

func (da *DependencyAnalyzer) findGoFiles(rootDir string) error {
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
			// Don't skip tools directory when analyzing from the project root
			return nil
		}

		// Only analyze .go files (not test files)
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			fmt.Printf("DEBUG: Found Go file: %s\n", path)
			da.allFiles = append(da.allFiles, path)
		}

		return nil
	})
}

func (da *DependencyAnalyzer) analyzeProject(rootDir string) error {
	// Find all Go files
	if err := da.findGoFiles(rootDir); err != nil {
		return fmt.Errorf("failed to find Go files: %v", err)
	}

	// Analyze each file
	for _, file := range da.allFiles {
		if err := da.analyzeFile(file); err != nil {
			fmt.Printf("Warning: failed to analyze %s: %v\n", file, err)
		}
	}

	return nil
}

func (da *DependencyAnalyzer) getUsedDependencies() []string {
	var used []string
	for imp := range da.usedImports {
		used = append(used, imp)
	}
	sort.Strings(used)
	return used
}

func (da *DependencyAnalyzer) readGoModDependencies(goModPath string) ([]string, error) {
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

func (da *DependencyAnalyzer) getUnusedDependencies(goModPath string) ([]string, error) {
	allDeps, err := da.readGoModDependencies(goModPath)
	if err != nil {
		return nil, err
	}

	var unused []string
	for _, dep := range allDeps {
		if !da.usedImports[dep] {
			// Check if it's a standard library package
			if !strings.Contains(dep, ".") {
				continue // Skip stdlib
			}
			unused = append(unused, dep)
		}
	}

	sort.Strings(unused)
	return unused, nil
}

func (da *DependencyAnalyzer) getUnusedFiles() ([]string, error) {
	var unusedFiles []string

	// Check if files are imported/referenced
	for _, file := range da.allFiles {
		if isTestFile(file) {
			continue // Skip test files for this analysis
		}

		packageName := getPackageName(file)
		if packageName == "" {
			continue
		}

		// Check if this package is imported anywhere
		importPath := getImportPath(file)
		if importPath != "" {
			// Look for imports of this package
			found := false
			for _, otherFile := range da.allFiles {
				if otherFile == file {
					continue
				}

				if fileContainsImport(otherFile, importPath) {
					found = true
					break
				}
			}

			if !found {
				unusedFiles = append(unusedFiles, file)
			}
		}
	}

	return unusedFiles, nil
}

func isTestFile(file string) bool {
	return strings.HasSuffix(file, "_test.go")
}

func getPackageName(file string) string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.PackageClauseOnly)
	if err != nil {
		return ""
	}
	return f.Name.Name
}

func getImportPath(file string) string {
	// Extract import path from file path
	// This is a simplified approach - in real projects you'd use go modules info
	dir := filepath.Dir(file)
	parts := strings.Split(dir, string(filepath.Separator))

	// Find the first non-standard directory (after src, pkg, etc.)
	for i, part := range parts {
		if part == "internal" || part == "pkg" {
			return strings.Join(parts[i:], "/")
		}
	}

	return ""
}

func fileContainsImport(file, importPath string) bool {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return false
	}

	for _, imp := range f.Imports {
		if strings.Contains(strings.Trim(imp.Path.Value, `"`), importPath) {
			return true
		}
	}

	return false
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run tools/analyze_deps.go <project-root>")
		os.Exit(1)
	}

	rootDir := os.Args[1]
	analyzer := NewDependencyAnalyzer()

	fmt.Printf("🔍 Analyzing project: %s\n", rootDir)

	// Analyze the project
	if err := analyzer.analyzeProject(rootDir); err != nil {
		fmt.Printf("Error analyzing project: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📁 Found %d Go files\n", len(analyzer.allFiles))

	// Check for go.mod
	goModPath := filepath.Join(rootDir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		// Analyze dependencies
		unusedDeps, err := analyzer.getUnusedDependencies(goModPath)
		if err != nil {
			fmt.Printf("Error analyzing dependencies: %v\n", err)
		} else {
			fmt.Printf("\n📦 Dependency Analysis:\n")
			fmt.Printf("   Used imports: %d\n", len(analyzer.getUsedDependencies()))

			if len(unusedDeps) > 0 {
				fmt.Printf("   ⚠️  Potentially unused dependencies: %d\n", len(unusedDeps))
				fmt.Println("\n   Unused dependencies:")
				for _, dep := range unusedDeps {
					fmt.Printf("     - %s\n", dep)
				}
			} else {
				fmt.Printf("   ✅ No unused dependencies found\n")
			}
		}
	}

	// Analyze unused files
	fmt.Printf("\n🔍 Analyzing unused files...\n")
	unusedFiles, err := analyzer.getUnusedFiles()
	if err != nil {
		fmt.Printf("Error analyzing files: %v\n", err)
	} else {
		if len(unusedFiles) > 0 {
			fmt.Printf("   ⚠️  Potentially unused files: %d\n", len(unusedFiles))
			fmt.Println("\n   Unused files:")
			for _, file := range unusedFiles {
				fmt.Printf("     - %s\n", file)
			}
		} else {
			fmt.Printf("   ✅ No unused files found\n")
		}
	}

	fmt.Printf("\n📊 Analysis Complete!\n")
}
