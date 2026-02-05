//go:build L1
// +build L1

package orchestrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseCucumberJSON(t *testing.T) {
	tests := []struct {
		name           string
		cucumberJSON   string
		wantErrors     int
		wantSubstrings []string
	}{
		{
			name: "extracts error from failed scenario",
			cucumberJSON: `[{
				"uri": "features/test.feature",
				"id": "test-feature",
				"name": "Test Feature",
				"elements": [{
					"id": "test-scenario",
					"name": "Test scenario",
					"type": "scenario",
					"steps": [
						{"keyword": "Given ", "name": "a precondition", "result": {"status": "passed"}},
						{"keyword": "When ", "name": "an action", "result": {"status": "failed", "error_message": "expected 5 but got 3"}}
					]
				}]
			}]`,
			wantErrors:     1,
			wantSubstrings: []string{"Test scenario", "expected 5 but got 3"},
		},
		{
			name: "no errors for passing scenarios",
			cucumberJSON: `[{
				"uri": "features/test.feature",
				"id": "test-feature",
				"name": "Test Feature",
				"elements": [{
					"id": "test-scenario",
					"name": "Test scenario",
					"type": "scenario",
					"steps": [
						{"keyword": "Given ", "name": "a precondition", "result": {"status": "passed"}},
						{"keyword": "When ", "name": "an action", "result": {"status": "passed"}}
					]
				}]
			}]`,
			wantErrors:     0,
			wantSubstrings: nil,
		},
		{
			name: "handles multiple failed scenarios",
			cucumberJSON: `[{
				"uri": "features/test.feature",
				"id": "test-feature",
				"name": "Test Feature",
				"elements": [
					{
						"id": "scenario-1",
						"name": "First scenario",
						"type": "scenario",
						"steps": [{"keyword": "When ", "name": "step fails", "result": {"status": "failed", "error_message": "error one"}}]
					},
					{
						"id": "scenario-2",
						"name": "Second scenario",
						"type": "scenario",
						"steps": [{"keyword": "When ", "name": "step fails", "result": {"status": "failed", "error_message": "error two"}}]
					}
				]
			}]`,
			wantErrors:     2,
			wantSubstrings: []string{"First scenario", "error one", "Second scenario", "error two"},
		},
		{
			name: "reports all failed steps per scenario",
			cucumberJSON: `[{
				"elements": [{
					"name": "Multi-failure scenario",
					"steps": [
						{"result": {"status": "failed", "error_message": "first error"}},
						{"result": {"status": "failed", "error_message": "second error"}}
					]
				}]
			}]`,
			wantErrors:     2,
			wantSubstrings: []string{"first error", "second error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory with cucumber.json
			tmpDir := t.TempDir()
			cucumberPath := filepath.Join(tmpDir, "cucumber.json")
			if err := os.WriteFile(cucumberPath, []byte(tt.cucumberJSON), 0o644); err != nil {
				t.Fatalf("failed to write cucumber.json: %v", err)
			}

			warnings, errors := parseCucumberJSON(cucumberPath)

			// Check error count
			if len(errors) != tt.wantErrors {
				t.Errorf("got %d errors, want %d\nerrors: %v", len(errors), tt.wantErrors, errors)
			}

			// Check for expected substrings
			allErrors := strings.Join(errors, "\n")
			for _, sub := range tt.wantSubstrings {
				if !strings.Contains(allErrors, sub) {
					t.Errorf("errors missing expected substring %q\nerrors: %v", sub, errors)
				}
			}

			// Verify no warnings (cucumber doesn't produce warnings)
			if len(warnings) != 0 {
				t.Errorf("expected 0 warnings, got %d: %v", len(warnings), warnings)
			}
		})
	}
}

func TestParseCucumberJSON_MissingFile(t *testing.T) {
	warnings, errors := parseCucumberJSON("/nonexistent/path/cucumber.json")

	if len(warnings) != 0 || len(errors) != 0 {
		t.Errorf("expected empty results for missing file, got warnings=%v errors=%v", warnings, errors)
	}
}

func TestParseCucumberJSON_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cucumberPath := filepath.Join(tmpDir, "cucumber.json")
	if err := os.WriteFile(cucumberPath, []byte("not valid json"), 0o644); err != nil {
		t.Fatalf("failed to write cucumber.json: %v", err)
	}

	warnings, errors := parseCucumberJSON(cucumberPath)

	if len(warnings) != 0 || len(errors) != 0 {
		t.Errorf("expected empty results for invalid JSON, got warnings=%v errors=%v", warnings, errors)
	}
}

