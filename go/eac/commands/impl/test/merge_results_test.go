//go:build L1
// +build L1

package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==============================================================================
// MergeResults Command Tests
// ==============================================================================

func TestMergeResultsCommand_HappyPath(t *testing.T) {
	tests := []struct {
		name               string
		module             string
		version            string
		manualResults      ManualTestResults
		existingManifest   *TestManifest
		wantNewManifest    bool
		wantTestCount      int
		wantManualSuite    bool
		wantPassedCount    int
		wantFailedCount    int
		wantSkippedCount   int
	}{
		{
			name:    "merge into non-existent manifest creates new",
			module:  "test-module",
			version: "v1.0.0",
			manualResults: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:        "2026-01-19T12:00:00Z",
					Tester:          "tester@example.com",
					Module:          "test-module",
					ReleaseVersion:  "v1.0.0",
					SchemaVersion:   "1.0",
					DurationSeconds: 120,
				},
				Results: []ManualTestResult{
					{
						ScenarioID:      "test-module/feature1/manual-scenario",
						Status:          "passed",
						DurationSeconds: 30,
						Notes:           "Test passed successfully",
					},
				},
			},
			existingManifest: nil,
			wantNewManifest:  true,
			wantTestCount:    1,
			wantManualSuite:  true,
			wantPassedCount:  1,
			wantFailedCount:  0,
			wantSkippedCount: 0,
		},
		{
			name:    "merge into existing manifest appends tests",
			module:  "test-module",
			version: "v1.0.0",
			manualResults: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:        "2026-01-19T12:00:00Z",
					Tester:          "tester@example.com",
					Module:          "test-module",
					ReleaseVersion:  "v1.0.0",
					SchemaVersion:   "1.0",
					DurationSeconds: 60,
				},
				Results: []ManualTestResult{
					{
						ScenarioID:      "test-module/feature1/manual-scenario-1",
						Status:          "passed",
						DurationSeconds: 30,
					},
				},
			},
			existingManifest: &TestManifest{
				TestID:          "existing-test-id",
				TestAgent:       "devbox",
				Moniker:         "test-module",
				Type:            "go",
				TestTime:        "2026-01-19T11:00:00Z",
				DurationSeconds: 45.5,
				GitCommit:       "1234567890abcdef1234567890abcdef12345678",
				InputHash:       "abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab",
				Summary: TestSummary{
					Total:   2,
					Passed:  2,
					Failed:  0,
					Skipped: 0,
				},
				Suites: map[string]SuiteResult{
					"unit": {
						RunTime:         "2026-01-19T11:00:00Z",
						DurationSeconds: 45.5,
						Tests: TestSummary{
							Total:   2,
							Passed:  2,
							Failed:  0,
							Skipped: 0,
						},
					},
				},
				Tests: []TestEntry{
					{
						Name:       "TestExisting1",
						Package:    "github.com/example/test-module",
						Type:       "gotest",
						Suite:      "unit",
						Status:     "passed",
						DurationMs: 100,
						Tags:       []string{},
						FilePath:   "test_test.go",
					},
					{
						Name:       "TestExisting2",
						Package:    "github.com/example/test-module",
						Type:       "gotest",
						Suite:      "unit",
						Status:     "passed",
						DurationMs: 150,
						Tags:       []string{},
						FilePath:   "test_test.go",
					},
				},
				Artifacts: []ArtifactInfo{},
				Version:   "1.0",
			},
			wantNewManifest:  false,
			wantTestCount:    3, // 2 existing + 1 manual
			wantManualSuite:  true,
			wantPassedCount:  3,
			wantFailedCount:  0,
			wantSkippedCount: 0,
		},
		{
			name:    "merge with all statuses",
			module:  "test-module",
			version: "v1.0.0",
			manualResults: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:        "2026-01-19T12:00:00Z",
					Tester:          "tester@example.com",
					Module:          "test-module",
					ReleaseVersion:  "v1.0.0",
					SchemaVersion:   "1.0",
					DurationSeconds: 180,
				},
				Results: []ManualTestResult{
					{
						ScenarioID:      "test-module/feature1/scenario-1",
						Status:          "passed",
						DurationSeconds: 30,
					},
					{
						ScenarioID:      "test-module/feature1/scenario-2",
						Status:          "failed",
						DurationSeconds: 45,
						Error:           "Expected behavior not observed",
						Notes:           "Failed during verification step",
					},
					{
						ScenarioID:      "test-module/feature2/scenario-3",
						Status:          "skipped",
						DurationSeconds: 0,
						Notes:           "Dependent feature not available",
					},
				},
			},
			existingManifest: nil,
			wantNewManifest:  true,
			wantTestCount:    3,
			wantManualSuite:  true,
			wantPassedCount:  1,
			wantFailedCount:  1,
			wantSkippedCount: 1,
		},
		{
			name:    "merge with evidence references",
			module:  "test-module",
			version: "v1.0.0",
			manualResults: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:        "2026-01-19T12:00:00Z",
					Tester:          "tester@example.com",
					Module:          "test-module",
					ReleaseVersion:  "v1.0.0",
					SchemaVersion:   "1.0",
					DurationSeconds: 120,
				},
				Results: []ManualTestResult{
					{
						ScenarioID:      "test-module/feature1/manual-scenario",
						Status:          "passed",
						DurationSeconds: 30,
						Evidence: []EvidenceReference{
							{
								URL:         "https://example.com/screenshot.png",
								Type:        "screenshot",
								Description: "Login page screenshot",
								SHA256:      "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
							},
						},
					},
				},
			},
			existingManifest: nil,
			wantNewManifest:  true,
			wantTestCount:    1,
			wantManualSuite:  true,
			wantPassedCount:  1,
			wantFailedCount:  0,
			wantSkippedCount: 0,
		},
		{
			name:    "merge replaces existing manual suite",
			module:  "test-module",
			version: "v1.0.0",
			manualResults: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:        "2026-01-19T14:00:00Z",
					Tester:          "tester@example.com",
					Module:          "test-module",
					ReleaseVersion:  "v1.0.0",
					SchemaVersion:   "1.0",
					DurationSeconds: 90,
				},
				Results: []ManualTestResult{
					{
						ScenarioID:      "test-module/feature1/scenario-1",
						Status:          "passed",
						DurationSeconds: 45,
					},
					{
						ScenarioID:      "test-module/feature1/scenario-2",
						Status:          "passed",
						DurationSeconds: 45,
					},
				},
			},
			existingManifest: &TestManifest{
				TestID:          "existing-test-id",
				TestAgent:       "devbox",
				Moniker:         "test-module",
				Type:            "go",
				TestTime:        "2026-01-19T11:00:00Z",
				DurationSeconds: 45.5,
				GitCommit:       "1234567890abcdef1234567890abcdef12345678",
				InputHash:       "abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab",
				Summary: TestSummary{
					Total:   3,
					Passed:  2,
					Failed:  1,
					Skipped: 0,
				},
				Suites: map[string]SuiteResult{
					"unit": {
						RunTime:         "2026-01-19T11:00:00Z",
						DurationSeconds: 45.5,
						Tests: TestSummary{
							Total:   2,
							Passed:  2,
							Failed:  0,
							Skipped: 0,
						},
					},
					"manual": {
						RunTime:         "2026-01-19T12:00:00Z",
						DurationSeconds: 60,
						Tests: TestSummary{
							Total:   1,
							Passed:  0,
							Failed:  1,
							Skipped: 0,
						},
					},
				},
				Tests: []TestEntry{
					{
						Name:       "TestExisting1",
						Package:    "github.com/example/test-module",
						Type:       "gotest",
						Suite:      "unit",
						Status:     "passed",
						DurationMs: 100,
						Tags:       []string{},
						FilePath:   "test_test.go",
					},
					{
						Name:       "TestExisting2",
						Package:    "github.com/example/test-module",
						Type:       "gotest",
						Suite:      "unit",
						Status:     "passed",
						DurationMs: 150,
						Tags:       []string{},
						FilePath:   "test_test.go",
					},
					{
						Name:       "old-scenario",
						Package:    "manual",
						Type:       "manual",
						Suite:      "manual",
						Status:     "failed",
						DurationMs: 30000,
						Tags:       []string{},
						FilePath:   "",
						Error:      "Old error",
					},
				},
				Artifacts: []ArtifactInfo{},
				Version:   "1.0",
			},
			wantNewManifest:  false,
			wantTestCount:    4, // 2 unit + 2 manual (replaces 1 old manual)
			wantManualSuite:  true,
			wantPassedCount:  4,
			wantFailedCount:  0,
			wantSkippedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test directory
			tmpDir := t.TempDir()
			setupTestRepositoryForMerge(t, tmpDir, tt.module)

			// Change to temp directory
			oldWd, _ := os.Getwd()
			require.NoError(t, os.Chdir(tmpDir))
			defer func() { _ = os.Chdir(oldWd) }()

			// Create manual results file
			resultsDir := filepath.Join(tmpDir, "test-results", tt.module, tt.version)
			require.NoError(t, os.MkdirAll(resultsDir, 0755))
			resultsFile := filepath.Join(resultsDir, "manual-results.json")
			data, err := json.MarshalIndent(tt.manualResults, "", "  ")
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(resultsFile, data, 0644))

			// Create existing manifest if specified
			if tt.existingManifest != nil {
				manifestDir := filepath.Join(tmpDir, "out", "test", tt.module)
				require.NoError(t, os.MkdirAll(manifestDir, 0755))
				manifestFile := filepath.Join(manifestDir, "test.manifest.json")
				data, err := json.MarshalIndent(tt.existingManifest, "", "  ")
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(manifestFile, data, 0644))
			}

			// Run merge-results command
			exitCode := runMergeResults(tt.module, tt.version)

			// Should succeed
			assert.Equal(t, 0, exitCode, "command should succeed")

			// Verify manifest exists
			manifestPath := filepath.Join(tmpDir, "out", "test", tt.module, "test.manifest.json")
			require.FileExists(t, manifestPath, "manifest file should exist")

			// Verify manifest content
			verifyMergedManifest(t, manifestPath, tt.module, tt.wantTestCount, tt.wantManualSuite,
				tt.wantPassedCount, tt.wantFailedCount, tt.wantSkippedCount)
		})
	}
}

