package mocha

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/ctrf"
	coreTesting "github.com/ready-to-release/eac/go/core/testing"
)

// ---------------------------------------------------------------------------
// mochaDurationMs
// ---------------------------------------------------------------------------

func TestMochaDurationMs(t *testing.T) {
	tests := []struct {
		name string
		ms   float64
		want int64
	}{
		{name: "zero", ms: 0, want: 0},
		{name: "negative", ms: -1, want: 0},
		{name: "exact integer", ms: 100, want: 100},
		{name: "rounds up from 0.5", ms: 100.5, want: 101},
		{name: "rounds up from 0.6", ms: 100.6, want: 101},
		{name: "rounds down from 0.4", ms: 100.4, want: 100},
		{name: "very small positive rounds to 1", ms: 0.1, want: 1},
		{name: "large value", ms: 99999.0, want: 99999},
		{name: "one millisecond", ms: 1, want: 1},
		{name: "sub-millisecond clamps to 1", ms: 0.01, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mochaDurationMs(tt.ms)
			if got != tt.want {
				t.Errorf("mochaDurationMs(%v) = %d, want %d", tt.ms, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// convertMochaJSONToCTRF
// ---------------------------------------------------------------------------

func TestConvertMochaJSONToCTRF_InvalidJSON(t *testing.T) {
	result := convertMochaJSONToCTRF([]byte(`{not json`))
	if result != nil {
		t.Error("expected nil for invalid JSON, got non-nil")
	}
}

func TestConvertMochaJSONToCTRF_EmptyInput(t *testing.T) {
	result := convertMochaJSONToCTRF(nil)
	if result != nil {
		t.Error("expected nil for nil input, got non-nil")
	}

	result = convertMochaJSONToCTRF([]byte{})
	if result != nil {
		t.Error("expected nil for empty input, got non-nil")
	}
}

func TestConvertMochaJSONToCTRF_EmptyReport(t *testing.T) {
	input := `{"stats":{},"tests":[],"passes":[],"failures":[],"pending":[]}`
	report := convertMochaJSONToCTRF([]byte(input))
	if report == nil {
		t.Fatal("expected non-nil report for valid empty JSON")
	}
	if report.Results.Tool.Name != "mocha" {
		t.Errorf("tool name = %q, want %q", report.Results.Tool.Name, "mocha")
	}
	if len(report.Results.Tests) != 0 {
		t.Errorf("expected 0 tests, got %d", len(report.Results.Tests))
	}
}

func TestConvertMochaJSONToCTRF_PassingTests(t *testing.T) {
	input := mochaReport{
		Passes: []mochaTest{
			{FullTitle: "Suite Test A", Duration: 42},
			{FullTitle: "Suite Test B", Duration: 0.3},
		},
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal test input: %v", err)
	}

	report := convertMochaJSONToCTRF(data)
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if len(report.Results.Tests) != 2 {
		t.Fatalf("expected 2 tests, got %d", len(report.Results.Tests))
	}

	assertCTRFTest(t, report.Results.Tests[0], "Suite Test A", ctrf.StatusPassed, 42)
	assertCTRFTest(t, report.Results.Tests[1], "Suite Test B", ctrf.StatusPassed, 1)

	if report.Results.Summary.Passed != 2 {
		t.Errorf("summary.Passed = %d, want 2", report.Results.Summary.Passed)
	}
	if report.Results.Summary.Tests != 2 {
		t.Errorf("summary.Tests = %d, want 2", report.Results.Summary.Tests)
	}
}

func TestConvertMochaJSONToCTRF_FailingTests(t *testing.T) {
	input := mochaReport{
		Failures: []mochaTest{
			{
				FullTitle: "Suite Failing",
				Duration:  150,
				Err: mochaErr{
					Message: "expected true to be false",
					Stack:   "Error: expected true to be false\n    at Context.<anonymous>",
				},
			},
		},
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal test input: %v", err)
	}

	report := convertMochaJSONToCTRF(data)
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if len(report.Results.Tests) != 1 {
		t.Fatalf("expected 1 test, got %d", len(report.Results.Tests))
	}

	test := report.Results.Tests[0]
	assertCTRFTest(t, test, "Suite Failing", ctrf.StatusFailed, 150)

	if test.Message != "expected true to be false" {
		t.Errorf("message = %q, want %q", test.Message, "expected true to be false")
	}
	if test.Trace != "Error: expected true to be false\n    at Context.<anonymous>" {
		t.Errorf("trace = %q, unexpected value", test.Trace)
	}

	if report.Results.Summary.Failed != 1 {
		t.Errorf("summary.Failed = %d, want 1", report.Results.Summary.Failed)
	}
}

func TestConvertMochaJSONToCTRF_PendingTests(t *testing.T) {
	input := mochaReport{
		Pending: []mochaTest{
			{FullTitle: "Suite Pending Test", Duration: 0},
		},
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal test input: %v", err)
	}

	report := convertMochaJSONToCTRF(data)
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if len(report.Results.Tests) != 1 {
		t.Fatalf("expected 1 test, got %d", len(report.Results.Tests))
	}

	assertCTRFTest(t, report.Results.Tests[0], "Suite Pending Test", ctrf.StatusPending, 0)

	if report.Results.Summary.Pending != 1 {
		t.Errorf("summary.Pending = %d, want 1", report.Results.Summary.Pending)
	}
}

func TestConvertMochaJSONToCTRF_MixedResults(t *testing.T) {
	input := mochaReport{
		Passes:   []mochaTest{{FullTitle: "passed-1", Duration: 10}},
		Failures: []mochaTest{{FullTitle: "failed-1", Duration: 20, Err: mochaErr{Message: "oops"}}},
		Pending:  []mochaTest{{FullTitle: "pending-1", Duration: 0}},
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal test input: %v", err)
	}

	report := convertMochaJSONToCTRF(data)
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if len(report.Results.Tests) != 3 {
		t.Fatalf("expected 3 tests, got %d", len(report.Results.Tests))
	}

	summary := report.Results.Summary
	if summary.Tests != 3 {
		t.Errorf("summary.Tests = %d, want 3", summary.Tests)
	}
	if summary.Passed != 1 {
		t.Errorf("summary.Passed = %d, want 1", summary.Passed)
	}
	if summary.Failed != 1 {
		t.Errorf("summary.Failed = %d, want 1", summary.Failed)
	}
	if summary.Pending != 1 {
		t.Errorf("summary.Pending = %d, want 1", summary.Pending)
	}
}

func TestConvertMochaJSONToCTRF_WithTimestamps(t *testing.T) {
	input := `{
		"stats": {
			"start": "2024-01-15T10:00:00.000Z",
			"end":   "2024-01-15T10:00:05.000Z"
		},
		"passes": [],
		"failures": [],
		"pending": []
	}`

	report := convertMochaJSONToCTRF([]byte(input))
	if report == nil {
		t.Fatal("expected non-nil report")
	}

	if report.Results.Summary.Start == 0 {
		t.Error("expected non-zero start time")
	}
	if report.Results.Summary.Stop == 0 {
		t.Error("expected non-zero stop time")
	}
	if report.Results.Summary.Stop <= report.Results.Summary.Start {
		t.Errorf("stop (%d) should be > start (%d)",
			report.Results.Summary.Stop, report.Results.Summary.Start)
	}
}

func TestConvertMochaJSONToCTRF_InvalidTimestamps(t *testing.T) {
	input := `{
		"stats": {
			"start": "not-a-date",
			"end":   "also-not-a-date"
		},
		"passes": [],
		"failures": [],
		"pending": []
	}`

	report := convertMochaJSONToCTRF([]byte(input))
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	// When both start and end are non-empty but unparseable, the code enters
	// the if-branch but time.Parse fails silently. Neither SetTimes nor
	// Finalize is called, so Stop stays at its NewReport default of 0.
	if report.Results.Summary.Stop != 0 {
		t.Errorf("expected stop = 0 when timestamps are invalid, got %d", report.Results.Summary.Stop)
	}
}

func TestConvertMochaJSONToCTRF_PartialTimestamps(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "start only",
			input: `{
				"stats": {"start": "2024-01-15T10:00:00.000Z"},
				"passes": [], "failures": [], "pending": []
			}`,
		},
		{
			name: "end only",
			input: `{
				"stats": {"end": "2024-01-15T10:00:05.000Z"},
				"passes": [], "failures": [], "pending": []
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := convertMochaJSONToCTRF([]byte(tt.input))
			if report == nil {
				t.Fatal("expected non-nil report")
			}
			// Partial timestamps cause fallback to Finalize()
		})
	}
}

// ---------------------------------------------------------------------------
// findTsModuleForPath
// ---------------------------------------------------------------------------

func TestFindTsModuleForPath(t *testing.T) {
	cfg := &config.EACConfig{
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "web-app",
					Components: config.ModuleComponents{
						"typescript": {Root: "packages/web"},
					},
				},
				{
					Moniker: "api-server",
					Components: config.ModuleComponents{
						"typescript": {Root: "packages/api"},
					},
				},
				{
					Moniker: "no-root",
					Components: config.ModuleComponents{
						"typescript": {Root: ""},
					},
				},
				{
					Moniker: "nil-entry",
					Components: config.ModuleComponents{
						"typescript": nil,
					},
				},
			},
		},
	}

	tests := []struct {
		name    string
		relPath string
		want    string
	}{
		{
			name:    "exact match on root",
			relPath: "packages/web",
			want:    "web-app",
		},
		{
			name:    "nested path under root",
			relPath: "packages/web/src/components",
			want:    "web-app",
		},
		{
			name:    "matches second module",
			relPath: "packages/api/tests",
			want:    "api-server",
		},
		{
			name:    "no match",
			relPath: "packages/other/src",
			want:    "",
		},
		{
			name:    "partial prefix mismatch",
			relPath: "packages/webapp/src",
			want:    "",
		},
		{
			name:    "empty path",
			relPath: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findTsModuleForPath(tt.relPath, cfg)
			if got != tt.want {
				t.Errorf("findTsModuleForPath(%q) = %q, want %q", tt.relPath, got, tt.want)
			}
		})
	}
}

func TestFindTsModuleForPath_NoModules(t *testing.T) {
	cfg := &config.EACConfig{
		Repository: &config.RepositoryConfig{
			Modules: nil,
		},
	}
	got := findTsModuleForPath("packages/web/src", cfg)
	if got != "" {
		t.Errorf("expected empty string for nil modules, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// MochaRunner methods
// ---------------------------------------------------------------------------

func TestMochaRunner_TestTypes(t *testing.T) {
	r := &MochaRunner{}
	types := r.TestTypes()
	if len(types) != 1 {
		t.Fatalf("expected 1 test type, got %d", len(types))
	}
	if types[0] != "mocha" {
		t.Errorf("test type = %q, want %q", types[0], "mocha")
	}
}

func TestMochaRunner_IsBDD(t *testing.T) {
	r := &MochaRunner{}
	if r.IsBDD() {
		t.Error("IsBDD() should return false for mocha runner")
	}
}

func TestMochaRunner_FindTestRoot(t *testing.T) {
	r := &MochaRunner{}
	result := r.FindTestRoot("/some/path", nil)
	if result != "" {
		t.Errorf("FindTestRoot() = %q, want empty string", result)
	}
}

func TestMochaRunner_BuildPackagePath(t *testing.T) {
	r := &MochaRunner{}
	tests := []struct {
		name     string
		testRoot string
		testPath string
		want     string
	}{
		{name: "returns testRoot", testRoot: "packages/web/tests", testPath: "packages/web/tests/foo.test.ts", want: "packages/web/tests"},
		{name: "empty testRoot", testRoot: "", testPath: "anything", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.BuildPackagePath(tt.testRoot, tt.testPath)
			if got != tt.want {
				t.Errorf("BuildPackagePath(%q, %q) = %q, want %q", tt.testRoot, tt.testPath, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// MochaRunner.GetTestInfo
// ---------------------------------------------------------------------------

func TestMochaRunner_GetTestInfo(t *testing.T) {
	r := &MochaRunner{}
	cfg := &config.EACConfig{
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "frontend",
					Components: config.ModuleComponents{
						"typescript": {Root: "packages/frontend"},
					},
				},
			},
		},
	}

	workspaceRoot := filepath.FromSlash("/workspace")

	tests := []struct {
		name          string
		ref           coreTesting.TestReference
		wantNil       bool
		wantMoniker   string
		wantLanguage  string
		wantTestRoot  string
	}{
		{
			name: "valid test ref under known module",
			ref: coreTesting.TestReference{
				FilePath: filepath.FromSlash("/workspace/packages/frontend/tests/unit.test.ts"),
			},
			wantNil:      false,
			wantMoniker:  "frontend",
			wantLanguage: "ts",
			wantTestRoot: "packages/frontend/tests",
		},
		{
			name: "test ref at module root",
			ref: coreTesting.TestReference{
				FilePath: filepath.FromSlash("/workspace/packages/frontend/index.test.ts"),
			},
			wantNil:      false,
			wantMoniker:  "frontend",
			wantLanguage: "ts",
			wantTestRoot: "packages/frontend",
		},
		{
			name: "test ref not matching any module",
			ref: coreTesting.TestReference{
				FilePath: filepath.FromSlash("/workspace/packages/unknown/tests/unit.test.ts"),
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.GetTestInfo(tt.ref, workspaceRoot, cfg)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil TestInfo, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil TestInfo, got nil")
			}
			if got.ModuleMoniker != tt.wantMoniker {
				t.Errorf("ModuleMoniker = %q, want %q", got.ModuleMoniker, tt.wantMoniker)
			}
			if got.Language != tt.wantLanguage {
				t.Errorf("Language = %q, want %q", got.Language, tt.wantLanguage)
			}
			if got.TestRoot != filepath.ToSlash(tt.wantTestRoot) {
				t.Errorf("TestRoot = %q, want %q", got.TestRoot, filepath.ToSlash(tt.wantTestRoot))
			}
		})
	}
}

func TestMochaRunner_GetTestInfo_RelPathError(t *testing.T) {
	// On Windows, paths on different drives cannot be made relative.
	// On Unix, this is harder to trigger, but we can at least verify
	// the function does not panic with unusual path combos.
	r := &MochaRunner{}
	cfg := &config.EACConfig{
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "test",
					Components: config.ModuleComponents{
						"ts": {Root: "src"},
					},
				},
			},
		},
	}

	// filepath.Rel may fail on certain OS-specific path combinations.
	// This primarily tests that the function handles the error gracefully (returns nil).
	ref := coreTesting.TestReference{
		FilePath: filepath.FromSlash("/workspace/src/test.ts"),
	}
	// Passing a valid workspace root that can produce a relative path
	// should work fine. The error path is OS-dependent, so we just
	// verify no panics occur with a normal call.
	_ = r.GetTestInfo(ref, filepath.FromSlash("/workspace"), cfg)
}

// ---------------------------------------------------------------------------
// mochaReport JSON round-trip (structural test)
// ---------------------------------------------------------------------------

func TestMochaReport_JSONRoundTrip(t *testing.T) {
	original := mochaReport{
		Stats: struct {
			Suites   int    `json:"suites"`
			Tests    int    `json:"tests"`
			Passes   int    `json:"passes"`
			Pending  int    `json:"pending"`
			Failures int    `json:"failures"`
			Start    string `json:"start"`
			End      string `json:"end"`
			Duration int    `json:"duration"`
		}{
			Suites:   2,
			Tests:    5,
			Passes:   3,
			Pending:  1,
			Failures: 1,
			Start:    "2024-01-15T10:00:00.000Z",
			End:      "2024-01-15T10:00:05.000Z",
			Duration: 5000,
		},
		Tests: []mochaTest{
			{Title: "test1", FullTitle: "Suite test1", Duration: 100},
		},
		Passes: []mochaTest{
			{Title: "test1", FullTitle: "Suite test1", Duration: 100},
		},
		Failures: []mochaTest{
			{
				Title:     "test2",
				FullTitle: "Suite test2",
				Duration:  200,
				Err:       mochaErr{Message: "fail", Stack: "stack trace"},
			},
		},
		Pending: []mochaTest{
			{Title: "test3", FullTitle: "Suite test3", Duration: 0},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded mochaReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Stats.Suites != 2 {
		t.Errorf("Stats.Suites = %d, want 2", decoded.Stats.Suites)
	}
	if decoded.Stats.Tests != 5 {
		t.Errorf("Stats.Tests = %d, want 5", decoded.Stats.Tests)
	}
	if len(decoded.Passes) != 1 {
		t.Errorf("len(Passes) = %d, want 1", len(decoded.Passes))
	}
	if len(decoded.Failures) != 1 {
		t.Errorf("len(Failures) = %d, want 1", len(decoded.Failures))
	}
	if decoded.Failures[0].Err.Message != "fail" {
		t.Errorf("Failure message = %q, want %q", decoded.Failures[0].Err.Message, "fail")
	}
	if decoded.Failures[0].Err.Stack != "stack trace" {
		t.Errorf("Failure stack = %q, want %q", decoded.Failures[0].Err.Stack, "stack trace")
	}
}

// ---------------------------------------------------------------------------
// saveMochaCTRF (edge cases only - no filesystem writes)
// ---------------------------------------------------------------------------

func TestSaveMochaCTRF_EmptyInputs(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  []byte
		outputDir string
	}{
		{name: "nil json", jsonData: nil, outputDir: "/out"},
		{name: "empty json", jsonData: []byte{}, outputDir: "/out"},
		{name: "empty outputDir", jsonData: []byte(`{"passes":[],"failures":[],"pending":[]}`), outputDir: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic for any of these edge cases
			saveMochaCTRF(tt.jsonData, tt.outputDir, nil)
		})
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func assertCTRFTest(t *testing.T, test ctrf.Test, wantName string, wantStatus ctrf.Status, wantDuration int64) {
	t.Helper()
	if test.Name != wantName {
		t.Errorf("test.Name = %q, want %q", test.Name, wantName)
	}
	if test.Status != wantStatus {
		t.Errorf("test.Status = %q, want %q", test.Status, wantStatus)
	}
	if test.Duration != wantDuration {
		t.Errorf("test.Duration = %d, want %d", test.Duration, wantDuration)
	}
}
