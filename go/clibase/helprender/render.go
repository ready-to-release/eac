package helprender

import (
	"fmt"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/render"
)

// RenderMarkdownHelp renders a command's metadata as heading-free markdown documentation.
// Output contains no '#' heading markers and no standalone bold lines.
// Section separators use horizontal rules (---).
func RenderMarkdownHelp(cmd core.CommandPort, reg core.CommandRegistryPort) string {
	if cmd == nil {
		return ""
	}

	meta := cmd.Metadata()
	name := cmd.Name()

	var sb strings.Builder

	// Description
	desc := strings.TrimSpace(meta.Long)
	if desc == "" {
		desc = meta.Short
	}
	if desc != "" {
		sb.WriteString(render.EscapeMarkdownHTML(render.EscapeJinja2(reflowForMarkdown(desc))))
		sb.WriteString("\n\n")
	}

	// Usage/Synopsis
	synopsis := BuildSynopsis(name, meta)
	if synopsis != "" {
		sb.WriteString(fmt.Sprintf("**Usage:** `%s`\n\n", synopsis))
	}

	// Subcommands (if parent command with registry)
	if reg != nil {
		entries := reg.SubcommandEntries(name)
		sb.WriteString(formatSubcommands(entries, name))
	}

	// Arguments (from meta.Args)
	if meta.Args != "" {
		sb.WriteString("\n---\n\n")
		tb := render.NewTableBuilder().
			WithMarkdown().
			WithHeaders("Argument", "Description").
			AddRow(fmt.Sprintf("`%s`", meta.Args), "")
		sb.WriteString(tb.BuildMarkdown())
		sb.WriteString("\n\n")
	}

	// Flags (from structured FlagSpec)
	if len(meta.Flags) > 0 {
		sb.WriteString("\n---\n\n")
		tb := render.NewTableBuilder().
			WithMarkdown().
			WithHeaders("Flag", "Description")
		for _, flag := range meta.Flags {
			tb.AddRow(
				fmt.Sprintf("`%s`", formatFlagDisplay(flag)),
				render.EscapeMarkdownHTML(render.EscapeJinja2(flag.Usage)),
			)
		}
		sb.WriteString(tb.BuildMarkdown())
		sb.WriteString("\n\n")
	}

	// Notes
	if meta.Notes != "" {
		sb.WriteString("\n---\n\n")
		sb.WriteString(render.EscapeMarkdownHTML(render.EscapeJinja2(reflowForMarkdown(meta.Notes))))
		sb.WriteString("\n\n")
	}

	// Examples
	if len(meta.Examples) > 0 {
		examples := strings.Join(meta.Examples, "\n")
		sb.WriteString("\n---\n\n")
		sb.WriteString("```bash\n")
		sb.WriteString(render.EscapeJinja2(examples))
		sb.WriteString("\n```\n")
	}

	return sb.String()
}

// reflowForMarkdown converts terminal-wrapped text to markdown-friendly text.
// Sentence boundaries (". \n") become paragraph breaks for readability.
// Mid-sentence soft wraps become spaces so markdown reflows naturally.
// Explicit paragraph breaks (\n\n) and list items (\n- ) are preserved.
func reflowForMarkdown(s string) string {
	// \x00 = paragraph break (becomes \n\n)
	// \x01 = single newline (stays \n, used for list items)

	// Preserve explicit paragraph breaks.
	s = strings.ReplaceAll(s, "\n\n", "\x00")
	// List items: preserve as single newlines (no blank lines between items),
	// but ensure a blank line before the first item.
	// Handles "\n  - " (indented), "\n- " (flush), and "\nN. " (numbered).
	s = strings.ReplaceAll(s, "\n  - ", "\x01- ")
	s = strings.ReplaceAll(s, "\n- ", "\x01- ")
	for d := '0'; d <= '9'; d++ {
		s = strings.ReplaceAll(s, "\n  "+string(d)+". ", "\x01"+string(d)+". ")
		s = strings.ReplaceAll(s, "\n"+string(d)+". ", "\x01"+string(d)+". ")
	}
	// Replace remaining soft-wrap newlines with spaces (join lines first).
	// Also collapse any leading whitespace on continuation lines.
	s = strings.ReplaceAll(s, "\n  ", "\n")
	s = strings.ReplaceAll(s, "\n", " ")
	// Sentence boundaries: ". Capital" / "). Capital" / ": Capital" → paragraph break.
	// Skip ". Capital" when preceded by a digit (numbered list items like "1. Foo").
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if i+2 < len(s) && s[i+1] == ' ' && s[i+2] >= 'A' && s[i+2] <= 'Z' {
			isSentenceEnd := false
			switch s[i] {
			case '.':
				// ". Capital" — but not "1. Capital" (numbered list)
				isSentenceEnd = i == 0 || s[i-1] < '0' || s[i-1] > '9'
			case ')':
				// "). Capital"
				isSentenceEnd = true
			case ':':
				// ": Capital" — only when followed by a newline-separated block.
				// Don't split here; colons introduce lists/content that should stay attached.
			}
			if isSentenceEnd {
				result.WriteByte(s[i])
				result.WriteString("\n\n")
				i++ // skip the space
				continue
			}
		}
		result.WriteByte(s[i])
	}
	s = result.String()
	// Restore paragraph breaks.
	s = strings.ReplaceAll(s, "\x00", "\n\n")
	// Restore list item newlines.
	s = strings.ReplaceAll(s, "\x01", "\n")
	// Ensure blank line before the first list item in each run.
	lines := strings.Split(s, "\n")
	var out []string
	prevIsList := false
	for i, line := range lines {
		cur := isListItem(line)
		if cur && !prevIsList && i > 0 && lines[i-1] != "" {
			out = append(out, "") // blank line before first list item
		}
		out = append(out, line)
		prevIsList = cur
	}
	return strings.Join(out, "\n")
}

func isListItem(line string) bool {
	if strings.HasPrefix(line, "- ") {
		return true
	}
	if len(line) >= 3 && line[0] >= '0' && line[0] <= '9' && line[1] == '.' && line[2] == ' ' {
		return true
	}
	return false
}
