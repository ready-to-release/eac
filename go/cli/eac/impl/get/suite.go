// Command: get suite
// Usage: get suite <suite-moniker>
//
//	--as-yaml: Output as YAML (default)
//	--as-json: Output as JSON
//	--as-toml: Output as TOML
//
// Long:
// Long: Expected Output:
// Long: YAML test suite definition and configuration, including:
// Long:   - Suite metadata (moniker, name, description)
// Long:   - Test selection criteria (tags, modules, patterns)
// Long:   - Suite-level configuration and settings
// Long:   - List of included tests with their metadata
package get

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/testdata"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	contractsreports "github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/testing"
)

func init() {
	registry.Register(GetSuite)
}

// suiteFlags defines valid flags for the get suite command

// GetSuite returns test suite information as structured data (YAML/JSON/TOML).
func GetSuite() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Check for help flag
	for _, arg := range os.Args[2:] {
		if arg == "--help" || arg == "-h" {
			fmt.Println("Get test suite definition and configuration")
			fmt.Println("\nUsage: get suite <suite-moniker> [flags]")
			fmt.Println("\nArguments:")
			fmt.Println("  suite-moniker    Test suite moniker")
			fmt.Println("\nFlags:")
			fmt.Println("  --as-yaml    Output as YAML (default)")
			fmt.Println("  --as-json    Output as JSON")
			fmt.Println("  --as-toml    Output as TOML")
			fmt.Println("  -h, --help   Show this help message")
			return 0
		}
	}

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
		fmt.Fprintln(os.Stderr, "Error: suite moniker required")
		fmt.Fprintln(os.Stderr, "Usage: get suite <suite-moniker> [--as-yaml|--as-json|--as-toml]")
		fmt.Fprintln(os.Stderr, "Available suites:")
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
		fmt.Fprintln(os.Stderr, "Error: suite moniker required")
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Available suites:")
		for _, moniker := range testing.ListSuites() {
			fmt.Fprintf(os.Stderr, "  - %s\n", moniker)
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
		fileModuleMap, err := testdata.BuildFileModuleMap(repoRoot)
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
