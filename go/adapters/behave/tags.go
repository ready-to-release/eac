package behave

import (
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// BehaveTagTranslator translates TagFilter to behave tag expression syntax.
// Behave syntax: --tags @tag1 --tags @tag2 (AND), --tags ~@tag (NOT).
type BehaveTagTranslator struct{}

// TranslateTagFilter converts a TagFilter to behave CLI arguments.
// Returns a slice of strings to be appended to the behave command args.
func (t *BehaveTagTranslator) TranslateTagFilter(filter core.TagFilter) []string {
	var args []string

	for _, selector := range filter.Selectors {
		// RequireTags: each tag becomes --tags @tag
		for _, tag := range selector.RequireTags {
			args = append(args, "--tags", tag)
		}

		// ExcludeTags: each becomes --tags ~@tag
		for _, tag := range selector.ExcludeTags {
			if strings.HasPrefix(tag, "@") {
				args = append(args, "--tags", "~"+tag)
			} else {
				args = append(args, "--tags", "~@"+tag)
			}
		}

		// AnyOfTags: joined with comma (OR) as single --tags arg
		if len(selector.AnyOfTags) > 0 {
			orExpr := strings.Join(selector.AnyOfTags, ",")
			args = append(args, "--tags", orExpr)
		}
	}

	return args
}
