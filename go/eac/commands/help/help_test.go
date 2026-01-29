package help

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

// TestPrintParentHelp tests help output for a parent command.
func TestPrintParentHelp(t *testing.T) {
	reg := &registry.CommandRegistration{
		ActualCommand: "get",
		Short:         "Retrieve repository data in structured format",
		IsParent:      true,
		Subcommands: []registry.SubcommandGroup{
			{Name: "Configuration", Subcommands: []string{"config"}},
			{Name: "Repository Structure", Subcommands: []string{"modules", "dependencies"}},
		},
		Examples: []string{
			"r2r get modules",
			"r2r get dependencies --format=yaml",
		},
	}

	// Create mock subcommand registrations
	subRegs := map[string]*registry.CommandRegistration{
		"get config":       {Short: "Get all EAC configuration"},
		"get modules":      {Short: "Get all module contracts"},
		"get dependencies": {Short: "Get module dependency graph"},
	}

	var buf bytes.Buffer
	PrintHelp(&buf, reg, subRegs)

	output := buf.String()

	// Check title
	if !strings.Contains(output, "get - Retrieve repository data in structured format") {
		t.Error("Expected title with command name and short description")
	}

	// Check usage
	if !strings.Contains(output, "Usage: r2r get <subcommand>") {
		t.Error("Expected usage line")
	}

	// Check groups are present
	if !strings.Contains(output, "Configuration:") {
		t.Error("Expected Configuration group")
	}
	if !strings.Contains(output, "Repository Structure:") {
		t.Error("Expected Repository Structure group")
	}

	// Check subcommands are listed with descriptions
	if !strings.Contains(output, "config") {
		t.Error("Expected config subcommand")
	}
	if !strings.Contains(output, "Get all EAC configuration") {
		t.Error("Expected config description")
	}

	// Check examples
	if !strings.Contains(output, "Examples:") {
		t.Error("Expected Examples section")
	}
	if !strings.Contains(output, "r2r get modules") {
		t.Error("Expected example 1")
	}
	if !strings.Contains(output, "r2r get dependencies --format=yaml") {
		t.Error("Expected example 2")
	}

	// Check footer
	if !strings.Contains(output, "Use 'r2r get <subcommand> --help'") {
		t.Error("Expected footer with help hint")
	}
}

// TestPrintLeafHelp tests help output for a leaf command (non-parent).
func TestPrintLeafHelp(t *testing.T) {
	reg := &registry.CommandRegistration{
		ActualCommand: "get modules",
		Short:         "Get all module contracts",
		Long:          "Retrieves all module contracts from the repository in structured format.",
		IsParent:      false,
		Flags: []registry.FlagMetadata{
			{Name: "format", Shorthand: "f", Type: "string", DefaultValue: "json", Usage: "Output format (json|yaml|table)"},
			{Name: "debug", Shorthand: "d", Type: "bool", Usage: "Enable debug output"},
		},
	}

	var buf bytes.Buffer
	PrintHelp(&buf, reg, nil)

	output := buf.String()

	// Check title
	if !strings.Contains(output, "get modules - Get all module contracts") {
		t.Error("Expected title with command name and short description")
	}

	// Check long description
	if !strings.Contains(output, "Retrieves all module contracts") {
		t.Error("Expected long description")
	}

	// Check flags
	if !strings.Contains(output, "Flags:") {
		t.Error("Expected Flags section")
	}
	if !strings.Contains(output, "-f, --format") {
		t.Error("Expected format flag with shorthand")
	}
	if !strings.Contains(output, "(default: json)") {
		t.Error("Expected default value for format")
	}
	if !strings.Contains(output, "-d, --debug") {
		t.Error("Expected debug flag with shorthand")
	}
}

// TestPrintHelpNoSubcommandDescriptions tests fallback when subcommand descriptions unavailable.
func TestPrintHelpNoSubcommandDescriptions(t *testing.T) {
	reg := &registry.CommandRegistration{
		ActualCommand: "work",
		Short:         "Manage work branches",
		IsParent:      true,
		Subcommands: []registry.SubcommandGroup{
			{Name: "Branch", Subcommands: []string{"create", "remove"}},
		},
	}

	var buf bytes.Buffer
	PrintHelp(&buf, reg, nil) // No subcommand registrations

	output := buf.String()

	// Subcommands should still be listed even without descriptions
	if !strings.Contains(output, "create") {
		t.Error("Expected create subcommand even without description")
	}
	if !strings.Contains(output, "remove") {
		t.Error("Expected remove subcommand even without description")
	}
}

// TestPrintHelpWithRequiredFlags tests that required flags are marked.
func TestPrintHelpWithRequiredFlags(t *testing.T) {
	reg := &registry.CommandRegistration{
		ActualCommand: "build",
		Short:         "Build modules",
		IsParent:      false,
		Flags: []registry.FlagMetadata{
			{Name: "module", Shorthand: "m", Type: "string", Required: true, Usage: "Module to build"},
		},
	}

	var buf bytes.Buffer
	PrintHelp(&buf, reg, nil)

	output := buf.String()

	if !strings.Contains(output, "[required]") {
		t.Error("Expected [required] marker for required flag")
	}
}

// TestPrintHelpEmptyExamples tests that examples section is omitted when empty.
func TestPrintHelpEmptyExamples(t *testing.T) {
	reg := &registry.CommandRegistration{
		ActualCommand: "simple",
		Short:         "A simple command",
		IsParent:      false,
		Examples:      nil,
	}

	var buf bytes.Buffer
	PrintHelp(&buf, reg, nil)

	output := buf.String()

	if strings.Contains(output, "Examples:") {
		t.Error("Expected no Examples section when examples are empty")
	}
}
