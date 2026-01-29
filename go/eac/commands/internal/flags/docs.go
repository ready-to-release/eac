package flags

import (
	"fmt"
	"strings"
)

// CommandDoc describes a command and its flag subscriptions.
type CommandDoc struct {
	Name        string
	Description string
	Execution   bool
	Output      bool
	Cache       bool
	Module      bool
	DryRun      bool
	// SkipDepmSupported indicates if --skip-depm is functional for this command.
	// If false and Module is true, --skip-depm throws "not implemented".
	SkipDepmSupported bool
	// SpecificFlags are command-specific flags not in shared flag sets.
	SpecificFlags []FlagDef
}

// FlagDocGenerator generates documentation for flag sets.
type FlagDocGenerator struct {
	commands []CommandDoc
}

// NewFlagDocGenerator creates a new documentation generator.
func NewFlagDocGenerator() *FlagDocGenerator {
	return &FlagDocGenerator{
		commands: []CommandDoc{
			{
				Name:              "build",
				Description:       "Build components",
				Execution:         true,
				Output:            true,
				Cache:             true,
				Module:            true,
				DryRun:            true,
				SkipDepmSupported: true,
				SpecificFlags: []FlagDef{
					{Name: "tidy-first", Type: "bool", Usage: "Run go mod tidy before building"},
					{Name: "layered-build", Type: "bool", Usage: "Enable layered build mode"},
					{Name: "version", Type: "string", Usage: "Version string for build"},
					{Name: "reproducible", Type: "bool", Usage: "Enable reproducible builds"},
					{Name: "all", Type: "bool", Usage: "Build all components"},
				},
			},
			{
				Name:              "test",
				Description:       "Run tests",
				Execution:         true,
				Output:            true,
				Cache:             true,
				Module:            true,
				DryRun:            true,
				SkipDepmSupported: true,
				SpecificFlags: []FlagDef{
					{Name: "tags", Type: "string", Usage: "Filter test suites by tags (e.g., @smoke, @integration)"},
					{Name: "list-only", Type: "bool", Usage: "List tests without running them"},
					{Name: "suite", Type: "string", Usage: "Run specific test suite"},
					{Name: "coverage", Type: "bool", Usage: "Generate coverage report"},
				},
			},
			{
				Name:              "lint",
				Description:       "Run linters",
				Execution:         true,
				Output:            true,
				Cache:             true,
				Module:            true,
				DryRun:            true,
				SkipDepmSupported: false, // --skip-depm not implemented
			},
			{
				Name:              "scan",
				Description:       "Run security scans",
				Execution:         true,
				Output:            true,
				Cache:             true,
				Module:            true,
				DryRun:            true,
				SkipDepmSupported: false, // --skip-depm not implemented
			},
		},
	}
}

// GenerateFlagMatrix generates a markdown table showing flag availability.
func (g *FlagDocGenerator) GenerateFlagMatrix() string {
	var sb strings.Builder

	sb.WriteString("# Shared Flags\n\n")
	sb.WriteString("This document describes the flags shared across `build`, `test`, `lint`, and `scan` commands.\n\n")

	// Execution flags
	sb.WriteString("## Execution Control\n\n")
	sb.WriteString("Control parallel execution behavior.\n\n")
	g.writeFlagSetTable(&sb, "execution")

	// Output flags
	sb.WriteString("\n## Output Control\n\n")
	sb.WriteString("Control terminal user interface and output display.\n\n")
	g.writeFlagSetTable(&sb, "output")

	// Cache flags
	sb.WriteString("\n## Cache Control\n\n")
	sb.WriteString("Control incremental caching and dependency verification.\n\n")
	g.writeFlagSetTable(&sb, "cache")

	// Module flags
	sb.WriteString("\n## Module Selection\n\n")
	sb.WriteString("Control module selection and dependency processing.\n\n")
	g.writeFlagSetTable(&sb, "module")

	// DryRun flags
	sb.WriteString("\n## Dry Run\n\n")
	sb.WriteString("Simulate execution without making changes.\n\n")
	g.writeFlagSetTable(&sb, "dryrun")

	// Command availability matrix
	sb.WriteString("\n## Command Availability Matrix\n\n")
	g.writeAvailabilityMatrix(&sb)

	// Command-specific flags
	sb.WriteString("\n## Command-Specific Flags\n\n")
	sb.WriteString("These flags are specific to individual commands and not shared.\n\n")
	g.writeCommandSpecificFlags(&sb)

	return sb.String()
}

