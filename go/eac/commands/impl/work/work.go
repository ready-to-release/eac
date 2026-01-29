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

	"github.com/ready-to-release/eac/go/eac/commands/help"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/commands/tui"
	"github.com/ready-to-release/eac/go/eac/core/logging"

	// Import default TUI to register it
	_ "github.com/ready-to-release/eac/go/eac/commands/tui/default"
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
func runInteractiveTUI() int {
	console := tui.NewForCommand("work", tui.Config{
		CommandName: "work",
		Height:      15,
	})

	// Set subcommands if the console supports it
	if ic, ok := console.(tui.InteractiveConsole); ok {
		ic.SetSubcommands(subcommands)
	}

	// Run the TUI
	if err := console.Start(context.Background()); err != nil {
		log.Errorf("TUI error: %v", err)
		return 1
	}

	// Get selected command
	if ic, ok := console.(tui.InteractiveConsole); ok {
		selected, params := ic.GetSelectedCommand()
		if selected != "" {
			// Re-execute with selected subcommand
			newArgs := []string{os.Args[0], "work", selected}
			if args, ok := params["args"]; ok && args != "" {
				newArgs = append(newArgs, args)
			}
			os.Args = newArgs
			return Work()
		}
	}

	return 0
}