func TestMergeResultsCommand_ErrorCases(t *testing.T) {
	tests := []struct {
		name         string
		module       string
		version      string
		setupResults bool
		invalidJSON  bool
		wantExitCode int
		wantError    string
	}{
		{
			name:         "manual results file doesn't exist",
			module:       "test-module",
			version:      "v1.0.0",
			setupResults: false,
			wantExitCode: 1,
			wantError:    "manual results file not found",
		},
		{
			name:         "manual results file invalid JSON",
			module:       "test-module",
			version:      "v1.0.0",
			setupResults: true,
			invalidJSON:  true,
			wantExitCode: 1,
			wantError:    "invalid JSON",
		},
		{
			name:         "module validation fails",
			module:       "nonexistent-module",
			version:      "v1.0.0",
			setupResults: true,
			wantExitCode: 1,
			wantError:    "unknown module",
		},
		{
			name:         "invalid version format",
			module:       "test-module",
			version:      "not-a-version",
			setupResults: true,
			wantExitCode: 1,
			wantError:    "invalid version format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test directory
			tmpDir := t.TempDir()
			setupTestRepositoryForMerge(t, tmpDir, "test-module")

			// Change to temp directory
			oldWd, _ := os.Getwd()
			require.NoError(t, os.Chdir(tmpDir))
			defer func() { _ = os.Chdir(oldWd) }()

			if tt.setupResults {
				resultsDir := filepath.Join(tmpDir, "test-results", tt.module, tt.version)
				require.NoError(t, os.MkdirAll(resultsDir, 0755))
				resultsFile := filepath.Join(resultsDir, "manual-results.json")

				if tt.invalidJSON {
					// Write invalid JSON
					require.NoError(t, os.WriteFile(resultsFile, []byte("{invalid json}"), 0644))
				} else {
					// Write valid manual results
					results := ManualTestResults{
						ImportMetadata: ImportMetadata{
							TestTime:        "2026-01-19T12:00:00Z",
							Tester:          "tester@example.com",
							Module:          tt.module,
							ReleaseVersion:  tt.version,
							SchemaVersion:   "1.0",
							DurationSeconds: 120,
						},
						Results: []ManualTestResult{
							{
								ScenarioID:      tt.module + "/feature1/scenario",
								Status:          "passed",
								DurationSeconds: 30,
							},
						},
					}
					data, err := json.MarshalIndent(results, "", "  ")
					require.NoError(t, err)
					require.NoError(t, os.WriteFile(resultsFile, data, 0644))
				}
			}

			// Run merge-results command
			exitCode := runMergeResults(tt.module, tt.version)

			// Should fail with expected exit code
			assert.Equal(t, tt.wantExitCode, exitCode, "command should fail with expected exit code")
		})
	}
}

