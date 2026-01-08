// Command: get test-results
// Short: Get test execution results from test manifests
// Long: Get test execution results from test manifests in structured format.
// Long:
// Long: This command reads all test.manifest.json files from out/test/ and provides
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
// Long:   get test-results --as-json
// Long:   get test-results --as-yaml
// Flag.as-yaml: type=bool, usage=Output as YAML (default format)
// Flag.as-json: type=bool, usage=Output as JSON
// Flag.as-toml: type=bool, usage=Output as TOML
package get

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/manifests"
	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/testdata"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(GetTestResults)
}

func GetTestResults() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
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

		// Load test manifests
		manifestList, err := manifests.LoadAllTestManifests(repoRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to load manifests: %w", err)
		}

		if len(manifestList) == 0 {
			return nil, fmt.Errorf("no test manifests found in %s/out/test/ (run tests first)", repoRoot)
		}

		// Use shared function to build complete test data
		data := manifests.BuildCompleteTestData(manifestList)

		// Return complete data structure
		return data, nil
	})
}
