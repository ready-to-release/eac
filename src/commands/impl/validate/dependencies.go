// Command: validate dependencies
// Description: Validate module dependencies from go.mod files against contracts
package validate

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
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
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	contractVersion := "0.1.0"
	baseModulePath := "github.com/ready-to-release/eac"

	// Load module contracts
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot, contractVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading contracts: %v\n", err)
		return 1
	}

	// Build dependency graph from go.mod files
	excludeDirs := []string{"out", "vendor", ".git", "node_modules"}
	graph, err := gomod.BuildFromDirectory(workspaceRoot, moduleRegistry, baseModulePath, excludeDirs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building dependency graph: %v\n", err)
		return 1
	}

	// Validate
	validator := gomod.NewValidator(graph, moduleRegistry)
	report := validator.Validate()

	// Print report
	fmt.Println(validator.FormatReport(report))

	// Return appropriate exit code
	if report.HasDiscrepancies() {
		return 1
	}

	return 0
}
