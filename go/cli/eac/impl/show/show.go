// Command: show
// Short: Display repository information in human-readable format
// IsParent: true
// Group.Configuration: config
// Group.Documentation: books
// Group.Repository Structure: modules, components, units, dependencies
// Group.Files and Changes: files, files-changed, files-staged
// Group.Workspaces: workspaces
// Group.Build: build-summary, build-times
// Group.Testing: test-results, test-summary, tests, suite, test-timings
// Group.Environment: environments
// Group.Commands: valid-commands
// Group.Specifications: specs, changelog
// Group.Approvals: approval-comments, approve-summary
// Group.Artifacts: artifacts, component-types
// Group.Lint: lint-summary
// Group.Scan: scan-summary
// Group.Release: release-summary, release-notes
// Group.CI: ci-summary, dependency-ci-summary, deps-setup-summary, trigger-summary
// Example: r2r show modules
// Example: r2r show files-changed
// Example: r2r show suite integration
package show

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
	registry.Register(Show)
}

// subcommands defines all available show subcommands.
var subcommands = []tui.SubcommandInfo{
	{Name: "config", Description: "Show all EAC configuration summary"},
	{Name: "books", Description: "Show all configured books"},
	{Name: "modules", Description: "Show all module contracts"},
	{Name: "components", Description: "Show all components with phase and dependency info"},
	{Name: "units", Description: "Show units of work for a framework (build|test|lint|scan)"},
	{Name: "dependencies", Description: "Show module dependencies"},
	{Name: "files", Description: "Show repository files with module ownership"},
	{Name: "files-changed", Description: "Show modified files with module ownership"},
	{Name: "files-staged", Description: "Show staged files with module ownership"},
	{Name: "workspaces", Description: "Show all git worktrees and their status"},
	{Name: "build-summary", Description: "Generate pretty build summary for GitHub Actions"},
	{Name: "build-times", Description: "Show build timing analysis"},
	{Name: "test-results", Description: "Display test execution results"},
	{Name: "test-summary", Description: "Generate pretty test summary for GitHub Actions"},
	{Name: "tests", Description: "Show all tests in the repository"},
	{Name: "suite", Description: "Show detailed test suite information"},
	{Name: "test-timings", Description: "Show test timing analysis"},
	{Name: "environments", Description: "Show all environment contracts"},
	{Name: "valid-commands", Description: "Show all valid commands in a table"},
	{Name: "specs", Description: "Show specification files"},
	{Name: "changelog", Description: "Show changelog entries"},
	{Name: "approval-comments", Description: "Show approval comments"},
	{Name: "artifacts", Description: "Show artifacts"},
	{Name: "component-types", Description: "Show component types"},
	{Name: "lint-summary", Description: "Show lint summary"},
	{Name: "scan-summary", Description: "Show scan summary"},
	{Name: "release-summary", Description: "Show release summary"},
	{Name: "release-notes", Description: "Show release notes"},
	{Name: "trigger-summary", Description: "Show trigger summary"},
	{Name: "ci-summary", Description: "Show CI summary"},
	{Name: "dependency-ci-summary", Description: "Show dependency CI summary"},
	{Name: "deps-setup-summary", Description: "Show deps setup summary"},
	{Name: "approve-summary", Description: "Show approve summary"},
	{Name: "help", Description: "Show help information"},
}

// printHelp prints the help for the show command using registry metadata.
func printHelp() {
	reg := registry.GetCommand("show")
	help.PrintHelp(os.Stdout, reg, registry.GetCommandRegistry())
}

// Show command entry point.
func Show() int {
	args := os.Args[2:] // Skip program name and "show"

	// Check for help flag first
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printHelp()
		return 0
	}

	// If no subcommand and interactive, show TUI
	if len(args) == 0 {
		if tui.ShouldUseTUI("show", false, false) {
			return runInteractiveTUI()
		}
		printHelp()
		return 1
	}

	// Check for valid subcommand
	switch args[0] {
	case "approval-comments", "approve-summary", "artifacts", "books", "build-summary",
		"build-times", "changelog", "ci-summary", "component-types", "components", "config",
		"dependencies", "dependency-ci-summary", "deps-setup-summary", "environments",
		"files", "files-changed", "files-staged", "help", "lint-summary", "modules",
		"release-notes", "release-summary", "scan-summary", "specs", "suite",
		"test-results", "test-summary", "test-timings", "tests", "trigger-summary",
		"units", "valid-commands", "workspaces":
		// Handled by separate registrations in respective files
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n", args[0])
		printHelp()
		return 1
	}
}

// runInteractiveTUI shows the interactive TUI for subcommand selection.
// Uses the new SelectorConsole pattern: TUI just picks a command, caller executes.
func runInteractiveTUI() int {
	options := tui.SubcommandsToOptions(subcommands)
	selected, args, cancelled := selector.RunSelector(context.Background(), options)

	if cancelled || selected == "" {
		return 0
	}

	newArgs := []string{os.Args[0], "show", selected}
	if args != "" {
		newArgs = append(newArgs, strings.Fields(args)...)
	}
	os.Args = newArgs
	return Show()
}

