// Command: get tests
//
//	--as-yaml: Output as YAML (default)
//	--as-json: Output as JSON
//	--as-toml: Output as TOML
//
// Long:
// Long: Expected Output:
// Long: YAML list of all discovered tests with metadata, including:
// Long:   - Test name and location (module, package, file)
// Long:   - Test type (unit, integration, e2e, etc.)
// Long:   - Tags and markers
// Long:   - Total test count
// Long:   - Aggregations by module and type
package get

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/testdata"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(GetTests)
}

// testsFlags defines valid flags for the get tests command

func GetTests() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	return internal.ExecuteGetCommand(func() (interface{}, error) {
		// Get repository root
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}

		repoRoot, err := testdata.FindRepoRoot(cwd)
		if err != nil {
			return nil, fmt.Errorf("failed to find repository root: %w", err)
		}

		// Get all tests with metadata and aggregations
		data, err := testdata.GetAllTests(repoRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to get tests: %w", err)
		}

		return map[string]interface{}{
			"total_tests": data.TotalCount,
			"tests":       data.Tests,
		}, nil
	})
}
