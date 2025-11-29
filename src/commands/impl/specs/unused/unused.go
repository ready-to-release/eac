// Command: specs unused-steps
// Description: Find step definitions not used by any feature file
// Short: Detect unused godog step definitions
// Long: The specs unused-steps command scans step definition files in src/specs/impl/
// Long: and compares them against feature files in specs/ to find step definitions
// Long: that are not matched by any Gherkin step.
// Long: This helps identify dead code and maintain a clean test codebase.
// Long: Shared steps from src/specs/internal/steps.go are checked against all pairs.
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed output including all scanned files
// Flag.module: type=string, shorthand=m, default=, usage=Only analyze a specific module (e.g., src-commands)
// Usage: specs unused-steps [--verbose] [--module=<name>]
// HasSideEffects: false
package unused

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(SpecsUnusedSteps)
}

// SpecsUnusedSteps is the entry point for the specs unused-steps command.
func SpecsUnusedSteps() int {
	args := os.Args[3:] // Skip program name, "specs", and "unused-steps"

	verbose := false
	moduleFilter := ""

	// Parse flags
	for _, arg := range args {
		switch {
		case arg == "-v" || arg == "--verbose":
			verbose = true
		case strings.HasPrefix(arg, "--module="):
			moduleFilter = strings.TrimPrefix(arg, "--module=")
		case strings.HasPrefix(arg, "-m="):
			moduleFilter = strings.TrimPrefix(arg, "-m=")
		case arg == "-h" || arg == "--help":
			printUsage()
			return 0
		}
	}

	// Find repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Run analysis
	exitCode := runAnalysis(repoRoot, verbose, moduleFilter)
	return exitCode
}

func runAnalysis(repoRoot string, verbose bool, moduleFilter string) int {
	// Discover all impl↔specs pairs
	pairs, err := DiscoverPairs(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to discover pairs: %v\n", err)
		return 1
	}

	if len(pairs) == 0 {
		fmt.Println("No impl↔specs pairs found in src/specs/impl/")
		return 0
	}

	// Filter by module if specified
	if moduleFilter != "" {
		var filtered []ImplSpecsPair
		for _, pair := range pairs {
			if strings.Contains(pair.ImplDir, moduleFilter) {
				filtered = append(filtered, pair)
			}
		}
		if len(filtered) == 0 {
			fmt.Fprintf(os.Stderr, "Error: no pairs found matching module filter '%s'\n", moduleFilter)
			return 1
		}
		pairs = filtered
	}

	// Get shared steps file
	sharedStepsFile := GetInternalStepsFile(repoRoot)
	if _, err := os.Stat(sharedStepsFile); os.IsNotExist(err) {
		sharedStepsFile = "" // File doesn't exist
	}

	// Analyze for unused steps
	result, err := FindUnusedSteps(pairs, sharedStepsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: analysis failed: %v\n", err)
		return 1
	}

	// Print results
	printResults(result, repoRoot, verbose)

	if result.TotalUnused > 0 {
		return 1 // Exit with error if unused steps found
	}
	return 0
}

func printResults(result *AnalysisResult, repoRoot string, verbose bool) {
	if verbose {
		fmt.Printf("Scanned %d impl↔specs pairs\n", len(result.Pairs))
		fmt.Printf("Total step definitions: %d\n", result.TotalSteps)
		fmt.Printf("Total feature files: %d\n", result.TotalFeatures)
		fmt.Println()
	}

	hasUnused := false

	// Print unused steps per pair
	for _, pairResult := range result.Pairs {
		if len(pairResult.UnusedModuleSteps) == 0 {
			continue
		}

		hasUnused = true
		relImplDir := relativePath(repoRoot, pairResult.Pair.ImplDir)
		fmt.Printf("\n%s:\n", relImplDir)

		// Group by file
		byFile := make(map[string][]UnusedStep)
		for _, step := range pairResult.UnusedModuleSteps {
			byFile[step.File] = append(byFile[step.File], step)
		}

		for file, steps := range byFile {
			relFile := relativePath(repoRoot, file)
			fmt.Printf("  %s:\n", filepath.Base(relFile))
			for _, step := range steps {
				fmt.Printf("    Line %d: %s\n", step.Line, truncatePattern(step.Pattern, 60))
			}
		}
	}

	// Print globally unused shared steps
	if len(result.UnusedShared) > 0 {
		hasUnused = true
		fmt.Printf("\nShared steps (src/specs/internal/steps.go) - unused by ALL pairs:\n")
		for _, step := range result.UnusedShared {
			fmt.Printf("  Line %d: %s\n", step.Line, truncatePattern(step.Pattern, 60))
		}
	}

	// Print summary
	fmt.Println()
	if hasUnused {
		fmt.Printf("Summary: %d unused step(s) found\n", result.TotalUnused)
	} else {
		fmt.Println("No unused step definitions found.")
	}
}

func printUsage() {
	fmt.Println("Find step definitions not used by any feature file")
	fmt.Println()
	fmt.Println("Usage: r2r specs unused-steps [--verbose] [--module=<name>]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -v, --verbose        Show detailed output including all scanned files")
	fmt.Println("  -m, --module=<name>  Only analyze a specific module (e.g., src-commands)")
	fmt.Println("  -h, --help           Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  r2r specs unused-steps")
	fmt.Println("  r2r specs unused-steps --verbose")
	fmt.Println("  r2r specs unused-steps --module=src-commands")
}

func relativePath(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}

func truncatePattern(pattern string, maxLen int) string {
	if len(pattern) <= maxLen {
		return pattern
	}
	return pattern[:maxLen-3] + "..."
}
