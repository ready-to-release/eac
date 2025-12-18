// Command: get specs unused-steps
// Description: Find step definitions not used by any feature file
// Short: Detect unused godog step definitions
// Long: The get specs unused-steps command scans step definition files in go/eac/specs/impl/
// Long: and compares them against feature files in specs/ to find step definitions
// Long: that are not matched by any Gherkin step.
// Long: This helps identify dead code and maintain a clean test codebase.
// Long: Shared steps from go/eac/specs/internal/steps.go are checked against all pairs.
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed output including all scanned files
// Flag.module: type=string, shorthand=m, default=, usage=Only analyze a specific module (e.g., eac-commands)
// Usage: specs unused-steps [--verbose] [--module=<name>]
package unused

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// commandFlags defines valid flags for the specs unused command

var log = logging.C()

func init() {
	registry.Register(SpecsUnusedSteps)
}

// SpecsUnusedSteps is the entry point for the specs unused-steps command.
func SpecsUnusedSteps() int {
	args := os.Args[3:] // Skip program name, "specs", and "unused-steps"

	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

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
		log.Errorf("Error: failed to find repository root: %v", err)
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
		log.Errorf("Error: failed to discover pairs: %v", err)
		return 1
	}

	if len(pairs) == 0 {
		log.Info("No impl↔specs pairs found in go/eac/specs/impl/")
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
			log.Errorf("Error: no pairs found matching module filter '%s'", moduleFilter)
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
		log.Errorf("Error: analysis failed: %v", err)
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
		log.Infof("Scanned %d impl↔specs pairs", len(result.Pairs))
		log.Infof("Total step definitions: %d", result.TotalSteps)
		log.Infof("Total feature files: %d", result.TotalFeatures)
		log.Info("")
	}

	hasUnused := false

	// Print unused steps per pair
	for _, pairResult := range result.Pairs {
		if len(pairResult.UnusedModuleSteps) == 0 {
			continue
		}

		hasUnused = true
		relImplDir := relativePath(repoRoot, pairResult.Pair.ImplDir)
		log.Infof("\n%s:", relImplDir)

		// Group by file
		byFile := make(map[string][]UnusedStep)
		for _, step := range pairResult.UnusedModuleSteps {
			byFile[step.File] = append(byFile[step.File], step)
		}

		for file, steps := range byFile {
			relFile := relativePath(repoRoot, file)
			log.Infof("  %s:", filepath.Base(relFile))
			for _, step := range steps {
				log.Infof("    Line %d: %s", step.Line, truncatePattern(step.Pattern, 60))
			}
		}
	}

	// Print globally unused shared steps
	if len(result.UnusedShared) > 0 {
		hasUnused = true
		log.Info("\nShared steps (go/eac/specs/internal/steps.go) - unused by ALL pairs:")
		for _, step := range result.UnusedShared {
			log.Infof("  Line %d: %s", step.Line, truncatePattern(step.Pattern, 60))
		}
	}

	// Print summary
	log.Info("")
	if hasUnused {
		log.Infof("Summary: %d unused step(s) found", result.TotalUnused)
	} else {
		log.Info("No unused step definitions found.")
	}
}

func printUsage() {
	log.Info("Find step definitions not used by any feature file")
	log.Info("")
	log.Info("Usage: r2r specs unused-steps [--verbose] [--module=<name>]")
	log.Info("")
	log.Info("Flags:")
	log.Info("  -v, --verbose        Show detailed output including all scanned files")
	log.Info("  -m, --module=<name>  Only analyze a specific module (e.g., eac-commands)")
	log.Info("  -h, --help           Show this help message")
	log.Info("")
	log.Info("Examples:")
	log.Info("  r2r specs unused-steps")
	log.Info("  r2r specs unused-steps --verbose")
	log.Info("  r2r specs unused-steps --module=eac-commands")
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
