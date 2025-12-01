// Command: validate dependencies
// Short: Validate module dependencies from go.mod files against contracts
// Long: Validate module dependencies from go.mod files against contracts.
// Long:
// Long: This command checks that actual dependencies in go.mod files match the dependencies
// Long: declared in module contracts. It helps ensure consistency between contract definitions
// Long: and actual implementation dependencies.
// Long:
// Long: Validation failures indicate mismatches that should be resolved by either updating
// Long: the contract or fixing the go.mod file.
// Long:
// Long: Example:
// Long:   validate dependencies
package validate

import (
	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/ready-to-release/eac/src/core/repository/gomod"
)

func init() {
	registry.Register(ValidateDependencies)
}

func ValidateDependencies() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	baseModulePath := "github.com/ready-to-release/eac"

	// Load module contracts
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		log.Errorf("Error loading contracts: %v", err)
		return 1
	}

	// Build dependency graph from go.mod files
	excludeDirs := []string{"out", "vendor", ".git", "node_modules"}
	graph, err := gomod.BuildFromDirectory(workspaceRoot, moduleRegistry, baseModulePath, excludeDirs)
	if err != nil {
		log.Errorf("Error building dependency graph: %v", err)
		return 1
	}

	// Validate
	validator := gomod.NewValidator(graph, moduleRegistry)
	report := validator.Validate()

	// Print report
	log.Info(validator.FormatReport(report))

	// Return appropriate exit code
	if report.HasDiscrepancies() {
		return 1
	}

	return 0
}
