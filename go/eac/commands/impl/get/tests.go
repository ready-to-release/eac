// Command: get tests
// Description: Get all tests in the repository
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
package get

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/testdata"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(GetTests)
}

func GetTests() int {
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
