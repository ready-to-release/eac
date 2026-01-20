//go:build L1
// +build L1

package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/specs/gherkin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==============================================================================
// ExportManual Command Tests
// ==============================================================================

func TestExportManualCommand_HappyPath(t *testing.T) {
	tests := []struct {
		name            string
		module          string
		release         string
		format          string
		featureFiles    map[string]string
		expectedScenarios int
		wantOutputFile  string
	}{
		{
			name:    "export single manual scenario as JSON",
			module:  "test-module",
			release: "v1.0.0",
			format:  "json",
			featureFiles: map[string]string{
				"specs/test-module/feature1/spec.feature": `@deps:go @L2 @ov
Feature: test-module_feature1

  As a tester
  I want to test manually
  So that I can verify behavior

  Rule: Manual testing

    @Manual @L2 @ov
    Scenario: Manual test scenario
      Given a precondition
      When I perform an action
      Then I expect a result
`,
			},
			expectedScenarios: 1,
			wantOutputFile:   "manual-test-scenarios.json",
		},
		{
			name:    "export multiple manual scenarios as JSON",
			module:  "test-module",
			release: "v1.0.0",
			format:  "json",
			featureFiles: map[string]string{
				"specs/test-module/feature1/spec.feature": `@deps:go @L2 @ov
Feature: test-module_feature1

  As a tester
  I want to test manually
  So that I can verify behavior

  Rule: Manual testing

    @Manual @L2 @ov
    Scenario: First manual scenario
      Given a precondition
      When I perform an action
      Then I expect a result

    @Manual @L2 @ov
    Scenario: Second manual scenario
      Given another precondition
      When I perform another action
      Then I expect another result
`,
			},
			expectedScenarios: 2,
			wantOutputFile:   "manual-test-scenarios.json",
		},
		{
			name:    "export manual scenarios mixed with automated",
			module:  "test-module",
			release: "v1.0.0",
			format:  "json",
			featureFiles: map[string]string{
				"specs/test-module/feature1/spec.feature": `@deps:go @L2 @ov
Feature: test-module_feature1

  As a tester
  I want to test both manually and automatically
  So that I can verify behavior

  Rule: Mixed testing

    @L2 @ov
    Scenario: Automated scenario
      Given an automated precondition
      When I perform automated action
      Then automated result expected

    @Manual @L2 @ov
    Scenario: Manual scenario
      Given a manual precondition
      When I perform manual action
      Then manual result expected
`,
			},
			expectedScenarios: 1, // Only manual scenarios
			wantOutputFile:   "manual-test-scenarios.json",
		},
		{
			name:    "export manual scenarios from multiple feature files",
			module:  "test-module",
			release: "v1.0.0",
			format:  "json",
			featureFiles: map[string]string{
				"specs/test-module/feature1/spec.feature": `@deps:go @L2 @ov
Feature: test-module_feature1

  As a tester
  I want to test manually
  So that I can verify behavior

  Rule: Manual testing

    @Manual @L2 @ov
    Scenario: Manual scenario in feature1
      Given a precondition
      When I perform an action
      Then I expect a result
`,
				"specs/test-module/feature2/spec.feature": `@deps:go @L2 @ov
Feature: test-module_feature2

  As a tester
  I want to test manually
  So that I can verify behavior

  Rule: Manual testing

    @Manual @L2 @ov
    Scenario: Manual scenario in feature2
      Given a precondition
      When I perform an action
      Then I expect a result
`,
			},
			expectedScenarios: 2,
			wantOutputFile:   "manual-test-scenarios.json",
		},
		{
			name:    "export as CSV format",
			module:  "test-module",
			release: "v1.0.0",
			format:  "csv",
			featureFiles: map[string]string{
				"specs/test-module/feature1/spec.feature": `@deps:go @L2 @ov
Feature: test-module_feature1

  As a tester
  I want to test manually
  So that I can verify behavior

  Rule: Manual testing

    @Manual @L2 @ov
    Scenario: Manual test scenario
      Given a precondition
      When I perform an action
      Then I expect a result
`,
			},
			expectedScenarios: 1,
			wantOutputFile:   "manual-test-scenarios.csv",
		},
		{
			name:    "export as Markdown format",
			module:  "test-module",
			release: "v1.0.0",
			format:  "markdown",
			featureFiles: map[string]string{
				"specs/test-module/feature1/spec.feature": `@deps:go @L2 @ov
Feature: test-module_feature1

  As a tester
  I want to test manually
  So that I can verify behavior

  Rule: Manual testing

    @Manual @L2 @ov
    Scenario: Manual test scenario
      Given a precondition
      When I perform an action
      Then I expect a result
`,
			},
			expectedScenarios: 1,
			wantOutputFile:   "manual-test-scenarios.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test directory
			tmpDir := t.TempDir()
			setupTestRepository(t, tmpDir, tt.module, tt.featureFiles)

			// Change to temp directory
			oldWd, _ := os.Getwd()
			require.NoError(t, os.Chdir(tmpDir))
			defer func() { _ = os.Chdir(oldWd) }()

			// Run export-manual command
			exitCode := runExportManual(tt.module, tt.release, tt.format)

			// Should succeed
			assert.Equal(t, 0, exitCode, "command should succeed")

			// Verify output file exists
			outputPath := filepath.Join(tmpDir, tt.wantOutputFile)
			require.FileExists(t, outputPath, "output file should exist")

			// Verify content based on format
			content, err := os.ReadFile(outputPath)
			require.NoError(t, err)

			if tt.format == "json" {
				verifyJSONExport(t, content, tt.module, tt.release, tt.expectedScenarios)
			} else if tt.format == "csv" {
				verifyCSVExport(t, string(content), tt.expectedScenarios)
			} else if tt.format == "markdown" {
				verifyMarkdownExport(t, string(content), tt.expectedScenarios)
			}
		})
	}
}

