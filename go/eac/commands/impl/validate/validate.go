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
// Example: r2r validate contracts
// Example: r2r validate dependencies
// Example: r2r validate test-tags
// Example: r2r validate module-hierarchy
package validate

import (
	"context"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/help"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/commands/tui"

	// Import default TUI to register it
	_ "github.com/ready-to-release/eac/go/eac/commands/tui/default"
)

func init() {
	registry.Register(Validate)
}

// subcommands defines all available validate subcommands.
var subcommands = []tui.SubcommandInfo{
	{Name: "books", Description: "Validate books.yml configuration"},
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
	case "artifacts", "books", "contracts", "control-tags", "dependencies",
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
func runInteractiveTUI() int {
	console := tui.NewForCommand("validate", tui.Config{
		CommandName: "validate",
		Height:      20,
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
			newArgs := []string{os.Args[0], "validate", selected}
			if args, ok := params["args"]; ok && args != "" {
				newArgs = append(newArgs, args)
			}
			os.Args = newArgs
			return Validate()
		}
	}

	return 0
}

