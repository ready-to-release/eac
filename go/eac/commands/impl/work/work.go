// Command: work
// Short: Workspace management for parallel development using git worktrees
// IsParent: true
// Group.Workspace Lifecycle: create, list, remove
// Group.Development Workflow: commit, pull
// Group.Completion: merge, pr
// Example: r2r work create feature/authentication
// Example: r2r work commit --all
// Example: r2r work pull
// Example: r2r work merge
// Example: r2r work pr
// Example: r2r work list
package work

import (
	"context"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/help"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/adapters/tui"
	"github.com/ready-to-release/eac/go/eac/adapters/tui/selector"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(Work)
}

// subcommands defines all available work subcommands.
var subcommands = []tui.SubcommandInfo{
	{Name: "create", Description: "Create new workspace for parallel development"},
	{Name: "list", Description: "List all workspaces and their status"},
	{Name: "remove", Description: "Remove workspace and optionally delete branches"},
	{Name: "commit", Description: "Commit changes with AI-generated messages"},
	{Name: "pull", Description: "Sync workspace with latest main via rebase"},
	{Name: "merge", Description: "Merge workspace to main (squash by default)"},
	{Name: "pr", Description: "Create pull request with AI-generated description"},
}

// printHelp prints the help for the work command using registry metadata.
func printHelp() {
	reg := registry.GetCommand("work")
	help.PrintHelp(os.Stdout, reg, registry.GetCommandRegistry())
}

// Work command entry point.
func Work() int {
	args := os.Args[2:] // Skip program name and "work"

	// Check for help flag first
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printHelp()
		return 0
	}

	// If no subcommand and interactive, show TUI
	if len(args) == 0 {
		if tui.ShouldUseTUI("work", false, false) {
			return runInteractiveTUI()
		}
		printHelp()
		return 1
	}

	// Check for valid subcommand
	switch args[0] {
	case "create", "list", "commit", "pull", "merge", "pr", "remove":
		// Handled by separate registrations in respective files
		return 0
	default:
		log.Errorf("Error: unknown subcommand: %s", args[0])
		log.Info("")
		printHelp()
		return 1
	}
}

// runInteractiveTUI shows the interactive TUI for subcommand selection.
// Uses the new SelectorConsole pattern: TUI just picks a command, caller executes.
func runInteractiveTUI() int {
	// Convert subcommands to CommandOptions
	options := tui.SubcommandsToOptions(subcommands)

	// Run the selector - it shows the list, user picks, returns selection
	selected, args, cancelled := selector.RunSelector(context.Background(), options)

	// Handle cancellation
	if cancelled {
		return 0
	}

	// Handle empty selection (shouldn't happen, but be safe)
	if selected == "" {
		return 0
	}

	// Build new args and re-execute with selected subcommand
	newArgs := []string{os.Args[0], "work", selected}
	if args != "" {
		// Split args on whitespace to handle multiple arguments
		newArgs = append(newArgs, strings.Fields(args)...)
	}
	os.Args = newArgs
	return Work()
}

