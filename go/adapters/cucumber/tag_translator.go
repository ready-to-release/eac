package cucumber

import (
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// CucumberTagTranslator translates TagFilter to cucumber-js tag expression syntax.
// Cucumber-js syntax: (@tag1 or @tag2) (OR), @tag1 and @tag2 (AND), not @tag (NOT).
type CucumberTagTranslator struct{}

// TranslateTagFilter converts a TagFilter to a cucumber-js-compatible tag expression string.
func (t *CucumberTagTranslator) TranslateTagFilter(filter core.TagFilter) string {
	var parts []string

	for _, selector := range filter.Selectors {
		var selectorParts []string

		// RequireTags: each tag is AND'd
		for _, tag := range selector.RequireTags {
			selectorParts = append(selectorParts, tag)
		}

		// ExcludeTags: each becomes "not @tag"
		for _, tag := range selector.ExcludeTags {
			selectorParts = append(selectorParts, "not "+tag)
		}

		// AnyOfTags: joined with " or " inside parentheses
		if len(selector.AnyOfTags) > 0 {
			orExpr := "(" + strings.Join(selector.AnyOfTags, " or ") + ")"
			selectorParts = append(selectorParts, orExpr)
		}

		if len(selectorParts) > 0 {
			parts = append(parts, strings.Join(selectorParts, " and "))
		}
	}

	if len(parts) > 1 {
		return "(" + strings.Join(parts, ") or (") + ")"
	} else if len(parts) == 1 {
		return parts[0]
	}
	return ""
}
