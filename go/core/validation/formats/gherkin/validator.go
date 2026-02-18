package gherkin

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/validation"
)

// Validator validates Gherkin specifications against structure and tag domain.
type Validator struct {
	contract   *domain.Contract
	tagsConfig *config.TestingTagsConfig
}

// NewValidator creates a new Gherkin specification validator.
// Both parameters are required - tags config may be empty but must not be nil.
//
// Example:
//
//	loader := config.NewContractLoader(templateRoot, "gherkin-spec", paths.DefaultsVersion)
//	contract, _ := loader.LoadContract()
//	cfg, _ := config.Load(config.DefaultLoadOptions())
//	validator := gherkin.NewValidator(contract, cfg.TestingTags)
func NewValidator(contract *domain.Contract, tagsConfig *config.TestingTagsConfig) *Validator {
	return &Validator{
		contract:   contract,
		tagsConfig: tagsConfig,
	}
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
// Returns a list of validation errors (empty if valid).
func (v *Validator) Validate(output string, context map[string]interface{}) []validation.ValidationError {
	if strings.TrimSpace(output) == "" {
		return []validation.ValidationError{
			*validation.NewValidationError(validation.ErrEmptyOutput, "Generated output is empty", 0),
		}
	}

	lines := strings.Split(output, "\n")
	state := newValidationState()

	// Phase 1: Parse structure line-by-line
	errors := v.parseGherkinStructure(lines, state)

	// Phase 2: Validate required elements are present
	errors = append(errors, v.validateRequiredElements(state, output)...)

	// Phase 3: Validate rule-scenario relationships
	errors = append(errors, validateRuleScenarioRelationships(state)...)

	// Phase 4: Validate file complexity (rule/scenario counts)
	errors = append(errors, validateFileComplexity(state)...)

	// Phase 5: Validate tags on scenarios (verification, format, constraints)
	errors = append(errors, v.validateScenarioTags(lines)...)

	return errors
}

// newValidationState creates an initialized gherkinValidationState.
func newValidationState() *gherkinValidationState {
	return &gherkinValidationState{
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
}

// parseGherkinStructure walks each line of the Gherkin content, populating the
// validation state and returning any structural errors found during parsing.
func (v *Validator) parseGherkinStructure(lines []string, state *gherkinValidationState) []validation.ValidationError {
	var errors []validation.ValidationError

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Phase 0: Detect ACD artifact comments (must appear before Feature:)
		if !state.seenFeature {
			if strings.HasPrefix(trimmed, "# Intent:") {
				if !state.intentFound {
					content := strings.TrimSpace(strings.TrimPrefix(trimmed, "# Intent:"))
					state.intentFound = len(content) > 0
				}
				continue
			}
			if strings.HasPrefix(trimmed, "# Architecture:") {
				if !state.architectureFound {
					content := strings.TrimSpace(strings.TrimPrefix(trimmed, "# Architecture:"))
					state.architectureFound = len(content) > 0
				}
				continue
			}
		}

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "Feature:"):
			errors = append(errors, v.parseFeatureLine(trimmed, lineNum, state)...)
		case strings.HasPrefix(trimmed, "Background:"):
			errors = append(errors, parseBackgroundLine(lineNum, state)...)
		case strings.HasPrefix(trimmed, "Rule:"):
			errors = append(errors, parseRuleLine(line, trimmed, lineNum, state)...)
		case strings.HasPrefix(trimmed, "Scenario:"),
			strings.HasPrefix(trimmed, "Scenario Outline:"):
			errors = append(errors, parseScenarioLine(line, trimmed, lineNum, state)...)
		case strings.HasPrefix(trimmed, "Examples:"):
			errors = append(errors, parseExamplesLine(lineNum, state)...)
		}
	}

	return errors
}

// parseFeatureLine handles a Feature: declaration during structure parsing.
func (v *Validator) parseFeatureLine(trimmed string, lineNum int, state *gherkinValidationState) []validation.ValidationError {
	var errors []validation.ValidationError

	if state.seenFeature {
		errors = append(errors, *validation.NewValidationError(validation.ErrMultipleFeatures, "Multiple Feature declarations found (only one allowed per file)", lineNum))
	} else {
		state.seenFeature = true
		featureName := strings.TrimSpace(strings.TrimPrefix(trimmed, "Feature:"))

		// Validate naming convention
		namingError := v.validateFeatureNaming(featureName, lineNum)
		if namingError.Code != nil { // Non-empty error
			errors = append(errors, namingError)
		}
	}

	return errors
}

// parseBackgroundLine handles a Background: declaration during structure parsing.
func parseBackgroundLine(lineNum int, state *gherkinValidationState) []validation.ValidationError {
	var errors []validation.ValidationError

	if !state.seenFeature {
		errors = append(errors, *validation.NewValidationError(validation.ErrBackgroundBeforeFeature, "Background must come after Feature declaration", lineNum))
	}

	return errors
}

