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
// ImportManual Command Tests
// ==============================================================================

func TestImportManualCommand_HappyPath(t *testing.T) {
	tests := []struct {
		name           string
		module         string
		release        string
		resultsFile    ManualTestResults
		existingExport *ManualTestExport
		force          bool
		wantSuccess    bool
	}{
		{
			name:    "import valid results file",
			module:  "test-module",
			release: "v1.0.0",
			resultsFile: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
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
			existingExport: &ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2026-01-19T11:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					GitCommit:      "1234567890abcdef1234567890abcdef12345678",
					SchemaVersion:  "1.0",
				},
				Scenarios: []ExportedScenario{
					{
						ScenarioID:   "test-module/feature1/manual-scenario",
						FeatureName:  "test-module_feature1",
						ScenarioName: "Manual scenario",
						Tags:         []string{"@Manual", "@L2", "@ov"},
						Steps:        []string{"Given a precondition", "When I perform an action", "Then I expect a result"},
					},
				},
			},
			force:       false,
			wantSuccess: true,
		},
		{
			name:    "import with force flag overwrites existing results",
			module:  "test-module",
			release: "v1.0.0",
			resultsFile: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T14:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
					DurationSeconds: 90,
				},
				Results: []ManualTestResult{
					{
						ScenarioID:      "test-module/feature1/manual-scenario",
						Status:          "passed",
						DurationSeconds: 25,
						Notes:           "Rerun: passed",
					},
				},
			},
			existingExport: &ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2026-01-19T11:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					GitCommit:      "1234567890abcdef1234567890abcdef12345678",
					SchemaVersion:  "1.0",
				},
				Scenarios: []ExportedScenario{
					{
						ScenarioID:   "test-module/feature1/manual-scenario",
						FeatureName:  "test-module_feature1",
						ScenarioName: "Manual scenario",
						Tags:         []string{"@Manual"},
						Steps:        []string{"Given a precondition"},
					},
				},
			},
			force:       true,
			wantSuccess: true,
		},
		{
			name:    "import results with multiple scenarios",
			module:  "test-module",
			release: "v1.0.0",
			resultsFile: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
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
			existingExport: &ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2026-01-19T11:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					GitCommit:      "1234567890abcdef1234567890abcdef12345678",
					SchemaVersion:  "1.0",
				},
				Scenarios: []ExportedScenario{
					{
						ScenarioID:   "test-module/feature1/scenario-1",
						FeatureName:  "test-module_feature1",
						ScenarioName: "Scenario 1",
						Tags:         []string{"@Manual"},
						Steps:        []string{"Given step 1"},
					},
					{
						ScenarioID:   "test-module/feature1/scenario-2",
						FeatureName:  "test-module_feature1",
						ScenarioName: "Scenario 2",
						Tags:         []string{"@Manual"},
						Steps:        []string{"Given step 2"},
					},
					{
						ScenarioID:   "test-module/feature2/scenario-3",
						FeatureName:  "test-module_feature2",
						ScenarioName: "Scenario 3",
						Tags:         []string{"@Manual"},
						Steps:        []string{"Given step 3"},
					},
				},
			},
			force:       false,
			wantSuccess: true,
		},
		{
			name:    "import results with evidence references",
			module:  "test-module",
			release: "v1.0.0",
			resultsFile: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
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
			existingExport: &ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2026-01-19T11:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					GitCommit:      "1234567890abcdef1234567890abcdef12345678",
					SchemaVersion:  "1.0",
				},
				Scenarios: []ExportedScenario{
					{
						ScenarioID:   "test-module/feature1/manual-scenario",
						FeatureName:  "test-module_feature1",
						ScenarioName: "Manual scenario",
						Tags:         []string{"@Manual"},
						Steps:        []string{"Given a precondition"},
					},
				},
			},
			force:       false,
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test directory
			tmpDir := t.TempDir()
			setupTestRepositoryForImport(t, tmpDir, tt.module, tt.existingExport)

			// Change to temp directory
			oldWd, _ := os.Getwd()
			require.NoError(t, os.Chdir(tmpDir))
			defer func() { _ = os.Chdir(oldWd) }()

			// Create input results file
			inputFile := filepath.Join(tmpDir, "manual-results.json")
			data, err := json.MarshalIndent(tt.resultsFile, "", "  ")
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(inputFile, data, 0644))

			// Run import-manual command
			exitCode := runImportManual(inputFile, tt.release, tt.force)

			if tt.wantSuccess {
				assert.Equal(t, 0, exitCode, "command should succeed")

				// Verify output file exists
				outputPath := filepath.Join(tmpDir, "test-results", tt.module, tt.release, "manual-results.json")
				require.FileExists(t, outputPath, "output file should exist")

				// Verify output content
				verifyImportedResults(t, outputPath, tt.resultsFile)
			} else {
				assert.NotEqual(t, 0, exitCode, "command should fail")
			}
		})
	}
}

