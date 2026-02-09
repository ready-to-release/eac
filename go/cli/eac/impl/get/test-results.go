// Command: get test-results
// Short: Get test execution results from test manifests
// Long: Get test execution results from test manifests in structured format.
// Long:
// Long: This command reads test UoW manifests from out/test/ and provides
// Long: comprehensive test execution data including results, timing, coverage, and
// Long: security control mappings. The output can be formatted as YAML, JSON, or TOML.
// Long:
// Long: Output includes:
// Long:   - Test results with status (passed/failed/skipped)
// Long:   - Suite assignment and timing data
// Long:   - Specification coverage for godog tests
// Long:   - Control tag summaries
// Long:
// Long: Example:
// Long:   get test-results
// Long:   get test-results eac-ext
// Long:   get test-results eac-ext clie
// Long:   get test-results eac-ext --as-json
// Args: [module...]
// Flag.as-yaml: type=bool, usage=Output as YAML (default format)
// Flag.as-json: type=bool, usage=Output as JSON
// Flag.as-toml: type=bool, usage=Output as TOML
package get

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/manifests/testview"
	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/testdata"
	"github.com/ready-to-release/eac/go/clibase/flags"
)

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

		repoRoot, err := testdata.FindRepoRoot(cwd)
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
