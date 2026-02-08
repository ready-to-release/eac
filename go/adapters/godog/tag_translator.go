package godog

import (
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// GodogTagTranslator translates TagFilter to godog tag expression syntax.
//
// # CRITICAL: Godog Tag Expression Syntax
//
// Godog's tag parser has specific syntax requirements. Using incorrect syntax
// (like parentheses or "||") causes godog to SILENTLY return zero scenarios
// without any error, which can cause CI to falsely pass.
//
// Correct syntax:
//   - @tag1,@tag2    → OR  (comma, no space)
//   - @tag1 && @tag2 → AND (double ampersand with spaces)
//   - ~@tag          → NOT (tilde prefix)
//
// WRONG syntax (causes silent failure):
//   - (@tag1 || @tag2)  ← parentheses break the parser
//   - @tag1 || @tag2    ← "||" is not recognized
//   - @tag1 or @tag2    ← "or" keyword not supported
type GodogTagTranslator struct{}

// TranslateTagFilter converts a TagFilter to a godog-compatible tag expression string.
func (t *GodogTagTranslator) TranslateTagFilter(filter core.TagFilter) string {
	var parts []string

	for _, selector := range filter.Selectors {
		var selectorParts []string

		// RequireTags: each tag is AND'd
		selectorParts = append(selectorParts, selector.RequireTags...)

		// ExcludeTags: each becomes ~@tag
		for _, tag := range selector.ExcludeTags {
			selectorParts = append(selectorParts, "~"+tag)
		}

		// AnyOfTags: joined with comma (OR)
		if len(selector.AnyOfTags) > 0 {
			orExpr := strings.Join(selector.AnyOfTags, ",")
			selectorParts = append(selectorParts, orExpr)
		}

		if len(selectorParts) > 0 {
			parts = append(parts, strings.Join(selectorParts, " && "))
		}
	}

	if len(parts) > 1 {
		return strings.Join(parts, ",")
	} else if len(parts) == 1 {
		return parts[0]
	}
	return ""
}