func TestMergeResultsCommand_EdgeCases(t *testing.T) {
	tests := []struct {
		name             string
		module           string
		version          string
		manualResults    ManualTestResults
		wantTestCount    int
		wantPassedCount  int
		wantFailedCount  int
		wantSkippedCount int
	}{
		{
			name:    "empty results array still creates manifest",
			module:  "test-module",
			version: "v1.0.0",
			manualResults: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:        "2026-01-19T12:00:00Z",
					Tester:          "tester@example.com",
					Module:          "test-module",
					ReleaseVersion:  "v1.0.0",
					SchemaVersion:   "1.0",
					DurationSeconds: 0,
				},
				Results: []ManualTestResult{},
			},
			wantTestCount:    0,
			wantPassedCount:  0,
			wantFailedCount:  0,
			wantSkippedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test directory
			tmpDir := t.TempDir()
			setupTestRepositoryForMerge(t, tmpDir, tt.module)

			// Change to temp directory
			oldWd, _ := os.Getwd()
			require.NoError(t, os.Chdir(tmpDir))
			defer func() { _ = os.Chdir(oldWd) }()

			// Create manual results file
			resultsDir := filepath.Join(tmpDir, "test-results", tt.module, tt.version)
			require.NoError(t, os.MkdirAll(resultsDir, 0755))
			resultsFile := filepath.Join(resultsDir, "manual-results.json")
			data, err := json.MarshalIndent(tt.manualResults, "", "  ")
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(resultsFile, data, 0644))

			// Run merge-results command
			exitCode := runMergeResults(tt.module, tt.version)

			// For empty results, we may want it to succeed but with 0 tests
			// Or we may want it to fail - depends on requirements
			// For now, assume it succeeds
			assert.Equal(t, 0, exitCode, "command should succeed even with empty results")

			// Verify manifest exists
			manifestPath := filepath.Join(tmpDir, "out", "test", tt.module, "test.manifest.json")
			require.FileExists(t, manifestPath, "manifest file should exist")
		})
	}
}

