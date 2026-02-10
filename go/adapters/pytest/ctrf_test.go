package pytest

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/ctrf"
)

func TestConvertPytestJSONToCTRF(t *testing.T) {
	tests := []struct {
		name           string
		jsonData       string
		wantNil        bool
		wantTests      int
		wantPassed     int
		wantFailed     int
		wantSkipped    int
		wantTool       string
		checkTestNames []string
		checkStatuses  []ctrf.Status
		checkMessage   string
		checkTrace     string
	}{
		{
			name: "valid report with passed, failed, and skipped tests",
			jsonData: `{
				"summary": {"passed": 2, "failed": 1, "total": 3},
				"tests": [
					{
						"nodeid": "tests/test_calc.py::test_add",
						"outcome": "passed",
						"duration": 0.001,
						"setup": {"duration": 0.0001, "outcome": "passed"},
						"call": {"duration": 0.001, "outcome": "passed"},
						"teardown": {"duration": 0.0001, "outcome": "passed"}
					},
					{
						"nodeid": "tests/test_calc.py::test_sub",
						"outcome": "failed",
						"duration": 0.002,
						"setup": {"duration": 0.0001, "outcome": "passed"},
						"call": {
							"duration": 0.002,
							"outcome": "failed",
							"longrepr": "assert 1 == 2",
							"crash": {"path": "tests/test_calc.py", "lineno": 10, "message": "AssertionError"}
						},
						"teardown": {"duration": 0.0001, "outcome": "passed"}
					},
					{
						"nodeid": "tests/test_calc.py::test_skip",
						"outcome": "skipped",
						"duration": 0.0
					}
				]
			}`,
			wantNil:     false,
			wantTests:   3,
			wantPassed:  1,
			wantFailed:  1,
			wantSkipped: 1,
			wantTool:    "pytest",
			checkTestNames: []string{
				"tests/test_calc.py::test_add",
				"tests/test_calc.py::test_sub",
				"tests/test_calc.py::test_skip",
			},
			checkStatuses: []ctrf.Status{
				ctrf.StatusPassed,
				ctrf.StatusFailed,
				ctrf.StatusSkipped,
			},
			checkMessage: "AssertionError",
			checkTrace:   "assert 1 == 2",
		},
		{
			name:     "empty JSON object returns report with zero tests",
			jsonData: `{}`,
			wantNil:  false,
			wantTests: 0,
			wantTool: "pytest",
		},
		{
			name:     "valid report with no tests array returns zero tests",
			jsonData: `{"summary": {"passed": 0, "failed": 0, "total": 0}}`,
			wantNil:  false,
			wantTests: 0,
			wantTool: "pytest",
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
			name: "test with unknown outcome maps to other status",
			jsonData: `{
				"summary": {"total": 1},
				"tests": [
					{
						"nodeid": "tests/test_calc.py::test_unknown",
						"outcome": "xfailed",
						"duration": 0.005
					}
				]
			}`,
			wantNil:    false,
			wantTests:  1,
			wantTool:   "pytest",
			checkStatuses: []ctrf.Status{ctrf.StatusOther},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertPytestJSONToCTRF([]byte(tt.jsonData))

			if tt.wantNil {
				if got != nil {
					t.Fatalf("convertPytestJSONToCTRF() = non-nil, want nil")
				}
				return
			}

			if got == nil {
				t.Fatalf("convertPytestJSONToCTRF() = nil, want non-nil")
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

			// Verify individual test names
			for i, wantName := range tt.checkTestNames {
				if i >= len(got.Results.Tests) {
					break
				}
				if got.Results.Tests[i].Name != wantName {
					t.Errorf("test[%d].name = %q, want %q", i, got.Results.Tests[i].Name, wantName)
				}
			}

			// Verify individual test statuses
			for i, wantStatus := range tt.checkStatuses {
				if i >= len(got.Results.Tests) {
					break
				}
				if got.Results.Tests[i].Status != wantStatus {
					t.Errorf("test[%d].status = %q, want %q", i, got.Results.Tests[i].Status, wantStatus)
				}
			}

			// Verify error message on the failed test (index 1)
			if tt.checkMessage != "" && len(got.Results.Tests) > 1 {
				if got.Results.Tests[1].Message != tt.checkMessage {
					t.Errorf("failed test message = %q, want %q", got.Results.Tests[1].Message, tt.checkMessage)
				}
			}

			// Verify trace on the failed test (index 1)
			if tt.checkTrace != "" && len(got.Results.Tests) > 1 {
				if got.Results.Tests[1].Trace != tt.checkTrace {
					t.Errorf("failed test trace = %q, want %q", got.Results.Tests[1].Trace, tt.checkTrace)
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

func TestDurationMs(t *testing.T) {
	tests := []struct {
		name    string
		seconds float64
		want    int64
	}{
		{name: "zero seconds", seconds: 0.0, want: 0},
		{name: "negative seconds", seconds: -1.0, want: 0},
		{name: "one second", seconds: 1.0, want: 1000},
		{name: "half second", seconds: 0.5, want: 500},
		{name: "sub-millisecond rounds to 1ms", seconds: 0.0001, want: 1},
		{name: "exactly one millisecond", seconds: 0.001, want: 1},
		{name: "two milliseconds", seconds: 0.002, want: 2},
		{name: "large duration", seconds: 60.5, want: 60500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := durationMs(tt.seconds)
			if got != tt.want {
				t.Errorf("durationMs(%v) = %d, want %d", tt.seconds, got, tt.want)
			}
		})
	}
}
