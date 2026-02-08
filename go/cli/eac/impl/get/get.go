// Command: get
// Short: Retrieve repository data in structured format
// IsParent: true
// Group.Configuration: config
// Group.Repository Structure: modules, components, units, dependencies
// Group.Files and Changes: files, files-by-module, changed-modules, changed-modules-ci, changed-modules-local, ghosts
// Group.CI/CD: ci-dispatch, ci-workflows
// Group.Build: build-times, build-deps, artifacts
// Group.Testing: test-results, tests, suite, test-timings
// Group.Utilities: token-size
// Group.Commands: valid-commands, documented-commands
// Group.Release: release-bundle, release-notes, release-status
// Group.Environment: environments
// Group.Specifications: specs, changelog
// Group.Approvals: approval-comments
// Group.Books: book-description
// Group.Git: current-sha
// Group.Evidence: evidence-ci-runs, module-ci-workflow, module-trigger-reason
// Group.CLI: cli-release-notes
// Example: r2r get modules
// Example: r2r get dependencies
// Example: r2r get changed-modules
// Example: r2r get changed-modules-ci --as-json
package get

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/help"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/adapters/tui/selector"
	"github.com/ready-to-release/eac/go/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(Get)
}

// printHelp prints the help for the get command using registry metadata.
func printHelp() {
	reg := registry.GetCommand("get")
	help.PrintHelp(os.Stdout, reg, registry.GetCommandRegistry())
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
	if registry.IsValidSubcommand("get", args[0]) {
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

