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

type getSpecsCommand struct{}

var _ core.SimpleCommandPort = (*getSpecsCommand)(nil)

func (c *getSpecsCommand) Name() string { return "get specs" }

func (c *getSpecsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-specs",
		Short:         "Get specification files and status for a module",
		Long:          "Expected Output:\nYAML/JSON/TOML representation of specifications including:\n  - module: Module moniker\n  - version: Version number or \"Unreleased\"\n  - spec_files: Array of specification files with status and metadata\n  - added_count: Number of added specs\n  - modified_count: Number of modified specs\n  - deleted_count: Number of deleted specs\n  - total_scenarios: Total scenario count across all specs\n\nIf version is specified, returns specs for that version.",
		Args:          "module [version]",
	}
}

func (c *getSpecsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetSpecs()
}

// specsFlags defines valid flags for the get specs command

func GetSpecs() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Check for help flag
	for _, arg := range os.Args[2:] {
		if arg == "--help" || arg == "-h" {
			fmt.Println("Get specifications for a module")
			fmt.Println("\nUsage: get specs <module> [version] [flags]")
			fmt.Println("\nArguments:")
			fmt.Println("  module     Module moniker")
			fmt.Println("  version    Optional version number")
			fmt.Println("\nFlags:")
			fmt.Println("  --as-yaml    Output as YAML (default)")
			fmt.Println("  --as-json    Output as JSON")
			fmt.Println("  --as-toml    Output as TOML")
			fmt.Println("  --branch     Branch to query (default: main)")
			fmt.Println("  -h, --help   Show this help message")
			return 0
		}
	}

	// Parse arguments - expect module after "get specs"
	args := os.Args[1:]

	// Find where "get specs" ends
	cmdIdx := -1
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "get" && args[i+1] == "specs" {
			cmdIdx = i + 2
			break
		}
	}

	// Parse flags
	branch := ""
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--branch" {
			branch = args[i+1]
			break
		}
	}

	// Collect positional arguments (non-flag arguments after command)
	var positional []string
	if cmdIdx != -1 && cmdIdx < len(args) {
		for i := cmdIdx; i < len(args); i++ {
			if len(args[i]) > 0 && args[i][0] != '-' {
				positional = append(positional, args[i])
			}
		}
	}

	if len(positional) < 1 {
		fmt.Fprintf(os.Stderr, "Error: module argument required\n")
		fmt.Fprintf(os.Stderr, "Usage: get specs <module> [version]\n")
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
		report, err := reports.GetSpecs(reports.Deps(), workspaceRoot, module, version, branch)
		if err != nil {
			return nil, err
		}

		return report, nil
	})
}