func TestMergeResultsCommand_TransformLogic(t *testing.T) {
	tests := []struct {
		name          string
		manualResult  ManualTestResult
		wantTestEntry TestEntry
	}{
		{
			name: "passed scenario transforms correctly",
			manualResult: ManualTestResult{
				ScenarioID:      "test-module/feature1/manual-scenario",
				Status:          "passed",
				DurationSeconds: 30.5,
				Notes:           "All steps completed successfully",
			},
			wantTestEntry: TestEntry{
				Name:       "manual-scenario",
				Package:    "manual",
				Type:       "manual",
				Suite:      "manual",
				Status:     "passed",
				DurationMs: 30500, // 30.5 seconds -> milliseconds
				Tags:       []string{},
				FilePath:   "",
			},
		},
		{
			name: "failed scenario with error transforms correctly",
			manualResult: ManualTestResult{
				ScenarioID:      "test-module/feature2/failed-scenario",
				Status:          "failed",
				DurationSeconds: 45,
				Error:           "Expected behavior not observed",
				Notes:           "Failed during verification",
			},
			wantTestEntry: TestEntry{
				Name:       "failed-scenario",
				Package:    "manual",
				Type:       "manual",
				Suite:      "manual",
				Status:     "failed",
				DurationMs: 45000,
				Tags:       []string{},
				FilePath:   "",
				Error:      "Expected behavior not observed",
			},
		},
		{
			name: "skipped scenario transforms correctly",
			manualResult: ManualTestResult{
				ScenarioID:      "test-module/feature3/skipped-scenario",
				Status:          "skipped",
				DurationSeconds: 0,
				Notes:           "Dependent feature not available",
			},
			wantTestEntry: TestEntry{
				Name:       "skipped-scenario",
				Package:    "manual",
				Type:       "manual",
				Suite:      "manual",
				Status:     "skipped",
				DurationMs: 0,
				Tags:       []string{},
				FilePath:   "",
			},
		},
		{
			name: "scenario with zero duration",
			manualResult: ManualTestResult{
				ScenarioID: "test-module/feature1/quick-test",
				Status:     "passed",
			},
			wantTestEntry: TestEntry{
				Name:       "quick-test",
				Package:    "manual",
				Type:       "manual",
				Suite:      "manual",
				Status:     "passed",
				DurationMs: 0,
				Tags:       []string{},
				FilePath:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Transform the manual result
			testEntry := transformManualResultToTestEntry(tt.manualResult)

			// Verify transformation
			assert.Equal(t, tt.wantTestEntry.Name, testEntry.Name)
			assert.Equal(t, tt.wantTestEntry.Package, testEntry.Package)
			assert.Equal(t, tt.wantTestEntry.Type, testEntry.Type)
			assert.Equal(t, tt.wantTestEntry.Suite, testEntry.Suite)
			assert.Equal(t, tt.wantTestEntry.Status, testEntry.Status)
			assert.Equal(t, tt.wantTestEntry.DurationMs, testEntry.DurationMs)
			assert.Equal(t, tt.wantTestEntry.FilePath, testEntry.FilePath)

			if tt.wantTestEntry.Error != "" {
				assert.Equal(t, tt.wantTestEntry.Error, testEntry.Error)
			}
		})
	}
}

