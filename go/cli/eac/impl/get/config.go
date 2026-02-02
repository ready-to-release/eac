// Command: get config
//
//	--as-yaml: Output as YAML (default)
//	--as-json: Output as JSON
//	--as-toml: Output as TOML
//
// Long:
// Long: Expected Output:
// Long: YAML with all loaded configuration including:
// Long:   - modules: Module contracts with moniker, type, root path, dependencies
// Long:   - module_types: Module type definitions with build/test/deploy capabilities
// Long:   - environments: Environment definitions
// Long:   - testing: Testing configuration (tags and suites from eac-testing contract)
package get

import (
	"fmt"
	"os"

	getInternal "github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/config"
)

func init() {
	registry.Register(GetConfig)
}

// configFlags defines valid flags for the get config command

// ConfigOutput represents the structured output of all configs.
type ConfigOutput struct {
	Modules        interface{} `yaml:"modules" json:"modules"`
	ComponentTypes interface{} `yaml:"component_types" json:"component_types"`
	Environments   interface{} `yaml:"environments" json:"environments"`
	Testing        interface{} `yaml:"testing" json:"testing"`
}

func GetConfig() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Use the shared get command helper
	return getInternal.ExecuteGetCommand(func() (interface{}, error) {
		// Load all configs with defaults applied
		cfg, err := config.Load(config.DefaultLoadOptions())
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration: %w", err)
		}

		// Build output structure
		output := ConfigOutput{}

		if cfg.Repository != nil {
			output.Modules = cfg.Repository
		}

		if cfg.ComponentTypes != nil {
			output.ComponentTypes = cfg.ComponentTypes
		}

		if cfg.Environments != nil {
			output.Environments = cfg.Environments
		}

		if cfg.Testing != nil {
			output.Testing = cfg.Testing
		}

		return output, nil
	})
}
