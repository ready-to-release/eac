// Command: work list
// Short: List all workspaces and their status
// Long: Lists all git worktrees (workspaces) in a formatted table showing their path, branch, and status.
// Long:
// Long: The status column indicates whether the worktree has uncommitted changes:
// Long:   - clean: No uncommitted changes
// Long:   - dirty: Has uncommitted changes
// Long:
// Long: Use --verbose to see additional information including commit SHA.
// Long:
// Long: Example:
// Long:   work list
// Long:   work list --verbose
// Long:   work list -v
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed information including commit SHA
// HasSideEffects: false
package work

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/commands/impl/work/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/commands/internal/render"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(List)
}

// Intent: List all workspaces (git worktrees) with their status
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear flow: validate → get worktrees → format → display
//   - Table-based output for easy scanning
//   - Status indicators (clean/dirty) are self-explanatory
//
// Easy to change:
//   - Formatting isolated in buildWorktreeTable function
//   - Verbose mode easily extended with additional columns
//   - Uses shared worktree utilities
//
// Hard to break:
//   - Validates git repo before proceeding
//   - Handles empty worktree list
//   - Clear error messages for failure cases

// List displays all git worktrees in a formatted table
func List() int {
	// Phase 1: Parse configuration
	config, err := parseListConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 2: Validate environment
	if err := internal.EnsureInGitRepo(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 3: Get worktrees
	worktrees, err := internal.GetWorktrees(config.repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get worktrees: %v\n", err)
		return 1
	}

	// Phase 4: Display worktrees
	displayWorktrees(worktrees, config.verbose)
	return 0
}

// listConfig holds configuration for the list command
type listConfig struct {
	verbose  bool
	repoRoot string
}

// parseListConfig parses command line arguments
func parseListConfig() (*listConfig, error) {
	args := os.Args[3:] // Skip program name, "work", "list"

	config := &listConfig{
		verbose: false,
	}

	// Parse flags
	for _, arg := range args {
		if arg == "--verbose" || arg == "-v" {
			config.verbose = true
		}
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}
	config.repoRoot = repoRoot

	return config, nil
}

// displayWorktrees formats and displays worktrees in a table
func displayWorktrees(worktrees []internal.Worktree, verbose bool) {
	if len(worktrees) == 0 {
		fmt.Println("No worktrees found")
		return
	}

	// Build table
	table := buildWorktreeTable(worktrees, verbose)
	fmt.Println(table)
	fmt.Println()

	// Display summary
	if len(worktrees) == 1 {
		fmt.Println("1 worktree total")
	} else {
		fmt.Printf("%d worktrees total\n", len(worktrees))
	}
}

// buildWorktreeTable creates a formatted table of worktrees
func buildWorktreeTable(worktrees []internal.Worktree, verbose bool) string {
	// Determine headers based on verbose mode
	var headers []string
	if verbose {
		headers = []string{"Path", "Branch", "SHA", "Status"}
	} else {
		headers = []string{"Path", "Branch", "Status"}
	}

	tb := render.NewTableBuilder().WithHeaders(headers...)

	// Add rows
	for _, wt := range worktrees {
		status := "clean"
		if !wt.Clean {
			status = "dirty"
		}

		branch := wt.Branch
		if branch == "" {
			branch = "(detached)"
		}

		if verbose {
			// Show short SHA (first 7 chars)
			shortSHA := wt.SHA
			if len(shortSHA) > 7 {
				shortSHA = shortSHA[:7]
			}
			tb.AddRow(wt.Path, branch, shortSHA, status)
		} else {
			tb.AddRow(wt.Path, branch, status)
		}
	}

	return tb.Build()
}

// FormatWorktreeStatus returns a human-readable status string
func FormatWorktreeStatus(clean bool) string {
	if clean {
		return "clean"
	}
	return "dirty"
}

// ShortenSHA returns the first 7 characters of a SHA
func ShortenSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

// FormatBranch formats branch name, handling detached HEAD
func FormatBranch(branch string) string {
	if branch == "" || strings.TrimSpace(branch) == "" {
		return "(detached)"
	}
	return branch
}
