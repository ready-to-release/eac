// Package evidence provides test evidence loading for risk assessment.
package evidence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// FindLatestTestRun finds the most recent test run directory in out/test/.
// Test run directories are named with timestamps: "2025-11-29T14-30-00"
func FindLatestTestRun(workspaceRoot string) (string, error) {
	testDir := filepath.Join(workspaceRoot, "out", "test")

	entries, err := os.ReadDir(testDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no test runs found (directory does not exist)")
		}
		return "", fmt.Errorf("failed to read test directory: %w", err)
	}

	// Find timestamp directories (format: 2025-11-29T14-30-00)
	var timestampDirs []string
	for _, entry := range entries {
		if entry.IsDir() && strings.Contains(entry.Name(), "T") {
			timestampDirs = append(timestampDirs, entry.Name())
		}
	}

	if len(timestampDirs) == 0 {
		return "", fmt.Errorf("no test runs found in %s", testDir)
	}

	// Sort descending to get latest (timestamp format is lexicographically sortable)
	sort.Slice(timestampDirs, func(i, j int) bool {
		return timestampDirs[i] > timestampDirs[j]
	})

	return filepath.Join(testDir, timestampDirs[0]), nil
}

// FindTestResultsForModule discovers test result files for a given module.
// Test results are stored in: out/test/<timestamp>/<module>/
func FindTestResultsForModule(workspaceRoot, moduleName string) (*TestResults, error) {
	// Find latest test run directory
	latestTestRun, err := FindLatestTestRun(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("no test results found: %w", err)
	}

	// Look for module subdirectory
	moduleDir := filepath.Join(latestTestRun, moduleName)
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no test results for module '%s'", moduleName)
	}

	results := &TestResults{
		ModuleName:       moduleName,
		TestRunDirectory: latestTestRun,
	}

	// Collect test result files
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read module test directory: %w", err)
	}

	var latestModTime time.Time
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		fullPath := filepath.Join(moduleDir, filename)

		// Track modification time
		if info, err := entry.Info(); err == nil {
			if info.ModTime().After(latestModTime) {
				latestModTime = info.ModTime()
			}
		}

		// Categorize files
		if strings.HasSuffix(filename, ".cucumber.json") {
			results.AcceptanceFiles = append(results.AcceptanceFiles, fullPath)
		} else if strings.HasSuffix(filename, ".json") {
			results.UnitTestFiles = append(results.UnitTestFiles, fullPath)
		}
	}

	results.LastModified = latestModTime
	return results, nil
}

// FindAcceptanceTestResults finds acceptance test results for a module.
// Checks both the timestamped directory and the acceptance subdirectory.
func FindAcceptanceTestResults(workspaceRoot, moduleName string) ([]string, error) {
	var acceptanceFiles []string

	// Check timestamped test run directory
	results, err := FindTestResultsForModule(workspaceRoot, moduleName)
	if err == nil {
		acceptanceFiles = append(acceptanceFiles, results.AcceptanceFiles...)
	}

	// Also check acceptance directory: out/test/acceptance/<module>/
	acceptanceDir := filepath.Join(workspaceRoot, "out", "test", "acceptance", moduleName)
	if entries, err := os.ReadDir(acceptanceDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".cucumber.json") {
				acceptanceFiles = append(acceptanceFiles, filepath.Join(acceptanceDir, entry.Name()))
			}
		}
	}

	return acceptanceFiles, nil
}

// CucumberFeature represents a feature in Cucumber JSON format.
type CucumberFeature struct {
	URI         string            `json:"uri"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Tags        []CucumberTag     `json:"tags,omitempty"`
	Elements    []CucumberElement `json:"elements,omitempty"`
}

// CucumberTag represents a tag in Cucumber JSON.
type CucumberTag struct {
	Name string `json:"name"`
	Line int    `json:"line,omitempty"`
}

// CucumberElement represents a scenario or background in Cucumber JSON.
type CucumberElement struct {
	ID          string          `json:"id,omitempty"`
	Type        string          `json:"type"` // "scenario", "background"
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Tags        []CucumberTag   `json:"tags,omitempty"`
	Steps       []CucumberStep  `json:"steps,omitempty"`
}

// CucumberStep represents a step in Cucumber JSON.
type CucumberStep struct {
	Keyword string         `json:"keyword"`
	Name    string         `json:"name"`
	Result  CucumberResult `json:"result,omitempty"`
}

// CucumberResult represents the result of a step.
type CucumberResult struct {
	Status   string `json:"status"` // "passed", "failed", "skipped", "pending"
	Duration int64  `json:"duration,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// ParseCucumberResults parses a Cucumber JSON file.
func ParseCucumberResults(filePath string) ([]CucumberFeature, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cucumber file: %w", err)
	}

	var features []CucumberFeature
	if err := json.Unmarshal(data, &features); err != nil {
		return nil, fmt.Errorf("failed to parse cucumber JSON: %w", err)
	}

	return features, nil
}