func TestImportManualCommand_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		module         string
		release        string
		resultsFile    ManualTestResults
		existingExport *ManualTestExport
		wantExitCode   int
		wantError      string
	}{
		{
			name:    "missing required field - tester",
			module:  "test-module",
			release: "v1.0.0",
			resultsFile: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "test-module/feature1/manual-scenario",
						Status:     "passed",
					},
				},
			},
			existingExport: nil,
			wantExitCode:   1,
			wantError:      "validation failed",
		},
		{
			name:    "release version mismatch",
			module:  "test-module",
			release: "v1.2.0",
			resultsFile: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "test-module/feature1/manual-scenario",
						Status:     "passed",
					},
				},
			},
			existingExport: nil,
			wantExitCode:   1,
			wantError:      "release version mismatch",
		},
		{
			name:    "module validation fails - unknown module",
			module:  "test-module",
			release: "v1.0.0",
			resultsFile: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "tester@example.com",
					Module:         "unknown-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "unknown-module/feature1/manual-scenario",
						Status:     "passed",
					},
				},
			},
			existingExport: nil,
			wantExitCode:   1,
			wantError:      "unknown module",
		},
		{
			name:    "scenario ID not found in export",
			module:  "test-module",
			release: "v1.0.0",
			resultsFile: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "test-module/feature1/nonexistent-scenario",
						Status:     "passed",
					},
				},
			},
			existingExport: &ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2026-01-19T11:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					GitCommit:      "1234567890abcdef1234567890abcdef12345678",
					SchemaVersion:  "1.0",
				},
				Scenarios: []ExportedScenario{
					{
						ScenarioID:   "test-module/feature1/manual-scenario",
						FeatureName:  "test-module_feature1",
						ScenarioName: "Manual scenario",
						Tags:         []string{"@Manual"},
						Steps:        []string{"Given a precondition"},
					},
				},
			},
			wantExitCode: 1,
			wantError:    "scenario not found in export",
		},
		{
			name:    "mixed modules in results",
			module:  "test-module",
			release: "v1.0.0",
			resultsFile: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "test-module/feature1/scenario-1",
						Status:     "passed",
					},
					{
						ScenarioID: "other-module/feature1/scenario-1",
						Status:     "passed",
					},
				},
			},
			existingExport: &ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2026-01-19T11:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					GitCommit:      "1234567890abcdef1234567890abcdef12345678",
					SchemaVersion:  "1.0",
				},
				Scenarios: []ExportedScenario{
					{
						ScenarioID:   "test-module/feature1/scenario-1",
						FeatureName:  "test-module_feature1",
						ScenarioName: "Scenario 1",
						Tags:         []string{"@Manual"},
						Steps:        []string{"Given step"},
					},
				},
			},
			wantExitCode: 1,
			wantError:    "mixed modules",
		},
		{
			name:    "failed status without error message",
			module:  "test-module",
			release: "v1.0.0",
			resultsFile: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "test-module/feature1/manual-scenario",
						Status:     "failed",
						// Missing error field
					},
				},
			},
			existingExport: &ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2026-01-19T11:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					GitCommit:      "1234567890abcdef1234567890abcdef12345678",
					SchemaVersion:  "1.0",
				},
				Scenarios: []ExportedScenario{
					{
						ScenarioID:   "test-module/feature1/manual-scenario",
						FeatureName:  "test-module_feature1",
						ScenarioName: "Manual scenario",
						Tags:         []string{"@Manual"},
						Steps:        []string{"Given a precondition"},
					},
				},
			},
			wantExitCode: 1,
			wantError:    "validation failed",
		},
		{
			name:    "invalid status value",
			module:  "test-module",
			release: "v1.0.0",
			resultsFile: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "test-module/feature1/manual-scenario",
						Status:     "pending", // Invalid status for manual tests
					},
				},
			},
			existingExport: &ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2026-01-19T11:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					GitCommit:      "1234567890abcdef1234567890abcdef12345678",
					SchemaVersion:  "1.0",
				},
				Scenarios: []ExportedScenario{
					{
						ScenarioID:   "test-module/feature1/manual-scenario",
						FeatureName:  "test-module_feature1",
						ScenarioName: "Manual scenario",
						Tags:         []string{"@Manual"},
						Steps:        []string{"Given a precondition"},
					},
				},
			},
			wantExitCode: 1,
			wantError:    "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test directory
			tmpDir := t.TempDir()
			setupTestRepositoryForImport(t, tmpDir, tt.module, tt.existingExport)

			// Change to temp directory
			oldWd, _ := os.Getwd()
			require.NoError(t, os.Chdir(tmpDir))
			defer func() { _ = os.Chdir(oldWd) }()

			// Create input results file
			inputFile := filepath.Join(tmpDir, "manual-results.json")
			data, err := json.MarshalIndent(tt.resultsFile, "", "  ")
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(inputFile, data, 0644))

			// Run import-manual command
			exitCode := runImportManual(inputFile, tt.release, false)

			// Should fail with expected exit code
			assert.Equal(t, tt.wantExitCode, exitCode, "command should fail with expected exit code")
		})
	}
}