// parseRuleLine handles a Rule: declaration during structure parsing.
func parseRuleLine(line, trimmed string, lineNum int, state *gherkinValidationState) []validation.ValidationError {
	var errors []validation.ValidationError

	if !state.seenFeature {
		errors = append(errors, *validation.NewValidationError(validation.ErrRuleBeforeFeature, "Rule must come after Feature declaration", lineNum))
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

	return errors
}

// parseScenarioLine handles a Scenario: or Scenario Outline: declaration during
// structure parsing, including nesting and indentation checks.
func parseScenarioLine(line, trimmed string, lineNum int, state *gherkinValidationState) []validation.ValidationError {
	var errors []validation.ValidationError

	if !state.seenFeature {
		errors = append(errors, *validation.NewValidationError(validation.ErrScenarioBeforeFeature, "Scenario must come after Feature declaration", lineNum))
		return errors
	}

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
		return errors
	}

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
		errors = append(errors, *validation.NewValidationError(validation.ErrIncorrectIndentation, fmt.Sprintf("Scenario should be indented under its Rule (Rule at line %d has %d spaces, Scenario has %d spaces)", state.lastRuleLine, state.currentRuleIndent, indentLevel), lineNum))
	}

	return errors
}

// parseExamplesLine handles an Examples: declaration during structure parsing.
func parseExamplesLine(lineNum int, state *gherkinValidationState) []validation.ValidationError {
	var errors []validation.ValidationError

	if !state.seenScenario {
		errors = append(errors, *validation.NewValidationError(validation.ErrExamplesWithoutScenario, "Examples must come after Scenario Outline", lineNum))
	}

	return errors
}

// validateRequiredElements checks that the parsed state contains all required
// Gherkin elements (Feature, Rule, Scenario) and returns enhanced errors for
// any that are missing.
func (v *Validator) validateRequiredElements(state *gherkinValidationState, output string) []validation.ValidationError {
	var errors []validation.ValidationError
	formatter := validation.NewErrorFormatter()

	if !state.seenFeature {
		errors = append(errors, *formatter.FormatEnhancedError(
			validation.ErrMissingFeature,
			"Missing Feature: declaration - specification must start with a Feature",
			output,
			"Feature: <module>_<feature-name>",
			`Feature: eac-cli_risk-assessment

  Rule: Risk assessment generation
    Scenario: Generate executive summary
      Given a profile with security controls
      When AI generates a risk assessment
      Then the assessment should have an executive summary`,
			"Add a Feature: declaration at the start of the file with format '<module>_<feature-name>'.",
			0,
		))
	}

	if !state.seenRule {
		errors = append(errors, *formatter.FormatEnhancedError(
			validation.ErrMissingRule,
			"Missing Rule: declaration - features must contain acceptance criteria as Rules",
			"",
			"Feature with Rules",
			`Feature: eac-cli_risk-assessment

  Rule: Risk assessment generation
    Scenario: Generate executive summary
      ...

  Rule: Error handling
    Scenario: Handle invalid profile
      ...`,
			"Add Rule: declarations under the Feature to define acceptance criteria. Each Rule should have at least one Scenario.",
			0,
		))
	}

	if !state.seenScenario {
		errors = append(errors, *formatter.FormatEnhancedError(
			validation.ErrMissingScenario,
			"Missing Scenario: declaration - Rules must contain concrete examples as Scenarios",
			"",
			"Rule with Scenarios",
			`Rule: Risk assessment generation
  Scenario: Generate executive summary
    Given a profile with security controls
    When AI generates a risk assessment
    Then the assessment should have an executive summary

  Scenario: Include control statistics
    Given a profile with 21 controls
    When AI generates the summary
    Then it should include control satisfaction rates`,
			"Add Scenario: declarations under each Rule to provide concrete examples of the behavior.",
			0,
		))
	}

	if !state.intentFound {
		errors = append(errors, *formatter.FormatEnhancedError(
			validation.ErrMissingIntentComment,
			"Missing '# Intent:' comment — specification must declare its purpose on line 1",
			"",
			"# Intent: <one sentence describing why this change exists>",
			`# Intent: Enforce API validation so that breaking changes are caught before deployment
# Architecture: Affects eac-cli validate command; reads OpenAPI specs; depends on validation adapter

@deps:go @ov
Feature: eac-cli_api-contract-validation

  Rule: Contract validation gates merges

    @L2 @ov
    Scenario: Breaking change is rejected
      Given a spec with a removed endpoint
      When validation runs
      Then the merge is blocked`,
			"Add '# Intent: <why this change exists>' as the very first line of the file, before any tags or Feature: declaration.",
			1,
		))
	}
	if !state.architectureFound {
		errors = append(errors, *formatter.FormatEnhancedError(
			validation.ErrMissingArchitectureComment,
			"Missing '# Architecture:' comment — specification must declare the components it affects",
			"",
			"# Architecture: <affected components, systems, or constraints>",
			`# Intent: Enforce API validation so that breaking changes are caught before deployment
# Architecture: Affects eac-cli validate command; reads OpenAPI specs; depends on validation adapter

@deps:go @ov
Feature: eac-cli_api-contract-validation

  Rule: Contract validation gates merges

    @L2 @ov
    Scenario: Breaking change is rejected
      Given a spec with a removed endpoint
      When validation runs
      Then the merge is blocked`,
			"Add '# Architecture: <components affected>' on line 2, before any tags or Feature: declaration.",
			1,
		))
	}

	return errors
}

