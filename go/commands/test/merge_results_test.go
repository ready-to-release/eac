//go:build L1
// +build L1

package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==============================================================================
// MergeResults Command Tests
// ==============================================================================

func TestMergeResultsCommand_HappyPath(t *testing.T) {
	tests := []struct {
		name          string
		module        string
		version       string
		manualResults ManualTestResults
		wantTestCount int
	}{
		{
			name:    "merge creates UoW manifest with manual tests",
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
			wantTestCount: 1,
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
					},
					{
						ScenarioID:      "test-module/feature2/scenario-3",
						Status:          "skipped",
						DurationSeconds: 0,
					},
				},
			},
			wantTestCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			setupTestRepositoryForMerge(t, tmpDir, tt.module)

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

			exitCode := runMergeResults(tt.module, tt.version)
			assert.Equal(t, 0, exitCode, "command should succeed")

			// Verify UoW manifest was created
			uowManifestPath := filepath.Join(tmpDir, "out", "test", tt.module, "manual-manual-manual", "uow.manifest.json")
			require.FileExists(t, uowManifestPath, "UoW manifest should exist")

			// Verify UoW manifest content
			manifestData, err := os.ReadFile(uowManifestPath)
			require.NoError(t, err)

			var manifest coreoutput.UoWManifest
			require.NoError(t, json.Unmarshal(manifestData, &manifest))

			assert.Equal(t, "test", string(manifest.Action))
			assert.Equal(t, tt.module, manifest.Module)
			assert.Equal(t, "manual", manifest.Component)
			assert.Equal(t, "manual", manifest.Tool)
			assert.Equal(t, 0, manifest.ExitCode)
			assert.Len(t, manifest.Artifacts, 1)
			assert.Equal(t, "manual-report", manifest.Artifacts[0].Type)
			assert.Equal(t, "manual-tests.json", manifest.Artifacts[0].Path)
			assert.NotEmpty(t, manifest.Artifacts[0].SHA256)
			assert.NotEmpty(t, manifest.OutputHash)

			// Verify manual-tests.json was created
			manualTestsPath := filepath.Join(tmpDir, "out", "test", tt.module, "manual-manual-manual", "manual-tests.json")
			require.FileExists(t, manualTestsPath)

			// Verify manual-tests.json content
			mtData, err := os.ReadFile(manualTestsPath)
			require.NoError(t, err)

			var mtFile manualTestsJSONFile
			require.NoError(t, json.Unmarshal(mtData, &mtFile))
			assert.Len(t, mtFile.Tests, tt.wantTestCount)
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
	}{
		{
			name:         "manual results file doesn't exist",
			module:       "test-module",
			version:      "v1.0.0",
			setupResults: false,
			wantExitCode: 1,
		},
		{
			name:         "manual results file invalid JSON",
			module:       "test-module",
			version:      "v1.0.0",
			setupResults: true,
			invalidJSON:  true,
			wantExitCode: 1,
		},
		{
			name:         "module validation fails",
			module:       "nonexistent-module",
			version:      "v1.0.0",
			setupResults: true,
			wantExitCode: 1,
		},
		{
			name:         "invalid version format",
			module:       "test-module",
			version:      "not-a-version",
			setupResults: true,
			wantExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			setupTestRepositoryForMerge(t, tmpDir, "test-module")

			oldWd, _ := os.Getwd()
			require.NoError(t, os.Chdir(tmpDir))
			defer func() { _ = os.Chdir(oldWd) }()

			if tt.setupResults {
				resultsDir := filepath.Join(tmpDir, "test-results", tt.module, tt.version)
				require.NoError(t, os.MkdirAll(resultsDir, 0755))
				resultsFile := filepath.Join(resultsDir, "manual-results.json")

				if tt.invalidJSON {
					require.NoError(t, os.WriteFile(resultsFile, []byte("{invalid json}"), 0644))
				} else {
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

			exitCode := runMergeResults(tt.module, tt.version)
			assert.Equal(t, tt.wantExitCode, exitCode, "command should fail with expected exit code")
		})
	}
}

