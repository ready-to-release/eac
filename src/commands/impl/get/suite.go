// Command: get suite
// Description: Get test suite information as structured data
// Usage: get suite <suite-moniker>
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
// HasSideEffects: false
package get

import (
	"fmt"
	"os"

	get "github.com/ready-to-release/eac/src/commands/impl/get/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	contractsreports "github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/ready-to-release/eac/src/core/testing"
)

func init() {
	registry.Register(GetSuite)
}

// GetSuite returns test suite information as structured data (YAML/JSON/TOML)
func GetSuite() int {
	// Parse arguments - expect suite moniker after "get suite"
	args := os.Args[1:]

	// Find where "get suite" ends
	suiteIdx := -1
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "get" && args[i+1] == "suite" {
			suiteIdx = i + 2
			break
		}
	}

	if suiteIdx == -1 || suiteIdx >= len(args) {
		fmt.Fprintf(os.Stderr, "Error: suite moniker required\n\n")
		fmt.Fprintf(os.Stderr, "Usage: get suite <suite-moniker> [--as-yaml|--as-json|--as-toml]\n\n")
		fmt.Fprintf(os.Stderr, "Available suites:\n")
		for _, moniker := range testing.ListSuites() {
			fmt.Fprintf(os.Stderr, "  - %s\n", moniker)
		}
		return 1
	}

	// Extract suite moniker (stop at first flag)
	suiteMoniker := ""
	for i := suiteIdx; i < len(args); i++ {
		if len(args[i]) > 0 && args[i][0] == '-' {
			break
		}
		suiteMoniker = args[i]
		break
	}

	if suiteMoniker == "" {
		fmt.Fprintf(os.Stderr, "Error: suite moniker required\n")
		return 1
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not in a git repository: %v\n", err)
		return 1
	}

	// Get suite
	suite, err := testing.GetSuite(suiteMoniker)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		fmt.Fprintf(os.Stderr, "Available suites:\n")
		for _, moniker := range testing.ListSuites() {
			fmt.Fprintf(os.Stderr, "  - %s\n", moniker)
		}
		return 1
	}

	// Use the shared get command helper
	return get.ExecuteGetCommand(func() (interface{}, error) {
		// Load module registry
		moduleReport, err := contractsreports.GetModuleContracts(repoRoot, "0.1.0")
		if err != nil {
			// Non-fatal: continue without module registry
			moduleReport = nil
		}

		// Build file-to-module mapping
		fileModuleMap, err := buildFileModuleMap(repoRoot)
		if err != nil {
			// Non-fatal: use empty map
			fileModuleMap = make(map[string]string)
		}

		// Generate suite report using canonical data generator
		var moduleRegistry *modules.Registry
		if moduleReport != nil {
			moduleRegistry = moduleReport.Registry
		}

		report, err := testing.GenerateSuiteReport(suite, repoRoot, moduleRegistry, fileModuleMap)
		if err != nil {
			return nil, err
		}

		return report, nil
	})
}

// buildFileModuleMap creates a mapping from file paths to module monikers
// TODO: This is duplicated from show/suite.go - should be extracted to a shared location
func buildFileModuleMap(repoRoot string) (map[string]string, error) {
	files, err := repository.GetRepositoryFilesWithModules(true, false, false, repoRoot, "0.1.0")
	if err != nil {
		return nil, err
	}

	fileModuleMap := make(map[string]string)
	for _, file := range files {
		if len(file.Modules) > 0 {
			// Use first module if multiple
			fileModuleMap[file.Name] = file.Modules[0]
		}
	}

	return fileModuleMap, nil
}
