package gherkin

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/core/validation"
)

// validateScenarioTags validates all tag-related rules on scenarios:
// - Verification tags present (with inheritance from Feature and Rule)
// - Tag format validation
// - Skip reason validation
// - Mutual exclusion constraints
// - GxP tag requirements
// - Unknown tag warnings
//
// Tag inheritance follows Gherkin semantics:
// - Feature tags are inherited by all Rules and Scenarios
// - Rule tags are inherited by all Scenarios within that Rule.
func (v *Validator) validateScenarioTags(lines []string) []validation.ValidationError {
	var errors []validation.ValidationError

	// Inherited tags from Feature and Rule levels
	var featureTags []string     // Tags on the Feature (inherited by all)
	var currentRuleTags []string // Tags on current Rule (inherited by scenarios in that rule)
	var pendingTags []string     // Tags collected immediately before a scenario
	var pendingTagLines []int    // Line numbers for pending tags

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Collect tags (lines starting with @)
		if strings.HasPrefix(trimmed, "@") {
			// Extract all tags from the line (space-separated)
			tags := strings.Fields(trimmed)
			for _, tag := range tags {
				if strings.HasPrefix(tag, "@") {
					pendingTags = append(pendingTags, tag)
					pendingTagLines = append(pendingTagLines, lineNum)
				}
			}
			continue
		}

		// When we hit Feature, capture pending tags as feature-level tags
		if strings.HasPrefix(trimmed, "Feature:") {
			featureTags = append([]string{}, pendingTags...)
			pendingTags = []string{}
			pendingTagLines = []int{}
			continue
		}

		// When we hit Rule, capture pending tags as rule-level tags
		if strings.HasPrefix(trimmed, "Rule:") {
			currentRuleTags = append([]string{}, pendingTags...)
			pendingTags = []string{}
			pendingTagLines = []int{}
			continue
		}

		// When we hit a Scenario, validate with inherited tags
		if strings.HasPrefix(trimmed, "Scenario:") || strings.HasPrefix(trimmed, "Scenario Outline:") {
			// Combine inherited tags: Feature + Rule + Scenario
			allTags := v.combineInheritedTags(featureTags, currentRuleTags, pendingTags)
			scenarioErrors := v.validateTagsForScenario(allTags, pendingTagLines, lineNum)
			errors = append(errors, scenarioErrors...)

			// Reset only scenario-level pending tags (keep inherited)
			pendingTags = []string{}
			pendingTagLines = []int{}
			continue
		}

		// Background clears pending tags but doesn't affect inheritance
		if strings.HasPrefix(trimmed, "Background:") {
			pendingTags = []string{}
			pendingTagLines = []int{}
			continue
		}
	}

	return errors
}

// combineInheritedTags merges tags from Feature, Rule, and Scenario levels
// Returns a deduplicated list of all applicable tags.
func (v *Validator) combineInheritedTags(featureTags, ruleTags, scenarioTags []string) []string {
	seen := make(map[string]bool)
	var result []string

	// Add in order: Feature, Rule, Scenario (scenario tags take precedence visually)
	for _, tag := range featureTags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}
	for _, tag := range ruleTags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}
	for _, tag := range scenarioTags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}

	return result
}

// validateTagAgainstSchema validates a single tag against the schema definition.
func (v *Validator) validateTagAgainstSchema(tag string, lineNum int, isFeature bool) []validation.ValidationError {
	var errors []validation.ValidationError

	// Get tag definition from schema
	tagDef, known := v.tagsConfig.GetTag(tag)

	if !known {
		// Unknown tag - warning only
		errors = append(errors, *validation.NewValidationError(validation.ErrUnknownTag, fmt.Sprintf("Unknown tag '%s' not defined in testing-tags.yml", tag), lineNum))
		return errors
	}

	// Validate tag format using schema pattern
	if formatErr := v.tagsConfig.ValidateTag(tag); formatErr != nil {
		errors = append(errors, *validation.NewValidationError(validation.ErrInvalidTagFormat, formatErr.Error(), lineNum))
	}

	// Validate level constraint from schema
	if tagDef.Level == "scenario" && isFeature {
		errors = append(errors, *validation.NewValidationError(validation.ErrInvalidTagLevel, fmt.Sprintf("Tag '%s' can only be used on scenarios, not features (level: scenario)", tag), lineNum))
	}

	if tagDef.Level == "feature" && !isFeature {
		errors = append(errors, *validation.NewValidationError(validation.ErrInvalidTagLevel, fmt.Sprintf("Tag '%s' can only be used on features, not scenarios (level: feature)", tag), lineNum))
	}

	return errors
}

