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
// Long:   - testing_tags: Testing tag definitions
// Long:   - test_suites: Test suite configurations
package get

import (
	"fmt"
	"os"

	getInternal "github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

func init() {
	registry.Register(GetConfig)
}

// configFlags defines valid flags for the get config command

// ConfigOutput represents the structured output of all configs.
type ConfigOutput struct {
	Modules      interface{} `yaml:"modules" json:"modules"`
	ModuleTypes  interface{} `yaml:"module_types" json:"module_types"`
	Environments interface{} `yaml:"environments" json:"environments"`
	TestingTags  interface{} `yaml:"testing_tags" json:"testing_tags"`
	TestSuites   interface{} `yaml:"test_suites" json:"test_suites"`
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

		if cfg.ModuleTypes != nil {
			output.ModuleTypes = cfg.ModuleTypes
		}

		if cfg.Environments != nil {
			output.Environments = cfg.Environments
		}

		if cfg.TestingTags != nil {
			output.TestingTags = cfg.TestingTags
		}

		if cfg.TestSuites != nil {
			output.TestSuites = cfg.TestSuites
		}

		return output, nil
	})
}
