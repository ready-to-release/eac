package get

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/get/internal"
	testview "github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/commands/base"
	"github.com/ready-to-release/eac/go/clibase/flags"
)

type getTestResultsCommand struct{}

var _ core.SimpleCommandPort = (*getTestResultsCommand)(nil)

func (c *getTestResultsCommand) Name() string { return "get test-results" }

func (c *getTestResultsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-test-results",
		Short:         "Get test execution results from test manifests",
		Long: "Get test execution results from test manifests in structured format.\n" +
			"\n" +
			"This command reads test UoW manifests from out/test/ and provides\n" +
			"comprehensive test execution data including results, timing, coverage, and\n" +
			"security control mappings. The output can be formatted as YAML, JSON, or TOML.\n" +
			"\n" +
			"Output includes:\n" +
			"  - Test results with status (passed/failed/skipped)\n" +
			"  - Suite assignment and timing data\n" +
			"  - Specification coverage for godog tests\n" +
			"  - Control tag summaries\n" +
			"\n" +
			"Example:\n" +
			"  get test-results\n" +
			"  get test-results eac-ext\n" +
			"  get test-results eac-ext clie\n" +
			"  get test-results eac-ext --as-json",
		Args: "[module...]",
		Flags: []core.FlagSpec{
			{Name: "as-yaml", Type: "bool", Usage: "Output as YAML (default format)"},
			{Name: "as-json", Type: "bool", Usage: "Output as JSON"},
			{Name: "as-toml", Type: "bool", Usage: "Output as TOML"},
		},
	}
}

func (c *getTestResultsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetTestResults()
}

func GetTestResults() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Parse module arguments
	args := os.Args[1:]
	var monikers []string

	// Skip command name and collect module arguments
	inCommand := false
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Skip until we find "test-results"
		if !inCommand {
			if arg == "test-results" {
				inCommand = true
			}
			continue
		}

		// Skip flags
		if strings.HasPrefix(arg, "--") {
			continue
		}

		// Collect module monikers
		monikers = append(monikers, arg)
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

		// Load test views from UoW manifests
		var views []*testview.TestModuleView
		if len(monikers) == 0 {
			views, err = testview.LoadAllTestViews(repoRoot)
			if err != nil {
				return nil, fmt.Errorf("failed to load test data: %w", err)
			}
		} else {
			views, err = testview.LoadTestViewsForModules(repoRoot, monikers)
			if err != nil {
				return nil, fmt.Errorf("failed to load test data: %w", err)
			}
		}

		if len(views) == 0 {
			if len(monikers) == 0 {
				return nil, fmt.Errorf("no test manifests found (run tests first)")
			}
			return nil, fmt.Errorf("no test manifests found for modules: %s (run tests first)", strings.Join(monikers, ", "))
		}

		// Build complete test data from UoW-based views
		data := testview.BuildCompleteTestData(views)

		return data, nil
	})
}
