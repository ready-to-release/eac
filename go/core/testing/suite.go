package testing

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
)

// GetSuite retrieves a suite by its moniker from configuration.
// Returns error if config is unavailable (fail-closed - no hardcoded fallbacks).
// Supports composite suites with "+" separator (e.g., "unit+integration").
func GetSuite(moniker string) (*TestSuite, error) {
	cfg := config.Global()
	if cfg == nil || cfg.Testing == nil {
		return nil, fmt.Errorf("cannot get suite '%s': config unavailable (ensure config is loaded)", moniker)
	}

	// Handle composite suites with "+" separator (e.g., "unit+integration")
	if strings.Contains(moniker, "+") {
		return buildCompositeSuite(moniker, cfg)
	}

	suiteDef, ok := cfg.Testing.GetSuite(moniker)
	if ok {
		return convertSuitePort(suiteDef), nil
	}
	return nil, fmt.Errorf("suite not found: %s", moniker)
}

// buildCompositeSuite creates a combined suite from multiple suites joined by "+".
// Example: "unit+integration" combines tests from both suites.
func buildCompositeSuite(moniker string, cfg *config.EACConfig) (*TestSuite, error) {
	parts := strings.Split(moniker, "+")

	// Collect all L-tags from constituent suites
	var allAnyOfTags []string
	var names []string

	for _, part := range parts {
		suiteDef, ok := cfg.Testing.GetSuite(part)
		if !ok {
			return nil, fmt.Errorf("suite not found in composite '%s': %s", moniker, part)
		}
		names = append(names, suiteDef.Name())

		// Get L-tags from config
		ltags := cfg.Testing.GetSuiteLTags(part)
		allAnyOfTags = append(allAnyOfTags, ltags...)
	}

	// Build exclude list from L-tag to suite mapping (exclude anything not in our set)
	ltagMap := cfg.Testing.GetLTagToSuiteMap()
	var excludeTags []string
	for ltag := range ltagMap {
		found := false
		for _, included := range allAnyOfTags {
			if ltag == included {
				found = true
				break
			}
		}
		if !found {
			excludeTags = append(excludeTags, ltag)
		}
	}

	return &TestSuite{
		Moniker:     moniker,
		Name:        strings.Join(names, " + "),
		Description: fmt.Sprintf("Combined suite: %s", moniker),
		Selectors: []TagSelector{
			{
				AnyOfTags:   allAnyOfTags,
				ExcludeTags: excludeTags,
			},
		},
		Inferences: GetGlobalInferences(),
	}, nil
}

// ListSuites returns all available suite monikers.
// Returns empty list if config is unavailable.
func ListSuites() []string {
	cfg := config.Global()
	if cfg == nil || cfg.Testing == nil {
		return []string{} // Config unavailable - return empty list
	}
	monikers := cfg.Testing.ListSuites()
	sort.Strings(monikers)
	return monikers
}

// convertSuitePort converts a SuitePort interface to a runtime TestSuite.
func convertSuitePort(suite core.SuitePort) *TestSuite {
	selectors := make([]TagSelector, len(suite.Selectors()))
	for i, sel := range suite.Selectors() {
		selectors[i] = TagSelector{
			RequireTags: sel.RequireTags,
			AnyOfTags:   sel.AnyOfTags,
			ExcludeTags: sel.ExcludeTags,
		}
	}

	return &TestSuite{
		Moniker:     suite.Moniker(),
		Name:        suite.Name(),
		Description: suite.Description(),
		Selectors:   selectors,
		Inferences:  GetGlobalInferences(),
	}
}

// ToTagFilter converts this suite's selectors into a technology-agnostic TagFilter.
// Adapters translate the returned TagFilter to their technology-specific syntax
// (godog, cucumber-js, etc.) using a TagFilterTranslator.
func (suite *TestSuite) ToTagFilter() core.TagFilter {
	selectors := make([]core.TagFilterSelector, len(suite.Selectors))
	for i, sel := range suite.Selectors {
		selectors[i] = core.TagFilterSelector{
			RequireTags: sel.RequireTags,
			AnyOfTags:   sel.AnyOfTags,
			ExcludeTags: sel.ExcludeTags,
		}
	}
	return core.TagFilter{Selectors: selectors}
}

// ToTagFilterWithSkipTags converts this suite's selectors into a TagFilter
// with additional skip tag exclusions appended to every selector.
func (suite *TestSuite) ToTagFilterWithSkipTags(skipTags []string) core.TagFilter {
	filter := suite.ToTagFilter()
	for i := range filter.Selectors {
		filter.Selectors[i].ExcludeTags = append(
			filter.Selectors[i].ExcludeTags, skipTags...,
		)
	}
	return filter
}

// SelectionStats contains statistics about test selection.
type SelectionStats struct {
	TotalDiscovered  int // Total tests discovered
	Skipped          int // Tests tagged with @skip:<reason>
	NotMatchingSuite int // Tests that don't match suite selectors
	Selected         int // Tests selected for the suite
}

// SelectTestsWithStats applies suite selectors to filter tests and returns statistics.
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
	}

	return selected, stats
}

// SelectTests applies suite selectors to filter tests.
// Use SelectTestsWithStats if you need selection statistics.
func (suite *TestSuite) SelectTests(allTests []TestReference) []TestReference {
	selected, _ := suite.SelectTestsWithStats(allTests)
	return selected
}

// Matches checks if a test matches the suite's selectors.
func (suite *TestSuite) Matches(test TestReference) bool {
	// Test must match at least one selector
	for _, selector := range suite.Selectors {
		if matchesSelector(test.Tags, selector) {
			return true
		}
	}
	return false
}

// matchesSelector checks if tags match a selector.
func matchesSelector(tags []string, selector TagSelector) bool {
	// Check required tags (AND)
	for _, required := range selector.RequireTags {
		if !slices.Contains(tags, required) {
			return false
		}
	}

	// Check any-of tags (OR)
	if len(selector.AnyOfTags) > 0 {
		hasAny := false
		for _, anyTag := range selector.AnyOfTags {
			if slices.Contains(tags, anyTag) {
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
		if slices.Contains(tags, excluded) {
			return false
		}
	}

	return true
}

// GetSystemDependencies extracts all @deps:* tags from tests (excludes @depm:* and OS platform tags).
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

// GetModuleDependencies extracts all @depm:* tags from tests.
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

// GetManualTests returns only manual tests from a list.
func GetManualTests(tests []TestReference) []TestReference {
	manual := []TestReference{}
	for _, test := range tests {
		if test.IsManual {
			manual = append(manual, test)
		}
	}
	return manual
}

// GetGxPTests returns only GxP tests from a list.
func GetGxPTests(tests []TestReference) []TestReference {
	gxp := []TestReference{}
	for _, test := range tests {
		if test.IsGxP {
			gxp = append(gxp, test)
		}
	}
	return gxp
}
