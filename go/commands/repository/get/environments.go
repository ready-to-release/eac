package get

import (
	"context"
	"fmt"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/config"
)

type getEnvironmentsCommand struct{}

var _ core.SimpleCommandPort = (*getEnvironmentsCommand)(nil)

func (c *getEnvironmentsCommand) Name() string { return "get environments" }

func (c *getEnvironmentsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-environments",
		Short:         "Get environment definitions and configuration",
		Long:          "Expected Output:\nYAML list of environment definitions with configuration for each environment,\nincluding environment variables, deployment targets, and environment-specific settings.",
	}
}

func (c *getEnvironmentsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetEnvironments()
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
