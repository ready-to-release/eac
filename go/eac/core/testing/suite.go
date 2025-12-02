package testing

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

// GetSuite retrieves a suite by its moniker from configuration.
// Returns error if config is unavailable (fail-closed - no hardcoded fallbacks).
func GetSuite(moniker string) (*TestSuite, error) {
	cfg := config.Global()
	if cfg == nil || cfg.TestSuites == nil {
		return nil, fmt.Errorf("cannot get suite '%s': config unavailable (ensure config is loaded)", moniker)
	}

	suiteDef := cfg.TestSuites.Get(moniker)
	if suiteDef != nil {
		return convertSuiteDef(suiteDef), nil
	}
	return nil, fmt.Errorf("suite not found: %s", moniker)
}

// ListSuites returns all available suite monikers.
// Returns empty list if config is unavailable.
func ListSuites() []string {
	cfg := config.Global()
	if cfg == nil || cfg.TestSuites == nil {
		return []string{} // Config unavailable - return empty list
	}
	monikers := cfg.TestSuites.List()
	sort.Strings(monikers)
	return monikers
}

// convertSuiteDef converts a config definition to a runtime TestSuite
func convertSuiteDef(def *config.TestSuiteDef) *TestSuite {
	selectors := make([]TagSelector, len(def.Selectors))
	for i, sel := range def.Selectors {
		selectors[i] = TagSelector{
			RequireTags: sel.RequireTags,
			AnyOfTags:   sel.AnyOfTags,
			ExcludeTags: sel.ExcludeTags,
		}
	}

	return &TestSuite{
		Moniker:     def.Moniker,
		Name:        def.Name,
		Description: def.Description,
		Selectors:   selectors,
		Inferences:  GetGlobalInferences(),
	}
}

// BuildGodogTagFilter generates a godog-compatible tag expression for the suite.
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
//
// # Examples
//
// Commit suite (L0-L1 tests):
//
//	AnyOfTags:   ["@L0", "@L1"]
//	ExcludeTags: ["@L2", "@L3", "@L4"]
//	Output:      "@L0,@L1 && ~@L2 && ~@L3 && ~@L4"
//
// Acceptance suite (verification tests, excluding L0-L1):
//
//	AnyOfTags:   ["@iv", "@ov", "@pv"]
//	ExcludeTags: ["@L0", "@L1"]
//	Output:      "@iv,@ov,@pv && ~@L0 && ~@L1"
//
// BuildGodogTagFilterWithSkipTags builds a godog tag filter that includes skip tag exclusions.
// This is the primary method that should be used instead of BuildGodogTagFilter.
func (suite *TestSuite) BuildGodogTagFilterWithSkipTags(skipTags []string) string {
	// Add skip tags to each selector's ExcludeTags
	for i := range suite.Selectors {
		suite.Selectors[i].ExcludeTags = append(suite.Selectors[i].ExcludeTags, skipTags...)
	}

	filter := suite.BuildGodogTagFilter()

	// Remove skip tags from selectors to avoid mutation
	for i := range suite.Selectors {
		suite.Selectors[i].ExcludeTags = suite.Selectors[i].ExcludeTags[:len(suite.Selectors[i].ExcludeTags)-len(skipTags)]
	}

	return filter
}

func (suite *TestSuite) BuildGodogTagFilter() string {
	var parts []string

	for _, selector := range suite.Selectors {
		var selectorParts []string

		// RequireTags: each tag becomes an AND condition
		// Example: ["@smoke", "@critical"] → added as separate "&& @smoke && @critical"
		// NOTE: RequireTags are added FIRST to ensure they come before AnyOfTags
		for _, tag := range selector.RequireTags {
			selectorParts = append(selectorParts, tag)
		}

		// ExcludeTags: each tag becomes a NOT condition with tilde prefix
		// Example: ["@L0", "@L1"] → "&& ~@L0 && ~@L1"
		for _, tag := range selector.ExcludeTags {
			selectorParts = append(selectorParts, "~"+tag)
		}

		// AnyOfTags: join with comma (no space) for OR semantics
		// Example: ["@L0", "@L1", "@L2"] → "@L0,@L1,@L2"
		// NOTE: AnyOfTags is added LAST to avoid operator precedence issues
		if len(selector.AnyOfTags) > 0 {
			orExpr := strings.Join(selector.AnyOfTags, ",")
			selectorParts = append(selectorParts, orExpr)
		}

		// Combine all parts with AND (&&)
		if len(selectorParts) > 0 {
			parts = append(parts, strings.Join(selectorParts, " && "))
		}
	}

	// Multiple selectors are OR'd together using comma
	// This handles suites with multiple TagSelector entries
	if len(parts) > 1 {
		return strings.Join(parts, ",")
	} else if len(parts) == 1 {
		return parts[0]
	}
	return ""
}

