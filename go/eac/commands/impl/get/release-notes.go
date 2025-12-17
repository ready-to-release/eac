// Command: get release-notes
// Description: Get release notes data in structured format
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
// Args: module [version]
// Long:
// Long: Expected Output:
// Long: YAML/JSON/TOML representation of release notes including:
// Long:   - versions: Array of version entries
// Long:     - number: Version number
// Long:     - date: Release date
// Long:     - sections: Array of sections with headers and content
// Long:
// Long: If version is specified, returns only that version's data.
package get

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	getInternal "github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetReleaseNotes)
}

// releaseNotesFlags defines valid flags for the get release-notes command
var releaseNotesFlags = []flags.FlagDefinition{
	{Name: "--as-yaml", HasValue: false, ValueType: "bool"},
	{Name: "--as-json", HasValue: false, ValueType: "bool"},
	{Name: "--as-toml", HasValue: false, ValueType: "bool"},
}

func GetReleaseNotes() int {
	// Validate flags before parsing
	if err := flags.ValidateFlags(os.Args[3:], releaseNotesFlags); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Parse arguments - expect module after "get release-notes"
	args := os.Args[1:]

	// Find where "get release-notes" ends
	cmdIdx := -1
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "get" && args[i+1] == "release-notes" {
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
		fmt.Fprintf(os.Stderr, "Usage: get release-notes <module> [version]\n")
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
		report, err := reports.GetReleaseNotes(workspaceRoot, module)
		if err != nil {
			return nil, err
		}

		// If no version specified, return all versions
		if version == "" {
			return report.ReleaseNotes, nil
		}

		// Resolve version using common helper
		versionInfo, err := reports.ResolveVersion(workspaceRoot, module, version)
		if err != nil {
			return nil, err
		}

		// Return specific version
		ver, err := report.ReleaseNotes.GetVersion(versionInfo.VersionNumber)
		if err != nil {
			return nil, err
		}
		return ver, nil
	})
}
