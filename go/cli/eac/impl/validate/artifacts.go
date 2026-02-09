// Command: validate artifacts
// Short: Validate that build artifacts exist for a module and its dependencies
// Long: Validate that build artifacts exist for a module and all transitive dependencies.
// Long:
// Long: This command checks that all required build artifacts are present in out/build/
// Long: for the specified module and all modules it depends on. This ensures that the
// Long: build→test flow has all necessary artifacts before running tests.
// Long:
// Long: The validation includes:
// Long: - Target module artifacts (executables, files, directories, etc.)
// Long: - All transitive dependency artifacts (recursive check, unless --skip-depm)
// Long: - Platform-specific artifacts for current platform (or all if built)
// Long: - Marker files for modules with no traditional build outputs
// Long:
// Long: Validation failures indicate missing artifacts that must be built before testing.
// Long:
// Long: Expected Output:
// Long:   Displays validation results for target module and all dependency artifacts.
// Long:   Exit code 0 if all artifacts present, non-zero if any missing.
// Long:   Shows detailed table with artifact counts and missing artifact paths.
// Long:
// Long: Example:
// Long:   validate artifacts eac-cli
// Long:   validate artifacts clie --os linux --arch amd64
// Long:   validate artifacts docs --skip-depm     # Release context: skip module deps
//
// Args: module
//
// Flag.skip-depm: type=bool, default=false, usage=Skip validation of transitive module dependencies (for release workflows)
// Flag.os: type=string, default=runtime.GOOS, usage=Target OS for platform-specific artifacts
// Flag.arch: type=string, default=runtime.GOARCH, usage=Target architecture for platform-specific artifacts
package validate

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	implinternal "github.com/ready-to-release/eac/go/cli/eac/impl/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/repository"
)

func ValidateArtifacts() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "validate", and "artifacts"

	if len(args) == 0 {
		log.Errorf("Usage: validate artifacts <module>")
		return 1
	}

	moduleName := args[0]
	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH
	noDeps := false

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--os" && i+1 < len(args):
			targetOS = args[i+1]
			i++
		case arg == "--arch" && i+1 < len(args):
			targetArch = args[i+1]
			i++
		case arg == "--skip-depm":
			noDeps = true
		}
	}

	return validateArtifactsForModule(moduleName, targetOS, targetArch, noDeps)
}

func validateArtifactsForModule(moduleName, targetOS, targetArch string, noDeps bool) int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Load configuration
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("failed to load config: %v", err)
		return 1
	}

	// Load module registry
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		log.Errorf("failed to load module contracts: %v", err)
		return 1
	}

	// Check if module exists
	if _, exists := moduleRegistry.Get(moduleName); !exists {
		log.Errorf("module not found: %s", moduleName)
		return 1
	}

	var results *implinternal.ValidationResults
	if noDeps {
		// Validate only target module (for release workflows)
		results, err = implinternal.ValidateArtifactsTargetOnly(
			moduleName, cfg, moduleRegistry, targetOS, targetArch, workspaceRoot,
		)
	} else {
		// Validate artifacts for module and all dependencies
		results, err = implinternal.ValidateArtifactsWithDependencies(
			moduleName, cfg, moduleRegistry, targetOS, targetArch, workspaceRoot,
		)
	}
	if err != nil {
		log.Errorf("validation error: %v", err)
		return 1
	}

	// Format and display results
	output := formatValidationResults(results, targetOS, targetArch)
	log.Info(output)

	// Return exit code based on validation result
	if !results.Passed {
		return 1
	}

	return 0
}

// formatValidationResults formats validation results as a detailed table.
func formatValidationResults(results *implinternal.ValidationResults, targetOS, targetArch string) string {
	var output strings.Builder

	// Header
	output.WriteString(fmt.Sprintf("# Artifact Validation: %s\n\n", results.TargetModule))
	output.WriteString(fmt.Sprintf("Platform: %s/%s\n", targetOS, targetArch))
	output.WriteString(fmt.Sprintf("Modules checked: %d (%d passed, %d failed)\n\n",
		results.TotalModules, results.PassedCount, results.FailedCount))

	// Overall status
	if results.Passed {
		output.WriteString("✅ All artifacts present\n\n")
	} else {
		output.WriteString("❌ Missing artifacts detected\n\n")
	}

	// Module table
	tb := render.NewTableBuilder()
	tb.WithHeaders("Module", "Type", "Role", "Artifacts", "Missing", "Status")

	for _, modResult := range results.Modules {
		role := "Target"
		if modResult.IsDependency {
			role = "Dependency"
		}

		status := "✅ Pass"
		if modResult.Error != "" {
			status = fmt.Sprintf("❌ Error: %s", modResult.Error)
		} else if !modResult.HasBuildArtifacts {
			status = "⚪ No artifacts"
		} else if modResult.Summary.Missing > 0 {
			status = "❌ Missing"
		}

		artifactCount := "-"
		missingCount := "-"
		if modResult.HasBuildArtifacts && modResult.Summary != nil {
			artifactCount = fmt.Sprintf("%d", modResult.Summary.Total)
			missingCount = fmt.Sprintf("%d", modResult.Summary.Missing)
		}

		tb.AddRow(
			modResult.Moniker,
			modResult.Type,
			role,
			artifactCount,
			missingCount,
			status,
		)
	}

	output.WriteString(tb.Build())

	// Detailed missing artifacts section
	hasMissing := false
	for _, modResult := range results.Modules {
		if modResult.Summary != nil && modResult.Summary.Missing > 0 {
			if !hasMissing {
				output.WriteString("\n\n## Missing Artifacts\n\n")
				hasMissing = true
			}

			output.WriteString(fmt.Sprintf("### %s\n\n", modResult.Moniker))

			for i := range modResult.Artifacts {
				art := &modResult.Artifacts[i]
				if !art.Exists {
					output.WriteString(fmt.Sprintf("- ❌ %s: %s\n",
						art.ID, art.ResolvedPath))
				}
			}
			output.WriteString("\n")
		}
	}

	// Help text
	if !results.Passed {
		output.WriteString("\n## Resolution\n\n")
		output.WriteString(fmt.Sprintf("Run 'build %s' to generate missing artifacts\n", results.TargetModule))
	}

	return output.String()
}
