package test

import (
	"context"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/testing"
)

type testListSuitesCommand struct{}

var _ core.SimpleCommandPort = (*testListSuitesCommand)(nil)

func (c *testListSuitesCommand) Name() string { return "test list-suites" }

func (c *testListSuitesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "test-list-suites",
		Short:         "List all available test suites",
		Long:          "List all available test suites defined in the repository.\n\nTest suites are logical groupings of tests defined in domain.\nThis command displays all configured suites, making it easy to discover\nwhat test suites are available for execution.\n\nExpected Output:\n  - List of configured test suite names (unit, integration, acceptance, etc.)\n  - Suite name, display name, and description for each suite\n\nExample:\n  test list-suites",
	}
}

func (c *testListSuitesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ListSuites()
}

// ListSuites lists all available test suites.
func ListSuites() int {
	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	log.Info("Available test suites:")
	log.Info("")

	suites := testing.ListSuites()
	for _, moniker := range suites {
		suite, err := testing.GetSuite(moniker)
		if err != nil {
			log.Errorf("%v", err)
			continue
		}

		log.Infof("  %s", suite.Moniker)
		log.Infof("    Name: %s", suite.Name)
		log.Infof("    Description: %s", suite.Description)
		log.Info("")
	}

	return 0
}