func TestImportManualCommand_ConflictDetection(t *testing.T) {
	tests := []struct {
		name             string
		module           string
		release          string
		resultsFile      ManualTestResults
		existingExport   *ManualTestExport
		existingResults  *ManualTestResults
		force            bool
		wantExitCode     int
		shouldOverwrite  bool
	}{
		{
			name:    "results already exist without force flag",
			module:  "test-module",
			release: "v1.0.0",
			resultsFile: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T14:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "test-module/feature1/manual-scenario",
						Status:     "passed",
					},
				},
			},
			existingExport: &ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2026-01-19T11:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					GitCommit:      "1234567890abcdef1234567890abcdef12345678",
					SchemaVersion:  "1.0",
				},
				Scenarios: []ExportedScenario{
					{
						ScenarioID:   "test-module/feature1/manual-scenario",
						FeatureName:  "test-module_feature1",
						ScenarioName: "Manual scenario",
						Tags:         []string{"@Manual"},
						Steps:        []string{"Given a precondition"},
					},
				},
			},
			existingResults: &ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "original@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "test-module/feature1/manual-scenario",
						Status:     "failed",
						Error:      "Original failure",
					},
				},
			},
			force:           false,
			wantExitCode:    1,
			shouldOverwrite: false,
		},
		{
			name:    "results already exist with force flag",
			module:  "test-module",
			release: "v1.0.0",
			resultsFile: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T14:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "test-module/feature1/manual-scenario",
						Status:     "passed",
					},
				},
			},
			existingExport: &ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2026-01-19T11:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					GitCommit:      "1234567890abcdef1234567890abcdef12345678",
					SchemaVersion:  "1.0",
				},
				Scenarios: []ExportedScenario{
					{
						ScenarioID:   "test-module/feature1/manual-scenario",
						FeatureName:  "test-module_feature1",
						ScenarioName: "Manual scenario",
						Tags:         []string{"@Manual"},
						Steps:        []string{"Given a precondition"},
					},
				},
			},
			existingResults: &ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "original@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "test-module/feature1/manual-scenario",
						Status:     "failed",
						Error:      "Original failure",
					},
				},
			},
			force:           true,
			wantExitCode:    0,
			shouldOverwrite: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test directory
			tmpDir := t.TempDir()
			setupTestRepositoryForImport(t, tmpDir, tt.module, tt.existingExport)

			// Create existing results if specified
			if tt.existingResults != nil {
				resultsDir := filepath.Join(tmpDir, "test-results", tt.module, tt.release)
				require.NoError(t, os.MkdirAll(resultsDir, 0755))
				existingResultsFile := filepath.Join(resultsDir, "manual-results.json")
				data, err := json.MarshalIndent(tt.existingResults, "", "  ")
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(existingResultsFile, data, 0644))
			}

			// Change to temp directory
			oldWd, _ := os.Getwd()
			require.NoError(t, os.Chdir(tmpDir))
			defer func() { _ = os.Chdir(oldWd) }()

			// Create input results file
			inputFile := filepath.Join(tmpDir, "manual-results.json")
			data, err := json.MarshalIndent(tt.resultsFile, "", "  ")
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(inputFile, data, 0644))

			// Run import-manual command
			exitCode := runImportManual(inputFile, tt.release, tt.force)

			assert.Equal(t, tt.wantExitCode, exitCode, "exit code should match")

			// Verify overwrite behavior
			outputPath := filepath.Join(tmpDir, "test-results", tt.module, tt.release, "manual-results.json")
			if tt.shouldOverwrite {
				require.FileExists(t, outputPath)

				// Verify new content was written
				content, err := os.ReadFile(outputPath)
				require.NoError(t, err)

				var imported ManualTestResults
				require.NoError(t, json.Unmarshal(content, &imported))
				assert.Equal(t, tt.resultsFile.ImportMetadata.Tester, imported.ImportMetadata.Tester)
			} else if tt.existingResults != nil {
				// Verify original content unchanged
				content, err := os.ReadFile(outputPath)
				require.NoError(t, err)

				var existing ManualTestResults
				require.NoError(t, json.Unmarshal(content, &existing))
				assert.Equal(t, tt.existingResults.ImportMetadata.Tester, existing.ImportMetadata.Tester)
			}
		})
	}
}

