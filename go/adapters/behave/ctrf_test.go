package behave

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/ctrf"
)

func TestConvertBehaveJSONToCTRF(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		wantNil     bool
		wantTests   int
		wantPassed  int
		wantFailed  int
		wantSkipped int
		wantPending int
		wantTool    string
		checkNames  []string
		checkSuites []string
		checkStatus []ctrf.Status
		checkMsg    string
	}{
		{
			name: "valid behave JSON with passed and failed scenarios",
			jsonData: `[
				{
					"name": "Calculator",
					"keyword": "Feature",
					"status": "passed",
					"elements": [
						{
							"name": "Add two numbers",
							"keyword": "Scenario",
							"type": "scenario",
							"status": "passed",
							"steps": [
								{
									"name": "a calculator",
									"keyword": "Given",
									"status": "passed",
									"duration": 0.001,
									"result": {"status": "passed", "duration": 0.001}
								},
								{
									"name": "I add 1 and 2",
									"keyword": "When",
									"status": "passed",
									"duration": 0.002,
									"result": {"status": "passed", "duration": 0.002}
								}
							]
						},
						{
							"name": "Subtract two numbers",
							"keyword": "Scenario",
							"type": "scenario",
							"status": "failed",
							"steps": [
								{
									"name": "a calculator",
									"keyword": "Given",
									"status": "passed",
									"duration": 0.001,
									"result": {"status": "passed", "duration": 0.001}
								},
								{
									"name": "I subtract 5 from 3",
									"keyword": "When",
									"status": "failed",
									"duration": 0.003,
									"result": {"status": "failed", "duration": 0.003, "error_message": "expected -2 but got 2"}
								}
							]
						}
					]
				}
			]`,
			wantNil:    false,
			wantTests:  2,
			wantPassed: 1,
			wantFailed: 1,
			wantTool:   "behave",
			checkNames: []string{
				"Calculator: Add two numbers",
				"Calculator: Subtract two numbers",
			},
			checkSuites: []string{
				"Calculator",
				"Calculator",
			},
			checkStatus: []ctrf.Status{
				ctrf.StatusPassed,
				ctrf.StatusFailed,
			},
			checkMsg: "expected -2 but got 2",
		},
		{
			name: "multiple features with mixed statuses",
			jsonData: `[
				{
					"name": "Feature A",
					"keyword": "Feature",
					"status": "passed",
					"elements": [
						{
							"name": "Scenario A1",
							"keyword": "Scenario",
							"type": "scenario",
							"status": "passed",
							"steps": [
								{"name": "step 1", "keyword": "Given", "status": "passed", "duration": 0.01, "result": {"status": "passed", "duration": 0.01}}
							]
						}
					]
				},
				{
					"name": "Feature B",
					"keyword": "Feature",
					"status": "failed",
					"elements": [
						{
							"name": "Scenario B1",
							"keyword": "Scenario",
							"type": "scenario",
							"status": "skipped",
							"steps": [
								{"name": "step 1", "keyword": "Given", "status": "skipped", "duration": 0.0, "result": {"status": "skipped", "duration": 0.0}}
							]
						}
					]
				}
			]`,
			wantNil:     false,
			wantTests:   2,
			wantPassed:  1,
			wantSkipped: 1,
			wantTool:    "behave",
			checkNames: []string{
				"Feature A: Scenario A1",
				"Feature B: Scenario B1",
			},
			checkSuites: []string{
				"Feature A",
				"Feature B",
			},
			checkStatus: []ctrf.Status{
				ctrf.StatusPassed,
				ctrf.StatusSkipped,
			},
		},
		{
			name: "background elements are skipped",
			jsonData: `[
				{
					"name": "Feature C",
					"keyword": "Feature",
					"status": "passed",
					"elements": [
						{
							"name": "",
							"keyword": "Background",
							"type": "background",
							"status": "passed",
							"steps": [
								{"name": "setup step", "keyword": "Given", "status": "passed", "duration": 0.001, "result": {"status": "passed", "duration": 0.001}}
							]
						},
						{
							"name": "Real scenario",
							"keyword": "Scenario",
							"type": "scenario",
							"status": "passed",
							"steps": [
								{"name": "a step", "keyword": "When", "status": "passed", "duration": 0.005, "result": {"status": "passed", "duration": 0.005}}
							]
						}
					]
				}
			]`,
			wantNil:    false,
			wantTests:  1,
			wantPassed: 1,
			wantTool:   "behave",
			checkNames: []string{"Feature C: Real scenario"},
		},
		{
			name: "undefined scenario maps to pending status",
			jsonData: `[
				{
					"name": "Feature D",
					"keyword": "Feature",
					"status": "passed",
					"elements": [
						{
							"name": "Undefined scenario",
							"keyword": "Scenario",
							"type": "scenario",
							"status": "undefined",
							"steps": [
								{"name": "an undefined step", "keyword": "Given", "status": "undefined", "duration": 0.0, "result": {"status": "undefined", "duration": 0.0}}
							]
						}
					]
				}
			]`,
			wantNil:     false,
			wantTests:   1,
			wantPending: 1,
			wantTool:    "behave",
			checkStatus: []ctrf.Status{ctrf.StatusPending},
		},
		{
			name:      "empty array returns report with zero tests",
			jsonData:  `[]`,
			wantNil:   false,
			wantTests: 0,
			wantTool:  "behave",
		},
		{
			name:    "invalid JSON returns nil",
			jsonData: `{not valid json`,
			wantNil: true,
		},
		{
			name:    "empty input returns nil",
			jsonData: ``,
			wantNil: true,
		},
		{
			name:    "JSON object instead of array returns nil",
			jsonData: `{"name": "not an array"}`,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertBehaveJSONToCTRF([]byte(tt.jsonData))

			if tt.wantNil {
				if got != nil {
					t.Fatalf("convertBehaveJSONToCTRF() = non-nil, want nil")
				}
				return
			}

			if got == nil {
				t.Fatalf("convertBehaveJSONToCTRF() = nil, want non-nil")
			}

			// Verify tool name
			if tt.wantTool != "" && got.Results.Tool.Name != tt.wantTool {
				t.Errorf("tool name = %q, want %q", got.Results.Tool.Name, tt.wantTool)
			}

			// Verify test count
			if len(got.Results.Tests) != tt.wantTests {
				t.Errorf("test count = %d, want %d", len(got.Results.Tests), tt.wantTests)
			}

			// Verify summary counts
			if got.Results.Summary.Tests != tt.wantTests {
				t.Errorf("summary.tests = %d, want %d", got.Results.Summary.Tests, tt.wantTests)
			}
			if got.Results.Summary.Passed != tt.wantPassed {
				t.Errorf("summary.passed = %d, want %d", got.Results.Summary.Passed, tt.wantPassed)
			}
			if got.Results.Summary.Failed != tt.wantFailed {
				t.Errorf("summary.failed = %d, want %d", got.Results.Summary.Failed, tt.wantFailed)
			}
			if got.Results.Summary.Skipped != tt.wantSkipped {
				t.Errorf("summary.skipped = %d, want %d", got.Results.Summary.Skipped, tt.wantSkipped)
			}
			if got.Results.Summary.Pending != tt.wantPending {
				t.Errorf("summary.pending = %d, want %d", got.Results.Summary.Pending, tt.wantPending)
			}

			// Verify individual test names
			for i, wantName := range tt.checkNames {
				if i >= len(got.Results.Tests) {
					t.Errorf("expected test[%d] with name %q but only %d tests found", i, wantName, len(got.Results.Tests))
					break
				}
				if got.Results.Tests[i].Name != wantName {
					t.Errorf("test[%d].name = %q, want %q", i, got.Results.Tests[i].Name, wantName)
				}
			}

			// Verify individual test suites
			for i, wantSuite := range tt.checkSuites {
				if i >= len(got.Results.Tests) {
					break
				}
				if got.Results.Tests[i].Suite != wantSuite {
					t.Errorf("test[%d].suite = %q, want %q", i, got.Results.Tests[i].Suite, wantSuite)
				}
			}

			// Verify individual test statuses
			for i, wantStatus := range tt.checkStatus {
				if i >= len(got.Results.Tests) {
					break
				}
				if got.Results.Tests[i].Status != wantStatus {
					t.Errorf("test[%d].status = %q, want %q", i, got.Results.Tests[i].Status, wantStatus)
				}
			}

			// Verify error message on the failed test
			if tt.checkMsg != "" {
				found := false
				for _, test := range got.Results.Tests {
					if test.Message == tt.checkMsg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("no test found with message %q", tt.checkMsg)
				}
			}

			// Verify timestamps are set (Finalize was called)
			if got.Results.Summary.Start == 0 {
				t.Error("summary.start should be non-zero after NewReport")
			}
			if got.Results.Summary.Stop == 0 {
				t.Error("summary.stop should be non-zero after Finalize")
			}
		})
	}
}