func TestExportManualCommand_ErrorCases(t *testing.T) {
	tests := []struct {
		name         string
		module       string
		release      string
		format       string
		featureFiles map[string]string
		wantExitCode int
		wantError    string
	}{
		{
			name:         "no manual scenarios found",
			module:       "test-module",
			release:      "v1.0.0",
			format:       "json",
			featureFiles: map[string]string{
				"specs/test-module/feature1/spec.feature": `@deps:go @L2 @ov
Feature: test-module_feature1

  As a tester
  I want to test automatically
  So that I can verify behavior

  Rule: Automated testing

    @L2 @ov
    Scenario: Automated scenario
      Given a precondition
      When I perform an action
      Then I expect a result
`,
			},
			wantExitCode: 1,
			wantError:    "no manual test scenarios found",
		},
		{
			name:         "unknown module",
			module:       "nonexistent-module",
			release:      "v1.0.0",
			format:       "json",
			featureFiles: map[string]string{},
			wantExitCode: 1,
			wantError:    "module not found",
		},
		{
			name:    "missing release flag",
			module:  "test-module",
			release: "",
			format:  "json",
			featureFiles: map[string]string{
				"specs/test-module/feature1/spec.feature": `@deps:go @L2 @ov
Feature: test-module_feature1

  Rule: Manual testing

    @Manual @L2 @ov
    Scenario: Manual scenario
      Given a precondition
`,
			},
			wantExitCode: 1,
			wantError:    "release flag is required",
		},
		{
			name:    "invalid format",
			module:  "test-module",
			release: "v1.0.0",
			format:  "xml",
			featureFiles: map[string]string{
				"specs/test-module/feature1/spec.feature": `@deps:go @L2 @ov
Feature: test-module_feature1

  Rule: Manual testing

    @Manual @L2 @ov
    Scenario: Manual scenario
      Given a precondition
`,
			},
			wantExitCode: 1,
			wantError:    "invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test directory
			tmpDir := t.TempDir()
			if len(tt.featureFiles) > 0 {
				setupTestRepository(t, tmpDir, tt.module, tt.featureFiles)
			} else {
				setupMinimalRepository(t, tmpDir)
			}

			// Change to temp directory
			oldWd, _ := os.Getwd()
			require.NoError(t, os.Chdir(tmpDir))
			defer func() { _ = os.Chdir(oldWd) }()

			// Run export-manual command
			exitCode := runExportManual(tt.module, tt.release, tt.format)

			// Should fail with expected exit code
			assert.Equal(t, tt.wantExitCode, exitCode, "command should fail with expected exit code")
		})
	}
}

func TestExportManualCommand_ScenarioIDGeneration(t *testing.T) {
	tests := []struct {
		name         string
		featureName  string
		scenarioName string
		wantID       string
	}{
		{
			name:         "simple scenario ID",
			featureName:  "test-module_feature1",
			scenarioName: "Manual test scenario",
			wantID:       "test-modulefeature1:manual-test-scenario",  // Underscore removed by slugify
		},
		{
			name:         "scenario with special characters",
			featureName:  "test-module_feature1",
			scenarioName: "Test Scenario: With Special Chars!",
			wantID:       "test-modulefeature1:test-scenario-with-special-chars",  // Underscore removed by slugify
		},
		{
			name:         "scenario with multiple spaces",
			featureName:  "test-module_feature1",
			scenarioName: "Test   Multiple   Spaces",
			wantID:       "test-modulefeature1:test-multiple-spaces",  // Underscore removed by slugify
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario := gherkin.ScenarioDetail{
				FeatureName: tt.featureName,
				Name:        tt.scenarioName,
			}
			scenarioID := gherkin.GenerateScenarioID(scenario)
			assert.Equal(t, tt.wantID, scenarioID)
		})
	}
}