// ExtractControlTags extracts @control(...) tags from Cucumber features.
// Returns a map of control ID to scenarios containing that control tag.
func ExtractControlTags(features []CucumberFeature) map[string][]string {
	controlMap := make(map[string][]string)
	controlRegex := regexp.MustCompile(`@control\(([^)]+)\)`)

	for _, feature := range features {
		// Check feature-level tags
		for _, tag := range feature.Tags {
			if matches := controlRegex.FindStringSubmatch(tag.Name); len(matches) > 1 {
				controlIDs := strings.Split(matches[1], ",")
				for _, id := range controlIDs {
					id = strings.TrimSpace(id)
					controlMap[id] = append(controlMap[id], feature.URI)
				}
			}
		}

		// Check scenario-level tags
		for _, element := range feature.Elements {
			if element.Type != "scenario" {
				continue
			}

			for _, tag := range element.Tags {
				if matches := controlRegex.FindStringSubmatch(tag.Name); len(matches) > 1 {
					controlIDs := strings.Split(matches[1], ",")
					for _, id := range controlIDs {
						id = strings.TrimSpace(id)
						location := fmt.Sprintf("%s:%s", feature.URI, element.Name)
						controlMap[id] = append(controlMap[id], location)
					}
				}
			}
		}
	}

	return controlMap
}

// CalculateTestSummary calculates pass/fail counts from Cucumber features.
func CalculateTestSummary(features []CucumberFeature) *TestSummary {
	summary := &TestSummary{}

	for _, feature := range features {
		for _, element := range feature.Elements {
			if element.Type != "scenario" {
				continue
			}

			summary.Total++

			// Determine scenario status from step results
			scenarioPassed := true
			scenarioSkipped := true

			for _, step := range element.Steps {
				switch step.Result.Status {
				case "passed":
					scenarioSkipped = false
				case "failed":
					scenarioPassed = false
					scenarioSkipped = false
				case "skipped":
					scenarioPassed = false
				case "pending":
					scenarioPassed = false
					scenarioSkipped = false
				}
			}

			if scenarioSkipped {
				summary.Skipped++
			} else if scenarioPassed {
				summary.Passed++
			} else {
				summary.Failed++
			}
		}
	}

	return summary
}

// GetTestSummaryForModule calculates test summary for a module from all test results.
func GetTestSummaryForModule(workspaceRoot, moduleName string) (*TestSummary, error) {
	acceptanceFiles, err := FindAcceptanceTestResults(workspaceRoot, moduleName)
	if err != nil || len(acceptanceFiles) == 0 {
		return nil, fmt.Errorf("no acceptance test results found for module '%s'", moduleName)
	}

	totalSummary := &TestSummary{}

	for _, file := range acceptanceFiles {
		features, err := ParseCucumberResults(file)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		summary := CalculateTestSummary(features)
		totalSummary.Total += summary.Total
		totalSummary.Passed += summary.Passed
		totalSummary.Failed += summary.Failed
		totalSummary.Skipped += summary.Skipped
	}

	return totalSummary, nil
}

// CollectAllTestEvidence collects test evidence from all result files.
func CollectAllTestEvidence(results *TestResults) ([]Evidence, error) {
	var evidence []Evidence

	for _, file := range results.UnitTestFiles {
		evidence = append(evidence, Evidence{
			Type:    EvidenceTypeUnitTest,
			Path:    file,
			ModTime: getFileModTime(file),
			Module:  results.ModuleName,
		})
	}

	for _, file := range results.AcceptanceFiles {
		evidence = append(evidence, Evidence{
			Type:    EvidenceTypeAcceptanceTest,
			Path:    file,
			ModTime: getFileModTime(file),
			Module:  results.ModuleName,
		})
	}

	return evidence, nil
}

// GetControlToTestMapping scans all test results to build control -> test mapping.
func GetControlToTestMapping(workspaceRoot, moduleName string) (map[string][]string, error) {
	acceptanceFiles, err := FindAcceptanceTestResults(workspaceRoot, moduleName)
	if err != nil {
		return nil, err
	}

	controlMap := make(map[string][]string)

	for _, file := range acceptanceFiles {
		features, err := ParseCucumberResults(file)
		if err != nil {
			continue
		}

		fileControlMap := ExtractControlTags(features)
		for controlID, locations := range fileControlMap {
			controlMap[controlID] = append(controlMap[controlID], locations...)
		}
	}

	return controlMap, nil
}