func (g *FlagDocGenerator) writeFlagSetTable(sb *strings.Builder, setName string) {
	var set FlagSet
	switch setName {
	case "execution":
		set = NewExecutionFlagSet()
	case "output":
		set = NewOutputFlagSet()
	case "cache":
		set = NewCacheFlagSet()
	case "module":
		set = NewModuleFlagSet()
	case "dryrun":
		set = NewDryRunFlagSet()
	default:
		return
	}

	flags := set.Flags()

	sb.WriteString("| Flag | Type | Default | Description |\n")
	sb.WriteString("|------|------|---------|-------------|\n")

	for _, f := range flags {
		flagName := fmt.Sprintf("`--%s`", f.Name)
		if f.Shorthand != "" {
			flagName += fmt.Sprintf(", `-%s`", f.Shorthand)
		}

		defaultVal := f.Default
		if defaultVal == "" || defaultVal == "false" {
			defaultVal = "-"
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			flagName, f.Type, defaultVal, f.Usage))
	}
}

func (g *FlagDocGenerator) writeAvailabilityMatrix(sb *strings.Builder) {
	sb.WriteString("| Flag Set | build | test | lint | scan |\n")
	sb.WriteString("|----------|-------|------|------|------|\n")

	sets := []struct {
		name  string
		check func(c CommandDoc) bool
	}{
		{"Execution", func(c CommandDoc) bool { return c.Execution }},
		{"Output", func(c CommandDoc) bool { return c.Output }},
		{"Cache", func(c CommandDoc) bool { return c.Cache }},
		{"Module", func(c CommandDoc) bool { return c.Module }},
		{"DryRun", func(c CommandDoc) bool { return c.DryRun }},
	}

	for _, s := range sets {
		sb.WriteString(fmt.Sprintf("| %s |", s.name))
		for _, cmd := range g.commands {
			if s.check(cmd) {
				sb.WriteString(" Yes |")
			} else {
				sb.WriteString(" No |")
			}
		}
		sb.WriteString("\n")
	}

	// Add skip-depm note
	sb.WriteString("\n**Note**: `--skip-depm` is only functional for `build` and `test` commands. ")
	sb.WriteString("Using it with `lint` or `scan` will result in a \"not implemented\" error.\n")
}

func (g *FlagDocGenerator) writeCommandSpecificFlags(sb *strings.Builder) {
	for _, cmd := range g.commands {
		if len(cmd.SpecificFlags) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("### %s\n\n", cmd.Name))
		sb.WriteString("| Flag | Type | Description |\n")
		sb.WriteString("|------|------|-------------|\n")

		for _, f := range cmd.SpecificFlags {
			sb.WriteString(fmt.Sprintf("| `--%s` | %s | %s |\n", f.Name, f.Type, f.Usage))
		}
		sb.WriteString("\n")
	}
}

// GetCommandDoc returns documentation for a specific command.
func (g *FlagDocGenerator) GetCommandDoc(name string) *CommandDoc {
	for i := range g.commands {
		if g.commands[i].Name == name {
			return &g.commands[i]
		}
	}
	return nil
}

// GenerateCommandUsage generates usage text for a specific command.
func (g *FlagDocGenerator) GenerateCommandUsage(name string) string {
	cmd := g.GetCommandDoc(name)
	if cmd == nil {
		return ""
	}

	config := CommandFlagConfig{
		Command:   name,
		Execution: cmd.Execution,
		Output:    cmd.Output,
		Cache:     cmd.Cache,
		Module:    cmd.Module,
		DryRun:    cmd.DryRun,
	}

	parser := NewParserWithEnv(config, nil)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Usage: eac %s [flags] [modules...]\n", name))
	sb.WriteString(fmt.Sprintf("\n%s\n", cmd.Description))
	sb.WriteString("\nShared Flags:")
	sb.WriteString(parser.GenerateUsage())

	if len(cmd.SpecificFlags) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s-specific Flags:\n", strings.Title(name)))
		for _, f := range cmd.SpecificFlags {
			flagStr := fmt.Sprintf("  --%s", f.Name)
			if f.Type != "bool" {
				flagStr += fmt.Sprintf(" <%s>", f.Type)
			}
			for len(flagStr) < 28 {
				flagStr += " "
			}
			flagStr += f.Usage
			sb.WriteString(flagStr + "\n")
		}
	}

	return sb.String()
}
