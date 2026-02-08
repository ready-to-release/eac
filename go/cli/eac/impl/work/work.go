// Command: work
// Short: Workspace management for parallel development using git worktrees
// IsParent: true
// Group.Workspace Lifecycle: create, list, remove
// Group.Development Workflow: commit, pull
// Group.Completion: merge, pr
// Example: clie work create feature/authentication
// Example: clie work commit --all
// Example: clie work pull
// Example: clie work merge
// Example: clie work pr
// Example: clie work list
package work

import (
	"context"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/help"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/adapters/tui/selector"
	"github.com/ready-to-release/eac/go/core/logging"
)

var log = logging.C()

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

	// Check for valid subcommand using dynamic discovery from registry
	if registry.IsValidSubcommand("work", args[0]) {
		// Handled by separate registrations in respective files
		return 0
	}

	log.Errorf("Error: unknown subcommand: %s", args[0])
	log.Info("")
	printHelp()
	return 1
}

// runInteractiveTUI shows the interactive TUI for subcommand selection.
// Uses the new SelectorConsole pattern: TUI just picks a command, caller executes.
func runInteractiveTUI() int {
	// Get options dynamically from registry
	options := tui.SubcommandsFromRegistry("work")

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

