package get

import (
	"context"
	"fmt"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	getInternal "github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

type getReleaseNotesCommand struct{}

var _ core.SimpleCommandPort = (*getReleaseNotesCommand)(nil)

func (c *getReleaseNotesCommand) Name() string { return "get release-notes" }

func (c *getReleaseNotesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-release-notes",
		Short:         "Get parsed release notes for a module",
		Long: "",
		Notes: "Expected Output:\nYAML/JSON/TOML representation of release notes including:\n  - versions: Array of version entries\n    - number: Version number\n    - date: Release date\n    - sections: Array of sections with headers and content\n\nIf version is specified, returns only that version's data.",
		Args: "module [version]",
	}
}

func (c *getReleaseNotesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetReleaseNotes()
}

func GetReleaseNotes() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Check for help flag
	for _, arg := range os.Args[2:] {
		if arg == "--help" || arg == "-h" {
			fmt.Println("Get release notes for a module")
			fmt.Println("\nUsage: get release-notes <module> [version] [flags]")
			fmt.Println("\nArguments:")
			fmt.Println("  module     Module moniker")
			fmt.Println("  version    Optional version number")
			fmt.Println("\nFlags:")
			fmt.Println("  --as-yaml    Output as YAML (default)")
			fmt.Println("  --as-json    Output as JSON")
			fmt.Println("  --as-toml    Output as TOML")
			fmt.Println("  -h, --help   Show this help message")
			return 0
		}
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