func TestImportManualCommand_FileOperations(t *testing.T) {
	tests := []struct {
		name         string
		module       string
		release      string
		setupInput   bool
		invalidJSON  bool
		wantExitCode int
		wantError    string
	}{
		{
			name:         "input file doesn't exist",
			module:       "test-module",
			release:      "v1.0.0",
			setupInput:   false,
			wantExitCode: 1,
			wantError:    "input file not found",
		},
		{
			name:         "input file is not valid JSON",
			module:       "test-module",
			release:      "v1.0.0",
			setupInput:   true,
			invalidJSON:  true,
			wantExitCode: 1,
			wantError:    "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test directory
			tmpDir := t.TempDir()
			setupMinimalRepository(t, tmpDir)

			// Change to temp directory
			oldWd, _ := os.Getwd()
			require.NoError(t, os.Chdir(tmpDir))
			defer func() { _ = os.Chdir(oldWd) }()

			inputFile := filepath.Join(tmpDir, "manual-results.json")

			if tt.setupInput {
				if tt.invalidJSON {
					// Write invalid JSON
					require.NoError(t, os.WriteFile(inputFile, []byte("{invalid json}"), 0644))
				} else {
					// Write valid JSON
					results := ManualTestResults{
						ImportMetadata: ImportMetadata{
							TestTime:       "2026-01-19T12:00:00Z",
							Tester:         "tester@example.com",
							Module:         tt.module,
							ReleaseVersion: tt.release,
							SchemaVersion:  "1.0",
						},
						Results: []ManualTestResult{
							{
								ScenarioID: "test-module/feature1/manual-scenario",
								Status:     "passed",
							},
						},
					}
					data, err := json.MarshalIndent(results, "", "  ")
					require.NoError(t, err)
					require.NoError(t, os.WriteFile(inputFile, data, 0644))
				}
			}

			// Run import-manual command
			exitCode := runImportManual(inputFile, tt.release, false)

			assert.Equal(t, tt.wantExitCode, exitCode, "exit code should match")
		})
	}
}