// validateRuleScenarioRelationships checks that every Rule has at least one
// Scenario and that no Scenarios exist outside of Rule blocks.
func validateRuleScenarioRelationships(state *gherkinValidationState) []validation.ValidationError {
	var errors []validation.ValidationError

	// Check for Rules without scenarios
	for i, rule := range state.allRules {
		ruleIndex := i + 1
		if !state.rulesWithScenarios[ruleIndex] {
			errors = append(errors, *validation.NewValidationError(validation.ErrRuleWithoutScenarios, fmt.Sprintf("Rule '%s' has no scenarios nested under it (each Rule must have at least one Scenario)", rule.Description), rule.Line))
		}
	}

	// Check for orphaned scenarios
	if len(state.scenariosOutsideRule) > 0 {
		errors = append(errors, *validation.NewValidationError(validation.ErrScenariosOutsideRule, fmt.Sprintf("Found %d scenario(s) not nested under any Rule (scenarios must be under Rule blocks)", len(state.scenariosOutsideRule)), state.scenariosOutsideRule[0]))
	}

	return errors
}

// validateFileComplexity checks rule and scenario counts against recommended
// thresholds and returns warnings or errors for oversized specifications.
func validateFileComplexity(state *gherkinValidationState) []validation.ValidationError {
	var errors []validation.ValidationError

	ruleCount := len(state.allRules)
	scenarioCount := len(state.allScenarios)

	// Rule count validation
	if ruleCount > 10 {
		errors = append(errors, *validation.NewValidationError(validation.ErrTooManyRules, fmt.Sprintf("File has %d Rules (>10 is too large - must split feature)", ruleCount), 0))
	} else if ruleCount > 6 {
		errors = append(errors, *validation.NewValidationError(validation.ErrLargeRuleCount, fmt.Sprintf("File has %d Rules (>6 is large - consider splitting feature)", ruleCount), 0))
	} else if ruleCount < 2 && ruleCount > 0 {
		errors = append(errors, *validation.NewValidationError(validation.ErrTooFewRules, fmt.Sprintf("File has %d Rule (2-6 Rules recommended for proper feature scope)", ruleCount), 0))
	}

	// Scenario count validation
	if scenarioCount > 30 {
		errors = append(errors, *validation.NewValidationError(validation.ErrTooManyScenarios, fmt.Sprintf("File has %d Scenarios (>30 is too large - must split feature)", scenarioCount), 0))
	} else if scenarioCount > 20 {
		errors = append(errors, *validation.NewValidationError(validation.ErrLargeScenarioCount, fmt.Sprintf("File has %d Scenarios (>20 should split for better maintainability)", scenarioCount), 0))
	}

	return errors
}

// getIndentLevel returns the indentation level (number of leading spaces/tabs) of a line.
func getIndentLevel(line string) int {
	count := 0
	for _, ch := range line {
		switch ch {
		case ' ':
			count++
		case '\t':
			count += 4 // Count tabs as 4 spaces
		default:
			return count // Found non-whitespace, return current count
		}
	}
	return count
}

// gherkinValidationState tracks Gherkin structure during validation.
type gherkinValidationState struct {
	seenFeature          bool
	seenRule             bool
	seenScenario         bool
	intentFound          bool           // # Intent: comment found before Feature:
	architectureFound    bool           // # Architecture: comment found before Feature:
	currentRuleIndex     int            // Track which Rule we're currently in
	rulesWithScenarios   map[int]bool   // Track which Rules have scenarios
	scenariosOutsideRule []int          // Track line numbers of scenarios not under any Rule
	allRules             []RuleInfo     // Track all Rules with metadata
	allScenarios         []ScenarioInfo // Track all Scenarios for counting
	currentRuleIndent    int            // Indentation level of current Rule
	lastRuleLine         int            // Line number of last Rule seen
}

// RuleInfo holds metadata about a Rule.
type RuleInfo struct {
	Line          int
	Description   string
	ScenarioCount int
	IndentLevel   int
}

// ScenarioInfo holds metadata about a Scenario.
type ScenarioInfo struct {
	Line        int
	Description string
	IndentLevel int
}
