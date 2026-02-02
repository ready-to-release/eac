// Command: get approval-comments
//
//	--as-yaml: Output as YAML (default)
//	--as-json: Output as JSON
//	--as-toml: Output as TOML
//	--include-all-reviews: Include all review states (not just APPROVED)
//	--branch: Branch to query (default: trunk branch from config, usually "main")
//
// Args: module [version]
// Long:
// Long: Expected Output:
// Long: YAML/JSON/TOML representation of PR approval comments including:
// Long:   - module: Module moniker
// Long:   - version: Version number or "Unreleased"
// Long:   - total_prs: Number of PRs with spec files
// Long:   - total_approvals: Total number of approval reviews
// Long:   - approvals: Array of approval reviews with PR details
// Long:
// Long: If version is specified, returns approvals for that version.
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
	registry.Register(GetApprovalComments)
}

// approvalCommentsFlags defines valid flags for the get approval-comments command

func GetApprovalComments() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	// Parse arguments - expect module after "get approval-comments"
	args := os.Args[1:]

	// Find where "get approval-comments" ends
	cmdIdx := -1
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "get" && args[i+1] == "approval-comments" {
			cmdIdx = i + 2
			break
		}
	}

	// Parse flags
	includeAllReviews := false
	branch := ""
	for i := 0; i < len(args); i++ {
		if args[i] == "--include-all-reviews" {
			includeAllReviews = true
		}
		if args[i] == "--branch" && i+1 < len(args) {
			branch = args[i+1]
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
		fmt.Fprintf(os.Stderr, "Usage: get approval-comments <module> [version] [--include-all-reviews]\n")
		return 1
	}

	module := positional[0]
	version := ""
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
		return reports.GetApprovalComments(workspaceRoot, module, version, includeAllReviews, branch)
	})
}
