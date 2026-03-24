package validate

import (
	"context"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/adapters/tui/selector"
	"github.com/ready-to-release/eac/go/commands/repository/internal/helputil"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

type validateCommand struct{}

var _ core.CommandPort = (*validateCommand)(nil)

func (c *validateCommand) Name() string { return "validate" }

func (c *validateCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "validate",
		Short:         "Validate repository contracts and dependencies",
		IsParent:      true,
		SubcommandGroups: []core.SubcommandGroup{
			{Name: "Contracts", Subcommands: []string{"contracts", "dependencies", "books"}},
			{Name: "Code Quality", Subcommands: []string{"dependabot", "go-tidy", "markdown", "module-files", "module-hierarchy"}},
			{Name: "Documentation", Subcommands: []string{"docs"}},
			{Name: "Specifications", Subcommands: []string{"specs", "test-tags"}},
			{Name: "Risk", Subcommands: []string{"risk-catalog", "risk-profile"}},
			{Name: "Release", Subcommands: []string{"release", "release-version", "version"}},
			{Name: "Design", Subcommands: []string{"design"}},
			{Name: "Artifacts", Subcommands: []string{"artifacts", "control-tags"}},
		},
		Examples: []string{
			"eac validate contracts",
			"eac validate dependencies",
			"eac validate test-tags",
			"eac validate module-hierarchy",
		},
	}
}

// subcommands defines all available validate subcommands.
var subcommands = []tui.SubcommandInfo{
	{Name: "books", Description: "Validate books.yml configuration"},
	{Name: "config", Description: "Validate effective configuration from all sources"},
	{Name: "contracts", Description: "Validate repository contracts against JSON schemas"},
	{Name: "dependabot", Description: "Validate dependabot.yml covers all dependency sources"},
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
	cmd, _ := registry.Global().Get("validate")
	helputil.PrintHelp(os.Stdout, cmd, registry.Global())
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
	case "artifacts", "books", "config", "contracts", "control-tags", "dependabot",
		"dependencies", "design", "docs", "go-tidy", "markdown", "module-files",
		"module-hierarchy", "release", "release-version", "risk", "risk-catalog",
		"risk-profile", "specs", "test-tags", "version":
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