// SelectionStats contains statistics about test selection
type SelectionStats struct {
	TotalDiscovered  int // Total tests discovered
	Skipped          int // Tests tagged with @skip:<reason>
	NotMatchingSuite int // Tests that don't match suite selectors
	Selected         int // Tests selected for the suite
}

// SelectTestsWithStats applies suite selectors to filter tests and returns statistics
func (suite *TestSuite) SelectTestsWithStats(allTests []TestReference) ([]TestReference, SelectionStats) {
	selected := []TestReference{}
	stats := SelectionStats{
		TotalDiscovered: len(allTests),
	}

	for _, test := range allTests {
		// Filter out skipped tests FIRST (before any other selection)
		if test.IsIgnored {
			stats.Skipped++
			continue
		}

		if suite.Matches(test) {
			selected = append(selected, test)
		} else {
			stats.NotMatchingSuite++
		}
	}

	stats.Selected = len(selected)

	// Log skipped tests if any
	if stats.Skipped > 0 {
		log.Infof("%d tests skipped (tagged with @skip:<reason>)", stats.Skipped)
	}

	return selected, stats
}

// SelectTests applies suite selectors to filter tests.
// Use SelectTestsWithStats if you need selection statistics.
func (suite *TestSuite) SelectTests(allTests []TestReference) []TestReference {
	selected, _ := suite.SelectTestsWithStats(allTests)
	return selected
}

// Matches checks if a test matches the suite's selectors
func (suite *TestSuite) Matches(test TestReference) bool {
	// Test must match at least one selector
	for _, selector := range suite.Selectors {
		if matchesSelector(test.Tags, selector) {
			return true
		}
	}
	return false
}

// matchesSelector checks if tags match a selector
func matchesSelector(tags []string, selector TagSelector) bool {
	// Check required tags (AND)
	for _, required := range selector.RequireTags {
		if !contains(tags, required) {
			return false
		}
	}

	// Check any-of tags (OR)
	if len(selector.AnyOfTags) > 0 {
		hasAny := false
		for _, anyTag := range selector.AnyOfTags {
			if contains(tags, anyTag) {
				hasAny = true
				break
			}
		}
		if !hasAny {
			return false
		}
	}

	// Check excluded tags (NOT)
	for _, excluded := range selector.ExcludeTags {
		if contains(tags, excluded) {
			return false
		}
	}

	return true
}

// GetSystemDependencies extracts all @deps:* tags from tests (excludes @depm:* and OS platform tags)
func GetSystemDependencies(tests []TestReference) []string {
	depsMap := make(map[string]bool)
	osPlatformTagsFull := GetOSPlatformTagsFull()

	for _, test := range tests {
		for _, tag := range test.Tags {
			// Only include @deps: tags, not @depm: (module dependencies)
			// Exclude OS platform tags (handled by OS filtering)
			if strings.HasPrefix(tag, "@deps:") && !osPlatformTagsFull[tag] {
				depsMap[tag] = true
			}
		}
	}

	// Convert map to sorted slice
	deps := make([]string, 0, len(depsMap))
	for dep := range depsMap {
		deps = append(deps, dep)
	}
	sort.Strings(deps)

	return deps
}

// GetModuleDependencies extracts all @depm:* tags from tests
func GetModuleDependencies(tests []TestReference) []string {
	depsMap := make(map[string]bool)

	for _, test := range tests {
		for _, tag := range test.Tags {
			if strings.HasPrefix(tag, "@depm:") {
				depsMap[tag] = true
			}
		}
	}

	// Convert map to sorted slice
	deps := make([]string, 0, len(depsMap))
	for dep := range depsMap {
		deps = append(deps, dep)
	}
	sort.Strings(deps)

	return deps
}

// GetManualTests returns only manual tests from a list
func GetManualTests(tests []TestReference) []TestReference {
	manual := []TestReference{}
	for _, test := range tests {
		if test.IsManual {
			manual = append(manual, test)
		}
	}
	return manual
}

// GetGxPTests returns only GxP tests from a list
func GetGxPTests(tests []TestReference) []TestReference {
	gxp := []TestReference{}
	for _, test := range tests {
		if test.IsGxP {
			gxp = append(gxp, test)
		}
	}
	return gxp
}
