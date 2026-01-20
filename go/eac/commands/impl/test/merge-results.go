// Command: test merge-results
// Short: Merge manual test results into test manifest
// Args:
// Long: Merge manual test results from test-results/<module>/<version>/manual-results.json
// Long: into the test manifest at out/test/<module>/test.manifest.json.
// Long:
// Long: This command transforms manual test results into test entries and updates
// Long: the test manifest with aggregated statistics. If no manifest exists, a new
// Long: one is created. If a manual suite already exists, it is replaced.
// Long:
// Long: Transformation:
// Long:   - Each ManualTestResult becomes a test entry with type "manual"
// Long:   - Scenario name extracted from scenario_id (last path component)
// Long:   - Duration converted from seconds to milliseconds
// Long:   - Package set to "manual", suite set to "manual"
// Long:
// Long: Expected Output:
// Long:   - Updated test.manifest.json with manual test entries
// Long:   - Manual suite added/updated in suites object
// Long:   - Summary counts updated
// Long:   - Exit code 0 on success, non-zero on error
// Long:
// Long: Example:
// Long:   test merge-results --module eac-commands --version v1.2.0
// Flag.module: type=string, usage=Module moniker to merge results for (required)
// Flag.version: type=string, usage=Release version (required)
package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/impl/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/fileutil"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

func init() {
	registry.Register(MergeResults)
}

// MergeResults merges manual test results into the test manifest.
func MergeResults() int {
	// Parse command line arguments
	args := os.Args[2:] // Skip "test" and "merge-results"

	var moduleFlag, versionFlag string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--module":
			if i+1 < len(args) {
				moduleFlag = args[i+1]
				i++
			}
		case "--version":
			if i+1 < len(args) {
				versionFlag = args[i+1]
				i++
			}
		}
	}

	// Validate required flags
	if moduleFlag == "" {
		log.Errorf("--module flag is required")
		return 1
	}
	if versionFlag == "" {
		log.Errorf("--version flag is required")
		return 1
	}

	// Get workspace root
	workspaceRoot, err := os.Getwd()
	if err != nil {
		log.Errorf("getting working directory: %v", err)
		return 1
	}

	// Validate version format (basic semver or calver check)
	if err := validateVersionFormat(versionFlag); err != nil {
		log.Errorf("invalid version format: %v", err)
		return 1
	}

	// Load repository config
	repoCfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		log.Errorf("loading repository config: %v", err)
		return 1
	}

	// Validate module exists
	moduleConfig, exists := repoCfg.Repository.GetModule(moduleFlag)
	if !exists {
		log.Errorf("unknown module: %s", moduleFlag)
		return 1
	}

	// Load manual results file
	resultsPath := filepath.Join(workspaceRoot, "test-results", moduleFlag, versionFlag, "manual-results.json")
	if !fileExists(resultsPath) {
		log.Errorf("manual results file not found: %s", resultsPath)
		return 1
	}

	resultsData, err := os.ReadFile(resultsPath)
	if err != nil {
		log.Errorf("reading manual results file: %v", err)
		return 1
	}

	// Parse manual results
	var manualResults ManualTestResults
	if err := json.Unmarshal(resultsData, &manualResults); err != nil {
		log.Errorf("parsing manual results JSON: %v", err)
		return 1
	}

	// Load or create test manifest
	manifestPath := filepath.Join(workspaceRoot, "out", "test", moduleFlag, "test.manifest.json")
	var manifest *internal.TestManifest

	if fileExists(manifestPath) {
		// Load existing manifest
		manifestData, err := os.ReadFile(manifestPath)
		if err != nil {
			log.Errorf("reading test manifest: %v", err)
			return 1
		}

		manifest = &internal.TestManifest{}
		if err := json.Unmarshal(manifestData, manifest); err != nil {
			log.Errorf("parsing test manifest JSON: %v", err)
			return 1
		}

		log.Infof("Loaded existing test manifest for %s", moduleFlag)
	} else {
		// Create new manifest
		manifest = createNewManifest(moduleFlag, moduleConfig.GetComponentTypesDisplay(), &manualResults.ImportMetadata)
		log.Infof("Creating new test manifest for %s", moduleFlag)
	}

	// Remove existing manual tests from manifest
	manifest.Tests = filterNonManualTests(manifest.Tests)

	// Transform manual results to test entries
	manualTests := transformManualResults(manualResults.Results)
	manifest.Tests = append(manifest.Tests, manualTests...)

	// Update manual suite
	updateManualSuite(manifest, &manualResults.ImportMetadata, manualTests)

	// Recalculate summary
	updateSummary(manifest)

	// Write updated manifest atomically with locking (prevents concurrent merge corruption)
	manifestDir := filepath.Dir(manifestPath)
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		log.Errorf("creating manifest directory: %v", err)
		return 1
	}

	if err := fileutil.AtomicWriteJSONWithLock(manifestPath, manifest, 0644); err != nil {
		log.Errorf("writing manifest file: %v", err)
		return 1
	}

	// Report results
	manualPassed := 0
	manualFailed := 0
	manualSkipped := 0
	for _, result := range manualResults.Results {
		switch result.Status {
		case "passed":
			manualPassed++
		case "failed":
			manualFailed++
		case "skipped":
			manualSkipped++
		}
	}

	log.Infof("Merged manual test results for %s %s", moduleFlag, versionFlag)
	log.Infof("  Location: %s", manifestPath)
	log.Infof("  Manual tests: %d passed, %d failed, %d skipped", manualPassed, manualFailed, manualSkipped)
	log.Infof("  Total tests in manifest: %d", manifest.Summary.Total)

	return 0
}