func TestExportManualCommand_SchemaValidation(t *testing.T) {
	tests := []struct {
		name    string
		export  ManualTestExport
		wantErr bool
	}{
		{
			name: "valid export passes schema validation",
			export: ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2024-01-19T12:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					GitCommit:      "1234567890abcdef1234567890abcdef12345678",
					SchemaVersion:  "1.0",
				},
				Scenarios: []ExportedScenario{
					{
						ScenarioID:   "test-module-feature1:manual-scenario",
						FeatureName:  "test-module_feature1",
						ScenarioName: "Manual scenario",
						Tags:         []string{"@Manual", "@L2", "@ov"},
						Steps: []string{
							"Given a precondition",
							"When I perform an action",
							"Then I expect a result",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing required field fails validation",
			export: ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2024-01-19T12:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "",
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
			wantErr: true,
		},
		{
			name: "invalid scenario ID format fails validation",
			export: ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2024-01-19T12:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					GitCommit:      "1234567890abcdef1234567890abcdef12345678",
					SchemaVersion:  "1.0",
				},
				Scenarios: []ExportedScenario{
					{
						ScenarioID:   "InvalidID",  // Missing colon separator
						FeatureName:  "test-module_feature1",
						ScenarioName: "Manual scenario",
						Tags:         []string{"@Manual"},
						Steps:        []string{"Given a precondition"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty scenarios array fails validation",
			export: ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2024-01-19T12:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					GitCommit:      "1234567890abcdef1234567890abcdef12345678",
					SchemaVersion:  "1.0",
				},
				Scenarios: []ExportedScenario{},
			},
			wantErr: true,
		},
		{
			name: "invalid git commit SHA fails validation",
			export: ManualTestExport{
				ExportMetadata: ExportMetadata{
					ExportTime:     "2024-01-19T12:00:00Z",
					Module:         "test-module",
					ReleaseVersion: "v1.0.0",
					GitCommit:      "invalid-sha",
					SchemaVersion:  "1.0",
				},
				Scenarios: []ExportedScenario{
					{
						ScenarioID:   "test-module-feature1:manual-scenario",
						FeatureName:  "test-module_feature1",
						ScenarioName: "Manual scenario",
						Tags:         []string{"@Manual"},
						Steps:        []string{"Given a precondition"},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create schema file in temp directory
			schemaDir := filepath.Join(tmpDir, "contracts", "eac-core", "0.1.0")
			require.NoError(t, os.MkdirAll(schemaDir, 0755))

			// Read the real schema from the repo
			workspaceRoot, err := os.Getwd()
			require.NoError(t, err)
			// Navigate up from go/eac/commands/impl/test to repo root
			repoRoot := filepath.Join(workspaceRoot, "..", "..", "..", "..", "..")
			realSchemaPath := filepath.Join(repoRoot, "contracts", "eac-core", "0.1.0", "manual-test-export.schema.json")
			schemaContent, err := os.ReadFile(realSchemaPath)
			require.NoError(t, err, "should read real schema file")

			// Write schema to temp directory
			schemaFile := filepath.Join(schemaDir, "manual-test-export.schema.json")
			require.NoError(t, os.WriteFile(schemaFile, schemaContent, 0644))

			// Write export to temp JSON file
			dataFile := filepath.Join(tmpDir, "export.json")
			data, err := json.MarshalIndent(tt.export, "", "  ")
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

func TestExportManualCommand_TagExtraction(t *testing.T) {
	tests := []struct {
		name         string
		scenarioTags []string
		featureTags  []string
		wantTags     []string
	}{
		{
			name:         "scenario tags only",
			scenarioTags: []string{"@Manual", "@L2", "@ov"},
			featureTags:  []string{},
			wantTags:     []string{"@Manual", "@L2", "@ov"},
		},
		{
			name:         "scenario and feature tags",
			scenarioTags: []string{"@Manual", "@L2"},
			featureTags:  []string{"@deps:go", "@ov"},
			wantTags:     []string{"@deps:go", "@ov", "@Manual", "@L2"},
		},
		{
			name:         "duplicate tags should be deduplicated",
			scenarioTags: []string{"@Manual", "@ov"},
			featureTags:  []string{"@ov", "@deps:go"},
			wantTags:     []string{"@ov", "@deps:go", "@Manual"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := gherkin.CombineTags(tt.featureTags, tt.scenarioTags)
			assert.ElementsMatch(t, tt.wantTags, tags)
		})
	}
}

// ==============================================================================
// Test Helper Functions
// ==============================================================================

// runExportManual runs the export-manual command with specific flags
func runExportManual(module, release, format string) int {
	// Save original args and restore after
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args for command
	os.Args = []string{"eac", "test", "export-manual", "--module", module, "--release", release, "--format", format}

	return ExportManual()
}

func setupTestRepository(t *testing.T, tmpDir, module string, featureFiles map[string]string) {
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

	// Create feature files
	for path, content := range featureFiles {
		fullPath := filepath.Join(tmpDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	// Create contract schema
	schemaDir := filepath.Join(tmpDir, "contracts", "eac-core", "0.1.0")
	require.NoError(t, os.MkdirAll(schemaDir, 0755))

	schemaContent := `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "export_metadata": {
      "type": "object",
      "properties": {
        "export_time": {"type": "string", "format": "date-time"},
        "module": {"type": "string"},
        "release_version": {"type": "string"},
        "git_commit": {"type": "string", "pattern": "^[a-f0-9]{40}$"},
        "schema_version": {"type": "string", "const": "1.0"}
      },
      "required": ["export_time", "module", "release_version", "git_commit", "schema_version"]
    },
    "scenarios": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "scenario_id": {"type": "string", "pattern": "^[a-z][a-z0-9-]*:[a-z][a-z0-9-]*$"},
          "feature_name": {"type": "string"},
          "scenario_name": {"type": "string"},
          "tags": {"type": "array", "items": {"type": "string"}},
          "steps": {"type": "array", "items": {"type": "string"}}
        },
        "required": ["scenario_id", "feature_name", "scenario_name", "tags", "steps"]
      },
      "minItems": 1
    }
  },
  "required": ["export_metadata", "scenarios"]
}`
	schemaFile := filepath.Join(schemaDir, "manual-test-export.schema.json")
	require.NoError(t, os.WriteFile(schemaFile, []byte(schemaContent), 0644))
}

func setupMinimalRepository(t *testing.T, tmpDir string) {
	t.Helper()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))

	// Create .r2r/eac directory
	r2rDir := filepath.Join(tmpDir, ".r2r", "eac")
	require.NoError(t, os.MkdirAll(r2rDir, 0755))

	// Create minimal repository.yml
	repoYAML := `repository:
  type: mono

modules: []
`
	repoFile := filepath.Join(r2rDir, "repository.yml")
	require.NoError(t, os.WriteFile(repoFile, []byte(repoYAML), 0644))
}

func verifyJSONExport(t *testing.T, content []byte, module, release string, expectedScenarios int) {
	t.Helper()

	var export ManualTestExport
	err := json.Unmarshal(content, &export)
	require.NoError(t, err, "should be valid JSON")

	// Verify metadata
	assert.Equal(t, module, export.ExportMetadata.Module)
	assert.Equal(t, release, export.ExportMetadata.ReleaseVersion)
	assert.Equal(t, "1.0", export.ExportMetadata.SchemaVersion)
	assert.NotEmpty(t, export.ExportMetadata.ExportTime)
	assert.NotEmpty(t, export.ExportMetadata.GitCommit)

	// Verify scenarios
	assert.Len(t, export.Scenarios, expectedScenarios)

	for _, scenario := range export.Scenarios {
		// Verify scenario ID format (feature-name:scenario-name)
		assert.Regexp(t, `^[a-z][a-z0-9-]*:[a-z][a-z0-9-]*$`, scenario.ScenarioID)

		// Verify required fields
		assert.NotEmpty(t, scenario.FeatureName)
		assert.NotEmpty(t, scenario.ScenarioName)
		assert.NotEmpty(t, scenario.Tags)
		assert.NotEmpty(t, scenario.Steps)

		// Verify @Manual tag is present
		assert.Contains(t, scenario.Tags, "@Manual")
	}
}

func verifyCSVExport(t *testing.T, content string, expectedScenarios int) {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(content), "\n")

	// Should have header + scenarios
	assert.GreaterOrEqual(t, len(lines), expectedScenarios+1, "should have header and scenario rows")

	// Verify header
	header := lines[0]
	assert.Contains(t, header, "scenario_id")
	assert.Contains(t, header, "feature_name")
	assert.Contains(t, header, "scenario_name")
	assert.Contains(t, header, "tags")
	assert.Contains(t, header, "steps")
}

func verifyMarkdownExport(t *testing.T, content string, expectedScenarios int) {
	t.Helper()

	// Should contain markdown elements
	assert.Contains(t, content, "#")
	assert.Contains(t, content, "Manual Test Scenarios")

	// Count scenario sections (format: "## 1. Scenario Name")
	lines := strings.Split(content, "\n")
	scenarioCount := 0
	for _, line := range lines {
		// Match lines like "## 1. ", "## 2. ", etc.
		if strings.HasPrefix(line, "## ") && len(line) > 3 {
			rest := line[3:]
			if len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
				scenarioCount++
			}
		}
	}
	assert.Equal(t, expectedScenarios, scenarioCount)
}
