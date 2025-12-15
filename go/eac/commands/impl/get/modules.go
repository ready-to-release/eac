// Command: get modules
// Description: Get all module contracts in the repository
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
//   --as-<name>: Output using custom renderer (e.g., --as-summary, --as-count)
// Long:
// Long: Expected Output:
// Long: YAML list of all module contracts, each containing:
// Long:   - moniker: Unique module identifier
// Long:   - type: Module type (e.g., go, container, typescript, static)
// Long:   - root: Root path relative to repository
// Long:   - depends_on: List of dependency module monikers
// Long:   - Additional metadata (books, files, etc.)
package get

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetModules)
}

func GetModules() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Use the shared get command helper
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		report, err := reports.GetModuleContracts(workspaceRoot)
		if err != nil {
			return nil, err
		}
		return report.Modules, nil
	})
}
