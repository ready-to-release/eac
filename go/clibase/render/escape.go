package render

import "strings"

// EscapeJinja2 escapes Jinja2 template syntax to prevent MkDocs macros from
// processing content as template expressions. Escapes {{ }}, {% %}.
func EscapeJinja2(s string) string {
	s = strings.ReplaceAll(s, "{{", "{ {")
	s = strings.ReplaceAll(s, "}}", "} }")
	s = strings.ReplaceAll(s, "{%", "{ %")
	s = strings.ReplaceAll(s, "%}", "% }")
	return s
}

// EscapeMarkdownHTML wraps bare <placeholder> tokens in backticks so they render
// as code instead of being interpreted as HTML tags by mkdocs.
// Preserves content already inside backtick code spans.
func EscapeMarkdownHTML(s string) string {
	var b strings.Builder
	inCode := false
	for i := 0; i < len(s); i++ {
		if s[i] == '`' {
			inCode = !inCode
			b.WriteByte('`')
			continue
		}
		if !inCode && s[i] == '<' {
			// Find matching > — wrap <placeholder> in backticks to prevent HTML interpretation.
			end := strings.IndexByte(s[i+1:], '>')
			if end > 0 {
				b.WriteByte('`')
				b.WriteString(s[i : i+2+end]) // includes < and >
				b.WriteByte('`')
				i += end + 1 // skip past >
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// EscapeTableCell escapes pipe characters and newlines for markdown table cells.
func EscapeTableCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
