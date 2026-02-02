// Command: get
// Short: Retrieve repository data in structured format
// IsParent: true
// Group.Configuration: config
// Group.Repository Structure: modules, components, units, dependencies, execution-order
// Group.Files and Changes: files, files-by-module, changed-modules, changed-modules-ci, changed-modules-local
// Group.CI/CD: ci-dispatch, ci-dispatch-layers, ci-workflows
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

// subcommands defines all available get subcommands.
var subcommands = []tui.SubcommandInfo{
	{Name: "config", Description: "Get all EAC configuration"},
	{Name: "modules", Description: "Get all module contracts"},
	{Name: "components", Description: "Get all components with phase and dependency info"},
	{Name: "units", Description: "Get units of work for a framework (build|test|lint|scan)"},
	{Name: "dependencies", Description: "Get module dependency graph"},
	{Name: "execution-order", Description: "Get build order for modules"},
	{Name: "files", Description: "Get repository files with module mappings"},
	{Name: "files-by-module", Description: "Get files grouped by module"},
	{Name: "changed-modules", Description: "Get modules affected by changes"},
	{Name: "changed-modules-ci", Description: "Get modules requiring CI rebuild"},
	{Name: "changed-modules-local", Description: "Get modules requiring local rebuild"},
	{Name: "ci-dispatch", Description: "Filter modules for CI dispatch"},
	{Name: "ci-dispatch-layers", Description: "Get CI dispatch layers"},
	{Name: "ci-workflows", Description: "Get CI workflow configurations"},
	{Name: "build-times", Description: "Get build timing data"},
	{Name: "build-deps", Description: "Get build dependencies for a module"},
	{Name: "artifacts", Description: "Get resolved artifacts for a module"},
	{Name: "test-results", Description: "Get test execution results"},
	{Name: "tests", Description: "Get all tests in structured format"},
	{Name: "suite", Description: "Get test suite information"},
	{Name: "test-timings", Description: "Get test timing data"},
	{Name: "token-size", Description: "Estimate token counts for files"},
	{Name: "valid-commands", Description: "Get all valid commands"},
	{Name: "documented-commands", Description: "Get commands documented in markdown"},
	{Name: "release-bundle", Description: "Get release bundle configuration"},
	{Name: "release-notes", Description: "Get release notes"},
	{Name: "release-status", Description: "Get release status"},
	{Name: "environments", Description: "Get all environment contracts"},
	{Name: "specs", Description: "Get specification files"},
	{Name: "changelog", Description: "Get changelog entries"},
	{Name: "approval-comments", Description: "Get approval comments"},
	{Name: "book-description", Description: "Get book descriptions"},
	{Name: "cli-release-notes", Description: "Get CLI release notes"},
	{Name: "current-sha", Description: "Get current git SHA"},
	{Name: "evidence-ci-runs", Description: "Get evidence CI runs"},
	{Name: "module-ci-workflow", Description: "Get module CI workflow"},
	{Name: "module-trigger-reason", Description: "Get module trigger reason"},
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

	// Check for valid subcommand
	switch args[0] {
	case "artifacts", "build-deps", "build-times", "changed-modules", "changed-modules-ci",
		"changed-modules-local", "ci-dispatch", "ci-dispatch-layers", "ci-workflows",
		"components", "config", "dependencies", "documented-commands", "environments",
		"execution-order", "files", "files-by-module", "modules", "release-bundle",
		"release-notes", "release-status", "suite", "test-results", "tests", "test-timings",
		"token-size", "units", "valid-commands", "specs", "changelog", "approval-comments",
		"book-description", "cli-release-notes", "current-sha", "evidence-ci-runs",
		"module-ci-workflow", "module-trigger-reason":
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

	newArgs := []string{os.Args[0], "get", selected}
	if args != "" {
		newArgs = append(newArgs, strings.Fields(args)...)
	}
	os.Args = newArgs
	return Get()
}

