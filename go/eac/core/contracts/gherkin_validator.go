package contracts

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// GherkinValidator validates Gherkin specifications against structure and tag contracts
type GherkinValidator struct {
	contract       *Contract
	tagsConfig     *config.TestingTagsConfig
	antiCorruption *AntiCorruptionRules
}

// NewGherkinValidator creates a new Gherkin specification validator
func NewGherkinValidator(contract *Contract, antiCorruption *AntiCorruptionRules) *GherkinValidator {
	return &GherkinValidator{
		contract:       contract,
		antiCorruption: antiCorruption,
	}
}

// NewGherkinValidatorWithTags creates a validator with both structure and tag configs
func NewGherkinValidatorWithTags(contract *Contract, tagsConfig *config.TestingTagsConfig, antiCorruption *AntiCorruptionRules) *GherkinValidator {
	return &GherkinValidator{
		contract:       contract,
		tagsConfig:     tagsConfig,
		antiCorruption: antiCorruption,
	}
}

// NewGherkinValidatorFromConfig creates a validator from the unified AI config
func NewGherkinValidatorFromConfig(loader *AIConfigLoader, typeName string, antiCorruption *AntiCorruptionRules) *GherkinValidator {
	// Get validation rules from the type config
	validation, err := loader.GetValidation(typeName)
	if err != nil {
		// Return validator without contract data
		return &GherkinValidator{
			antiCorruption: antiCorruption,
		}
	}

	// Convert validation map to Contract for backward compatibility
	contract := &Contract{
		Version: "0.1.0",
		Name:    "Gherkin Specification Structure",
		RawData: make(map[string]interface{}),
	}

	// Extract feature_naming_pattern from patterns.feature_naming
	if patterns := ExtractMap(validation, "patterns"); patterns != nil {
		if featureNaming := ExtractString(patterns, "feature_naming"); featureNaming != "" {
			contract.RawData["feature_naming_pattern"] = featureNaming
		}
	}

	// Extract required_verification_tags from required_tags
	if requiredTags := ExtractStringList(validation, "required_tags"); len(requiredTags) > 0 {
		// Convert []string to []interface{} for RawData compatibility
		tags := make([]interface{}, len(requiredTags))
		for i, t := range requiredTags {
			tags[i] = t
		}
		contract.RawData["required_verification_tags"] = tags
	}

	return &GherkinValidator{
		contract:       contract,
		antiCorruption: antiCorruption,
	}
}

// SetTagsConfig sets the tags config for tag validation
func (v *GherkinValidator) SetTagsConfig(tagsConfig *config.TestingTagsConfig) {
	v.tagsConfig = tagsConfig
}

