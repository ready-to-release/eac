package risk

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

// ControlEvidence represents a control with its evidence scenarios
type ControlEvidence struct {
	ControlID string
	Scenarios []ScenarioEvidence
}

// ScenarioEvidence represents evidence from a single scenario
type ScenarioEvidence struct {
	FeaturePath  string   // Relative path to .feature file
	FeatureName  string   // Feature name from Feature: line
	ScenarioName string   // Scenario name
	LineNumber   int      // Line number in file
	Tags         []string // All tags on scenario
	TestStatus   string   // passed, failed, skipped (populated from test results)
}

var (
	controlTagPattern  = regexp.MustCompile(`@control:([a-z]{2,4}-[0-9]+(?:\([0-9]+\))?)`)
	controlsTagPattern = regexp.MustCompile(`@controls:((?:[a-z]{2,4}-[0-9]+(?:\([0-9]+\))?,)*[a-z]{2,4}-[0-9]+(?:\([0-9]+\))?)`)
)

// ExtractControlEvidence scans feature files and extracts control tag evidence
func ExtractControlEvidence(workspaceRoot string, moduleName string) (map[string]*ControlEvidence, error) {
	evidence := make(map[string]*ControlEvidence)

	// Find all .feature files for module
	moduleFiles, err := findModuleFeatureFiles(workspaceRoot, moduleName)
	if err != nil {
		return nil, err
	}

	// Scan each file
	for _, featurePath := range moduleFiles {
		fileEvidence, err := extractFromFile(featurePath, workspaceRoot)
		if err != nil {
			log.Errorf("Warning: Failed to extract from %s: %v", featurePath, err)
			continue
		}

		// Merge into evidence map
		for controlID, scenarios := range fileEvidence {
			if evidence[controlID] == nil {
				evidence[controlID] = &ControlEvidence{
					ControlID: controlID,
					Scenarios: []ScenarioEvidence{},
				}
			}
			evidence[controlID].Scenarios = append(evidence[controlID].Scenarios, scenarios...)
		}
	}

	return evidence, nil
}

// extractFromFile extracts control evidence from a single feature file
func extractFromFile(featurePath string, workspaceRoot string) (map[string][]ScenarioEvidence, error) {
	content, err := os.ReadFile(featurePath)
	if err != nil {
		return nil, err
	}

	evidence := make(map[string][]ScenarioEvidence)
	lines := strings.Split(string(content), "\n")

	var (
		featureName  string
		scenarioTags []string
	)

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Extract feature name
		if strings.HasPrefix(trimmed, "Feature:") {
			featureName = strings.TrimSpace(strings.TrimPrefix(trimmed, "Feature:"))
			continue
		}

		// Collect tags (lines starting with @)
		if strings.HasPrefix(trimmed, "@") {
			tags := extractTags(trimmed)
			scenarioTags = append(scenarioTags, tags...)
			continue
		}

		// Scenario declaration
		if strings.HasPrefix(trimmed, "Scenario:") || strings.HasPrefix(trimmed, "Scenario Outline:") {
			scenarioName := strings.TrimSpace(strings.TrimPrefix(
				strings.TrimPrefix(trimmed, "Scenario Outline:"), "Scenario:"))

			// Check if scenario tags contain control tags
			controlIDs := extractControlIDsFromTags(scenarioTags)

			if len(controlIDs) > 0 {
				// Create scenario evidence
				relPath, _ := filepath.Rel(workspaceRoot, featurePath)
				for _, controlID := range controlIDs {
					scenario := ScenarioEvidence{
						FeaturePath:  relPath,
						FeatureName:  featureName,
						ScenarioName: scenarioName,
						LineNumber:   lineNum + 1,
						Tags:         scenarioTags,
						TestStatus:   "unknown", // Will be populated from test results
					}

					evidence[controlID] = append(evidence[controlID], scenario)
				}
			}

			// Reset for next scenario
			scenarioTags = []string{}
			continue
		}

		// Empty line or step - reset tags if not in scenario
		if trimmed == "" || isStep(trimmed) {
			scenarioTags = []string{}
		}
	}

	return evidence, nil
}

// extractTags extracts all tags from a line
func extractTags(line string) []string {
	var tags []string
	parts := strings.Fields(line)
	for _, part := range parts {
		if strings.HasPrefix(part, "@") {
			tags = append(tags, part)
		}
	}
	return tags
}

// extractControlIDsFromTags extracts control IDs from tag list
func extractControlIDsFromTags(tags []string) []string {
	var ids []string

	for _, tag := range tags {
		// Check @control:<id>
		if matches := controlTagPattern.FindStringSubmatch(tag); len(matches) > 1 {
			ids = append(ids, matches[1])
		}

		// Check @controls:<id1>,<id2>
		if matches := controlsTagPattern.FindStringSubmatch(tag); len(matches) > 1 {
			controlList := strings.Split(matches[1], ",")
			ids = append(ids, controlList...)
		}
	}

	return ids
}

// isStep checks if line is a Gherkin step
func isStep(line string) bool {
	return strings.HasPrefix(line, "Given ") ||
		strings.HasPrefix(line, "When ") ||
		strings.HasPrefix(line, "Then ") ||
		strings.HasPrefix(line, "And ") ||
		strings.HasPrefix(line, "But ")
}

// findModuleFeatureFiles finds all .feature files for a module
func findModuleFeatureFiles(workspaceRoot, moduleName string) ([]string, error) {
	// Use module contract to determine module's spec files
	// For now, assume specs/<module>/**/*.feature pattern

	specsDir := filepath.Join(workspaceRoot, "specs", moduleName)
	var files []string

	err := filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".feature") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}
