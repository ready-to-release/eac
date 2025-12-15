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

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// FindLatestTestRun finds the most recent test run directory in out/test/.
// Test run directories are named with timestamps: "2025-11-29T14-30-00"
func FindLatestTestRun(workspaceRoot string) (string, error) {
	// Use helper function for clean fallback to defaults in test environments
	testDir := config.GetTestOutputPath(workspaceRoot)

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

// FindTestResultsForModuleInSuite discovers test result files for a given module in a specific test suite.
// Test results are stored in: out/test/<suite>/<module>/
// This function recursively scans subdirectories to handle test-package organization.
func FindTestResultsForModuleInSuite(workspaceRoot, moduleName, suiteName string) (*TestResults, error) {
	// Use helper function for clean fallback to defaults in test environments
	testRunDir := config.GetTestSuiteOutputPath(workspaceRoot, suiteName)

	// Look for module test results in the suite directory
	moduleDir := filepath.Join(testRunDir, moduleName)

	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no test results for module '%s' in suite '%s'", moduleName, suiteName)
	}

	results := &TestResults{
		ModuleName:       moduleName,
		TestRunDirectory: moduleDir, // Use module-specific directory for freshness checks
	}

	// Recursively collect test result files from module directory and subdirectories
	var latestModTime time.Time
	err := filepath.WalkDir(moduleDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		filename := d.Name()

		// Track modification time
		if info, err := d.Info(); err == nil {
			if info.ModTime().After(latestModTime) {
				latestModTime = info.ModTime()
			}
		}

		// Categorize files
		// Accept both naming patterns: cucumber-*.json and *.cucumber.json
		if isCucumberFile(filename) {
			results.AcceptanceFiles = append(results.AcceptanceFiles, path)
		} else if strings.HasSuffix(filename, ".json") {
			results.UnitTestFiles = append(results.UnitTestFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan module test directory: %w", err)
	}

	results.LastModified = latestModTime
	return results, nil
}

// isCucumberFile checks if a filename is a cucumber test result file.
// Accepts both patterns: cucumber-*.json and *.cucumber.json
func isCucumberFile(filename string) bool {
	if !strings.HasSuffix(filename, ".json") {
		return false
	}
	// Pattern 1: *.cucumber.json (e.g., "repository.cucumber.json")
	if strings.HasSuffix(filename, ".cucumber.json") {
		return true
	}
	// Pattern 2: cucumber-*.json (e.g., "cucumber-repository.json")
	if strings.HasPrefix(filename, "cucumber-") {
		return true
	}
	return false
}

// FindTestResultsForModule discovers test result files for a given module.
// Test results are stored in: out/test/<timestamp>/<module>/
// This function recursively scans subdirectories to handle test-package organization.
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

	// Recursively collect test result files from module directory and subdirectories
	var latestModTime time.Time
	err = filepath.WalkDir(moduleDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		filename := d.Name()

		// Track modification time
		if info, err := d.Info(); err == nil {
			if info.ModTime().After(latestModTime) {
				latestModTime = info.ModTime()
			}
		}

		// Categorize files
		// Accept both naming patterns: cucumber-*.json and *.cucumber.json
		if isCucumberFile(filename) {
			results.AcceptanceFiles = append(results.AcceptanceFiles, path)
		} else if strings.HasSuffix(filename, ".json") {
			results.UnitTestFiles = append(results.UnitTestFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan module test directory: %w", err)
	}

	results.LastModified = latestModTime
	return results, nil
}

// FindAcceptanceTestResults finds acceptance test results for a module.
// Checks both the timestamped directory and the acceptance subdirectory.
// Recursively scans subdirectories to handle test-package organization.
func FindAcceptanceTestResults(workspaceRoot, moduleName string) ([]string, error) {
	var acceptanceFiles []string

	// Check timestamped test run directory
	results, err := FindTestResultsForModule(workspaceRoot, moduleName)
	if err == nil {
		acceptanceFiles = append(acceptanceFiles, results.AcceptanceFiles...)
	}

	// Also check acceptance directory: out/test/acceptance/<module>/
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		return acceptanceFiles, nil // Return what we have so far
	}
	acceptanceDir := cfg.Repository.TestModuleOutputPathAbs(workspaceRoot, "acceptance", moduleName)

	// Recursively scan acceptance directory for cucumber files
	if _, err := os.Stat(acceptanceDir); err == nil {
		filepath.WalkDir(acceptanceDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // Continue on errors
			}
			if !d.IsDir() && isCucumberFile(d.Name()) {
				acceptanceFiles = append(acceptanceFiles, path)
			}
			return nil
		})
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

// ExtractControlTags extracts @control:<id> and @controls:<id1>,<id2> tags from Cucumber features.
// Returns a map of control ID to scenarios containing that control tag.
// Supports new OSCAL tag format: @control:ac-2 and @controls:ac-2,au-3
func ExtractControlTags(features []CucumberFeature) map[string][]string {
	controlMap := make(map[string][]string)

	// New OSCAL tag patterns
	controlTagPattern := regexp.MustCompile(`@control:([a-z]{2,4}-[0-9]+(?:\([0-9]+\))?)`)
	controlsTagPattern := regexp.MustCompile(`@controls:((?:[a-z]{2,4}-[0-9]+(?:\([0-9]+\))?,)*[a-z]{2,4}-[0-9]+(?:\([0-9]+\))?)`)

	for _, feature := range features {
		// Check feature-level tags
		for _, tag := range feature.Tags {
			// Check @control:<id>
			if matches := controlTagPattern.FindStringSubmatch(tag.Name); len(matches) > 1 {
				controlID := matches[1]
				controlMap[controlID] = append(controlMap[controlID], feature.URI)
			}

			// Check @controls:<id1>,<id2>
			if matches := controlsTagPattern.FindStringSubmatch(tag.Name); len(matches) > 1 {
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
				location := fmt.Sprintf("%s:%s", feature.URI, element.Name)

				// Check @control:<id>
				if matches := controlTagPattern.FindStringSubmatch(tag.Name); len(matches) > 1 {
					controlID := matches[1]
					controlMap[controlID] = append(controlMap[controlID], location)
				}

				// Check @controls:<id1>,<id2>
				if matches := controlsTagPattern.FindStringSubmatch(tag.Name); len(matches) > 1 {
					controlIDs := strings.Split(matches[1], ",")
					for _, id := range controlIDs {
						id = strings.TrimSpace(id)
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

// GetTestSummaryForModule calculates test summary for a module from test results in a specific suite.
func GetTestSummaryForModule(workspaceRoot, moduleName, testSuite string) (*TestSummary, error) {
	// Use suite-based lookup
	testResults, err := FindTestResultsForModuleInSuite(workspaceRoot, moduleName, testSuite)
	if err != nil {
		return nil, fmt.Errorf("no test results found for module '%s' in suite '%s'", moduleName, testSuite)
	}

	totalSummary := &TestSummary{}

	// Process cucumber files
	for _, file := range testResults.AcceptanceFiles {
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

	// Also count unit test files (Go tests)
	// Each .json file represents one test package
	for _, file := range testResults.UnitTestFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		// Parse Go test JSON (one JSON object per line)
		lines := strings.Split(string(data), "\n")
		testsPassed := 0
		testsFailed := 0
		testsSkipped := 0
		seenTests := make(map[string]bool)

		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			var result struct {
				Action string `json:"Action"`
				Test   string `json:"Test,omitempty"`
			}
			if err := json.Unmarshal([]byte(line), &result); err != nil {
				continue
			}

			// Only count test-level results (not package-level)
			if result.Test == "" {
				continue
			}

			// Avoid double-counting tests that have multiple action lines
			testKey := result.Test
			if result.Action == "pass" && !seenTests[testKey] {
				testsPassed++
				seenTests[testKey] = true
			} else if result.Action == "fail" && !seenTests[testKey] {
				testsFailed++
				seenTests[testKey] = true
			} else if result.Action == "skip" && !seenTests[testKey] {
				testsSkipped++
				seenTests[testKey] = true
			}
		}

		totalSummary.Total += testsPassed + testsFailed + testsSkipped
		totalSummary.Passed += testsPassed
		totalSummary.Failed += testsFailed
		totalSummary.Skipped += testsSkipped
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

// GetControlTestEvidence builds comprehensive control evidence directly from cucumber test results.
// This function extracts @control: tags from cucumber JSON and builds evidence with test status.
// This approach is more reliable than parsing feature files separately because:
// - Cucumber already parses and validates the Gherkin
// - Tags and test results are in the same file (no path matching needed)
// - Only includes tests that actually ran in the suite
func GetControlTestEvidence(workspaceRoot, moduleName, testSuite string) (map[string]*ControlTestEvidence, error) {
	evidenceMap := make(map[string]*ControlTestEvidence)

	// Load test results from the specified test suite
	testResults, err := FindTestResultsForModuleInSuite(workspaceRoot, moduleName, testSuite)
	if err != nil {
		// No test results means no evidence (module may not have tests in this suite)
		return evidenceMap, nil
	}

	// Build evidence directly from cucumber JSON files
	for _, cucumberFile := range testResults.AcceptanceFiles {
		features, err := ParseCucumberResults(cucumberFile)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		// Extract control evidence from each feature
		extractControlEvidenceFromCucumber(evidenceMap, features)
	}

	// Calculate test counts
	for _, evidence := range evidenceMap {
		evidence.TotalTests = len(evidence.Tests)
		evidence.HasEvidence = evidence.TotalTests > 0

		for _, test := range evidence.Tests {
			switch test.Status {
			case "passed":
				evidence.PassedTests++
			case "failed":
				evidence.FailedTests++
			case "skipped":
				evidence.SkippedTests++
			}
		}
	}

	return evidenceMap, nil
}

// extractControlEvidenceFromCucumber extracts control tags and test evidence from cucumber JSON.
func extractControlEvidenceFromCucumber(evidenceMap map[string]*ControlTestEvidence, features []CucumberFeature) {
	controlTagPattern := regexp.MustCompile(`@control:([a-z]{2,4}-[0-9]+(?:\([0-9]+\))?)`)
	controlsTagPattern := regexp.MustCompile(`@controls:((?:[a-z]{2,4}-[0-9]+(?:\([0-9]+\))?,)*[a-z]{2,4}-[0-9]+(?:\([0-9]+\))?)`)

	for _, feature := range features {
		// Extract feature-level control tags
		featureControls := extractControlsFromTags(feature.Tags, controlTagPattern, controlsTagPattern)

		for _, element := range feature.Elements {
			if element.Type != "scenario" {
				continue
			}

			// Extract scenario-level control tags
			scenarioControls := extractControlsFromTags(element.Tags, controlTagPattern, controlsTagPattern)

			// Combine feature and scenario controls (scenario tags override)
			allControls := make(map[string]bool)
			for _, ctrl := range featureControls {
				allControls[ctrl] = true
			}
			for _, ctrl := range scenarioControls {
				allControls[ctrl] = true
			}

			// Determine test status from steps
			status := determineScenarioStatus(element.Steps)

			// Create test evidence for each control
			for controlID := range allControls {
				testEv := TestEvidence{
					FeaturePath:  feature.URI,
					FeatureName:  feature.Name,
					ScenarioName: element.Name,
					Status:       status,
					Location:     fmt.Sprintf("%s:%s", filepath.Base(feature.URI), element.Name),
				}

				// Initialize control evidence if not exists
				if evidenceMap[controlID] == nil {
					evidenceMap[controlID] = &ControlTestEvidence{
						ControlID: controlID,
						Tests:     []TestEvidence{},
					}
				}

				// Add test evidence
				evidenceMap[controlID].Tests = append(evidenceMap[controlID].Tests, testEv)
			}
		}
	}
}

// extractControlsFromTags extracts control IDs from cucumber tags.
func extractControlsFromTags(tags []CucumberTag, controlPattern, controlsPattern *regexp.Regexp) []string {
	var controls []string

	for _, tag := range tags {
		// Check @control:<id>
		if matches := controlPattern.FindStringSubmatch(tag.Name); len(matches) > 1 {
			controls = append(controls, matches[1])
		}

		// Check @controls:<id1>,<id2>
		if matches := controlsPattern.FindStringSubmatch(tag.Name); len(matches) > 1 {
			ids := strings.Split(matches[1], ",")
			for _, id := range ids {
				controls = append(controls, strings.TrimSpace(id))
			}
		}
	}

	return controls
}

// determineScenarioStatus determines the status of a scenario from its steps.
func determineScenarioStatus(steps []CucumberStep) string {
	if len(steps) == 0 {
		return "not-run"
	}

	status := "passed"
	allSkipped := true

	for _, step := range steps {
		if step.Result.Status == "failed" {
			return "failed"
		} else if step.Result.Status == "passed" {
			allSkipped = false
		} else if step.Result.Status == "skipped" || step.Result.Status == "pending" {
			status = "skipped"
		}
	}

	if allSkipped {
		return "skipped"
	}

	return status
}

