// Command: validate dependencies
// Description: Validate module dependencies from go.mod files against contracts
package main

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/internal/contracts/modules"
	"github.com/ready-to-release/eac/src/internal/repository/gomod"
)

func init() {
	Register("validate dependencies", ValidateDependencies)
}

func ValidateDependencies() int {
	workspaceRoot := "../.."
	contractVersion := "0.1.0"
	baseModulePath := "github.com/ready-to-release/eac"

	// Load module contracts
	registry, err := modules.LoadFromWorkspace(workspaceRoot, contractVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading contracts: %v\n", err)
		return 1
	}

	// Build dependency graph from go.mod files
	excludeDirs := []string{"out", "vendor", ".git", "node_modules"}
	graph, err := gomod.BuildFromDirectory(workspaceRoot, registry, baseModulePath, excludeDirs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building dependency graph: %v\n", err)
		return 1
	}

	// Validate
	validator := gomod.NewValidator(graph, registry)
	report := validator.Validate()

	// Print report
	fmt.Println(validator.FormatReport(report))

	// Return appropriate exit code
	if report.HasDiscrepancies() {
		return 1
	}

	return 0
}
