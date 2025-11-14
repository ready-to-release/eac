package testing

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// FeatureFile represents parsed metadata from a .feature file
type FeatureFile struct {
	FilePath    string
	Module      string
	FeatureName string
	Title       string
	Description string
	Rules       []FeatureRule
	Scenarios   []string
}

// FeatureRule represents a Gherkin Rule block (ATDD)
type FeatureRule struct {
	Name        string
	Description string
}

// ParseFeatureFile extracts metadata from a .feature file
func ParseFeatureFile(path string) (*FeatureFile, error) {
	// Extract module and feature from path
	// Example: specs/src-cli/verify-configuration/specification.feature
	// -> Module: "src-cli", Feature: "verify-configuration"
	module, featureName := extractModuleAndFeatureFromPath(path)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	feature := &FeatureFile{
		FilePath:    path,
		Module:      module,
		FeatureName: featureName,
		Scenarios:   []string{},
		Rules:       []FeatureRule{},
	}

	scanner := bufio.NewScanner(file)
	inFeatureDescription := false
	inRuleDescription := false
	var currentRule *FeatureRule

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip comments and tags
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "@") {
			continue
		}

		// Feature title
		if strings.HasPrefix(trimmed, "Feature:") {
			feature.Title = strings.TrimSpace(strings.TrimPrefix(trimmed, "Feature:"))
			inFeatureDescription = true
			inRuleDescription = false
			continue
		}

		// Scenario or Scenario Outline
		if strings.HasPrefix(trimmed, "Scenario:") || strings.HasPrefix(trimmed, "Scenario Outline:") {
			scenarioName := strings.TrimPrefix(trimmed, "Scenario Outline:")
			scenarioName = strings.TrimPrefix(scenarioName, "Scenario:")
			scenarioName = strings.TrimSpace(scenarioName)
			feature.Scenarios = append(feature.Scenarios, scenarioName)
			inFeatureDescription = false
			inRuleDescription = false
			continue
		}

		// Rule
		if strings.HasPrefix(trimmed, "Rule:") {
			ruleName := strings.TrimSpace(strings.TrimPrefix(trimmed, "Rule:"))
			currentRule = &FeatureRule{Name: ruleName, Description: ""}
			feature.Rules = append(feature.Rules, *currentRule)
			inRuleDescription = true
			inFeatureDescription = false
			continue
		}

		// Background
		if strings.HasPrefix(trimmed, "Background:") {
			inFeatureDescription = false
			inRuleDescription = false
			continue
		}

		// Given/When/Then/And/But (step keywords) - stop description collection
		if isStepKeyword(trimmed) {
			inFeatureDescription = false
			inRuleDescription = false
			continue
		}

		// Feature description lines
		if inFeatureDescription {
			if feature.Description != "" {
				feature.Description += "\n"
			}
			feature.Description += "   " + trimmed
			continue
		}

		// Rule description lines
		if inRuleDescription && currentRule != nil {
			if currentRule.Description != "" {
				currentRule.Description += "\n"
			}
			currentRule.Description += "   " + trimmed
			// Update the rule in the slice
			feature.Rules[len(feature.Rules)-1] = *currentRule
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return feature, nil
}

// isStepKeyword checks if a line starts with a Gherkin step keyword
func isStepKeyword(line string) bool {
	keywords := []string{"Given", "When", "Then", "And", "But"}
	for _, kw := range keywords {
		if strings.HasPrefix(line, kw+" ") || strings.HasPrefix(line, kw+":") {
			return true
		}
	}
	return false
}

// extractModuleAndFeatureFromPath parses file path to extract module and feature names
// Expected format: specs/<module>/<feature>/specification.feature
// Example: specs/src-cli/verify-configuration/specification.feature
// OR absolute path: C:\projects\eac\specs\src-cli\verify-configuration\specification.feature
// Returns: "src-cli", "verify-configuration"
func extractModuleAndFeatureFromPath(path string) (module, feature string) {
	// Normalize to forward slashes
	path = filepath.ToSlash(path)

	// Split by "/"
	parts := strings.Split(path, "/")

	// Find "specs" in the path
	specsIndex := -1
	for i, part := range parts {
		if part == "specs" {
			specsIndex = i
			break
		}
	}

	// Expected format after "specs": <module>/<feature>/specification.feature
	if specsIndex >= 0 && len(parts) > specsIndex+2 {
		module = parts[specsIndex+1] // e.g., "src-cli"
		feature = parts[specsIndex+2] // e.g., "verify-configuration"
	} else {
		module = "unknown"
		feature = filepath.Base(filepath.Dir(path))
	}

	return module, feature
}

// FindFeatureFiles recursively finds all .feature files in a directory
func FindFeatureFiles(rootPath string) ([]string, error) {
	var featureFiles []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".feature") || strings.HasSuffix(path, ".spec")) {
			featureFiles = append(featureFiles, path)
		}

		return nil
	})

	return featureFiles, err
}
