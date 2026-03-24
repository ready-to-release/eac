package work

import (
	"context"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/adapters/tui/selector"
	"github.com/ready-to-release/eac/go/commands/repository/internal/helputil"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/logging"
)

type workCommand struct{}

var _ core.CommandPort = (*workCommand)(nil)

func (c *workCommand) Name() string { return "work" }

func (c *workCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "work",
		Short:         "Workspace management for parallel development using git worktrees",
		IsParent:      true,
		SubcommandGroups: []core.SubcommandGroup{
			{Name: "Workspace Lifecycle", Subcommands: []string{"create", "list", "remove"}},
			{Name: "Development Workflow", Subcommands: []string{"commit", "pull"}},
			{Name: "Completion", Subcommands: []string{"merge", "pr"}},
		},
		Examples: []string{
			"eac work create feature/authentication",
			"eac work commit --all",
			"eac work pull",
			"eac work merge",
			"eac work pr",
			"eac work list",
		},
	}
}

var log = logging.C()

// printHelp prints the help for the work command using registry metadata.
func printHelp() {
	cmd, _ := registry.Global().Get("work")
	helputil.PrintHelp(os.Stdout, cmd, registry.Global())
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
	if _, ok := registry.Global().Get("work " + args[0]); ok {
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