// validateVersionFormat validates that the version string follows a recognized format
func validateVersionFormat(version string) error {
	// Semver pattern: v1.2.3, v1.2.3-alpha.1, etc.
	semverPattern := `^v?\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?$`
	// Calver pattern: v2024.01.19, etc.
	calverPattern := `^v?\d{4}\.\d{2}\.\d{2}(-[a-zA-Z0-9.-]+)?$`

	semverRe := regexp.MustCompile(semverPattern)
	calverRe := regexp.MustCompile(calverPattern)

	if semverRe.MatchString(version) || calverRe.MatchString(version) {
		return nil
	}

	return fmt.Errorf("version must be in semver (v1.2.3) or calver (v2024.01.19) format")
}

// createNewManifest creates a new test manifest
func createNewManifest(module, moduleType string, importMetadata *ImportMetadata) *internal.TestManifest {
	gitCommit := getGitCommitSHA()
	manifest := internal.NewTestManifest(module, moduleType, gitCommit)
	manifest.DurationSeconds = importMetadata.DurationSeconds
	return manifest
}

// filterNonManualTests removes manual tests from the test list
func filterNonManualTests(tests []internal.TestEntry) []internal.TestEntry {
	filtered := make([]internal.TestEntry, 0, len(tests))
	for _, test := range tests {
		if test.Suite != "manual" {
			filtered = append(filtered, test)
		}
	}
	return filtered
}

// transformManualResults transforms manual test results into test entries
func transformManualResults(results []ManualTestResult) []internal.TestEntry {
	entries := make([]internal.TestEntry, 0, len(results))
	for _, result := range results {
		entry := internal.TestEntry{
			Name:       extractScenarioName(result.ScenarioID),
			Package:    "manual",
			Type:       "manual",
			Suite:      "manual",
			Status:     result.Status,
			DurationMs: int64(result.DurationSeconds * 1000),
			Tags:       []string{},
			FilePath:   "",
		}

		if result.Error != "" {
			entry.Error = result.Error
		}

		entries = append(entries, entry)
	}
	return entries
}

// extractScenarioName extracts the scenario name from a scenario ID
// Format: module/feature/scenario-slug
func extractScenarioName(scenarioID string) string {
	parts := strings.Split(scenarioID, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return scenarioID
}

// updateManualSuite updates or creates the manual suite in the manifest
func updateManualSuite(manifest *internal.TestManifest, importMetadata *ImportMetadata, manualTests []internal.TestEntry) {
	// Count manual test results by status
	summary := internal.TestSummary{
		Total:   len(manualTests),
		Passed:  0,
		Failed:  0,
		Skipped: 0,
	}

	for _, test := range manualTests {
		switch test.Status {
		case "passed":
			summary.Passed++
		case "failed":
			summary.Failed++
		case "skipped":
			summary.Skipped++
		}
	}

	// Parse test time (ISO 8601 format)
	testTime, err := time.Parse(time.RFC3339, importMetadata.TestTime)
	if err != nil {
		// Fallback to current time if parsing fails
		testTime = time.Now()
	}

	// Update or create manual suite
	manifest.Suites["manual"] = internal.SuiteResult{
		RunTime:         testTime,
		DurationSeconds: importMetadata.DurationSeconds,
		Tests:           summary,
	}
}

// updateSummary recalculates the manifest summary from all tests
func updateSummary(manifest *internal.TestManifest) {
	summary := internal.TestSummary{
		Total:     len(manifest.Tests),
		Passed:    0,
		Failed:    0,
		Skipped:   0,
		Pending:   0,
		Undefined: 0,
	}

	for _, test := range manifest.Tests {
		switch test.Status {
		case "passed":
			summary.Passed++
		case "failed":
			summary.Failed++
		case "skipped":
			summary.Skipped++
		case "pending":
			summary.Pending++
		case "undefined":
			summary.Undefined++
		}
	}

	manifest.Summary = summary
}