func TestParseCTRFJSON(t *testing.T) {
	tests := []struct {
		name           string
		ctrfJSON       string
		wantErrors     int
		wantSubstrings []string
	}{
		{
			name: "extracts error from failed test with trace",
			ctrfJSON: `{
				"results": {
					"tests": [
						{"name": "TestFoo", "status": "passed"},
						{"name": "TestBar", "status": "failed", "trace": "expected 1, got 2"}
					]
				}
			}`,
			wantErrors:     1,
			wantSubstrings: []string{"TestBar", "expected 1, got 2"},
		},
		{
			name: "uses message when trace is empty",
			ctrfJSON: `{
				"results": {
					"tests": [
						{"name": "TestBaz", "status": "failed", "message": "assertion error"}
					]
				}
			}`,
			wantErrors:     1,
			wantSubstrings: []string{"TestBaz", "assertion error"},
		},
		{
			name: "handles test with no error details",
			ctrfJSON: `{
				"results": {
					"tests": [
						{"name": "TestEmpty", "status": "failed"}
					]
				}
			}`,
			wantErrors:     1,
			wantSubstrings: []string{"TestEmpty: failed"},
		},
		{
			name: "no errors for passing tests",
			ctrfJSON: `{
				"results": {
					"tests": [
						{"name": "TestPass", "status": "passed"}
					]
				}
			}`,
			wantErrors:     0,
			wantSubstrings: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			ctrfPath := filepath.Join(tmpDir, "unit.json")
			if err := os.WriteFile(ctrfPath, []byte(tt.ctrfJSON), 0o644); err != nil {
				t.Fatalf("failed to write unit.json: %v", err)
			}

			warnings, errors := parseCTRFJSON(ctrfPath)

			if len(errors) != tt.wantErrors {
				t.Errorf("got %d errors, want %d\nerrors: %v", len(errors), tt.wantErrors, errors)
			}

			allErrors := strings.Join(errors, "\n")
			for _, sub := range tt.wantSubstrings {
				if !strings.Contains(allErrors, sub) {
					t.Errorf("errors missing expected substring %q\nerrors: %v", sub, errors)
				}
			}

			if len(warnings) != 0 {
				t.Errorf("expected 0 warnings, got %d: %v", len(warnings), warnings)
			}
		})
	}
}

func TestParseLogForIssues_PrioritizesCucumber(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	cucumberPath := filepath.Join(tmpDir, "cucumber.json")

	// Write log file with some content
	if err := os.WriteFile(logPath, []byte("some progress output...\n"), 0o644); err != nil {
		t.Fatalf("failed to write log: %v", err)
	}

	// Write cucumber.json with a failure
	cucumberJSON := `[{
		"elements": [{
			"name": "Failed test",
			"steps": [{"result": {"status": "failed", "error_message": "assertion failed"}}]
		}]
	}]`
	if err := os.WriteFile(cucumberPath, []byte(cucumberJSON), 0o644); err != nil {
		t.Fatalf("failed to write cucumber.json: %v", err)
	}

	warnings, errors := parseLogForIssues(logPath)

	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errors), errors)
	}
	if len(errors) > 0 && !strings.Contains(errors[0], "assertion failed") {
		t.Errorf("expected error to contain 'assertion failed', got: %s", errors[0])
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d: %v", len(warnings), warnings)
	}
}

func TestParseLogForIssues_PrioritizesCTRF(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	unitPath := filepath.Join(tmpDir, "unit.json")

	// Write log file
	if err := os.WriteFile(logPath, []byte("some output\n"), 0o644); err != nil {
		t.Fatalf("failed to write log: %v", err)
	}

	// Write unit.json with a failure
	ctrfJSON := `{
		"results": {
			"tests": [{"name": "TestUnit", "status": "failed", "trace": "unit test failed"}]
		}
	}`
	if err := os.WriteFile(unitPath, []byte(ctrfJSON), 0o644); err != nil {
		t.Fatalf("failed to write unit.json: %v", err)
	}

	warnings, errors := parseLogForIssues(logPath)

	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errors), errors)
	}
	if len(errors) > 0 && !strings.Contains(errors[0], "unit test failed") {
		t.Errorf("expected error to contain 'unit test failed', got: %s", errors[0])
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d: %v", len(warnings), warnings)
	}
}