// Validate validates Gherkin content against the specification contract
//
// This implements the Validator interface and checks:
// - Feature declaration exists and follows naming convention
// - Rule blocks present
// - Scenarios present under Rules
// - Proper keyword ordering
// - Hierarchical structure: Scenarios nested under Rules (Enhancement 1)
// - Each Rule has at least one Scenario (Enhancement 1)
// - No orphaned scenarios outside Rules (Enhancement 1)
// - File size constraints: 2-6 Rules ideal, >10 error (Enhancement 3)
// - Scenario count: 10-20 ideal, >30 error (Enhancement 3)
// - Proper indentation: Scenarios indented under Rules (Enhancement 5)
// - Verification tags present on scenarios (from testing-tags.yml)
// - Tag format validation (patterns from testing-tags.yml)
// - Skip reason validation
// - Mutual exclusion constraints (@Manual vs @L0-L4)
// - GxP tag requirements
// - Unknown tag warnings (validates against testing-tags.yml)
//
// Returns a list of validation errors (empty if valid)
func (v *GherkinValidator) Validate(output string, context map[string]interface{}) []ValidationError {
	var errors []ValidationError

	if strings.TrimSpace(output) == "" {
		errors = append(errors, ValidationError{
			Code:     "EMPTY_OUTPUT",
			Message:  "Generated output is empty",
			Severity: "error",
		})
		return errors
	}

	// Tags config is required for validation - fail immediately if not loaded
	if v.tagsConfig == nil {
		errors = append(errors, ValidationError{
			Code:     "MISSING_TAGS_CONFIG",
			Message:  "Testing tags configuration (testing-tags.yml) not loaded - cannot perform tag validation. Ensure .r2r/eac/testing-tags.yml exists and is valid.",
			Severity: "error",
		})
		return errors
	}

	lines := strings.Split(output, "\n")
	state := &gherkinValidationState{
		seenFeature:          false,
		seenRule:             false,
		seenScenario:         false,
		currentRuleIndex:     0,
		rulesWithScenarios:   make(map[int]bool),
		scenariosOutsideRule: []int{},
		allRules:             []RuleInfo{},
		allScenarios:         []ScenarioInfo{},
		currentRuleIndent:    -1,
		lastRuleLine:         0,
	}

	// Track feature name for naming convention validation
	var featureName string

	// Validate line by line
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check for Feature declaration
		if strings.HasPrefix(trimmed, "Feature:") {
			if state.seenFeature {
				errors = append(errors, ValidationError{
					Code:     "MULTIPLE_FEATURES",
					Message:  "Multiple Feature declarations found (only one allowed per file)",
					Line:     lineNum,
					Severity: "error",
				})
			} else {
				state.seenFeature = true
				featureName = strings.TrimSpace(strings.TrimPrefix(trimmed, "Feature:"))

				// Validate naming convention
				namingError := v.validateFeatureNaming(featureName, lineNum)
				if namingError.Code != "" { // Non-empty error
					errors = append(errors, namingError)
				}
			}
			continue
		}

		// Check for Background declaration
		if strings.HasPrefix(trimmed, "Background:") {
			if !state.seenFeature {
				errors = append(errors, ValidationError{
					Code:     "BACKGROUND_BEFORE_FEATURE",
					Message:  "Background must come after Feature declaration",
					Line:     lineNum,
					Severity: "error",
				})
			}
			continue
		}

		// Check for Rule declaration
		if strings.HasPrefix(trimmed, "Rule:") {
			if !state.seenFeature {
				errors = append(errors, ValidationError{
					Code:     "RULE_BEFORE_FEATURE",
					Message:  "Rule must come after Feature declaration",
					Line:     lineNum,
					Severity: "error",
				})
			} else {
				state.seenRule = true
				state.currentRuleIndex++
				state.lastRuleLine = lineNum

				ruleDesc := strings.TrimSpace(strings.TrimPrefix(trimmed, "Rule:"))
				indentLevel := getIndentLevel(line)
				state.currentRuleIndent = indentLevel

				state.allRules = append(state.allRules, RuleInfo{
					Line:          lineNum,
					Description:   ruleDesc,
					ScenarioCount: 0,
					IndentLevel:   indentLevel,
				})
			}
			continue
		}

		// Check for Scenario declaration
		if strings.HasPrefix(trimmed, "Scenario:") || strings.HasPrefix(trimmed, "Scenario Outline:") {
			if !state.seenFeature {
				errors = append(errors, ValidationError{
					Code:     "SCENARIO_BEFORE_FEATURE",
					Message:  "Scenario must come after Feature declaration",
					Line:     lineNum,
					Severity: "error",
				})
			} else {
				state.seenScenario = true

				scenarioDesc := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(trimmed, "Scenario Outline:"), "Scenario:"))
				indentLevel := getIndentLevel(line)

				// Track scenario metadata
				state.allScenarios = append(state.allScenarios, ScenarioInfo{
					Line:        lineNum,
					Description: scenarioDesc,
					IndentLevel: indentLevel,
				})

				// Enhancement 1: Check if scenario is under a Rule
				if !state.seenRule {
					// Scenario before any Rule
					state.scenariosOutsideRule = append(state.scenariosOutsideRule, lineNum)
				} else {
					// Check if scenario appears to be nested under current Rule
					// A scenario is considered "nested" if it appears after a Rule and before the next Rule
					if state.currentRuleIndex > 0 {
						state.rulesWithScenarios[state.currentRuleIndex] = true
						if state.currentRuleIndex <= len(state.allRules) {
							state.allRules[state.currentRuleIndex-1].ScenarioCount++
						}
					}

					// Enhancement 5: Check indentation (scenario should be indented more than Rule)
					if state.currentRuleIndent >= 0 && indentLevel <= state.currentRuleIndent {
						errors = append(errors, ValidationError{
							Code:     "INCORRECT_INDENTATION",
							Message:  fmt.Sprintf("Scenario should be indented under its Rule (Rule at line %d has %d spaces, Scenario has %d spaces)", state.lastRuleLine, state.currentRuleIndent, indentLevel),
							Line:     lineNum,
							Severity: "error",
						})
					}
				}
			}
			continue
		}

		// Check for Examples (must follow Scenario Outline)
		if strings.HasPrefix(trimmed, "Examples:") {
			if !state.seenScenario {
				errors = append(errors, ValidationError{
					Code:     "EXAMPLES_WITHOUT_SCENARIO",
					Message:  "Examples must come after Scenario Outline",
					Line:     lineNum,
					Severity: "error",
				})
			}
			continue
		}
	}

	// Final state validation
	if !state.seenFeature {
		errors = append(errors, ValidationError{
			Code:     "MISSING_FEATURE",
			Message:  "Missing Feature: declaration",
			Severity: "error",
		})
	}

	if !state.seenRule {
		errors = append(errors, ValidationError{
			Code:     "MISSING_RULE",
			Message:  "Missing Rule: declaration (required for acceptance criteria)",
			Severity: "error",
		})
	}

	if !state.seenScenario {
		errors = append(errors, ValidationError{
			Code:     "MISSING_SCENARIO",
			Message:  "Missing Scenario: declaration (required for behavior examples)",
			Severity: "error",
		})
	}

	// Enhancement 1: Check for Rules without scenarios
	for i, rule := range state.allRules {
		ruleIndex := i + 1
		if !state.rulesWithScenarios[ruleIndex] {
			errors = append(errors, ValidationError{
				Code:     "RULE_WITHOUT_SCENARIOS",
				Message:  fmt.Sprintf("Rule '%s' has no scenarios nested under it (each Rule must have at least one Scenario)", rule.Description),
				Line:     rule.Line,
				Severity: "error",
			})
		}
	}

	// Enhancement 1: Check for orphaned scenarios
	if len(state.scenariosOutsideRule) > 0 {
		errors = append(errors, ValidationError{
			Code:     "SCENARIOS_OUTSIDE_RULE",
			Message:  fmt.Sprintf("Found %d scenario(s) not nested under any Rule (scenarios must be under Rule blocks)", len(state.scenariosOutsideRule)),
			Line:     state.scenariosOutsideRule[0],
			Severity: "error",
		})
	}

	// Enhancement 3: File size and complexity warnings
	ruleCount := len(state.allRules)
	scenarioCount := len(state.allScenarios)

	// Rule count validation
	if ruleCount > 10 {
		errors = append(errors, ValidationError{
			Code:     "TOO_MANY_RULES",
			Message:  fmt.Sprintf("File has %d Rules (>10 is too large - must split feature)", ruleCount),
			Severity: "error",
		})
	} else if ruleCount > 6 {
		errors = append(errors, ValidationError{
			Code:     "LARGE_RULE_COUNT",
			Message:  fmt.Sprintf("File has %d Rules (>6 is large - consider splitting feature)", ruleCount),
			Severity: "error",
		})
	} else if ruleCount < 2 && ruleCount > 0 {
		errors = append(errors, ValidationError{
			Code:     "TOO_FEW_RULES",
			Message:  fmt.Sprintf("File has %d Rule (2-6 Rules recommended for proper feature scope)", ruleCount),
			Severity: "error",
		})
	}

	// Scenario count validation
	if scenarioCount > 30 {
		errors = append(errors, ValidationError{
			Code:     "TOO_MANY_SCENARIOS",
			Message:  fmt.Sprintf("File has %d Scenarios (>30 is too large - must split feature)", scenarioCount),
			Severity: "error",
		})
	} else if scenarioCount > 20 {
		errors = append(errors, ValidationError{
			Code:     "LARGE_SCENARIO_COUNT",
			Message:  fmt.Sprintf("File has %d Scenarios (>20 should split for better maintainability)", scenarioCount),
			Severity: "error",
		})
	}

	// Validate tags on scenarios (verification, format, constraints)
	tagErrors := v.validateScenarioTags(lines)
	errors = append(errors, tagErrors...)

	return errors
}

