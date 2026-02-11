package help

import (
	"bytes"
	"strings"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

// mockCommand implements core.CommandPort for testing.
type mockCommand struct {
	name     string
	metadata core.CommandMetadata
}

func (m *mockCommand) Name() string              { return m.name }
func (m *mockCommand) Metadata() core.CommandMetadata { return m.metadata }

// mockRegistry creates a registry with the given commands for testing.
func mockRegistry(cmds ...core.CommandPort) core.CommandRegistryPort {
	reg := registry.NewCommandRegistry()
	for _, cmd := range cmds {
		reg.MustRegister(cmd)
	}
	return reg
}

// TestPrintParentHelp tests help output for a parent command.
func TestPrintParentHelp(t *testing.T) {
	cmd := &mockCommand{
		name: "get",
		metadata: core.CommandMetadata{
			CanonicalName: "get",
			Short:         "Retrieve repository data in structured format",
			IsParent:      true,
			SubcommandGroups: []core.SubcommandGroup{
				{Name: "Configuration", Subcommands: []string{"config"}},
				{Name: "Repository Structure", Subcommands: []string{"modules", "dependencies"}},
			},
			Examples: []string{
				"clie get modules",
				"clie get dependencies --format=yaml",
			},
		},
	}

	// Create mock subcommand registrations
	subCmds := mockRegistry(
		&mockCommand{name: "get config", metadata: core.CommandMetadata{Short: "Get all EAC configuration"}},
		&mockCommand{name: "get modules", metadata: core.CommandMetadata{Short: "Get all module contracts"}},
		&mockCommand{name: "get dependencies", metadata: core.CommandMetadata{Short: "Get module dependency graph"}},
	)

	var buf bytes.Buffer
	PrintHelp(&buf, cmd, subCmds)

	output := buf.String()

	// Check title
	if !strings.Contains(output, "get - Retrieve repository data in structured format") {
		t.Error("Expected title with command name and short description")
	}

	// Check usage
	if !strings.Contains(output, "Usage: clie get <subcommand>") {
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
	if !strings.Contains(output, "clie get modules") {
		t.Error("Expected example 1")
	}
	if !strings.Contains(output, "clie get dependencies --format=yaml") {
		t.Error("Expected example 2")
	}

	// Check footer
	if !strings.Contains(output, "Use 'clie get <subcommand> --help'") {
		t.Error("Expected footer with help hint")
	}
}

// TestPrintLeafHelp tests help output for a leaf command (non-parent).
func TestPrintLeafHelp(t *testing.T) {
	cmd := &mockCommand{
		name: "get modules",
		metadata: core.CommandMetadata{
			CanonicalName: "get-modules",
			Short:         "Get all module contracts",
			Long:          "Retrieves all module contracts from the repository in structured format.",
			IsParent:      false,
			Flags: []core.FlagSpec{
				{Name: "format", Shorthand: "f", Type: "string", DefaultValue: "json", Usage: "Output format (json|yaml|table)"},
				{Name: "debug", Shorthand: "d", Type: "bool", Usage: "Enable debug output"},
			},
		},
	}

	var buf bytes.Buffer
	PrintHelp(&buf, cmd, nil)

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
	cmd := &mockCommand{
		name: "work",
		metadata: core.CommandMetadata{
			CanonicalName: "work",
			Short:         "Manage work branches",
			IsParent:      true,
			SubcommandGroups: []core.SubcommandGroup{
				{Name: "Branch", Subcommands: []string{"create", "remove"}},
			},
		},
	}

	var buf bytes.Buffer
	PrintHelp(&buf, cmd, nil) // No subcommand registrations

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
	cmd := &mockCommand{
		name: "build",
		metadata: core.CommandMetadata{
			CanonicalName: "build",
			Short:         "Build modules",
			IsParent:      false,
			Flags: []core.FlagSpec{
				{Name: "module", Shorthand: "m", Type: "string", Required: true, Usage: "Module to build"},
			},
		},
	}

	var buf bytes.Buffer
	PrintHelp(&buf, cmd, nil)

	output := buf.String()

	if !strings.Contains(output, "[required]") {
		t.Error("Expected [required] marker for required flag")
	}
}

// TestPrintHelpEmptyExamples tests that examples section is omitted when empty.
func TestPrintHelpEmptyExamples(t *testing.T) {
	cmd := &mockCommand{
		name: "simple",
		metadata: core.CommandMetadata{
			CanonicalName: "simple",
			Short:         "A simple command",
			IsParent:      false,
			Examples:      nil,
		},
	}

	var buf bytes.Buffer
	PrintHelp(&buf, cmd, nil)

	output := buf.String()

	if strings.Contains(output, "Examples:") {
		t.Error("Expected no Examples section when examples are empty")
	}
}

// TestPrintHelpWithDeclarativeFlags tests that behavior flags are grouped separately.
func TestPrintHelpWithDeclarativeFlags(t *testing.T) {
	cmd := &mockCommand{
		name: "build",
		metadata: core.CommandMetadata{
			CanonicalName: "build",
			Short:         "Build modules",
			IsParent:      false,
			Flags: []core.FlagSpec{
				{Name: "format", Shorthand: "f", Type: "string", DefaultValue: "json", Usage: "Output format"},
				{Name: "with-cache", Type: "bool", DefaultValue: "true", Usage: "Enable incremental caching"},
				{Name: "no-cache", Type: "bool", DefaultValue: "false", Usage: "Disable incremental caching"},
				{Name: "debug", Shorthand: "d", Type: "bool", Usage: "Enable debug output"},
			},
		},
	}

	var buf bytes.Buffer
	PrintHelp(&buf, cmd, nil)

	output := buf.String()

	// Check that behavior flags section appears
	if !strings.Contains(output, "Behavior Flags:") {
		t.Error("Expected 'Behavior Flags:' section")
	}

	// Check that flag pair is shown
	if !strings.Contains(output, "--with-cache / --no-cache") {
		t.Error("Expected behavior flag pair format")
	}

	// Check that regular Flags section still appears
	if !strings.Contains(output, "Flags:") {
		t.Error("Expected 'Flags:' section for regular flags")
	}

	// Check that regular flags are listed separately
	if !strings.Contains(output, "--format") {
		t.Error("Expected format flag in regular flags")
	}
	if !strings.Contains(output, "--debug") {
		t.Error("Expected debug flag in regular flags")
	}

	// Behavior Flags should appear before regular Flags
	behaviorIdx := strings.Index(output, "Behavior Flags:")
	regularIdx := strings.Index(output, "Flags:")
	if behaviorIdx > regularIdx {
		t.Error("Expected Behavior Flags to appear before regular Flags")
	}
}

// TestPrintHelpOnlyBehaviorFlags tests help output with only behavior flags (no regular flags).
func TestPrintHelpOnlyBehaviorFlags(t *testing.T) {
	cmd := &mockCommand{
		name: "test",
		metadata: core.CommandMetadata{
			CanonicalName: "test",
			Short:         "Run tests",
			IsParent:      false,
			Flags: []core.FlagSpec{
				{Name: "with-parallel", Type: "bool", DefaultValue: "true", Usage: "Parallel execution"},
				{Name: "no-parallel", Type: "bool", DefaultValue: "false"},
			},
		},
	}

	var buf bytes.Buffer
	PrintHelp(&buf, cmd, nil)

	output := buf.String()

	// Should have Behavior Flags section
	if !strings.Contains(output, "Behavior Flags:") {
		t.Error("Expected 'Behavior Flags:' section")
	}

	// Should NOT have regular Flags section (since there are none)
	// Count occurrences of "Flags:" - should only appear once in "Behavior Flags:"
	flagsCount := strings.Count(output, "Flags:")
	if flagsCount != 1 {
		t.Errorf("Expected exactly 1 'Flags:' occurrence (in 'Behavior Flags:'), got %d", flagsCount)
	}
}
