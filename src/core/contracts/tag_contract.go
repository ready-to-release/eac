package contracts

import (
	"fmt"
	"regexp"
	"strings"
)

// Intent: Define tag contract types and validation logic for Gherkin specifications
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear struct mapping to tags.yml schema
//   - Each tag type has explicit validation rules
//   - Pattern-based tags use compiled regex for clarity
//
// Easy to change:
//   - Tag definitions loaded from external contract (tags.yml)
//   - New tag types can be added without code changes
//   - Validation rules are data-driven, not hardcoded
//
// Hard to break:
//   - Compiled regex patterns for tag validation
//   - Explicit constraint definitions from contract
//   - Comprehensive error messages with context

// TagContract represents the test tagging system specification from tags.yml
type TagContract struct {
	Metadata    TagMetadata     `yaml:"metadata"`
	Tags        []TagDefinition `yaml:"tags"`
	Types       []TagType       `yaml:"types"`
	SkipReasons []SkipReason    `yaml:"skip_reasons"`

	// Compiled patterns for efficient validation (populated after loading)
	compiledPatterns map[string]*regexp.Regexp
	// Tag lookup by tag name for quick access
	tagLookup map[string]*TagDefinition
	// Tags grouped by type
	tagsByType map[string][]*TagDefinition
}

// TagMetadata contains contract metadata
type TagMetadata struct {
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Scope       string `yaml:"scope"`
}

