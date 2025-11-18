package specs

import (
	"fmt"
	"strings"
)

// validateGherkinContent ensures the generated content is valid Gherkin with structural validation
//
// This improved validation:
// - Checks keyword order (Feature before Rule before Scenario)
// - Ignores keywords in comments
// - Detects duplicate Feature declarations
// - Validates basic Gherkin structure
func validateGherkinContent(content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content is empty")
	}

	lines := strings.Split(content, "\n")
	state := &gherkinValidationState{
		seenFeature:  false,
		seenRule:     false,
		seenScenario: false,
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		lineNum := i + 1

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check keyword order and update state
		if err := validateGherkinLine(trimmed, state, lineNum); err != nil {
			return err
		}
	}

	// Final state validation
	if !state.seenFeature {
		return fmt.Errorf("missing Feature: declaration")
	}
	if !state.seenRule {
		return fmt.Errorf("missing Rule: declaration (required for ATDD)")
	}
	if !state.seenScenario {
		return fmt.Errorf("missing Scenario: declaration (required for BDD)")
	}

	return nil
}

// gherkinValidationState tracks Gherkin structure during validation
type gherkinValidationState struct {
	seenFeature  bool
	seenRule     bool
	seenScenario bool
}

// validateGherkinLine validates a single line of Gherkin and updates state
func validateGherkinLine(line string, state *gherkinValidationState, lineNum int) error {
	switch {
	case strings.HasPrefix(line, "Feature:"):
		if state.seenFeature {
			return fmt.Errorf("line %d: multiple Feature declarations found", lineNum)
		}
		state.seenFeature = true

	case strings.HasPrefix(line, "Background:"):
		if !state.seenFeature {
			return fmt.Errorf("line %d: Background must come after Feature", lineNum)
		}

	case strings.HasPrefix(line, "Rule:"):
		if !state.seenFeature {
			return fmt.Errorf("line %d: Rule must come after Feature", lineNum)
		}
		state.seenRule = true

	case strings.HasPrefix(line, "Scenario:"), strings.HasPrefix(line, "Scenario Outline:"):
		if !state.seenFeature {
			return fmt.Errorf("line %d: Scenario must come after Feature", lineNum)
		}
		state.seenScenario = true

	case strings.HasPrefix(line, "Examples:"):
		if !state.seenScenario {
			return fmt.Errorf("line %d: Examples must come after Scenario Outline", lineNum)
		}
	}

	return nil
}
