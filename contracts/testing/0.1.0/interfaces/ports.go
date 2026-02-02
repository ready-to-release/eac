// Package interfaces defines the testing contract interfaces.
// These interfaces define how test suites and tags are configured and accessed.
package interfaces

// TestConfigPort provides access to test suite and tag configuration.
// Implementations load suite definitions and tag configurations from files.
type TestConfigPort interface {
	// Suite access methods

	// GetSuite returns a test suite definition by moniker.
	// Returns false if the suite is not found.
	GetSuite(moniker string) (SuitePort, bool)

	// ListSuites returns all available suite monikers.
	ListSuites() []string

	// GetDefaultSuites returns the monikers of default (non-extended) suites.
	GetDefaultSuites() []string

	// ListNonProduction returns monikers of all suites except production (@L4).
	ListNonProduction() []string

	// GetSuiteLTags returns L-level tags for a suite (e.g., "@L0", "@L1").
	GetSuiteLTags(moniker string) []string

	// GetLTagToSuiteMap returns a mapping from L-level tags to suite monikers.
	GetLTagToSuiteMap() map[string]string

	// GetSuiteForLTag returns the suite moniker for a given L-level tag.
	GetSuiteForLTag(ltag string) string

	// Tag access methods

	// GetTag returns a tag definition by tag value.
	// Returns false if the tag is not found.
	GetTag(tag string) (TagPort, bool)

	// ListTags returns all available tag values.
	ListTags() []string

	// GetTagsByType returns tags of a specific type (e.g., "taxonomy-level").
	GetTagsByType(tagType string) []string

	// IsKnownTag checks if a tag is known (exact match or pattern match).
	IsKnownTag(tag string) bool

	// GetTaxonomyLevelTags returns all taxonomy-level tags (@L0-@L4, @HE2E).
	GetTaxonomyLevelTags() []string

	// GetVerificationTags returns all verification type tags.
	GetVerificationTags() []string

	// Tag validation methods

	// ValidateTag validates a tag and returns an error if invalid.
	ValidateTag(tag string) error

	// HasConstraint checks if a tag has a specific constraint.
	HasConstraint(tag, constraint string) bool

	// Skip reason methods

	// GetSkipReasons returns the list of valid skip reasons.
	GetSkipReasons() []SkipReason

	// GetValidSkipReasons returns all valid skip reason codes as strings.
	GetValidSkipReasons() []string

	// ValidateSkipReason checks if a skip reason code is valid.
	ValidateSkipReason(code string) (SkipReason, bool)

	// BuildGodogSkipTagFilter builds a Godog tag filter excluding @skip:<reason> tags.
	BuildGodogSkipTagFilter() string

	// GetSkipTagsForSuite returns skip tags as a slice for test suite selectors.
	GetSkipTagsForSuite() []string
}

// SuitePort defines a test suite's configuration.
type SuitePort interface {
	// Moniker returns the unique suite identifier (e.g., "unit").
	Moniker() string

	// Name returns the human-readable suite name.
	Name() string

	// Description returns a detailed description of the suite.
	Description() string

	// IsExtended returns true if this is an extended suite (not run by default).
	IsExtended() bool

	// Selectors returns the tag selectors for this suite.
	Selectors() []SelectorDef
}

// TagPort defines a test tag's configuration.
type TagPort interface {
	// Tag returns the tag value (e.g., "@L0", "@deps:<name>").
	Tag() string

	// Name returns the human-readable tag name.
	Name() string

	// Description returns a detailed description.
	Description() string

	// Type returns the tag type (e.g., "taxonomy-level", "system_dependency").
	Type() string

	// Pattern returns the regex pattern for parameterized tags (optional).
	Pattern() string

	// Level returns the tag scope level (e.g., "feature-or-scenario", "scenario").
	Level() string

	// Constraint returns any constraint defined for this tag (e.g., "mutually_exclusive_with_taxonomy_levels").
	Constraint() string
}

// SelectorDef defines a tag selector for suite filtering.
type SelectorDef struct {
	// AnyOfTags matches if any of these tags are present.
	AnyOfTags []string `yaml:"any_of_tags,omitempty"`

	// RequireTags matches only if all of these tags are present.
	RequireTags []string `yaml:"require_tags,omitempty"`

	// ExcludeTags excludes scenarios with any of these tags.
	ExcludeTags []string `yaml:"exclude_tags,omitempty"`
}

// SkipReason defines a valid reason code for skipping tests.
type SkipReason struct {
	// Code is the short identifier used in @skip:<code> tags.
	Code string `yaml:"code"`

	// Name is the human-readable name.
	Name string `yaml:"name"`

	// Description explains when to use this skip reason.
	Description string `yaml:"description"`
}