// TagDefinition represents a single tag in the contract
type TagDefinition struct {
	Tag         string `yaml:"tag"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
	Level       string `yaml:"level,omitempty"`   // feature-or-scenario, scenario
	Pattern     string `yaml:"pattern,omitempty"` // Regex pattern for parameterized tags
	Example     string `yaml:"example,omitempty"`
	Inferred    bool   `yaml:"inferred,omitempty"`
	Constraint  string `yaml:"constraint,omitempty"` // e.g., "mutually_exclusive_with_taxonomy_levels"
}

// TagType represents a tag type definition
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

// Initialize compiles patterns and builds lookup maps after loading
func (tc *TagContract) Initialize() error {
	tc.compiledPatterns = make(map[string]*regexp.Regexp)
	tc.tagLookup = make(map[string]*TagDefinition)
	tc.tagsByType = make(map[string][]*TagDefinition)

	for i := range tc.Tags {
		tag := &tc.Tags[i]

		// Build lookup by tag name (for exact match tags)
		if !strings.Contains(tag.Tag, "<") {
			tc.tagLookup[tag.Tag] = tag
		}

		// Group by type
		tc.tagsByType[tag.Type] = append(tc.tagsByType[tag.Type], tag)

		// Compile patterns for parameterized tags
		if tag.Pattern != "" {
			compiled, err := regexp.Compile(tag.Pattern)
			if err != nil {
				return fmt.Errorf("invalid pattern for tag %s: %w", tag.Tag, err)
			}
			tc.compiledPatterns[tag.Tag] = compiled
		}
	}

	return nil
}

// GetVerificationTags returns all verification type tags
func (tc *TagContract) GetVerificationTags() []string {
	tags := []string{}
	for _, tag := range tc.tagsByType["verification"] {
		tags = append(tags, tag.Tag)
	}
	return tags
}

// GetTaxonomyLevelTags returns all taxonomy-level tags (@L0-@L4, @HE2E)
func (tc *TagContract) GetTaxonomyLevelTags() []string {
	tags := []string{}
	for _, tag := range tc.tagsByType["taxonomy-level"] {
		tags = append(tags, tag.Tag)
	}
	return tags
}

// GetValidSkipReasons returns all valid skip reason codes
func (tc *TagContract) GetValidSkipReasons() []string {
	reasons := []string{}
	for _, sr := range tc.SkipReasons {
		reasons = append(reasons, sr.Code)
	}
	return reasons
}

// IsKnownTag checks if a tag is known (either exact match or matches a pattern)
func (tc *TagContract) IsKnownTag(tag string) bool {
	// Check exact match first
	if _, ok := tc.tagLookup[tag]; ok {
		return true
	}

	// Check pattern-based tags
	for _, compiled := range tc.compiledPatterns {
		if compiled.MatchString(tag) {
			return true
		}
	}

	return false
}

// ValidateTag validates a single tag against the contract
// Returns nil if valid, or a ValidationError if invalid
func (tc *TagContract) ValidateTag(tag string, lineNum int) *ValidationError {
	// Check exact match first
	if _, ok := tc.tagLookup[tag]; ok {
		return nil // Valid exact match
	}

	// Check ALL pattern-based tags for a match first
	for _, compiled := range tc.compiledPatterns {
		if compiled.MatchString(tag) {
			// Additional validation for specific patterns
			if strings.HasPrefix(tag, "@skip:") {
				return tc.validateSkipTag(tag, lineNum)
			}
			return nil // Valid pattern match
		}
	}

	// If no pattern matched, check if tag starts with a known pattern prefix
	// This helps detect format errors (e.g., @skip:INVALID vs @skip:wip)
	for templateTag := range tc.compiledPatterns {
		if tc.tagStartsWithPatternPrefix(tag, templateTag) {
			def := tc.getTagDefinitionByTemplate(templateTag)
			if def != nil && def.Pattern != "" {
				return &ValidationError{
					Code: "INVALID_TAG_FORMAT",
					Message: fmt.Sprintf(
						"Tag '%s' has invalid format. Expected pattern: %s (example: %s)",
						tag, def.Pattern, def.Example,
					),
					Line:     lineNum,
					Severity: "error",
				}
			}
		}
	}

	return nil // Unknown tags are handled separately
}

// validateSkipTag validates @skip:<reason> tags
func (tc *TagContract) validateSkipTag(tag string, lineNum int) *ValidationError {
	// Extract reason from @skip:<reason>
	if !strings.HasPrefix(tag, "@skip:") {
		return nil
	}

	reason := strings.TrimPrefix(tag, "@skip:")
	validReasons := tc.GetValidSkipReasons()

	for _, valid := range validReasons {
		if reason == valid {
			return nil // Valid skip reason
		}
	}

	return &ValidationError{
		Code: "INVALID_SKIP_REASON",
		Message: fmt.Sprintf(
			"Invalid skip reason '%s'. Valid reasons: %s",
			reason, strings.Join(validReasons, ", "),
		),
		Line:     lineNum,
		Severity: "error",
	}
}

// tagStartsWithPatternPrefix checks if tag starts with a pattern's prefix
func (tc *TagContract) tagStartsWithPatternPrefix(tag, templateTag string) bool {
	// Extract prefix from template (e.g., "@skip:<reason>" -> "@skip:")
	if idx := strings.Index(templateTag, "<"); idx > 0 {
		prefix := templateTag[:idx]
		return strings.HasPrefix(tag, prefix)
	}
	return false
}

// getTagDefinitionByTemplate returns the tag definition for a template tag
func (tc *TagContract) getTagDefinitionByTemplate(templateTag string) *TagDefinition {
	for i := range tc.Tags {
		if tc.Tags[i].Tag == templateTag {
			return &tc.Tags[i]
		}
	}
	return nil
}

// GetTagDefinition returns the tag definition for a specific tag
func (tc *TagContract) GetTagDefinition(tag string) *TagDefinition {
	// Check exact match
	if def, ok := tc.tagLookup[tag]; ok {
		return def
	}

	// Check pattern match
	for templateTag, compiled := range tc.compiledPatterns {
		if compiled.MatchString(tag) {
			return tc.getTagDefinitionByTemplate(templateTag)
		}
	}

	return nil
}

// HasConstraint checks if a tag has a specific constraint
func (tc *TagContract) HasConstraint(tag, constraint string) bool {
	def := tc.GetTagDefinition(tag)
	if def == nil {
		return false
	}
	return def.Constraint == constraint
}

// LoadTagContract loads the tag contract from tags.yml
func LoadTagContract(loader *Loader) (*TagContract, error) {
	contract := &TagContract{}

	err := loader.LoadYAML("contracts/testing/0.1.0/tags.yml", contract)
	if err != nil {
		return nil, fmt.Errorf("failed to load tag contract: %w", err)
	}

	if err := contract.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize tag contract: %w", err)
	}

	return contract, nil
}
