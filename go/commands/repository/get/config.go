package get

import (
	"context"
	"fmt"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	getInternal "github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/config"
)

type getConfigCommand struct{}

var _ core.SimpleCommandPort = (*getConfigCommand)(nil)

func (c *getConfigCommand) Name() string { return "get config" }

func (c *getConfigCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-config",
		Short:         "Get loaded repository configuration",
		Long:          "Expected Output:\nYAML with all loaded configuration including:\n  - modules: Module contracts with moniker, type, root path, dependencies\n  - component_kinds: Component type definitions with build/test/deploy capabilities\n  - environments: Environment definitions\n  - testing: Testing configuration (tags and suites from eac-testing contract)",
	}
}

func (c *getConfigCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetConfig()
}

// configFlags defines valid flags for the get config command

// ConfigOutput represents the structured output of all configs.
type ConfigOutput struct {
	Modules        interface{} `yaml:"modules" json:"modules"`
	ComponentKinds interface{} `yaml:"component_kinds" json:"component_kinds"`
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

		if cfg.ComponentKinds != nil {
			output.ComponentKinds = cfg.ComponentKinds
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