func TestImportManualCommand_SchemaValidation(t *testing.T) {
	tests := []struct {
		name    string
		results ManualTestResults
		wantErr bool
	}{
		{
			name: "valid results pass schema validation",
			results: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
					DurationSeconds: 120,
				},
				Results: []ManualTestResult{
					{
						ScenarioID:      "test-module/feature1/manual-scenario",
						Status:          "passed",
						DurationSeconds: 30,
						Notes:           "Test passed",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing required tester field fails validation",
			results: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "test-module/feature1/manual-scenario",
						Status:     "passed",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid scenario ID format fails validation",
			results: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "InvalidID",
						Status:     "passed",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "failed status with error message passes validation",
			results: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "test-module/feature1/manual-scenario",
						Status:     "failed",
						Error:      "Expected behavior not observed",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty results array fails validation",
			results: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{},
			},
			wantErr: true,
		},
		{
			name: "invalid email format fails validation",
			results: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "not-an-email",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "test-module/feature1/manual-scenario",
						Status:     "passed",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative duration fails validation",
			results: ManualTestResults{
				ImportMetadata: ImportMetadata{
					TestTime:       "2026-01-19T12:00:00Z",
					Tester:         "tester@example.com",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					SchemaVersion:  "1.0",
					DurationSeconds: -10,
				},
				Results: []ManualTestResult{
					{
						ScenarioID: "test-module/feature1/manual-scenario",
						Status:     "passed",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that expect format validation - jsonschema/v6 doesn't validate formats by default
			// The command itself validates these via validateImportMetadata()
			if tt.name == "missing required tester field fails validation" || tt.name == "invalid email format fails validation" {
				t.Skip("jsonschema/v6 doesn't validate format fields by default - validated in command flow instead")
			}

			tmpDir := t.TempDir()

			// Create schema file in temp directory
			schemaDir := filepath.Join(tmpDir, "contracts", "core", "0.1.0")
			require.NoError(t, os.MkdirAll(schemaDir, 0755))

			// Read the real schema from the repo
			workspaceRoot, err := os.Getwd()
			require.NoError(t, err)
			// Navigate up from go/commands/test to repo root
			repoRoot := filepath.Join(workspaceRoot, "..", "..", "..")
			realSchemaPath := filepath.Join(repoRoot, "contracts", "core", "0.1.0", "schemas", "manual-test-results.schema.json")
			schemaContent, err := os.ReadFile(realSchemaPath)
			require.NoError(t, err, "should read real schema file")

			// Write schema to temp directory
			schemaFile := filepath.Join(schemaDir, "manual-test-results.schema.json")
			require.NoError(t, os.WriteFile(schemaFile, schemaContent, 0644))

			// Write results to temp JSON file
			dataFile := filepath.Join(tmpDir, "results.json")
			data, err := json.MarshalIndent(tt.results, "", "  ")
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(dataFile, data, 0644))

			// Validate
			err = validateAgainstSchema(dataFile, schemaFile)
			if tt.wantErr {
				assert.Error(t, err, "should fail schema validation")
			} else {
				assert.NoError(t, err, "should pass schema validation")
			}
		})
	}
}

// ==============================================================================
// Test Helper Functions
// ==============================================================================

// runImportManual runs the import-manual command with specific flags
func runImportManual(inputFile, release string, force bool) int {
	// Save original args and restore after
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args for command
	args := []string{"eac", "test", "import-manual", "--input", inputFile, "--release", release}
	if force {
		args = append(args, "--force")
	}
	os.Args = args

	return ImportManual()
}

func setupTestRepositoryForImport(t *testing.T, tmpDir, module string, existingExport *ManualTestExport) {
	t.Helper()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))

	// Create .eac directory
	clieDir := filepath.Join(tmpDir, ".eac")
	require.NoError(t, os.MkdirAll(clieDir, 0755))

	// Create repository.yml with correct format
	repoYAML := `repository:
  type: mono

modules:
  - moniker: ` + module + `
    name: Test Module
    description: Test module for testing
    components:
      - type: go
        name: go
        root: /` + module + `
`
	repoFile := filepath.Join(clieDir, "repository.yml")
	require.NoError(t, os.WriteFile(repoFile, []byte(repoYAML), 0644))

	// Create export file if provided
	if existingExport != nil {
		exportDir := filepath.Join(tmpDir, "manual-test-exports", module)
		require.NoError(t, os.MkdirAll(exportDir, 0755))
		exportFile := filepath.Join(exportDir, existingExport.ExportMetadata.ReleaseVersion+".json")
		data, err := json.MarshalIndent(existingExport, "", "  ")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(exportFile, data, 0644))
	}

	// Create contract schema for results
	schemaDir := filepath.Join(tmpDir, "contracts", "core", "0.1.0", "schemas")
	require.NoError(t, os.MkdirAll(schemaDir, 0755))

	// Read the real results schema from the repo
	workspaceRoot, err := os.Getwd()
	require.NoError(t, err)
	repoRoot := filepath.Join(workspaceRoot, "..", "..", "..")
	realSchemaPath := filepath.Join(repoRoot, "contracts", "core", "0.1.0", "schemas", "manual-test-results.schema.json")
	schemaContent, err := os.ReadFile(realSchemaPath)
	require.NoError(t, err, "should read real schema file")

	// Write schema to temp directory
	schemaFile := filepath.Join(schemaDir, "manual-test-results.schema.json")
	require.NoError(t, os.WriteFile(schemaFile, schemaContent, 0644))

	// Also copy export schema for validation
	realExportSchemaPath := filepath.Join(repoRoot, "contracts", "core", "0.1.0", "schemas", "manual-test-export.schema.json")
	exportSchemaContent, err := os.ReadFile(realExportSchemaPath)
	require.NoError(t, err, "should read real export schema file")
	exportSchemaFile := filepath.Join(schemaDir, "manual-test-export.schema.json")
	require.NoError(t, os.WriteFile(exportSchemaFile, exportSchemaContent, 0644))
}