func TestMergeResultsCommand_ScenarioNameExtraction(t *testing.T) {
	tests := []struct {
		name       string
		scenarioID string
		wantName   string
	}{
		{
			name:       "simple scenario ID",
			scenarioID: "test-module/feature1/manual-scenario",
			wantName:   "manual-scenario",
		},
		{
			name:       "scenario with hyphens",
			scenarioID: "test-module/feature1/test-scenario-with-hyphens",
			wantName:   "test-scenario-with-hyphens",
		},
		{
			name:       "scenario in nested feature",
			scenarioID: "test-module/feature1/subfeature/scenario",
			wantName:   "scenario",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := extractScenarioName(tt.scenarioID)
			assert.Equal(t, tt.wantName, name)
		})
	}
}

// ==============================================================================
// Test Helper Functions
// ==============================================================================

// TestManifest represents the test manifest structure
type TestManifest struct {
	TestID               string                 `json:"test_id"`
	TestAgent            string                 `json:"test_agent"`
	Moniker              string                 `json:"moniker"`
	Type                 string                 `json:"type"`
	TestTime             string                 `json:"test_time"`
	DurationSeconds      float64                `json:"duration_seconds"`
	GitCommit            string                 `json:"git_commit"`
	InputHash            string                 `json:"input_hash"`
	Summary              TestSummary            `json:"summary"`
	Suites               map[string]SuiteResult `json:"suites"`
	Tests                []TestEntry            `json:"tests"`
	Artifacts            []ArtifactInfo         `json:"artifacts"`
	VerifiedUnchangedAt  string                 `json:"verified_unchanged_at,omitempty"`
	Version              string                 `json:"version"`
}

// TestSummary aggregates test counts
type TestSummary struct {
	Total     int `json:"total"`
	Passed    int `json:"passed"`
	Failed    int `json:"failed"`
	Skipped   int `json:"skipped"`
	Pending   int `json:"pending,omitempty"`
	Undefined int `json:"undefined,omitempty"`
}

