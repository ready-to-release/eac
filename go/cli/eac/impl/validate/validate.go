// Command: validate
// Short: Validate repository contracts and dependencies
// IsParent: true
// Group.Contracts: contracts, dependencies, books
// Group.Code Quality: go-tidy, markdown, module-files, module-hierarchy
// Group.Documentation: docs
// Group.Specifications: specs, test-tags
// Group.Risk: risk-catalog, risk-profile
// Group.Release: release, release-version, version
// Group.Design: design
// Group.Artifacts: artifacts, control-tags
// Example: clie validate contracts
// Example: clie validate dependencies
// Example: clie validate test-tags
// Example: clie validate module-hierarchy
package validate

import (
	"context"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/help"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/adapters/tui/selector"
)

// subcommands defines all available validate subcommands.
var subcommands = []tui.SubcommandInfo{
	{Name: "books", Description: "Validate books.yml configuration"},
	{Name: "config", Description: "Validate effective configuration from all sources"},
	{Name: "contracts", Description: "Validate repository contracts against JSON schemas"},
	{Name: "dependencies", Description: "Validate module dependency contracts"},
	{Name: "docs", Description: "Validate documentation for obsolete references"},
	{Name: "go-tidy", Description: "Validate Go module dependencies are tidy"},
	{Name: "markdown", Description: "Validate markdown file syntax"},
	{Name: "module-files", Description: "Validate module file ownership"},
	{Name: "module-hierarchy", Description: "Validate module dependency graph structure"},
	{Name: "release-version", Description: "Validate release version format (semver)"},
	{Name: "risk-catalog", Description: "Validate OSCAL risk catalog"},
	{Name: "risk-profile", Description: "Validate OSCAL risk profiles"},
	{Name: "specs", Description: "Validate specification files"},
	{Name: "test-tags", Description: "Validate that all test tags are defined"},
	{Name: "artifacts", Description: "Validate artifacts"},
	{Name: "control-tags", Description: "Validate control tags"},
	{Name: "design", Description: "Validate design documents"},
	{Name: "release", Description: "Validate release configuration"},
	{Name: "version", Description: "Validate version format"},
}

// printHelp prints the help for the validate command using registry metadata.
func printHelp() {
	reg := registry.GetCommand("validate")
	help.PrintHelp(os.Stdout, reg, registry.GetCommandRegistry())
}

// Validate command entry point.
func Validate() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[2:] // Skip program name and "validate"

	// Check for help flag first
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printHelp()
		return 0
	}

	// If no subcommand and interactive, show TUI
	if len(args) == 0 {
		if tui.ShouldUseTUI("validate", false, false) {
			return runInteractiveTUI()
		}
		printHelp()
		return 1
	}

	// Check for valid subcommand
	switch args[0] {
	case "artifacts", "books", "config", "contracts", "control-tags", "dependencies",
		"design", "docs", "go-tidy", "markdown", "module-files", "module-hierarchy",
		"release", "release-version", "risk", "risk-catalog", "risk-profile",
		"specs", "test-tags", "version":
		// Handled by separate registrations in respective files
		return 0
	default:
		log.Errorf("Error: unknown subcommand: %s\n", args[0])
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

	newArgs := []string{os.Args[0], "validate", selected}
	if args != "" {
		newArgs = append(newArgs, strings.Fields(args)...)
	}
	os.Args = newArgs
	return Validate()
}

