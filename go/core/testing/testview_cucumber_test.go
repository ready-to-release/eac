package testing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCucumberFile_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "cucumber.json")

	content := `[
		{
			"uri": "specs/mymodule/login/specification.feature",
			"elements": [
				{
					"name": "User can log in",
					"type": "scenario",
					"tags": [{"name": "@L0"}, {"name": "@control:auth-1"}],
					"steps": [
						{"result": {"status": "passed", "duration": 1000000}},
						{"result": {"status": "passed", "duration": 2000000}}
					]
				},
				{
					"name": "User sees error on bad credentials",
					"type": "scenario",
					"tags": [{"name": "@L1"}],
					"steps": [
						{"result": {"status": "passed", "duration": 500000}},
						{"result": {"status": "failed", "duration": 300000}}
					]
				}
			]
		}
	]`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	entries := parseCucumberFile(path, "mymodule")

	assert.Len(t, entries, 2)

	// First scenario: passed
	assert.Equal(t, "User can log in", entries[0].Name)
	assert.Equal(t, "mymodule", entries[0].Module)
	assert.Equal(t, "godog", entries[0].Type)
	assert.Equal(t, StatusPassed, entries[0].Status)
	assert.Equal(t, "unit", entries[0].Suite) // @L0 -> unit
	assert.Equal(t, int64(3), entries[0].DurationMs) // 3000000ns / 1M = 3ms
	assert.Contains(t, entries[0].Tags, "@L0")
	assert.Contains(t, entries[0].Tags, "@control:auth-1")
	assert.Equal(t, "specs/mymodule/login/specification.feature", entries[0].FilePath)

	// Second scenario: failed
	assert.Equal(t, "User sees error on bad credentials", entries[1].Name)
	assert.Equal(t, StatusFailed, entries[1].Status)
	assert.Equal(t, "integration", entries[1].Suite) // @L1 -> integration
}

func TestParseCucumberFile_SkipsBackground(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "cucumber.json")

	content := `[
		{
			"uri": "test.feature",
			"elements": [
				{
					"name": "Background setup",
					"type": "background",
					"steps": [{"result": {"status": "passed", "duration": 100}}]
				},
				{
					"name": "Actual scenario",
					"type": "scenario",
					"tags": [],
					"steps": [{"result": {"status": "passed", "duration": 100}}]
				}
			]
		}
	]`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	entries := parseCucumberFile(path, "mod")
	assert.Len(t, entries, 1)
	assert.Equal(t, "Actual scenario", entries[0].Name)
}

func TestParseCucumberFile_NonexistentFile(t *testing.T) {
	entries := parseCucumberFile("/nonexistent/cucumber.json", "mod")
	assert.Nil(t, entries)
}

func TestParseCucumberFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "cucumber.json")
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0644))

	entries := parseCucumberFile(path, "mod")
	assert.Nil(t, entries)
}

func TestParseCucumberFile_EmptyArray(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "cucumber.json")
	require.NoError(t, os.WriteFile(path, []byte("[]"), 0644))

	entries := parseCucumberFile(path, "mod")
	assert.Empty(t, entries)
}

func TestDeriveScenarioStatus(t *testing.T) {
	tests := []struct {
		name     string
		steps    []cucumberStep
		expected string
	}{
		{
			name:     "all passed",
			steps:    []cucumberStep{{Result: cucumberResult{Status: "passed"}}},
			expected: StatusPassed,
		},
		{
			name:     "one failed",
			steps:    []cucumberStep{{Result: cucumberResult{Status: "passed"}}, {Result: cucumberResult{Status: "failed"}}},
			expected: StatusFailed,
		},
		{
			name:     "undefined",
			steps:    []cucumberStep{{Result: cucumberResult{Status: "undefined"}}},
			expected: StatusFailed,
		},
		{
			name:     "pending",
			steps:    []cucumberStep{{Result: cucumberResult{Status: "pending"}}},
			expected: StatusSkipped,
		},
		{
			name:     "empty steps",
			steps:    nil,
			expected: StatusPassed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, deriveScenarioStatus(tt.steps))
		})
	}
}

func TestDeriveSuiteFromTags(t *testing.T) {
	assert.Equal(t, "unit", deriveSuiteFromTags([]string{"@L0"}))
	assert.Equal(t, "integration", deriveSuiteFromTags([]string{"@L1"}))
	assert.Equal(t, "e2e", deriveSuiteFromTags([]string{"@L2"}))
	assert.Equal(t, "e2e", deriveSuiteFromTags([]string{"@HE2E"}))
	assert.Equal(t, "unit", deriveSuiteFromTags([]string{"@other"}))
	assert.Equal(t, "unit", deriveSuiteFromTags(nil))
}
