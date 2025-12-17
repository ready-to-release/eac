// Command: get environments
// Description: Get all environment contracts
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
// Long:
// Long: Expected Output:
// Long: YAML list of environment definitions with configuration for each environment,
// Long: including environment variables, deployment targets, and environment-specific settings.
package get

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

func init() {
	registry.Register(GetEnvironments)
}

// environmentsFlags defines valid flags for the get environments command

func GetEnvironments() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Use the shared get command helper
	return internal.ExecuteGetCommand(func() (interface{}, error) {
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
