// Command: get environments
// Description: Get all environment contracts
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
	"github.com/ready-to-release/eac/src/core/environments"
)

func init() {
	registry.Register(GetEnvironments)
}

func GetEnvironments() int {
	// Use the shared get command helper
	return get.ExecuteGetCommand(func() (interface{}, error) {
		contract, err := environments.LoadEnvironmentContract()
		if err != nil {
			return nil, fmt.Errorf("failed to load environment contract: %w", err)
		}

		// Validate contract
		if err := contract.ValidateContract(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: environment contract validation failed: %v\n", err)
		}

		return contract.Environments, nil
	})
}
