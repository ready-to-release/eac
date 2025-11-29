package config

import (
	"fmt"
	"regexp"
	"strings"
)

// TestingTagsConfig represents the testing-tags.yml configuration
type TestingTagsConfig struct {
	ModuleMonikers       MonikerExamples    `yaml:"module_monikers"`
	EnvironmentMonikers  MonikerExamples    `yaml:"environment_monikers"`
	SystemDependencies   []SystemDependency `yaml:"system_dependency_names"`
	OSPlatforms          []OSPlatform       `yaml:"os_platform_names"`
	Tags                 []TagDefinition    `yaml:"tags"`
	Types                []TagType          `yaml:"types"`
	SkipReasons          []SkipReason       `yaml:"skip_reasons"`

	// Compiled patterns for efficient validation (populated by Initialize)
	compiledPatterns map[string]*regexp.Regexp
	// Tag lookup by tag name for quick access
	tagLookup map[string]*TagDefinition
	// Tags grouped by type
	tagsByType map[string][]*TagDefinition
}

// MonikerExamples holds example monikers for documentation
type MonikerExamples struct {
	Examples []string `yaml:"examples"`
}

// SystemDependency represents a system dependency
type SystemDependency struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// OSPlatform represents an OS platform
type OSPlatform struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// TagDefinition represents a single tag definition
type TagDefinition struct {
	Tag         string `yaml:"tag"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
	Level       string `yaml:"level,omitempty"`      // feature-or-scenario, scenario
	Pattern     string `yaml:"pattern,omitempty"`    // Regex pattern for parameterized tags
	Example     string `yaml:"example,omitempty"`
	Note        string `yaml:"note,omitempty"`
	Constraint  string `yaml:"constraint,omitempty"` // e.g., "mutually_exclusive_with_taxonomy_levels"
}

// TagType represents a tag type category
type TagType struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}

// SkipReason represents a valid skip reason code
type SkipReason struct {
	Code        string `yaml:"code"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// Initialize compiles patterns and builds lookup maps
func (c *TestingTagsConfig) Initialize() error {
	c.compiledPatterns = make(map[string]*regexp.Regexp)
	c.tagLookup = make(map[string]*TagDefinition)
	c.tagsByType = make(map[string][]*TagDefinition)

	for i := range c.Tags {
		tag := &c.Tags[i]

		// Build lookup by tag name (for exact match tags)
		if !strings.Contains(tag.Tag, "<") {
			c.tagLookup[tag.Tag] = tag
		}

		// Group by type
		c.tagsByType[tag.Type] = append(c.tagsByType[tag.Type], tag)

		// Compile patterns for parameterized tags
		if tag.Pattern != "" {
			compiled, err := regexp.Compile(tag.Pattern)
			if err != nil {
				return fmt.Errorf("invalid pattern for tag %s: %w", tag.Tag, err)
			}
			c.compiledPatterns[tag.Tag] = compiled
		}
	}

	return nil
}

// GetTagsByType returns all tags of a specific type
func (c *TestingTagsConfig) GetTagsByType(tagType string) []*TagDefinition {
	return c.tagsByType[tagType]
}

// GetTag returns a tag definition by exact match or pattern match
func (c *TestingTagsConfig) GetTag(tag string) (*TagDefinition, bool) {
	// Check exact match
	if def, ok := c.tagLookup[tag]; ok {
		return def, true
	}

	// Check pattern match
	for templateTag, compiled := range c.compiledPatterns {
		if compiled.MatchString(tag) {
			for i := range c.Tags {
				if c.Tags[i].Tag == templateTag {
					return &c.Tags[i], true
				}
			}
		}
	}

	return nil, false
}

// IsKnownTag checks if a tag is known (exact match or pattern match)
func (c *TestingTagsConfig) IsKnownTag(tag string) bool {
	_, ok := c.GetTag(tag)
	return ok
}

// GetTaxonomyLevelTags returns all taxonomy-level tags (@L0-@L4, @HE2E)
func (c *TestingTagsConfig) GetTaxonomyLevelTags() []string {
	var tags []string
	for _, tag := range c.tagsByType["taxonomy-level"] {
		tags = append(tags, tag.Tag)
	}
	return tags
}

// GetVerificationTags returns all verification type tags
func (c *TestingTagsConfig) GetVerificationTags() []string {
	var tags []string
	for _, tag := range c.tagsByType["verification"] {
		tags = append(tags, tag.Tag)
	}
	return tags
}

// GetValidSkipReasons returns all valid skip reason codes
func (c *TestingTagsConfig) GetValidSkipReasons() []string {
	reasons := make([]string, len(c.SkipReasons))
	for i, sr := range c.SkipReasons {
		reasons[i] = sr.Code
	}
	return reasons
}

// GetSkipReasons returns a map of skip reason code to SkipReason
func (c *TestingTagsConfig) GetSkipReasons() map[string]SkipReason {
	reasons := make(map[string]SkipReason)
	for _, sr := range c.SkipReasons {
		reasons[sr.Code] = sr
	}
	return reasons
}

// ValidateSkipReason checks if a skip reason code is valid
func (c *TestingTagsConfig) ValidateSkipReason(code string) (*SkipReason, bool) {
	for i := range c.SkipReasons {
		if c.SkipReasons[i].Code == code {
			return &c.SkipReasons[i], true
		}
	}
	return nil, false
}

// BuildGodogSkipTagFilter builds a Godog tag filter expression that excludes all @skip:<reason> tags
func (c *TestingTagsConfig) BuildGodogSkipTagFilter() string {
	if len(c.SkipReasons) == 0 {
		return ""
	}

	var parts []string
	for _, reason := range c.SkipReasons {
		parts = append(parts, fmt.Sprintf("~@skip:%s", reason.Code))
	}
	return strings.Join(parts, " && ")
}

// GetSkipTagsForSuite returns skip tags as a slice suitable for test suite selectors
func (c *TestingTagsConfig) GetSkipTagsForSuite() []string {
	tags := make([]string, 0, len(c.SkipReasons)+1)
	for _, reason := range c.SkipReasons {
		tags = append(tags, fmt.Sprintf("@skip:%s", reason.Code))
	}
	tags = append(tags, "@pending")
	return tags
}

// HasConstraint checks if a tag has a specific constraint
func (c *TestingTagsConfig) HasConstraint(tag, constraint string) bool {
	def, ok := c.GetTag(tag)
	if !ok {
		return false
	}
	return def.Constraint == constraint
}

// ValidateTag validates a tag and returns an error message if invalid
func (c *TestingTagsConfig) ValidateTag(tag string) error {
	// Check exact match
	if _, ok := c.tagLookup[tag]; ok {
		return nil
	}

	// Check pattern match
	for _, compiled := range c.compiledPatterns {
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
	}

	// Check if it starts with a known prefix but has wrong format
	for templateTag := range c.compiledPatterns {
		if idx := strings.Index(templateTag, "<"); idx > 0 {
			prefix := templateTag[:idx]
			if strings.HasPrefix(tag, prefix) {
				// Find the tag definition in Tags slice
				for i := range c.Tags {
					if c.Tags[i].Tag == templateTag {
						return fmt.Errorf("tag '%s' has invalid format, expected pattern: %s (example: %s)", tag, c.Tags[i].Pattern, c.Tags[i].Example)
					}
				}
			}
		}
	}

	return nil // Unknown tags are not an error here (handled separately)
}
