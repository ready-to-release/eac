package get

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/adapters/tui/selector"
	"github.com/ready-to-release/eac/go/commands/repository/internal/helputil"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/logging"
)

type getCommand struct{}

var _ core.CommandPort = (*getCommand)(nil)

func (c *getCommand) Name() string { return "get" }

func (c *getCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get",
		Short:         "Retrieve repository data in structured format",
		IsParent:      true,
		SubcommandGroups: []core.SubcommandGroup{
			{Name: "Configuration", Subcommands: []string{"config"}},
			{Name: "Repository Structure", Subcommands: []string{"modules", "components", "units", "dependencies"}},
			{Name: "Files and Changes", Subcommands: []string{"files", "files-by-module", "changed-modules", "changed-modules-ci", "changed-modules-local", "ghosts"}},
			{Name: "CI/CD", Subcommands: []string{"ci-dispatch", "ci-workflows"}},
			{Name: "Build", Subcommands: []string{"build-times", "build-deps", "artifacts"}},
			{Name: "Testing", Subcommands: []string{"test-results", "tests", "suite", "test-timings"}},
			{Name: "Utilities", Subcommands: []string{"token-size"}},
			{Name: "Commands", Subcommands: []string{"valid-commands", "documented-commands"}},
			{Name: "Release", Subcommands: []string{"release-bundle", "release-notes", "release-status"}},
			{Name: "Environment", Subcommands: []string{"environments"}},
			{Name: "Specifications", Subcommands: []string{"specs", "changelog"}},
			{Name: "Approvals", Subcommands: []string{"approval-comments"}},
			{Name: "Books", Subcommands: []string{"book-description"}},
			{Name: "Git", Subcommands: []string{"current-sha"}},
			{Name: "Evidence", Subcommands: []string{"evidence-ci-runs", "module-ci-workflow", "module-trigger-reason"}},
			{Name: "CLI", Subcommands: []string{"cli-release-notes"}},
		},
		Examples: []string{
			"clie get modules",
			"clie get dependencies",
			"clie get changed-modules",
			"clie get changed-modules-ci --as-json",
		},
	}
}

var log = logging.C()

// printHelp prints the help for the get command using registry metadata.
func printHelp() {
	cmd, _ := registry.Global().Get("get")
	helputil.PrintHelp(os.Stdout, cmd, registry.Global())
}

// Get command entry point.
func Get() int {
	args := os.Args[2:] // Skip program name and "get"

	// Check for help flag first
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printHelp()
		return 0
	}

	// If no subcommand and interactive, show TUI
	if len(args) == 0 {
		if tui.ShouldUseTUI("get", false, false) {
			return runInteractiveTUI()
		}
		printHelp()
		return 1
	}

	// Check for valid subcommand using dynamic discovery from registry
	if _, ok := registry.Global().Get("get " + args[0]); ok {
		// Handled by separate registrations in respective files
		return 0
	}

	fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n", args[0])
	printHelp()
	return 1
}

// runInteractiveTUI shows the interactive TUI for subcommand selection.
// Uses the new SelectorConsole pattern: TUI just picks a command, caller executes.
func runInteractiveTUI() int {
	// Get options dynamically from registry
	options := tui.SubcommandsFromRegistry("get")
	selected, args, cancelled := selector.RunSelector(context.Background(), options)

	if cancelled || selected == "" {
		return 0
	}

	newArgs := []string{os.Args[0], "get", selected}
	if args != "" {
		newArgs = append(newArgs, strings.Fields(args)...)
	}
	os.Args = newArgs
	return Get()
}

