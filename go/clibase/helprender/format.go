package helprender

import (
	"fmt"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/render"
)

// formatFlagDisplay formats a FlagSpec into a display string for the Flag column.
// Example outputs:
//
//	"--dry-run (optional)"
//	"-v, --verbose (optional)"
//	"--output <string> (optional, default: json)"
//	"--count <int> (required)"
func formatFlagDisplay(flag core.FlagSpec) string {
	var parts []string

	if flag.Shorthand != "" {
		parts = append(parts, "-"+flag.Shorthand+",")
	}

	name := "--" + flag.Name
	if flag.Type != "" && flag.Type != "bool" {
		name += " <" + flag.Type + ">"
	}
	parts = append(parts, name)

	var attrs []string
	if flag.Required {
		attrs = append(attrs, "required")
	} else {
		attrs = append(attrs, "optional")
	}
	if flag.DefaultValue != "" {
		attrs = append(attrs, "default: "+flag.DefaultValue)
	}
	parts = append(parts, "("+strings.Join(attrs, ", ")+")")

	return strings.Join(parts, " ")
}

// BuildSynopsis constructs the synopsis line from command name and metadata.
func BuildSynopsis(name string, meta core.CommandMetadata) string {
	synopsis := name
	if len(meta.Flags) > 0 {
		synopsis += " [flags]"
	}
	if meta.Args != "" {
		synopsis += " <" + meta.Args + ">"
	}
	return synopsis
}

// formatSubcommands renders subcommand groups as a markdown table with a --- separator.
// Uses SubcommandEntry keys for correct labels (handles aliases).
func formatSubcommands(entries []core.SubcommandEntry, parentName string) string {
	if len(entries) == 0 {
		return ""
	}

	tb := render.NewTableBuilder().
		WithMarkdown().
		WithHeaders("Command", "Description")
	for _, entry := range entries {
		subPart := strings.TrimPrefix(entry.Key, parentName+" ")
		desc := entry.Cmd.Metadata().Short
		if entry.Key != entry.Cmd.Name() {
			desc = desc + " (-> " + entry.Cmd.Name() + ")"
		}
		tb.AddRow(fmt.Sprintf("`%s`", subPart), desc)
	}

	return fmt.Sprintf("\n---\n\n%s\n\n", tb.BuildMarkdown())
}
