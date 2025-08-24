package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("🔍 Testing file finding...")

	count := 0
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			count++
			if count <= 10 {
				fmt.Printf("Found: %s\n", path)
			}
		}

		return nil
	})

	fmt.Printf("Total Go files found: %d\n", count)
}
