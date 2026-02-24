package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// TestingConfig implements core.TestConfigPort.
// It loads test suite definitions and tag configurations from the eac-testing contract.
type TestingConfig struct {
	suites      map[string]*core.SuiteDefinition
	suiteOrder  []string // Preserve order from YAML
	tags        map[string]*core.TagDefinition
	tagTypes    map[string]core.TagType
	skipReasons []core.SkipReason

	// Compiled patterns for efficient validation (for parameterized tags)
	compiledPatterns map[string]*regexp.Regexp
	// Tag lookup by exact name for quick access (excludes parameterized templates)
	tagLookup map[string]*core.TagDefinition
}

// Verify TestingConfig implements TestConfigPort.
var _ core.TestConfigPort = (*TestingConfig)(nil)

// LoadTestingConfig loads the testing configuration from the eac-testing contract.
// It loads both suites.yml and tags.yml, merging defaults with user overrides.
func LoadTestingConfig(repoRoot, configRoot string) (*TestingConfig, error) {
	cfg := &TestingConfig{
		suites:           make(map[string]*core.SuiteDefinition),
		tags:             make(map[string]*core.TagDefinition),
		tagTypes:         make(map[string]core.TagType),
		compiledPatterns: make(map[string]*regexp.Regexp),
		tagLookup:        make(map[string]*core.TagDefinition),
	}

	// Load suites
	if err := cfg.loadSuites(repoRoot, configRoot); err != nil {
		return nil, err
	}

	// Load tags
	if err := cfg.loadTags(repoRoot, configRoot); err != nil {
		return nil, err
	}

	// Initialize compiled patterns for parameterized tag matching
	if err := cfg.initialize(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// initialize compiles regex patterns and builds lookup maps.
func (c *TestingConfig) initialize() error {
	for tagValue, tag := range c.tags {
		// Build lookup for exact match tags (non-parameterized)
		if !strings.Contains(tagValue, "<") {
			c.tagLookup[tagValue] = tag
		}

		// Compile patterns for parameterized tags
		if tag.PatternValue != "" {
			compiled, err := regexp.Compile(tag.PatternValue)
			if err != nil {
				return fmt.Errorf("invalid pattern for tag %s: %w", tagValue, err)
			}
			c.compiledPatterns[tagValue] = compiled
		}
	}
	return nil
}

// loadSuites loads suite definitions from suites.yml.
func (c *TestingConfig) loadSuites(repoRoot, configRoot string) error {
	// Load contract defaults from embedded FS — available in all projects regardless of
	// whether the contracts/ directory is present on disk (external projects, CI, etc.).
	var defaults *core.SuitesConfig
	if data, err := core.FS.ReadFile(core.DefaultPath("test-suites.yml")); err == nil {
		var cfg core.SuitesConfig
		if unmarshalErr := yaml.Unmarshal(data, &cfg); unmarshalErr == nil {
			defaults = &cfg
		}
	}

	// Load user overrides
	userPath := filepath.Join(configRoot, "suites.yml")
	user, err := loadSuitesFile(userPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Merge: start with defaults, override with user
	if defaults != nil {
		for _, suite := range defaults.Suites {
			c.suites[suite.MonikerValue] = suite
			c.suiteOrder = append(c.suiteOrder, suite.MonikerValue)
		}
	}
	if user != nil {
		for _, suite := range user.Suites {
			if _, exists := c.suites[suite.MonikerValue]; !exists {
				c.suiteOrder = append(c.suiteOrder, suite.MonikerValue)
			}
			c.suites[suite.MonikerValue] = suite
		}
	}

	return nil
}

// loadTags loads tag definitions from tags.yml.
func (c *TestingConfig) loadTags(repoRoot, configRoot string) error {
	// Load contract defaults from embedded FS — available in all projects regardless of
	// whether the contracts/ directory is present on disk (external projects, CI, etc.).
	var defaults *core.TagsConfig
	if data, err := core.FS.ReadFile(core.DefaultPath("testing-tags.yml")); err == nil {
		var cfg core.TagsConfig
		if unmarshalErr := yaml.Unmarshal(data, &cfg); unmarshalErr == nil {
			defaults = &cfg
		}
	}

	// Load user overrides
	userPath := filepath.Join(configRoot, "tags.yml")
	user, err := loadTagsFile(userPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Merge: start with defaults, override with user
	if defaults != nil {
		for _, tag := range defaults.Tags {
			c.tags[tag.TagValue] = tag
		}
		for _, tagType := range defaults.Types {
			c.tagTypes[tagType.Type] = tagType
		}
		c.skipReasons = defaults.SkipReasons
	}
	if user != nil {
		for _, tag := range user.Tags {
			c.tags[tag.TagValue] = tag
		}
		for _, tagType := range user.Types {
			c.tagTypes[tagType.Type] = tagType
		}
		if len(user.SkipReasons) > 0 {
			c.skipReasons = user.SkipReasons
		}
	}

	return nil
}

// GetSuite returns a suite definition by moniker.
func (c *TestingConfig) GetSuite(moniker string) (core.SuitePort, bool) {
	suite, ok := c.suites[moniker]
	if !ok {
		return nil, false
	}
	return suite, true
}

// ListSuites returns all available suite monikers.
func (c *TestingConfig) ListSuites() []string {
	// Return in original order
	return append([]string{}, c.suiteOrder...)
}

// GetDefaultSuites returns the monikers of default (non-extended) suites.
func (c *TestingConfig) GetDefaultSuites() []string {
	var defaults []string
	for _, moniker := range c.suiteOrder {
		if suite, ok := c.suites[moniker]; ok && !suite.ExtendedSuite {
			defaults = append(defaults, moniker)
		}
	}
	return defaults
}

// GetTag returns a tag definition by tag value.
// Supports both exact match and pattern-based matching for parameterized tags.
func (c *TestingConfig) GetTag(tag string) (core.TagPort, bool) {
	// Check exact match first
	if tagDef, ok := c.tagLookup[tag]; ok {
		return tagDef, true
	}

	// Check pattern match for parameterized tags
	for templateTag, compiled := range c.compiledPatterns {
		if compiled.MatchString(tag) {
			if tagDef, ok := c.tags[templateTag]; ok {
				return tagDef, true
			}
		}
	}

	return nil, false
}

// ListTags returns all available tag values.
func (c *TestingConfig) ListTags() []string {
	tags := make([]string, 0, len(c.tags))
	for tag := range c.tags {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

// GetTagsByType returns tags of a specific type (e.g., "taxonomy-level").
func (c *TestingConfig) GetTagsByType(tagType string) []string {
	var tags []string
	for _, tag := range c.tags {
		if tag.TypeValue == tagType {
			tags = append(tags, tag.TagValue)
		}
	}
	sort.Strings(tags)
	return tags
}

// GetSkipReasons returns the list of valid skip reason codes.
func (c *TestingConfig) GetSkipReasons() []core.SkipReason {
	return c.skipReasons
}

// ListNonProduction returns monikers of all suites except production (@L4).
func (c *TestingConfig) ListNonProduction() []string {
	var suites []string
	for _, moniker := range c.suiteOrder {
		suite, ok := c.suites[moniker]
		if !ok {
			continue
		}
		// Skip suites that require @L4 tag (production verification)
		isProduction := false
		for _, sel := range suite.SelectorsValue {
			for _, tag := range sel.RequireTags {
				if tag == "@L4" {
					isProduction = true
					break
				}
			}
			for _, tag := range sel.AnyOfTags {
				if tag == "@L4" {
					isProduction = true
					break
				}
			}
			if isProduction {
				break
			}
		}
		if !isProduction {
			suites = append(suites, moniker)
		}
	}
	return suites
}

// GetSuiteLTags returns L-level tags for a suite (e.g., "@L0", "@L1").
func (c *TestingConfig) GetSuiteLTags(moniker string) []string {
	suite, ok := c.suites[moniker]
	if !ok {
		return nil
	}
	var ltags []string
	for _, sel := range suite.SelectorsValue {
		for _, tag := range sel.AnyOfTags {
			if len(tag) >= 2 && tag[0] == '@' && tag[1] == 'L' {
				ltags = append(ltags, tag)
			}
		}
	}
	return ltags
}

// GetLTagToSuiteMap returns a mapping from L-level tags to suite monikers.
func (c *TestingConfig) GetLTagToSuiteMap() map[string]string {
	tagMap := make(map[string]string)
	for moniker, suite := range c.suites {
		for _, sel := range suite.SelectorsValue {
			for _, tag := range sel.AnyOfTags {
				// Only map L-level tags
				if len(tag) >= 2 && tag[0] == '@' && tag[1] == 'L' {
					tagMap[tag] = moniker
				}
			}
		}
	}
	return tagMap
}

// GetSuiteForLTag returns the suite moniker for a given L-level tag.
// Returns empty string if no suite matches.
func (c *TestingConfig) GetSuiteForLTag(ltag string) string {
	for _, suite := range c.suites {
		for _, sel := range suite.SelectorsValue {
			for _, tag := range sel.AnyOfTags {
				if tag == ltag {
					return suite.MonikerValue
				}
			}
		}
	}
	return ""
}

// IsKnownTag checks if a tag is known (exact match or pattern match).
func (c *TestingConfig) IsKnownTag(tag string) bool {
	_, ok := c.GetTag(tag)
	return ok
}

// GetTaxonomyLevelTags returns all taxonomy-level tags (@L0-@L4, @HE2E).
func (c *TestingConfig) GetTaxonomyLevelTags() []string {
	return c.GetTagsByType(core.TagTypeTaxonomyLevel)
}

// GetVerificationTags returns all verification type tags.
func (c *TestingConfig) GetVerificationTags() []string {
	return c.GetTagsByType(core.TagTypeVerification)
}

// ValidateTag validates a tag and returns an error if invalid.
func (c *TestingConfig) ValidateTag(tag string) error {
	// Check exact match
	if _, ok := c.tagLookup[tag]; ok {
		return nil
	}

	// Check pattern match
	for templateTag, compiled := range c.compiledPatterns {
		if compiled.MatchString(tag) {
			// Additional validation for @skip tags
			if strings.HasPrefix(tag, "@skip:") {
				reason := strings.TrimPrefix(tag, "@skip:")
				if _, ok := c.ValidateSkipReason(reason); !ok {
					return fmt.Errorf("invalid skip reason '%s', valid: %s", reason, strings.Join(c.GetValidSkipReasons(), ", "))
				}
			}
			return nil
		}

		// Check if it starts with a known prefix but has wrong format
		if idx := strings.Index(templateTag, "<"); idx > 0 {
			prefix := templateTag[:idx]
			if strings.HasPrefix(tag, prefix) {
				tagDef := c.tags[templateTag]
				return fmt.Errorf("tag '%s' has invalid format, expected pattern: %s (example: %s)", tag, tagDef.PatternValue, tagDef.Example)
			}
		}
	}

	return nil // Unknown tags are not an error here (handled separately)
}

// HasConstraint checks if a tag has a specific constraint.
func (c *TestingConfig) HasConstraint(tag, constraint string) bool {
	tagDef, ok := c.GetTag(tag)
	if !ok {
		return false
	}
	return tagDef.Constraint() == constraint
}

// GetValidSkipReasons returns all valid skip reason codes as strings.
func (c *TestingConfig) GetValidSkipReasons() []string {
	reasons := make([]string, len(c.skipReasons))
	for i, sr := range c.skipReasons {
		reasons[i] = sr.Code
	}
	return reasons
}

// ValidateSkipReason checks if a skip reason code is valid.
func (c *TestingConfig) ValidateSkipReason(code string) (core.SkipReason, bool) {
	for _, sr := range c.skipReasons {
		if sr.Code == code {
			return sr, true
		}
	}
	return core.SkipReason{}, false
}

// GetSkipTags returns skip reason tags as a raw slice (e.g., ["@skip:wip", "@skip:broken"]).
func (c *TestingConfig) GetSkipTags() []string {
	tags := make([]string, 0, len(c.skipReasons))
	for _, reason := range c.skipReasons {
		tags = append(tags, fmt.Sprintf("@skip:%s", reason.Code))
	}
	return tags
}

// GetSkipTagsForSuite returns skip tags as a slice for test suite selectors.
func (c *TestingConfig) GetSkipTagsForSuite() []string {
	tags := make([]string, 0, len(c.skipReasons)+1)
	for _, reason := range c.skipReasons {
		tags = append(tags, fmt.Sprintf("@skip:%s", reason.Code))
	}
	tags = append(tags, "@pending")
	return tags
}

// Helper functions

func loadSuitesFile(path string) (*core.SuitesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg core.SuitesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadTagsFile(path string) (*core.TagsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg core.TagsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
