// Command: get test-results
// Short: Get test execution results from test manifests
// Long: Get test execution results from test manifests in structured format.
// Long:
// Long: This command reads test.manifest.json files from out/test/ and provides
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
// Long:   get test-results ext-eac
// Long:   get test-results ext-eac r2r-cli
// Long:   get test-results ext-eac --as-json
// Args: [module...]
// Flag.as-yaml: type=bool, usage=Output as YAML (default format)
// Flag.as-json: type=bool, usage=Output as JSON
// Flag.as-toml: type=bool, usage=Output as TOML
package get

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/manifests"
	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/testdata"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
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

		// Load test manifests for specified modules (or all if no modules specified)
		var manifestList []*implinternal.TestManifest
		if len(monikers) == 0 {
			// No modules specified - scan all modules with test manifests
			manifestList, err = manifests.LoadAllTestManifests(repoRoot)
			if err != nil {
				return nil, fmt.Errorf("failed to load manifests: %w", err)
			}
		} else {
			// Specific modules requested
			manifestList, err = manifests.LoadTestManifestsForModules(repoRoot, monikers)
			if err != nil {
				return nil, fmt.Errorf("failed to load manifests: %w", err)
			}
		}

		if len(manifestList) == 0 {
			if len(monikers) == 0 {
				return nil, fmt.Errorf("no test manifests found (run tests first)")
			}
			return nil, fmt.Errorf("no test manifests found for modules: %s (run tests first)", strings.Join(monikers, ", "))
		}

		// Use shared function to build complete test data
		data := manifests.BuildCompleteTestData(manifestList)

		// Return complete data structure
		return data, nil
	})
}