func TestMergeResultsCommand_EdgeCases(t *testing.T) {
	t.Run("empty results array still creates manifest", func(t *testing.T) {
		tmpDir := t.TempDir()
		setupTestRepositoryForMerge(t, tmpDir, "test-module")

		oldWd, _ := os.Getwd()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir(oldWd) }()

		manualResults := ManualTestResults{
			ImportMetadata: ImportMetadata{
				TestTime:        "2026-01-19T12:00:00Z",
				Tester:          "tester@example.com",
				Module:          "test-module",
				ReleaseVersion:  "v1.0.0",
				SchemaVersion:   "1.0",
				DurationSeconds: 0,
			},
			Results: []ManualTestResult{},
		}

		resultsDir := filepath.Join(tmpDir, "test-results", "test-module", "v1.0.0")
		require.NoError(t, os.MkdirAll(resultsDir, 0755))
		resultsFile := filepath.Join(resultsDir, "manual-results.json")
		data, err := json.MarshalIndent(manualResults, "", "  ")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(resultsFile, data, 0644))

		exitCode := runMergeResults("test-module", "v1.0.0")
		assert.Equal(t, 0, exitCode, "command should succeed even with empty results")

		uowManifestPath := filepath.Join(tmpDir, "out", "test", "test-module", "manual-manual-manual", "uow.manifest.json")
		require.FileExists(t, uowManifestPath, "UoW manifest should exist")
	})
}

func TestMergeResultsCommand_TransformLogic(t *testing.T) {
	tests := []struct {
		name         string
		manualResult ManualTestResult
		wantName     string
		wantStatus   string
		wantDuration int64
		wantError    string
	}{
		{
			name: "passed scenario transforms correctly",
			manualResult: ManualTestResult{
				ScenarioID:      "test-module/feature1/manual-scenario",
				Status:          "passed",
				DurationSeconds: 30.5,
			},
			wantName:     "manual-scenario",
			wantStatus:   "passed",
			wantDuration: 30500,
		},
		{
			name: "failed scenario with error transforms correctly",
			manualResult: ManualTestResult{
				ScenarioID:      "test-module/feature2/failed-scenario",
				Status:          "failed",
				DurationSeconds: 45,
				Error:           "Expected behavior not observed",
			},
			wantName:     "failed-scenario",
			wantStatus:   "failed",
			wantDuration: 45000,
			wantError:    "Expected behavior not observed",
		},
		{
			name: "skipped scenario transforms correctly",
			manualResult: ManualTestResult{
				ScenarioID:      "test-module/feature3/skipped-scenario",
				Status:          "skipped",
				DurationSeconds: 0,
			},
			wantName:     "skipped-scenario",
			wantStatus:   "skipped",
			wantDuration: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformToManualTestsJSON([]ManualTestResult{tt.manualResult})
			require.Len(t, result.Tests, 1)
			entry := result.Tests[0]
			assert.Equal(t, tt.wantName, entry.Name)
			assert.Equal(t, tt.wantStatus, entry.Status)
			assert.Equal(t, tt.wantDuration, entry.DurationMs)
			assert.Equal(t, tt.wantError, entry.Error)
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

// runMergeResults runs the merge-results command with specific flags
func runMergeResults(module, version string) int {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"eac", "test", "merge-results", "--module", module, "--version", version}
	return MergeResults()
}

func setupTestRepositoryForMerge(t *testing.T, tmpDir, module string) {
	t.Helper()

	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))

	clieDir := filepath.Join(tmpDir, ".eac")
	require.NoError(t, os.MkdirAll(clieDir, 0755))

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
	repoFile := filepath.Join(clieDir, "repository.yml")
	require.NoError(t, os.WriteFile(repoFile, []byte(repoYAML), 0644))
}