// validateSchemaConstraints validates constraint rules defined in the schema.
func (v *Validator) validateSchemaConstraints(tags []string, lineNum int) []validation.ValidationError {
	var errors []validation.ValidationError

	// Find tags with constraints
	var constrainedTags []string
	for _, tag := range tags {
		if v.tagsConfig.HasConstraint(tag, "mutually_exclusive_with_taxonomy_levels") {
			constrainedTags = append(constrainedTags, tag)
		}
	}

	if len(constrainedTags) == 0 {
		return errors
	}

	// Check if any taxonomy-level tags are present
	taxonomyTags := v.tagsConfig.GetTaxonomyLevelTags()
	for _, tag := range tags {
		for _, taxonomyTag := range taxonomyTags {
			if tag == taxonomyTag {
				// Found a taxonomy tag - this violates the constraint
				errors = append(errors, *validation.NewValidationError(validation.ErrMutualExclusionViolation, fmt.Sprintf("Tags %v cannot be used with taxonomy level tags (found: %s)", constrainedTags, tag), lineNum))
				return errors
			}
		}
	}

	return errors
}

// validateGxPTagsFromSchema validates GxP-related tags using schema definitions.
func (v *Validator) validateGxPTagsFromSchema(tags []string, lineNum int) []validation.ValidationError {
	var errors []validation.ValidationError

	hasGxP := false
	hasGmpCriticalAspect := false
	hasControlTag := false

	for _, tag := range tags {
		tagDef, ok := v.tagsConfig.GetTag(tag)
		if !ok {
			continue
		}

		// Check for GxP regulatory tags
		if tagDef.Type == "gxp_regulatory" {
			if tag == "@gxp" {
				hasGxP = true
			}
			if tag == "@gmp-critical-aspect" {
				hasGmpCriticalAspect = true
			}
		}

		// Check for OSCAL control tags
		if tagDef.Type == "oscal_control" || tagDef.Type == "oscal_control_multi" {
			hasControlTag = true
		}
	}

	// @gmp-critical-aspect requires @gxp
	if hasGmpCriticalAspect && !hasGxP {
		errors = append(errors, *validation.NewValidationError(validation.ErrCriticalAspectRequiresGxP, "@gmp-critical-aspect tag requires @gxp tag to be present", lineNum))
	}

	// @gxp should have @control: tag (warning, not error per schema notes)
	if hasGxP && !hasControlTag {
		errors = append(errors, *validation.NewValidationError(validation.ErrGxPMissingControl, "@gxp tag should link to OSCAL controls using @control: tag (see testing-tags.yml note)", lineNum))
	}

	return errors
}

// validateTagsForScenario validates all tags for a single scenario.
func (v *Validator) validateTagsForScenario(tags []string, tagLines []int, scenarioLine int) []validation.ValidationError {
	var errors []validation.ValidationError

	// Tags config is required - fail if not available
	if v.tagsConfig == nil {
		errors = append(errors, *validation.NewValidationError(validation.ErrMissingTagsConfig, "Testing tags configuration not loaded - cannot validate tags", scenarioLine))
		return errors
	}

	// 1. Check verification tags (required)
	if !v.hasVerificationTag(tags) {
		formatter := validation.NewErrorFormatter()
		verificationTags := v.tagsConfig.GetVerificationTags()
		currentTags := ""
		if len(tags) > 0 {
			currentTags = strings.Join(tags, ", ")
		} else {
			currentTags = "(none)"
		}
		errors = append(errors, *formatter.FormatEnhancedError(
			validation.ErrMissingVerificationTag,
			"Scenario missing verification tag - must specify test level",
			fmt.Sprintf("Current tags: %s", currentTags),
			"Add one verification tag",
			fmt.Sprintf(`Available verification tags:
  %s

Example:
  @L0 @Automated
  Scenario: Generate executive summary
    Given a profile with security controls
    When AI generates a risk assessment
    Then the assessment should have an executive summary`, strings.Join(verificationTags, ", ")),
			fmt.Sprintf("Add one of these verification tags before the Scenario: %s", strings.Join(verificationTags, ", ")),
			scenarioLine,
		))
	}

	// Schema-driven tag validation
	for i, tag := range tags {
		lineNum := scenarioLine
		if i < len(tagLines) {
			lineNum = tagLines[i]
		}

		// Validate each tag against schema
		tagErrors := v.validateTagAgainstSchema(tag, lineNum, false)
		errors = append(errors, tagErrors...)
	}

	// Schema-driven constraint validation
	constraintErrors := v.validateSchemaConstraints(tags, scenarioLine)
	errors = append(errors, constraintErrors...)

	// Schema-driven GxP validation
	gxpErrors := v.validateGxPTagsFromSchema(tags, scenarioLine)
	errors = append(errors, gxpErrors...)

	return errors
}

// hasVerificationTag checks if tag list contains at least one verification tag.
func (v *Validator) hasVerificationTag(tags []string) bool {
	// Tags config must be available
	if v.tagsConfig == nil {
		return false
	}

	verificationTags := v.tagsConfig.GetVerificationTags()
	tagMap := make(map[string]bool)
	for _, vTag := range verificationTags {
		tagMap[vTag] = true
	}

	for _, tag := range tags {
		if tagMap[tag] {
			return true
		}
	}
	return false
}