// SuiteResult contains results for a test suite
type SuiteResult struct {
	RunTime         string      `json:"run_time"`
	DurationSeconds float64     `json:"duration_seconds"`
	Tests           TestSummary `json:"tests"`
}

// TestEntry represents an individual test
type TestEntry struct {
	Name       string   `json:"name"`
	Package    string   `json:"package"`
	Type       string   `json:"type"`
	Suite      string   `json:"suite"`
	Status     string   `json:"status"`
	DurationMs int      `json:"duration_ms,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Error      string   `json:"error,omitempty"`
	FilePath   string   `json:"file_path,omitempty"`
}

// ArtifactInfo represents test artifact information
type ArtifactInfo struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	Size   int    `json:"size,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}

// runMergeResults runs the merge-results command with specific flags
func runMergeResults(module, version string) int {
	// Save original args and restore after
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args for command
	os.Args = []string{"eac", "test", "merge-results", "--module", module, "--version", version}

	return MergeResults()
}

func setupTestRepositoryForMerge(t *testing.T, tmpDir, module string) {
	t.Helper()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))

	// Create .r2r/eac directory
	r2rDir := filepath.Join(tmpDir, ".r2r", "eac")
	require.NoError(t, os.MkdirAll(r2rDir, 0755))

	// Create repository.yml with correct format
	repoYAML := `repository:
  type: mono

modules:
  - moniker: ` + module + `
    name: Test Module
    description: Test module for testing
    components:
      go:
        type: go
        root: /` + module + `
`
	repoFile := filepath.Join(r2rDir, "repository.yml")
	require.NoError(t, os.WriteFile(repoFile, []byte(repoYAML), 0644))
}

func verifyMergedManifest(t *testing.T, manifestPath, module string, wantTestCount int, wantManualSuite bool,
	wantPassedCount, wantFailedCount, wantSkippedCount int) {
	t.Helper()

	content, err := os.ReadFile(manifestPath)
	require.NoError(t, err, "should read manifest file")

	var manifest TestManifest
	err = json.Unmarshal(content, &manifest)
	require.NoError(t, err, "should be valid JSON")

	// Verify basic metadata
	assert.Equal(t, module, manifest.Moniker)
	assert.Equal(t, "1.0", manifest.Version)
	assert.NotEmpty(t, manifest.TestID)
	assert.NotEmpty(t, manifest.TestTime)

	// Verify test counts
	assert.Len(t, manifest.Tests, wantTestCount, "should have correct test count")

	// Verify summary counts
	assert.Equal(t, wantTestCount, manifest.Summary.Total)
	assert.Equal(t, wantPassedCount, manifest.Summary.Passed)
	assert.Equal(t, wantFailedCount, manifest.Summary.Failed)
	assert.Equal(t, wantSkippedCount, manifest.Summary.Skipped)

	// Verify manual suite exists if expected
	if wantManualSuite {
		manualSuite, exists := manifest.Suites["manual"]
		assert.True(t, exists, "manual suite should exist")
		assert.NotEmpty(t, manualSuite.RunTime)

		// Verify manual tests
		manualTestCount := 0
		for _, test := range manifest.Tests {
			if test.Suite == "manual" {
				manualTestCount++
				assert.Equal(t, "manual", test.Package)
				assert.Equal(t, "manual", test.Type)
				assert.Empty(t, test.FilePath)
			}
		}
		assert.Greater(t, manualTestCount, 0, "should have at least one manual test")
	}
}

// transformManualResultToTestEntry transforms a manual test result into a test entry
func transformManualResultToTestEntry(result ManualTestResult) TestEntry {
	return TestEntry{
		Name:       extractScenarioName(result.ScenarioID),
		Package:    "manual",
		Type:       "manual",
		Suite:      "manual",
		Status:     result.Status,
		DurationMs: int(result.DurationSeconds * 1000),
		Tags:       []string{},
		FilePath:   "",
		Error:      result.Error,
	}
}
