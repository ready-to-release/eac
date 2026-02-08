// Command: test merge-results
// Short: Merge manual test results into test manifest
// Args:
// Long: Merge manual test results from test-results/<module>/<version>/manual-results.json
// Long: into the test output at out/test/<module>/.
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
// Long:   - Updated UoW manifest with manual test entries
// Long:   - Manual suite added/updated in suites object
// Long:   - Summary counts updated
// Long:   - Exit code 0 on success, non-zero on error
// Long:
// Long: Example:
// Long:   test merge-results --module eac-cli --version v1.2.0
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

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/fileutil"
	"github.com/ready-to-release/eac/go/core/config"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
)

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
	_, exists := repoCfg.Repository.GetModule(moduleFlag)
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

	// Transform manual results to manual-tests.json format (for testview parser)
	manualTestEntries := transformToManualTestsJSON(manualResults.Results)

	// Write manual-tests.json into UoW output dir
	uowDir := filepath.Join(workspaceRoot, "out", "test", moduleFlag, "manual-manual-manual")
	if err := os.MkdirAll(uowDir, 0755); err != nil {
		log.Errorf("creating UoW directory: %v", err)
		return 1
	}

	manualTestsPath := filepath.Join(uowDir, "manual-tests.json")
	if err := fileutil.AtomicWriteJSONWithLock(manualTestsPath, manualTestEntries, 0644); err != nil {
		log.Errorf("writing manual tests file: %v", err)
		return 1
	}

	// Compute artifact hash
	size, hash, err := coreoutput.HashFile(manualTestsPath)
	if err != nil {
		log.Errorf("hashing manual tests file: %v", err)
		return 1
	}

	// Parse test time
	testTime, parseErr := time.Parse(time.RFC3339, manualResults.ImportMetadata.TestTime)
	if parseErr != nil {
		testTime = time.Now()
	}

	// Create and save UoW manifest
	artifacts := []coreoutput.Artifact{
		{
			ID:     "manual-report",
			Path:   "manual-tests.json",
			SHA256: hash,
			Size:   size,
			Type:   "manual-report",
		},
	}

	uowManifest := &coreoutput.UoWManifest{
		Action:    core.ActionTest,
		Module:     moduleFlag,
		Component:  "manual",
		Tool:       "manual",
		Extra:      map[string]string{"testname": "manual"},
		ExitCode:   0,
		InputHash:  hash,
		ExecutedAt: testTime,
		Duration:   time.Duration(manualResults.ImportMetadata.DurationSeconds * float64(time.Second)),
		Artifacts:  artifacts,
		OutputHash: coreoutput.ComputeOutputHash(artifacts),
		Version:    "1.0.0",
	}

	if err := uowManifest.Save(workspaceRoot); err != nil {
		log.Errorf("writing UoW manifest: %v", err)
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
	log.Infof("  Location: %s", uowManifest.ManifestPath(workspaceRoot))
	log.Infof("  Manual tests: %d passed, %d failed, %d skipped", manualPassed, manualFailed, manualSkipped)

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

// manualTestsJSONFile is the format for manual-tests.json (read by testview.parseManualTestFile).
type manualTestsJSONFile struct {
	Tests []manualTestsJSONEntry `json:"tests"`
}

type manualTestsJSONEntry struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	DurationMs int64  `json:"duration_ms,omitempty"`
	Error      string `json:"error,omitempty"`
}

// transformToManualTestsJSON converts manual results to the format expected by testview's manual parser.
func transformToManualTestsJSON(results []ManualTestResult) *manualTestsJSONFile {
	file := &manualTestsJSONFile{
		Tests: make([]manualTestsJSONEntry, 0, len(results)),
	}
	for _, result := range results {
		file.Tests = append(file.Tests, manualTestsJSONEntry{
			Name:       extractScenarioName(result.ScenarioID),
			Status:     result.Status,
			DurationMs: int64(result.DurationSeconds * 1000),
			Error:      result.Error,
		})
	}
	return file
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

