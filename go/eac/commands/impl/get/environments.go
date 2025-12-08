// Command: get environments
// Description: Get all environment contracts
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
package get

import (
	"fmt"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

func init() {
	registry.Register(GetEnvironments)
}

func GetEnvironments() int {
	// Use the shared get command helper
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		cfg, err := config.Load(config.DefaultLoadOptions())
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}

		// Validate contract
		if err := cfg.Environments.Validate(); err != nil {
			log.Warnf("Warning: environment contract validation failed: %v", err)
		}

		return cfg.Environments.Environments, nil
	})
}
