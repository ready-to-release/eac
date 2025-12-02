package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// TemplateTagsConfig represents merged config from testing-tags.yml and template-tags.yml
type TemplateTagsConfig struct {
	Tags                 []TemplateTagDefinition `yaml:"tags"`
	Types                []TemplateTagType       `yaml:"types"`
	SkipReasons          []ValueDefinition       `yaml:"skip_reasons"`
	IndustryCodes        []ValueDefinition       `yaml:"industry_codes"`
	SeverityLevels       []ValueDefinition       `yaml:"severity_levels"`
	RiskTypes            []ValueDefinition       `yaml:"risk_types"`
	ControlTypes         []ValueDefinition       `yaml:"control_types"`
	ImplementationLevels []ValueDefinition       `yaml:"implementation_levels"`
	AutomationLevels     []ValueDefinition       `yaml:"automation_levels"`

	// Compiled state
	compiledPatterns map[string]*regexp.Regexp
	exactTags        map[string]*TemplateTagDefinition
	valueValidators  map[string]map[string]bool // prefix -> valid values
}

// TemplateTagDefinition represents a single tag definition
type TemplateTagDefinition struct {
	Tag         string `yaml:"tag"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
	Pattern     string `yaml:"pattern,omitempty"`
	Example     string `yaml:"example,omitempty"`
	Context     string `yaml:"context,omitempty"` // "comment" or "gherkin"
}

// TemplateTagType represents a tag type category
type TemplateTagType struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}

// ValueDefinition represents a valid value for parameterized tags
type ValueDefinition struct {
	Code        string `yaml:"code"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// TagValidationResult contains validation results for a single tag
type TagValidationResult struct {
	Tag         string
	IsValid     bool
	Definition  *TemplateTagDefinition
	Error       string
	Suggestions []string
}

// ValidationReport contains full validation results
type ValidationReport struct {
	TotalTags    int
	ValidTags    int
	InvalidTags  int
	Results      []TagValidationResult
	ByCategory   map[string]*CategoryValidation
}

// CategoryValidation tracks validation by tag category
type CategoryValidation struct {
	Category    string
	TotalCount  int
	ValidCount  int
	InvalidTags []string
}

// LoadTemplateTagsConfig loads both testing-tags.yml and template-tags.yml
// as separate read-only sources to validate template tags against.
// This is an isolated validation tool - does not modify production configs.
func LoadTemplateTagsConfig(repoRoot string) (*TemplateTagsConfig, error) {
	cfg := &TemplateTagsConfig{}

	// 1. Load testing-tags.yml (read-only - defines valid test taxonomy tags)
	testingPath := filepath.Join(repoRoot, ".r2r", "eac", "testing-tags.yml")
	if data, err := os.ReadFile(testingPath); err == nil {
		var testingCfg struct {
			Tags        []TemplateTagDefinition `yaml:"tags"`
			SkipReasons []ValueDefinition       `yaml:"skip_reasons"`
		}
		if err := yaml.Unmarshal(data, &testingCfg); err != nil {
			return nil, fmt.Errorf("failed to parse testing-tags.yml: %w", err)
		}
		cfg.Tags = append(cfg.Tags, testingCfg.Tags...)
		cfg.SkipReasons = testingCfg.SkipReasons
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read testing-tags.yml: %w", err)
	}

	// 2. Load template-tags.yml (template-specific metadata tags)
	templatePath := filepath.Join(repoRoot, ".r2r", "eac", "template-tags.yml")
	if data, err := os.ReadFile(templatePath); err == nil {
		var templateCfg TemplateTagsConfig
		if err := yaml.Unmarshal(data, &templateCfg); err != nil {
			return nil, fmt.Errorf("failed to parse template-tags.yml: %w", err)
		}
		// Merge template-specific tags
		cfg.Tags = append(cfg.Tags, templateCfg.Tags...)
		cfg.Types = append(cfg.Types, templateCfg.Types...)
		cfg.IndustryCodes = templateCfg.IndustryCodes
		cfg.SeverityLevels = templateCfg.SeverityLevels
		cfg.RiskTypes = templateCfg.RiskTypes
		cfg.ControlTypes = templateCfg.ControlTypes
		cfg.ImplementationLevels = templateCfg.ImplementationLevels
		cfg.AutomationLevels = templateCfg.AutomationLevels
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read template-tags.yml: %w", err)
	}

	if err := cfg.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize tags config: %w", err)
	}

	return cfg, nil
}

// Initialize compiles patterns and builds lookup maps
func (c *TemplateTagsConfig) Initialize() error {
	c.compiledPatterns = make(map[string]*regexp.Regexp)
	c.exactTags = make(map[string]*TemplateTagDefinition)
	c.valueValidators = make(map[string]map[string]bool)

	// Build exact tag lookup and compile patterns
	for i := range c.Tags {
		tag := &c.Tags[i]

		// Exact tags (no placeholders)
		if !strings.Contains(tag.Tag, "<") {
			c.exactTags[tag.Tag] = tag
		}

		// Compile patterns
		if tag.Pattern != "" {
			compiled, err := regexp.Compile(tag.Pattern)
			if err != nil {
				return fmt.Errorf("invalid pattern for tag %s: %w", tag.Tag, err)
			}
			c.compiledPatterns[tag.Tag] = compiled
		}
	}

	// Build value validators from valid values lists
	c.buildValueValidator("industry", c.IndustryCodes)
	c.buildValueValidator("severity", c.SeverityLevels)
	c.buildValueValidator("risk-type", c.RiskTypes)
	c.buildValueValidator("control-type", c.ControlTypes)
	c.buildValueValidator("implementation", c.ImplementationLevels)
	c.buildValueValidator("automation", c.AutomationLevels)
	c.buildValueValidator("skip", c.SkipReasons)

	return nil
}

