package get

import (
	"context"
	"fmt"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/commands/base"
	"github.com/ready-to-release/eac/go/clibase/flags"
)

type getTestsCommand struct{}

var _ core.SimpleCommandPort = (*getTestsCommand)(nil)

func (c *getTestsCommand) Name() string { return "get tests" }

func (c *getTestsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-tests",
		Short:         "Get all discovered tests with metadata",
		Long: "",
		Notes: "Expected Output:\nYAML list of all discovered tests with metadata, including:\n  - Test name and location (module, package, file)\n  - Test type (unit, integration, e2e, etc.)\n  - Tags and markers\n  - Total test count\n  - Aggregations by module and type",
	}
}

func (c *getTestsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetTests()
}

// testsFlags defines valid flags for the get tests command

func GetTests() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	return internal.ExecuteGetCommand(func() (interface{}, error) {
		// Get repository root
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}

		repoRoot, err := base.FindRepoRoot(cwd)
		if err != nil {
			return nil, fmt.Errorf("failed to find repository root: %w", err)
		}

		// Get all tests with metadata and aggregations
		data, err := base.GetAllTests(repoRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to get tests: %w", err)
		}

		return map[string]interface{}{
			"total_tests": data.TotalCount,
			"tests":       data.Tests,
		}, nil
	})
}
