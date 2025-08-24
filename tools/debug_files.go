package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run tools/debug_files.go <project-root>")
		os.Exit(1)
	}

	rootDir := os.Args[1]

	fmt.Printf("🔍 Debugging file finding in: %s\n", rootDir)

	fileCount := 0
	goFileCount := 0

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fileCount++

		// Debug first few files
		if fileCount <= 20 {
			fmt.Printf("File %d: %s (isDir: %v)\n", fileCount, path, info.IsDir())
		}

		// Count Go files
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			goFileCount++
			if goFileCount <= 10 {
				fmt.Printf("Go file %d: %s\n", goFileCount, path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("   Total files: %d\n", fileCount)
	fmt.Printf("   Go files: %d\n", goFileCount)
}
