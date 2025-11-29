package unused

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// FeatureStep represents a step from a Gherkin feature file.
type FeatureStep struct {
	Text     string // step text without keyword (e.g., "I run the command \"foo\"")
	Keyword  string // Given, When, Then, And, But
	File     string // source file path
	Line     int    // line number
	Original string // original line for debugging
}

// stepKeywords are the Gherkin keywords that introduce steps.
var stepKeywords = []string{"Given ", "When ", "Then ", "And ", "But ", "* "}

// ParseFeatureFiles extracts all steps from the given feature files.
func ParseFeatureFiles(files []string) ([]FeatureStep, error) {
	var allSteps []FeatureStep

	for _, file := range files {
		steps, err := parseFeatureFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", file, err)
		}
		allSteps = append(allSteps, steps...)
	}

	return allSteps, nil
}

// parseFeatureFile parses a single .feature file for steps.
func parseFeatureFile(filePath string) ([]FeatureStep, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var steps []FeatureStep
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inDocString := false
	inDataTable := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			inDataTable = false // data tables end on empty line
			continue
		}

		// Skip comments
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Handle doc strings (""")
		if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, "'''") {
			inDocString = !inDocString
			continue
		}
		if inDocString {
			continue
		}

		// Handle data tables (lines starting with |)
		if strings.HasPrefix(trimmed, "|") {
			inDataTable = true
			continue
		}
		if inDataTable && !strings.HasPrefix(trimmed, "|") {
			inDataTable = false
		}

		// Skip non-step keywords
		if isNonStepKeyword(trimmed) {
			continue
		}

		// Check for step keywords
		for _, keyword := range stepKeywords {
			if strings.HasPrefix(trimmed, keyword) {
				stepText := strings.TrimPrefix(trimmed, keyword)
				steps = append(steps, FeatureStep{
					Text:     stepText,
					Keyword:  strings.TrimSpace(keyword),
					File:     filePath,
					Line:     lineNum,
					Original: line,
				})
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return steps, nil
}

// isNonStepKeyword checks if the line starts with a Gherkin keyword that is not a step.
func isNonStepKeyword(line string) bool {
	nonStepKeywords := []string{
		"Feature:", "Scenario:", "Scenario Outline:", "Background:",
		"Rule:", "Examples:", "@", "Feature :", "Scenario :",
	}
	for _, kw := range nonStepKeywords {
		if strings.HasPrefix(line, kw) {
			return true
		}
	}
	return false
}

// NormalizeForMatching prepares a step text for regex matching.
// It handles Scenario Outline placeholders by converting them to regex wildcards.
func NormalizeForMatching(stepText string) string {
	// Replace <placeholder> with a pattern that matches anything
	// This allows step definitions to match Scenario Outline steps
	placeholderPattern := regexp.MustCompile(`<[^>]+>`)
	normalized := placeholderPattern.ReplaceAllString(stepText, `[^"]*`)
	return normalized
}

// MatchesStep checks if a step definition matches a feature step.
func MatchesStep(stepDef StepDefinition, featureStep FeatureStep) bool {
	if stepDef.Regex == nil {
		return false
	}

	// Try direct match first
	if stepDef.Regex.MatchString(featureStep.Text) {
		return true
	}

	// Try with normalized placeholders (for Scenario Outline)
	normalized := NormalizeForMatching(featureStep.Text)
	if normalized != featureStep.Text {
		// Only try if normalization changed something
		if stepDef.Regex.MatchString(normalized) {
			return true
		}
	}

	return false
}
