// Command: get suite
// Description: Get test suite information as structured data
// Usage: get suite <suite-moniker>
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
package get

import (
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	contractsreports "github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"github.com/ready-to-release/eac/go/eac/core/testing"
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
		log.Error("Error: suite moniker required\n")
		log.Info("Usage: get suite <suite-moniker> [--as-yaml|--as-json|--as-toml]\n")
		log.Info("Available suites:")
		for _, moniker := range testing.ListSuites() {
			log.Infof("  - %s", moniker)
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
		log.Error("Error: suite moniker required")
		return 1
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot(".")
	if err != nil {
		log.Errorf("Error: not in a git repository: %v", err)
		return 1
	}

	// Get suite
	suite, err := testing.GetSuite(suiteMoniker)
	if err != nil {
		log.Errorf("Error: %v\n", err)
		log.Info("Available suites:")
		for _, moniker := range testing.ListSuites() {
			log.Infof("  - %s", moniker)
		}
		return 1
	}

	// Use the shared get command helper
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		// Load module registry
		moduleReport, err := contractsreports.GetModuleContracts(repoRoot)
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
	// Open git repository
	repo, err := git.Open(repoRoot)
	if err != nil {
		return nil, err
	}

	files, err := repository.GetRepositoryFilesWithModules(repo, true, false, false)
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
