package content

import (
	"fmt"
	"strings"
)

// FormatDescription joins description lines, preserving paragraph breaks.
func FormatDescription(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	var paragraphs []string
	var currentParagraph []string

	for _, line := range lines {
		if line == "" {
			if len(currentParagraph) > 0 {
				paragraphs = append(paragraphs, strings.Join(currentParagraph, " "))
				currentParagraph = nil
			}
		} else {
			currentParagraph = append(currentParagraph, line)
		}
	}

	if len(currentParagraph) > 0 {
		paragraphs = append(paragraphs, strings.Join(currentParagraph, " "))
	}

	return strings.TrimSpace(strings.Join(paragraphs, "\n\n"))
}

// ParseFlagLine parses a flag/argument line like "  --dry-run    Description here".
func ParseFlagLine(line string) *FlagArg {
	if !strings.HasPrefix(line, "  ") {
		return nil
	}

	line = strings.TrimPrefix(line, "  ")
	if line == "" {
		return nil
	}

	parts := flagSplitPattern.Split(line, 2)
	if len(parts) < 1 {
		return nil
	}

	name := strings.TrimSpace(parts[0])
	desc := ""
	if len(parts) > 1 {
		desc = strings.TrimSpace(parts[1])
	}

	if name == "" {
		return nil
	}

	return &FlagArg{
		Name:        name,
		Description: desc,
	}
}

// FormatCommandHelp formats a CommandHelp as markdown.
func FormatCommandHelp(help *CommandHelp, headingLevel int, includeTitle bool) string {
	var sb strings.Builder
	heading := strings.Repeat("#", headingLevel)
	subHeading := strings.Repeat("#", headingLevel+1)

	if includeTitle {
		sb.WriteString(fmt.Sprintf("%s %s\n\n", heading, help.Name))
	}

	if help.Description != "" {
		sb.WriteString(EscapeJinja2(help.Description))
		sb.WriteString("\n\n")
	}

	if help.Usage != "" {
		sb.WriteString(fmt.Sprintf("**Usage:** `%s`\n\n", help.Usage))
	}

	if len(help.Arguments) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Arguments\n\n", subHeading))
		sb.WriteString("| Argument | Description |\n")
		sb.WriteString("|----------|-------------|\n")
		for _, arg := range help.Arguments {
			sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", arg.Name, EscapeTableCell(arg.Description)))
		}
		sb.WriteString("\n")
	}

	if len(help.Flags) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Flags\n\n", subHeading))
		sb.WriteString("| Flag | Description |\n")
		sb.WriteString("|------|-------------|\n")
		for _, flag := range help.Flags {
			sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", flag.Name, EscapeTableCell(flag.Description)))
		}
		sb.WriteString("\n")
	}

	if help.Notes != "" {
		sb.WriteString(fmt.Sprintf("\n%s Notes\n\n", subHeading))
		sb.WriteString(EscapeJinja2(help.Notes))
		sb.WriteString("\n\n")
	}

	if help.Examples != "" {
		sb.WriteString(fmt.Sprintf("\n%s Examples\n\n", subHeading))
		sb.WriteString("```bash\n")
		sb.WriteString(EscapeJinja2(DedentExamples(help.Examples)))
		sb.WriteString("\n```\n")
	}

	return sb.String()
}

// EscapeTableCell escapes pipe characters and newlines for markdown tables.
func EscapeTableCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// EscapeJinja2 escapes Jinja2 template syntax to prevent MkDocs macros from processing
// command help output as template expressions.
func EscapeJinja2(s string) string {
	s = strings.ReplaceAll(s, "{{", "{ {")
	s = strings.ReplaceAll(s, "}}", "} }")
	s = strings.ReplaceAll(s, "{%", "{ %")
	s = strings.ReplaceAll(s, "%}", "% }")
	return s
}

// DedentExamples removes common leading whitespace from example lines.
func DedentExamples(examples string) string {
	lines := strings.Split(strings.TrimSpace(examples), "\n")
	if len(lines) == 0 {
		return examples
	}

	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return strings.TrimSpace(examples)
	}

	var result []string
	for _, line := range lines {
		if len(line) >= minIndent {
			result = append(result, line[minIndent:])
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