func (c *TemplateTagsConfig) buildValueValidator(prefix string, values []ValueDefinition) {
	if len(values) == 0 {
		return
	}

	validMap := make(map[string]bool)
	for _, v := range values {
		validMap[v.Code] = true
	}
	c.valueValidators[prefix] = validMap
}

// ValidateTag validates a single tag against the configuration
func (c *TemplateTagsConfig) ValidateTag(tag string) TagValidationResult {
	result := TagValidationResult{
		Tag:     tag,
		IsValid: false,
	}

	// Check exact match first
	if def, ok := c.exactTags[tag]; ok {
		result.IsValid = true
		result.Definition = def
		return result
	}

	// Check pattern matches
	for templateTag, compiled := range c.compiledPatterns {
		if compiled.MatchString(tag) {
			// Find the definition
			for i := range c.Tags {
				if c.Tags[i].Tag == templateTag {
					result.Definition = &c.Tags[i]
					break
				}
			}

			// Additional validation for parameterized tags with known values
			if err := c.validateParameterizedTag(tag); err != nil {
				result.Error = err.Error()
				result.Suggestions = c.getSuggestions(tag)
				return result
			}

			result.IsValid = true
			return result
		}
	}

	// Unknown tag
	result.Error = "tag not defined in template-tags.yml"
	result.Suggestions = c.findSimilarTags(tag)
	return result
}

// validateParameterizedTag validates the value part of a parameterized tag
func (c *TemplateTagsConfig) validateParameterizedTag(tag string) error {
	if !strings.Contains(tag, ":") {
		return nil
	}

	parts := strings.SplitN(tag, ":", 2)
	prefix := strings.TrimPrefix(parts[0], "@")
	value := parts[1]

	// Check if we have a validator for this prefix
	if validValues, ok := c.valueValidators[prefix]; ok {
		if !validValues[value] {
			return fmt.Errorf("invalid value '%s' for @%s tag", value, prefix)
		}
	}

	return nil
}

// getSuggestions returns valid values for a tag prefix
func (c *TemplateTagsConfig) getSuggestions(tag string) []string {
	if !strings.Contains(tag, ":") {
		return nil
	}

	parts := strings.SplitN(tag, ":", 2)
	prefix := strings.TrimPrefix(parts[0], "@")

	if validValues, ok := c.valueValidators[prefix]; ok {
		var suggestions []string
		for val := range validValues {
			suggestions = append(suggestions, "@"+prefix+":"+val)
		}
		return suggestions
	}

	return nil
}

// findSimilarTags finds tags with similar prefixes
func (c *TemplateTagsConfig) findSimilarTags(tag string) []string {
	var similar []string
	tagLower := strings.ToLower(tag)

	// Check exact tags
	for exactTag := range c.exactTags {
		if strings.Contains(strings.ToLower(exactTag), strings.TrimPrefix(tagLower, "@")) {
			similar = append(similar, exactTag)
		}
	}

	// Suggest pattern examples
	if strings.Contains(tag, ":") {
		parts := strings.SplitN(tag, ":", 2)
		prefix := parts[0] + ":"
		for _, def := range c.Tags {
			if strings.HasPrefix(def.Tag, prefix) && def.Example != "" {
				similar = append(similar, def.Example)
				break
			}
		}
	}

	return similar
}

// ValidateTags validates multiple tags and returns a report
func (c *TemplateTagsConfig) ValidateTags(tags []TagInfo) *ValidationReport {
	report := &ValidationReport{
		ByCategory: make(map[string]*CategoryValidation),
	}

	// Deduplicate tags for validation
	seen := make(map[string]bool)

	for _, tagInfo := range tags {
		if seen[tagInfo.Tag] {
			continue
		}
		seen[tagInfo.Tag] = true

		report.TotalTags++
		result := c.ValidateTag(tagInfo.Tag)
		report.Results = append(report.Results, result)

		// Determine category
		category := categorizeTag(tagInfo.Tag)
		if report.ByCategory[category] == nil {
			report.ByCategory[category] = &CategoryValidation{Category: category}
		}
		catVal := report.ByCategory[category]
		catVal.TotalCount++

		if result.IsValid {
			report.ValidTags++
			catVal.ValidCount++
		} else {
			report.InvalidTags++
			catVal.InvalidTags = append(catVal.InvalidTags, tagInfo.Tag)
		}
	}

	return report
}

// GetTagsByType returns all tag definitions of a specific type
func (c *TemplateTagsConfig) GetTagsByType(tagType string) []*TemplateTagDefinition {
	var tags []*TemplateTagDefinition
	for i := range c.Tags {
		if c.Tags[i].Type == tagType {
			tags = append(tags, &c.Tags[i])
		}
	}
	return tags
}

// GetValidValues returns valid values for a parameterized tag prefix
func (c *TemplateTagsConfig) GetValidValues(prefix string) []string {
	if validValues, ok := c.valueValidators[prefix]; ok {
		var values []string
		for v := range validValues {
			values = append(values, v)
		}
		return values
	}
	return nil
}

// IsKnownTag checks if a tag is known (exact match or pattern match)
func (c *TemplateTagsConfig) IsKnownTag(tag string) bool {
	result := c.ValidateTag(tag)
	return result.IsValid
}
