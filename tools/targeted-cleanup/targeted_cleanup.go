package main

import (
	"fmt"
	"os/exec"
)

func main() {
	fmt.Printf("🎯 Targeted Dependency Cleanup\n")
	fmt.Printf("===============================\n\n")

	// Dependencies that are confirmed safe to remove
	safeToRemove := []string{
		"github.com/prometheus",     // Not used in code
		"github.com/rs/zerolog",     // Not used (we use Charm logger)
		"github.com/sirupsen/logrus", // Not used
	}

	// Dependencies to keep (confirmed necessary or used in CI)
	keepThese := []string{
		"github.com/stretchr/testify", // Used in 25+ test files
		"github.com/air-verse/air",    // Used in CI
		"github.com/golangci",         // Used in CI
		"honnef.co/go/tools",          // Static analysis tool
		"github.com/4meepo/tagalign",  // Code formatting
		"github.com/Abirdcfly/dupword", // Duplicate word checker
	}

	fmt.Printf("📋 Dependencies to Remove:\n")
	for _, dep := range safeToRemove {
		fmt.Printf("   🗑️  %s\n", dep)
	}

	fmt.Printf("\n📋 Dependencies to Keep:\n")
	for _, dep := range keepThese {
		fmt.Printf("   🔒 %s\n", dep)
	}

	fmt.Printf("\n🚀 Starting Removal Process...\n")

	removed := 0
	failed := 0

	for _, dep := range safeToRemove {
		fmt.Printf("\n🗑️  Removing: %s\n", dep)

		// Use go get to remove the dependency
		cmd := exec.Command("go", "get", dep+"@none")
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
			if len(output) > 0 {
				fmt.Printf("      Output: %s\n", string(output))
			}
			failed++
		} else {
			fmt.Printf("   ✅ Removed successfully\n")
			removed++
		}
	}

	// Run go mod tidy
	fmt.Printf("\n🔧 Running go mod tidy...\n")
	tidyCmd := exec.Command("go", "mod", "tidy")
	if tidyOutput, tidyErr := tidyCmd.CombinedOutput(); tidyErr != nil {
		fmt.Printf("⚠️  go mod tidy warning: %v\n", tidyErr)
		if len(tidyOutput) > 0 {
			fmt.Printf("   Output: %s\n", string(tidyOutput))
		}
	} else {
		fmt.Printf("✅ go mod tidy completed\n")
	}

	// Check final dependency count
	fmt.Printf("\n📊 Final Status:\n")
	fmt.Printf("   ✅ Successfully removed: %d dependencies\n", removed)
	fmt.Printf("   ❌ Failed to remove: %d dependencies\n", failed)

	// Get current dependency count
	countCmd := exec.Command("sh", "-c", "go list -m all | grep -v github.com/mrtkrcm/ZeroUI | wc -l")
	if countOutput, countErr := countCmd.Output(); countErr == nil {
		var currentCount int
		if _, parseErr := fmt.Sscanf(string(countOutput), "%d", &currentCount); parseErr == nil {
			fmt.Printf("   📈 Current dependencies: %d\n", currentCount)
		}
	}

	fmt.Printf("\n✅ Targeted cleanup complete!\n")
	fmt.Printf("📝 Next steps:\n")
	fmt.Printf("   1. Test the application: go build && go test ./...\n")
	fmt.Printf("   2. Verify CI tools still work: make help (if available)\n")
	fmt.Printf("   3. If issues occur, restore from backup branch\n")
}
