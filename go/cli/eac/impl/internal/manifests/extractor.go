package manifests

import (
	"strings"

	implinternal "github.com/ready-to-release/eac/go/cli/eac/impl/internal"
)

// TestResult represents a test result extracted from manifests.
type TestResult struct {
	Name        string   `json:"name" yaml:"name"`
	Module      string   `json:"module" yaml:"module"`
	Package     string   `json:"package" yaml:"package"`
	Type        string   `json:"type" yaml:"type"`
	Suite       string   `json:"suite" yaml:"suite"`
	Status      string   `json:"status" yaml:"status"`
	DurationMs  int64    `json:"duration_ms,omitempty" yaml:"durationms,omitempty"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	FilePath    string   `json:"file_path,omitempty" yaml:"filepath,omitempty"`
	ControlTags []string `json:"control_tags,omitempty" yaml:"controltags,omitempty"`

	// Godog-specific fields
	FeatureName string `json:"feature_name,omitempty" yaml:"featurename,omitempty"`
	FeaturePath string `json:"feature_path,omitempty" yaml:"featurepath,omitempty"`
}

// ExtractTestResults extracts all test results from manifests.
func ExtractTestResults(manifests []*implinternal.TestManifest) []TestResult {
	var results []TestResult

	for _, manifest := range manifests {
		for i := range manifest.Tests {
			test := &manifest.Tests[i]
			result := TestResult{
				Name:        test.Name,
				Module:      manifest.Moniker,
				Package:     test.Package,
				Type:        test.Type,
				Suite:       test.Suite,
				Status:      test.Status,
				DurationMs:  test.DurationMs,
				Tags:        StripTagPrefixes(test.Tags),
				FilePath:    test.FilePath,
				ControlTags: ExtractControlTags(test.Tags),
			}

			// For godog tests, extract feature information
			if test.Type == "godog" && test.FilePath != "" {
				result.FeatureName = ExtractFeatureName(test.FilePath)
				result.FeaturePath = test.FilePath
			}

			results = append(results, result)
		}
	}

	return results
}

// StripTagPrefixes removes @ prefix from tags for display
// Tags like "@L0", "@control:ai-2" become "L0", "control:ai-2".
func StripTagPrefixes(tags []string) []string {
	result := make([]string, len(tags))
	for i, tag := range tags {
		if tag != "" && tag[0] == '@' {
			result[i] = tag[1:]
		} else {
			result[i] = tag
		}
	}
	return result
}

// ExtractControlTags extracts control IDs from tags array.
// Tags like "@control:ai-2" return ["ai-2"].
func ExtractControlTags(tags []string) []string {
	var controls []string

	for _, tag := range tags {
		if strings.HasPrefix(tag, "@control:") {
			// Extract control ID after "@control:"
			control := strings.TrimPrefix(tag, "@control:")
			if control != "" {
				controls = append(controls, control)
			}
		}
	}

	return controls
}

// ExtractFeatureName extracts the feature name from a feature file path.
// Examples:
//   - "specs/eac-cli/create-design/specification.feature" -> "create-design"
//   - "specs/core/logging/spec.feature" -> "logging"
func ExtractFeatureName(filePath string) string {
	if filePath == "" {
		return ""
	}

	// Normalize path separators
	// On Linux/container, filepath.ToSlash doesn't convert Windows backslashes
	// because they're not the OS separator. Explicitly replace them.
	filePath = strings.ReplaceAll(filePath, "\\", "/")

	// Check if it's a specs path
	if !strings.Contains(filePath, "/specs/") && !strings.HasPrefix(filePath, "specs/") {
		return ""
	}

	// Split by path separator
	parts := strings.Split(filePath, "/")

	// Find "specs" in the path
	specsIndex := -1
	for i, part := range parts {
		if part == "specs" {
			specsIndex = i
			break
		}
	}

	// Feature name is typically: specs/<module>/<feature-name>/file.feature
	// So it's 2 positions after "specs"
	if specsIndex >= 0 && specsIndex+2 < len(parts) {
		return parts[specsIndex+2]
	}

	return ""
}
