package ctrf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// ---------------------------------------------------------------------------
// Parse
// ---------------------------------------------------------------------------

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Report
		wantErr bool
	}{
		{
			name: "valid minimal report",
			input: `{
				"results": {
					"tool": {"name": "gotest"},
					"summary": {"tests":1,"passed":1,"failed":0,"pending":0,"skipped":0,"other":0,"start":1000,"stop":2000},
					"tests": [{"name":"TestA","status":"passed","duration":100}]
				}
			}`,
			want: &Report{
				Results: Results{
					Tool: Tool{Name: "gotest"},
					Summary: Summary{
						Tests: 1, Passed: 1, Failed: 0,
						Pending: 0, Skipped: 0, Other: 0,
						Start: 1000, Stop: 2000,
					},
					Tests: []Test{
						{Name: "TestA", Status: StatusPassed, Duration: 100},
					},
				},
			},
		},
		{
			name: "report with environment",
			input: `{
				"results": {
					"tool": {"name": "gotest","version":"1.21"},
					"summary": {"tests":0,"passed":0,"failed":0,"pending":0,"skipped":0,"other":0,"start":0,"stop":0},
					"tests": [],
					"environment": {"os":"linux","arch":"amd64","language":"go"}
				}
			}`,
			want: &Report{
				Results: Results{
					Tool:    Tool{Name: "gotest", Version: "1.21"},
					Summary: Summary{},
					Tests:   []Test{},
					Environment: &Environment{
						OS: "linux", Arch: "amd64", Language: "go",
					},
				},
			},
		},
		{
			name: "report with all test statuses",
			input: `{
				"results": {
					"tool": {"name": "t"},
					"summary": {"tests":5,"passed":1,"failed":1,"pending":1,"skipped":1,"other":1,"start":0,"stop":0},
					"tests": [
						{"name":"A","status":"passed","duration":1},
						{"name":"B","status":"failed","duration":2,"message":"boom","trace":"stack"},
						{"name":"C","status":"pending","duration":0},
						{"name":"D","status":"skipped","duration":0},
						{"name":"E","status":"other","duration":0,"rawStatus":"timeout"}
					]
				}
			}`,
			want: &Report{
				Results: Results{
					Tool: Tool{Name: "t"},
					Summary: Summary{
						Tests: 5, Passed: 1, Failed: 1,
						Pending: 1, Skipped: 1, Other: 1,
					},
					Tests: []Test{
						{Name: "A", Status: StatusPassed, Duration: 1},
						{Name: "B", Status: StatusFailed, Duration: 2, Message: "boom", Trace: "stack"},
						{Name: "C", Status: StatusPending, Duration: 0},
						{Name: "D", Status: StatusSkipped, Duration: 0},
						{Name: "E", Status: StatusOther, Duration: 0, RawStatus: "timeout"},
					},
				},
			},
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "malformed JSON",
			input:   "{not json",
			wantErr: true,
		},
		{
			name:    "invalid JSON type for results",
			input:   `{"results": "string"}`,
			wantErr: true,
		},
		{
			name:  "empty JSON object",
			input: `{}`,
			want: &Report{
				Results: Results{},
			},
		},
		{
			name:  "empty results object",
			input: `{"results": {"tool":{"name":""},"summary":{"tests":0,"passed":0,"failed":0,"pending":0,"skipped":0,"other":0,"start":0,"stop":0},"tests":null}}`,
			want: &Report{
				Results: Results{
					Tool:    Tool{Name: ""},
					Summary: Summary{},
					Tests:   nil,
				},
			},
		},
		{
			name: "test with suite and filePath",
			input: `{
				"results": {
					"tool": {"name": "gotest"},
					"summary": {"tests":1,"passed":1,"failed":0,"pending":0,"skipped":0,"other":0,"start":0,"stop":0},
					"tests": [{"name":"TestX","status":"passed","duration":50,"suite":"pkg/foo","filePath":"foo_test.go"}]
				}
			}`,
			want: &Report{
				Results: Results{
					Tool:    Tool{Name: "gotest"},
					Summary: Summary{Tests: 1, Passed: 1},
					Tests: []Test{
						{Name: "TestX", Status: StatusPassed, Duration: 50, Suite: "pkg/foo", FilePath: "foo_test.go"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ParseFile
// ---------------------------------------------------------------------------

func TestParseFile(t *testing.T) {
	t.Run("parses valid file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "report.json")

		report := &Report{
			Results: Results{
				Tool:    Tool{Name: "gotest"},
				Summary: Summary{Tests: 1, Passed: 1, Start: 100, Stop: 200},
				Tests:   []Test{{Name: "TestA", Status: StatusPassed, Duration: 50}},
			},
		}
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal test data: %v", err)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		got, err := ParseFile(path)
		if err != nil {
			t.Fatalf("ParseFile() error = %v", err)
		}
		if !reflect.DeepEqual(got, report) {
			t.Errorf("ParseFile() = %+v, want %+v", got, report)
		}
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		_, err := ParseFile("/nonexistent/path/report.json")
		if err == nil {
			t.Fatal("ParseFile() expected error for nonexistent file, got nil")
		}
	})

	t.Run("returns error for malformed file content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "bad.json")
		if err := os.WriteFile(path, []byte("{bad json}"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := ParseFile(path)
		if err == nil {
			t.Fatal("ParseFile() expected error for malformed JSON, got nil")
		}
	})

	t.Run("returns error for empty file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.json")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := ParseFile(path)
		if err == nil {
			t.Fatal("ParseFile() expected error for empty file, got nil")
		}
	})
}

// ---------------------------------------------------------------------------
// AddTest
// ---------------------------------------------------------------------------

func TestAddTest(t *testing.T) {
	tests := []struct {
		name        string
		addTests    []Test
		wantSummary Summary
		wantCount   int
	}{
		{
			name:        "no tests added",
			addTests:    nil,
			wantSummary: Summary{Start: 1000},
			wantCount:   0,
		},
		{
			name: "single passed test",
			addTests: []Test{
				{Name: "TestA", Status: StatusPassed, Duration: 100},
			},
			wantSummary: Summary{Tests: 1, Passed: 1, Start: 1000},
			wantCount:   1,
		},
		{
			name: "single failed test",
			addTests: []Test{
				{Name: "TestB", Status: StatusFailed, Duration: 200, Message: "assertion failed"},
			},
			wantSummary: Summary{Tests: 1, Failed: 1, Start: 1000},
			wantCount:   1,
		},
		{
			name: "single skipped test",
			addTests: []Test{
				{Name: "TestC", Status: StatusSkipped},
			},
			wantSummary: Summary{Tests: 1, Skipped: 1, Start: 1000},
			wantCount:   1,
		},
		{
			name: "single pending test",
			addTests: []Test{
				{Name: "TestD", Status: StatusPending},
			},
			wantSummary: Summary{Tests: 1, Pending: 1, Start: 1000},
			wantCount:   1,
		},
		{
			name: "unknown status counts as other",
			addTests: []Test{
				{Name: "TestE", Status: "flaky"},
			},
			wantSummary: Summary{Tests: 1, Other: 1, Start: 1000},
			wantCount:   1,
		},
		{
			name: "empty string status counts as other",
			addTests: []Test{
				{Name: "TestF", Status: ""},
			},
			wantSummary: Summary{Tests: 1, Other: 1, Start: 1000},
			wantCount:   1,
		},
		{
			name: "mixed statuses accumulate correctly",
			addTests: []Test{
				{Name: "T1", Status: StatusPassed},
				{Name: "T2", Status: StatusPassed},
				{Name: "T3", Status: StatusFailed},
				{Name: "T4", Status: StatusSkipped},
				{Name: "T5", Status: StatusPending},
				{Name: "T6", Status: "other"},
			},
			wantSummary: Summary{Tests: 6, Passed: 2, Failed: 1, Skipped: 1, Pending: 1, Other: 1, Start: 1000},
			wantCount:   6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Report{
				Results: Results{
					Tool:    Tool{Name: "test"},
					Tests:   make([]Test, 0),
					Summary: Summary{Start: 1000},
				},
			}

			for _, test := range tt.addTests {
				r.AddTest(test)
			}

			if r.Results.Summary != tt.wantSummary {
				t.Errorf("Summary = %+v, want %+v", r.Results.Summary, tt.wantSummary)
			}
			if len(r.Results.Tests) != tt.wantCount {
				t.Errorf("len(Tests) = %d, want %d", len(r.Results.Tests), tt.wantCount)
			}
			// Verify tests are appended in order
			for i, test := range tt.addTests {
				if i < len(r.Results.Tests) && r.Results.Tests[i].Name != test.Name {
					t.Errorf("Tests[%d].Name = %q, want %q", i, r.Results.Tests[i].Name, test.Name)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Merge
// ---------------------------------------------------------------------------

func TestMerge(t *testing.T) {
	tests := []struct {
		name        string
		base        *Report
		other       *Report
		wantSummary Summary
		wantCount   int
	}{
		{
			name: "merge empty into empty",
			base: &Report{
				Results: Results{
					Tool:    Tool{Name: "t"},
					Tests:   make([]Test, 0),
					Summary: Summary{Start: 5000, Stop: 6000},
				},
			},
			other: &Report{
				Results: Results{
					Tool:    Tool{Name: "t"},
					Tests:   make([]Test, 0),
					Summary: Summary{Start: 7000, Stop: 8000},
				},
			},
			wantSummary: Summary{Start: 5000, Stop: 8000},
			wantCount:   0,
		},
		{
			name: "merge tests into empty base",
			base: &Report{
				Results: Results{
					Tool:    Tool{Name: "t"},
					Tests:   make([]Test, 0),
					Summary: Summary{Start: 5000, Stop: 6000},
				},
			},
			other: &Report{
				Results: Results{
					Tool: Tool{Name: "t"},
					Tests: []Test{
						{Name: "T1", Status: StatusPassed, Duration: 100},
						{Name: "T2", Status: StatusFailed, Duration: 200},
					},
					Summary: Summary{Tests: 2, Passed: 1, Failed: 1, Start: 4000, Stop: 7000},
				},
			},
			wantSummary: Summary{Tests: 2, Passed: 1, Failed: 1, Start: 4000, Stop: 7000},
			wantCount:   2,
		},
		{
			name: "merge combines test counts",
			base: &Report{
				Results: Results{
					Tool: Tool{Name: "t"},
					Tests: []Test{
						{Name: "T1", Status: StatusPassed},
					},
					Summary: Summary{Tests: 1, Passed: 1, Start: 1000, Stop: 2000},
				},
			},
			other: &Report{
				Results: Results{
					Tool: Tool{Name: "t"},
					Tests: []Test{
						{Name: "T2", Status: StatusFailed},
						{Name: "T3", Status: StatusSkipped},
					},
					Summary: Summary{Tests: 2, Failed: 1, Skipped: 1, Start: 3000, Stop: 5000},
				},
			},
			// base already has Passed=1, Tests=1 from its own initial state.
			// Merge calls AddTest for T2 (failed) and T3 (skipped), adding to the base counts.
			wantSummary: Summary{Tests: 3, Passed: 1, Failed: 1, Skipped: 1, Start: 1000, Stop: 5000},
			wantCount:   3,
		},
		{
			name: "merge picks earliest start time",
			base: &Report{
				Results: Results{
					Tool:    Tool{Name: "t"},
					Tests:   make([]Test, 0),
					Summary: Summary{Start: 9000, Stop: 10000},
				},
			},
			other: &Report{
				Results: Results{
					Tool:    Tool{Name: "t"},
					Tests:   make([]Test, 0),
					Summary: Summary{Start: 1000, Stop: 2000},
				},
			},
			wantSummary: Summary{Start: 1000, Stop: 10000},
			wantCount:   0,
		},
		{
			name: "merge picks latest stop time",
			base: &Report{
				Results: Results{
					Tool:    Tool{Name: "t"},
					Tests:   make([]Test, 0),
					Summary: Summary{Start: 1000, Stop: 2000},
				},
			},
			other: &Report{
				Results: Results{
					Tool:    Tool{Name: "t"},
					Tests:   make([]Test, 0),
					Summary: Summary{Start: 3000, Stop: 50000},
				},
			},
			wantSummary: Summary{Start: 1000, Stop: 50000},
			wantCount:   0,
		},
		{
			name: "merge with other having no tests but earlier start",
			base: &Report{
				Results: Results{
					Tool: Tool{Name: "t"},
					Tests: []Test{
						{Name: "T1", Status: StatusPassed},
					},
					Summary: Summary{Tests: 1, Passed: 1, Start: 5000, Stop: 6000},
				},
			},
			other: &Report{
				Results: Results{
					Tool:    Tool{Name: "t"},
					Tests:   []Test{},
					Summary: Summary{Start: 1000, Stop: 2000},
				},
			},
			wantSummary: Summary{Tests: 1, Passed: 1, Start: 1000, Stop: 6000},
			wantCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.base.Merge(tt.other)

			if tt.base.Results.Summary != tt.wantSummary {
				t.Errorf("Summary after Merge = %+v, want %+v", tt.base.Results.Summary, tt.wantSummary)
			}
			if len(tt.base.Results.Tests) != tt.wantCount {
				t.Errorf("len(Tests) = %d, want %d", len(tt.base.Results.Tests), tt.wantCount)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// NewReport
// ---------------------------------------------------------------------------

func TestNewReport(t *testing.T) {
	r := NewReport("gotest")

	if r.Results.Tool.Name != "gotest" {
		t.Errorf("Tool.Name = %q, want %q", r.Results.Tool.Name, "gotest")
	}
	if r.Results.Tests == nil {
		t.Fatal("Tests slice should be initialized, got nil")
	}
	if len(r.Results.Tests) != 0 {
		t.Errorf("len(Tests) = %d, want 0", len(r.Results.Tests))
	}
	if r.Results.Summary.Start <= 0 {
		t.Errorf("Start should be set to current time in millis, got %d", r.Results.Summary.Start)
	}
	if r.Results.Summary.Stop != 0 {
		t.Errorf("Stop should be 0 before Finalize, got %d", r.Results.Summary.Stop)
	}
}

// ---------------------------------------------------------------------------
// NewEmptyReport
// ---------------------------------------------------------------------------

func TestNewEmptyReport(t *testing.T) {
	r := NewEmptyReport("aggregator")

	if r.Results.Tool.Name != "aggregator" {
		t.Errorf("Tool.Name = %q, want %q", r.Results.Tool.Name, "aggregator")
	}
	if r.Results.Tests == nil {
		t.Fatal("Tests slice should be initialized, got nil")
	}
	if len(r.Results.Tests) != 0 {
		t.Errorf("len(Tests) = %d, want 0", len(r.Results.Tests))
	}
	// Start should be set to a very large sentinel value
	expectedStart := int64(1<<62 - 1)
	if r.Results.Summary.Start != expectedStart {
		t.Errorf("Start = %d, want %d (sentinel max)", r.Results.Summary.Start, expectedStart)
	}
	if r.Results.Summary.Stop != 0 {
		t.Errorf("Stop = %d, want 0 (sentinel min)", r.Results.Summary.Stop)
	}
}

// ---------------------------------------------------------------------------
// NewEmptyReport + Merge (aggregation pattern)
// ---------------------------------------------------------------------------

func TestNewEmptyReport_MergeUpdatesTimeBounds(t *testing.T) {
	agg := NewEmptyReport("agg")

	r1 := &Report{
		Results: Results{
			Tool: Tool{Name: "t"},
			Tests: []Test{
				{Name: "T1", Status: StatusPassed},
			},
			Summary: Summary{Tests: 1, Passed: 1, Start: 3000, Stop: 4000},
		},
	}
	r2 := &Report{
		Results: Results{
			Tool: Tool{Name: "t"},
			Tests: []Test{
				{Name: "T2", Status: StatusFailed},
			},
			Summary: Summary{Tests: 1, Failed: 1, Start: 1000, Stop: 5000},
		},
	}

	agg.Merge(r1)
	agg.Merge(r2)

	if agg.Results.Summary.Start != 1000 {
		t.Errorf("Start = %d, want 1000 (earliest)", agg.Results.Summary.Start)
	}
	if agg.Results.Summary.Stop != 5000 {
		t.Errorf("Stop = %d, want 5000 (latest)", agg.Results.Summary.Stop)
	}
	if agg.Results.Summary.Tests != 2 {
		t.Errorf("Tests = %d, want 2", agg.Results.Summary.Tests)
	}
	if agg.Results.Summary.Passed != 1 {
		t.Errorf("Passed = %d, want 1", agg.Results.Summary.Passed)
	}
	if agg.Results.Summary.Failed != 1 {
		t.Errorf("Failed = %d, want 1", agg.Results.Summary.Failed)
	}
}

// ---------------------------------------------------------------------------
// SetTimes
// ---------------------------------------------------------------------------

func TestSetTimes(t *testing.T) {
	r := NewReport("t")
	r.SetTimes(42, 99)

	if r.Results.Summary.Start != 42 {
		t.Errorf("Start = %d, want 42", r.Results.Summary.Start)
	}
	if r.Results.Summary.Stop != 99 {
		t.Errorf("Stop = %d, want 99", r.Results.Summary.Stop)
	}
}

// ---------------------------------------------------------------------------
// Finalize
// ---------------------------------------------------------------------------

func TestFinalize(t *testing.T) {
	r := NewReport("t")
	r.Finalize()

	if r.Results.Summary.Stop <= 0 {
		t.Errorf("Stop should be set to current time after Finalize, got %d", r.Results.Summary.Stop)
	}
	if r.Results.Summary.Stop < r.Results.Summary.Start {
		t.Errorf("Stop (%d) should be >= Start (%d)", r.Results.Summary.Stop, r.Results.Summary.Start)
	}
}

// ---------------------------------------------------------------------------
// ToJSON
// ---------------------------------------------------------------------------

func TestToJSON(t *testing.T) {
	t.Run("round-trip produces equivalent report", func(t *testing.T) {
		original := &Report{
			Results: Results{
				Tool:    Tool{Name: "gotest", Version: "1.21"},
				Summary: Summary{Tests: 1, Passed: 1, Start: 1000, Stop: 2000},
				Tests: []Test{
					{Name: "TestA", Status: StatusPassed, Duration: 150},
				},
				Environment: &Environment{OS: "linux", Arch: "amd64", Language: "go"},
			},
		}

		data, err := original.ToJSON()
		if err != nil {
			t.Fatalf("ToJSON() error = %v", err)
		}

		parsed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parse() round-trip error = %v", err)
		}

		if !reflect.DeepEqual(original, parsed) {
			t.Errorf("round-trip mismatch:\n  got:  %+v\n  want: %+v", parsed, original)
		}
	})

	t.Run("empty report serializes without error", func(t *testing.T) {
		r := NewReport("t")
		r.SetTimes(0, 0)

		data, err := r.ToJSON()
		if err != nil {
			t.Fatalf("ToJSON() error = %v", err)
		}
		if len(data) == 0 {
			t.Error("ToJSON() returned empty bytes")
		}
	})

	t.Run("output is valid JSON", func(t *testing.T) {
		r := NewReport("t")
		r.AddTest(Test{Name: "T1", Status: StatusPassed})
		r.SetTimes(100, 200)

		data, err := r.ToJSON()
		if err != nil {
			t.Fatalf("ToJSON() error = %v", err)
		}

		if !json.Valid(data) {
			t.Error("ToJSON() output is not valid JSON")
		}
	})
}

// ---------------------------------------------------------------------------
// Status constants
// ---------------------------------------------------------------------------

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusPassed, "passed"},
		{StatusFailed, "failed"},
		{StatusSkipped, "skipped"},
		{StatusPending, "pending"},
		{StatusOther, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("Status = %q, want %q", tt.status, tt.want)
			}
		})
	}
}