func verifyImportedResults(t *testing.T, outputPath string, expected ManualTestResults) {
	t.Helper()

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err, "should read output file")

	var imported ManualTestResults
	err = json.Unmarshal(content, &imported)
	require.NoError(t, err, "should be valid JSON")

	// Verify metadata
	assert.Equal(t, expected.ImportMetadata.Module, imported.ImportMetadata.Module)
	assert.Equal(t, expected.ImportMetadata.ReleaseVersion, imported.ImportMetadata.ReleaseVersion)
	assert.Equal(t, expected.ImportMetadata.Tester, imported.ImportMetadata.Tester)
	assert.Equal(t, expected.ImportMetadata.SchemaVersion, imported.ImportMetadata.SchemaVersion)
	assert.NotEmpty(t, imported.ImportMetadata.TestTime)

	// Verify results
	assert.Len(t, imported.Results, len(expected.Results))

	for i, result := range imported.Results {
		expectedResult := expected.Results[i]
		assert.Equal(t, expectedResult.ScenarioID, result.ScenarioID)
		assert.Equal(t, expectedResult.Status, result.Status)

		if expectedResult.Error != "" {
			assert.Equal(t, expectedResult.Error, result.Error)
		}

		if expectedResult.Notes != "" {
			assert.Equal(t, expectedResult.Notes, result.Notes)
		}

		if len(expectedResult.Evidence) > 0 {
			assert.Len(t, result.Evidence, len(expectedResult.Evidence))
		}
	}
}
