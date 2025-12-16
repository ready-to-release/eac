// Command: get changelog
// Description: Get changelog data in structured format
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
// Args: module [version]
// Long:
// Long: Expected Output:
// Long: YAML/JSON/TOML representation of changelog data including:
// Long:   - module: Module moniker
// Long:   - title: Changelog title
// Long:   - version_type: semver or calver
// Long:   - unreleased: Unreleased changes (if any)
// Long:   - versions: Array of version entries with dates and changes
// Long:
// Long: If version is specified, returns only that version's data.
package get

import (
	"fmt"
	"os"

	getInternal "github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetChangelog)
}

func GetChangelog() int {
	// Parse arguments - expect module after "get changelog"
	args := os.Args[1:]

	// Find where "get changelog" ends
	cmdIdx := -1
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "get" && args[i+1] == "changelog" {
			cmdIdx = i + 2
			break
		}
	}

	// Collect positional arguments (non-flag arguments after command)
	var positional []string
	if cmdIdx != -1 && cmdIdx < len(args) {
		for i := cmdIdx; i < len(args); i++ {
			if args[i][0] != '-' {
				positional = append(positional, args[i])
			}
		}
	}

	if len(positional) < 1 {
		fmt.Fprintf(os.Stderr, "Error: module argument required\n")
		fmt.Fprintf(os.Stderr, "Usage: get changelog <module> [version]\n")
		return 1
	}

	module := positional[0]
	var version string
	if len(positional) > 1 {
		version = positional[1]
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Use the shared get command helper
	return getInternal.ExecuteGetCommand(func() (interface{}, error) {
		report, err := reports.GetChangelog(workspaceRoot, module)
		if err != nil {
			return nil, err
		}

		// If no version specified, return full changelog
		if version == "" {
			return report.Changelog, nil
		}

		// Resolve version using common helper
		versionInfo, err := reports.ResolveVersion(workspaceRoot, module, version)
		if err != nil {
			return nil, err
		}

		// Return data based on version type
		if versionInfo.IsUnreleased {
			if report.Changelog.Unreleased == nil {
				return nil, fmt.Errorf("no unreleased changes found")
			}
			return report.Changelog.Unreleased, nil
		}

		ver := report.Changelog.GetVersion(versionInfo.VersionNumber)
		if ver == nil {
			return nil, fmt.Errorf("version not found: %s", versionInfo.VersionNumber)
		}
		return ver, nil
	})
}
