// Package cucumber provides appendix rendering for feature specifications
package cucumber

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RenderAppendixA renders Appendix A with level 3 headers for each feature
// Format:
//
//	## Appendix A: Specifications and Test Results
//
//	### feature-name
//
//	```gherkin
//	<feature file content>
//	```
//
//	#### Test result for feature-name
//
//	For test results see cucumber.json in release artifact.
func RenderAppendixA(report CucumberReport, workspaceRoot string) string {
	var buf strings.Builder

	buf.WriteString("## Appendix A: Specifications and Test Results\n\n")

	for _, feature := range report {
		featureID := feature.GetFeatureID()

		// Extract feature name from Feature ID (e.g., "src-commands_ai-commit-generation" -> "ai-commit-generation")
		featureName := extractFeatureName(featureID)

		// Level 3 header with feature name
		buf.WriteString(fmt.Sprintf("### %s {#%s}\n\n", featureName, featureName))

		// Read and include the feature file content
		featureContent := readFeatureFile(feature.URI, workspaceRoot)
		if featureContent != "" {
			buf.WriteString("```gherkin\n")
			buf.WriteString(featureContent)
			buf.WriteString("\n```\n\n")
		}

		// Test results reference
		buf.WriteString(fmt.Sprintf("#### Test result for %s\n\n", featureName))
		buf.WriteString("For test results see cucumber.json in release artifact.\n\n")
	}

	return buf.String()
}

// extractFeatureName extracts the feature name from a Feature ID
// Example: "src-commands_ai-commit-generation" -> "ai-commit-generation"
func extractFeatureName(featureID string) string {
	// Split on underscore and take the last part
	parts := strings.Split(featureID, "_")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return featureID
}

// readFeatureFile reads the content of a feature file
// Handles relative paths from cucumber.json (e.g., "../../../specs/...")
func readFeatureFile(uri string, workspaceRoot string) string {
	// Normalize the path
	filePath := uri

	// If path is relative (starts with ../), resolve from workspace root
	if strings.HasPrefix(filePath, "../") {
		// Remove all ../ prefixes and treat as relative to workspace root
		for strings.HasPrefix(filePath, "../") {
			filePath = strings.TrimPrefix(filePath, "../")
		}
		filePath = filepath.Join(workspaceRoot, filePath)
	}

	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Sprintf("# Error reading feature file: %v", err)
	}

	return string(content)
}
