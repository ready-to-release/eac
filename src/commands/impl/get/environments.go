// Command: get environments
// Description: Get all environment contracts
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
package get

import (
	"fmt"
	"os"

	get "github.com/ready-to-release/eac/src/commands/impl/get/internal"
	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/config"
)

func init() {
	registry.Register(GetEnvironments)
}

func GetEnvironments() int {
	// Use the shared get command helper
	return get.ExecuteGetCommand(func() (interface{}, error) {
		cfg, err := config.Load(config.DefaultLoadOptions())
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}

		// Validate contract
		if err := cfg.Environments.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: environment contract validation failed: %v\n", err)
		}

		return cfg.Environments.Environments, nil
	})
}
