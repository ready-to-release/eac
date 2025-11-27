package testing

import (
	"fmt"
	"sort"
	"strings"
)

// SuiteRegistry holds all defined test suites
var SuiteRegistry = map[string]*TestSuite{
	"commit":                  NewCommitSuite(),
	"acceptance":              NewAcceptanceSuite(),
	"production-verification": NewProductionVerificationSuite(),
}

// NewCommitSuite creates the commit test suite (L0-L2)
func NewCommitSuite() *TestSuite {
	return &TestSuite{
		Moniker:     "commit",
		Name:        "Commit Tests",
		Description: "Fast tests for Stage 2-4 (Pre-commit, MR, Commit) - L0-L2",
		Selectors: []TagSelector{
			{
				AnyOfTags: []string{"@L0", "@L1", "@L2"},
			},
		},
		Inferences: GetGlobalInferences(),
	}
}

// NewAcceptanceSuite creates the acceptance test suite (L3 IV, OV, PV)
// Only includes L3+ tests to avoid overlap with commit suite (L0-L2)
func NewAcceptanceSuite() *TestSuite {
	return &TestSuite{
		Moniker:     "acceptance",
		Name:        "PLTE Acceptance Tests",
		Description: "Stage 5-6 - L3 Installation, Operational, and Performance Verification",
		Selectors: []TagSelector{
			{
				AnyOfTags:   []string{"@iv", "@ov", "@pv"},
				ExcludeTags: []string{"@L0", "@L1", "@L2"},
			},
		},
		Inferences: GetGlobalInferences(),
	}
}

// NewProductionVerificationSuite creates the production verification suite (L4 + PIV)
func NewProductionVerificationSuite() *TestSuite {
	return &TestSuite{
		Moniker:     "production-verification",
		Name:        "Production Installation Verification",
		Description: "Stage 11-12 - Production smoke tests",
		Selectors: []TagSelector{
			{
				RequireTags: []string{"@L4", "@piv"},
			},
		},
		Inferences: GetGlobalInferences(),
	}
}

// GetSuite retrieves a suite by its moniker
func GetSuite(moniker string) (*TestSuite, error) {
	suite, exists := SuiteRegistry[moniker]
	if !exists {
		return nil, fmt.Errorf("suite not found: %s", moniker)
	}
	return suite, nil
}

// ListSuites returns all available suite monikers
func ListSuites() []string {
	monikers := make([]string, 0, len(SuiteRegistry))
	for moniker := range SuiteRegistry {
		monikers = append(monikers, moniker)
	}
	sort.Strings(monikers)
	return monikers
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
// Commit suite (L0-L2 tests):
//
//	AnyOfTags: ["@L0", "@L1", "@L2"]
//	Output:    "@L0,@L1,@L2"
//
// Acceptance suite (verification tests, excluding L0-L2):
//
//	AnyOfTags:   ["@iv", "@ov", "@pv"]
//	ExcludeTags: ["@L0", "@L1", "@L2"]
//	Output:      "@iv,@ov,@pv && ~@L0 && ~@L1 && ~@L2"
//
func (suite *TestSuite) BuildGodogTagFilter() string {
	var parts []string

	for _, selector := range suite.Selectors {
		var selectorParts []string

		// AnyOfTags: join with comma (no space) for OR semantics
		// Example: ["@L0", "@L1", "@L2"] → "@L0,@L1,@L2"
		if len(selector.AnyOfTags) > 0 {
			orExpr := strings.Join(selector.AnyOfTags, ",")
			selectorParts = append(selectorParts, orExpr)
		}

		// RequireTags: each tag becomes an AND condition
		// Example: ["@smoke", "@critical"] → added as separate "&& @smoke && @critical"
		for _, tag := range selector.RequireTags {
			selectorParts = append(selectorParts, tag)
		}

		// ExcludeTags: each tag becomes a NOT condition with tilde prefix
		// Example: ["@L0", "@L1"] → "&& ~@L0 && ~@L1"
		for _, tag := range selector.ExcludeTags {
			selectorParts = append(selectorParts, "~"+tag)
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

// SelectTests applies suite selectors to filter tests
func (suite *TestSuite) SelectTests(allTests []TestReference) []TestReference {
	selected := []TestReference{}
	ignoredCount := 0

	for _, test := range allTests {
		// Filter out ignored tests FIRST (before any other selection)
		if test.IsIgnored {
			ignoredCount++
			continue
		}

		if suite.Matches(test) {
			selected = append(selected, test)
		}
	}

	// Log ignored tests if any
	if ignoredCount > 0 {
		fmt.Printf("INFO: %d tests ignored (tagged with @ignore)\n", ignoredCount)
	}

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

	for _, test := range tests {
		for _, tag := range test.Tags {
			// Only include @deps: tags, not @depm: (module dependencies)
			// Also exclude OS platform tags (handled by OS filtering)
			if strings.HasPrefix(tag, "@deps:") && !OsPlatformTagsFull[tag] {
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