// validateFeatureNaming checks if feature name follows the naming convention
//
// Expected format: <module>_<feature-name>
// Pattern is read from contract.feature_naming_pattern
func (v *GherkinValidator) validateFeatureNaming(featureName string, lineNum int) ValidationError {
	// Get naming pattern from contract
	pattern := `^[a-z][a-z0-9-]*_[a-z][a-z0-9-]*$` // Default fallback
	if v.contract != nil && v.contract.RawData != nil {
		if patternVal, ok := v.contract.RawData["feature_naming_pattern"].(string); ok && patternVal != "" {
			pattern = patternVal
		}
	}

	matched, err := regexp.MatchString(pattern, featureName)
	if err != nil || !matched {
		return ValidationError{
			Code: "INVALID_FEATURE_NAMING",
			Message: fmt.Sprintf(
				"Feature name '%s' does not follow naming convention '<module>_<feature-name>' (lowercase with hyphens, e.g., 'eac-commands_user-auth')",
				featureName,
			),
			Line:     lineNum,
			Severity: "error",
		}
	}

	return ValidationError{} // No error (empty struct)
}

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
// - Rule tags are inherited by all Scenarios within that Rule
func (v *GherkinValidator) validateScenarioTags(lines []string) []ValidationError {
	var errors []ValidationError

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
// Returns a deduplicated list of all applicable tags
func (v *GherkinValidator) combineInheritedTags(featureTags, ruleTags, scenarioTags []string) []string {
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

// validateTagAgainstSchema validates a single tag against the schema definition
func (v *GherkinValidator) validateTagAgainstSchema(tag string, lineNum int, isFeature bool) []ValidationError {
	var errors []ValidationError

	// Get tag definition from schema
	tagDef, known := v.tagsConfig.GetTag(tag)

	if !known {
		// Unknown tag - warning only
		errors = append(errors, ValidationError{
			Code:     "UNKNOWN_TAG",
			Message:  fmt.Sprintf("Unknown tag '%s' not defined in testing-tags.yml", tag),
			Line:     lineNum,
			Severity: "warning",
		})
		return errors
	}

	// Validate tag format using schema pattern
	if formatErr := v.tagsConfig.ValidateTag(tag); formatErr != nil {
		errors = append(errors, ValidationError{
			Code:     "INVALID_TAG_FORMAT",
			Message:  formatErr.Error(),
			Line:     lineNum,
			Severity: "error",
		})
	}

	// Validate level constraint from schema
	if tagDef.Level == "scenario" && isFeature {
		errors = append(errors, ValidationError{
			Code:     "INVALID_TAG_LEVEL",
			Message:  fmt.Sprintf("Tag '%s' can only be used on scenarios, not features (level: scenario)", tag),
			Line:     lineNum,
			Severity: "error",
		})
	}

	if tagDef.Level == "feature" && !isFeature {
		errors = append(errors, ValidationError{
			Code:     "INVALID_TAG_LEVEL",
			Message:  fmt.Sprintf("Tag '%s' can only be used on features, not scenarios (level: feature)", tag),
			Line:     lineNum,
			Severity: "error",
		})
	}

	return errors
}

// validateSchemaConstraints validates constraint rules defined in the schema
func (v *GherkinValidator) validateSchemaConstraints(tags []string, lineNum int) []ValidationError {
	var errors []ValidationError

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
				errors = append(errors, ValidationError{
					Code:     "MUTUAL_EXCLUSION_VIOLATION",
					Message:  fmt.Sprintf("Tags %v cannot be used with taxonomy level tags (found: %s)", constrainedTags, tag),
					Line:     lineNum,
					Severity: "error",
				})
				return errors
			}
		}
	}

	return errors
}

