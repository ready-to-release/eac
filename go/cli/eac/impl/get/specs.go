// Command: get specs
// Short: Get specification files and status for a module
//
//	--as-yaml: Output as YAML (default)
//	--as-json: Output as JSON
//	--as-toml: Output as TOML
//	--branch: Branch to query (default: trunk branch from config, usually "main")
//
// Args: module [version]
// Long:
// Long: Expected Output:
// Long: YAML/JSON/TOML representation of specifications including:
// Long:   - module: Module moniker
// Long:   - version: Version number or "Unreleased"
// Long:   - spec_files: Array of specification files with status and metadata
// Long:   - added_count: Number of added specs
// Long:   - modified_count: Number of modified specs
// Long:   - deleted_count: Number of deleted specs
// Long:   - total_scenarios: Total scenario count across all specs
// Long:
// Long: If version is specified, returns specs for that version.
package get

import (
	"fmt"
	"os"

	getInternal "github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(GetSpecs)
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
		report, err := reports.GetSpecs(workspaceRoot, module, version, branch)
		if err != nil {
			return nil, err
		}

		return report, nil
	})
}