// validateGxPTagsFromSchema validates GxP-related tags using schema definitions
func (v *GherkinValidator) validateGxPTagsFromSchema(tags []string, lineNum int) []ValidationError {
	var errors []ValidationError

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
		errors = append(errors, ValidationError{
			Code:     "CRITICAL_ASPECT_REQUIRES_GXP",
			Message:  "@gmp-critical-aspect tag requires @gxp tag to be present",
			Line:     lineNum,
			Severity: "error",
		})
	}

	// @gxp should have @control: tag (warning, not error per schema notes)
	if hasGxP && !hasControlTag {
		errors = append(errors, ValidationError{
			Code:     "GXP_MISSING_CONTROL",
			Message:  "@gxp tag should link to OSCAL controls using @control: tag (see testing-tags.yml note)",
			Line:     lineNum,
			Severity: "warning",
		})
	}

	return errors
}

// validateTagsForScenario validates all tags for a single scenario
func (v *GherkinValidator) validateTagsForScenario(tags []string, tagLines []int, scenarioLine int) []ValidationError {
	var errors []ValidationError

	// Tags config is required - fail if not available
	if v.tagsConfig == nil {
		errors = append(errors, ValidationError{
			Code:     "MISSING_TAGS_CONFIG",
			Message:  "Testing tags configuration not loaded - cannot validate tags",
			Line:     scenarioLine,
			Severity: "error",
		})
		return errors
	}

	// 1. Check verification tags (required)
	if !v.hasVerificationTag(tags) {
		verificationTags := v.tagsConfig.GetVerificationTags()
		errors = append(errors, ValidationError{
			Code:     "MISSING_VERIFICATION_TAG",
			Message:  fmt.Sprintf("Scenario missing verification tag (required: %s)", strings.Join(verificationTags, ", ")),
			Line:     scenarioLine,
			Severity: "error",
		})
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

// hasVerificationTag checks if tag list contains at least one verification tag
func (v *GherkinValidator) hasVerificationTag(tags []string) bool {
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

// isInTagList checks if a tag is in the given list
func (v *GherkinValidator) isInTagList(tag string, list []string) bool {
	for _, t := range list {
		if t == tag {
			return true
		}
	}
	return false
}

// getIndentLevel returns the indentation level (number of leading spaces/tabs) of a line
func getIndentLevel(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4 // Count tabs as 4 spaces
		} else {
			break
		}
	}
	return count
}

// VerifyImplementation verifies that the validator implements all contract rules
//
// This is a self-check to ensure the validator stays in sync with the contract
func (v *GherkinValidator) VerifyImplementation() []ValidationError {
	var errors []ValidationError

	// Check that contract is loaded
	if v.contract == nil {
		errors = append(errors, ValidationError{
			Code:     "NO_CONTRACT",
			Message:  "Contract not loaded",
			Severity: "error",
		})
		return errors
	}

	// Verify contract version matches
	expectedVersion := "0.1.0"
	if v.contract.Version != expectedVersion {
		errors = append(errors, ValidationError{
			Code:     "CONTRACT_VERSION_MISMATCH",
			Message:  fmt.Sprintf("Expected contract version %s, got %s", expectedVersion, v.contract.Version),
			Severity: "error",
		})
	}

	// Verify contract name
	if v.contract.Name != "Gherkin Specification Structure" {
		errors = append(errors, ValidationError{
			Code:     "CONTRACT_NAME_MISMATCH",
			Message:  fmt.Sprintf("Expected contract name 'Gherkin Specification Structure', got '%s'", v.contract.Name),
			Severity: "warning",
		})
	}

	// Verify tag contract is loaded (warning only, not required)
	if v.tagsConfig == nil {
		errors = append(errors, ValidationError{
			Code:     "NO_TAG_CONTRACT",
			Message:  "Tag contract not loaded - advanced tag validation disabled",
			Severity: "warning",
		})
	}

	return errors
}

// gherkinValidationState tracks Gherkin structure during validation
type gherkinValidationState struct {
	seenFeature          bool
	seenRule             bool
	seenScenario         bool
	currentRuleIndex     int              // Track which Rule we're currently in
	rulesWithScenarios   map[int]bool     // Track which Rules have scenarios
	scenariosOutsideRule []int            // Track line numbers of scenarios not under any Rule
	allRules             []RuleInfo       // Track all Rules with metadata
	allScenarios         []ScenarioInfo   // Track all Scenarios for counting
	currentRuleIndent    int              // Indentation level of current Rule
	lastRuleLine         int              // Line number of last Rule seen
}

// RuleInfo holds metadata about a Rule
type RuleInfo struct {
	Line          int
	Description   string
	ScenarioCount int
	IndentLevel   int
}

// ScenarioInfo holds metadata about a Scenario
type ScenarioInfo struct {
	Line        int
	Description string
	IndentLevel int
}
