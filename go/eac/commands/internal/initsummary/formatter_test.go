package initsummary

import (
	"strings"
	"testing"
	"time"
)

func TestFormatCompact(t *testing.T) {
	s := New("build").
		SetRequest([]string{"a"}, []string{"a", "b", "c"}).
		SetExecutionPlan([][]string{{"a"}, {"b", "c"}}).
		SetExecutionContext("local").
		SetDepmStatus(DepmStatus{Verified: true, Total: 2, Resolved: []string{"b", "c"}}).
		SetDepsStatus(DepsStatus{Verified: true, Required: []string{"go"}, Available: []DepsResult{{Name: "go", Available: true}}}).
		SetOutputDir("out/build/")

	output := FormatCompact(s)

	// Should contain key info in compact form
	if !strings.Contains(output, "1 requested → 3 total (+2 depm)") {
		t.Errorf("FormatCompact missing module summary\n\nGot:\n%s", output)
	}
	if !strings.Contains(output, "Layers: 2") {
		t.Errorf("FormatCompact missing layers\n\nGot:\n%s", output)
	}
	if !strings.Contains(output, "Depm: ✅") {
		t.Errorf("FormatCompact missing depm status\n\nGot:\n%s", output)
	}
	if !strings.Contains(output, "Deps: ✅") {
		t.Errorf("FormatCompact missing deps status\n\nGot:\n%s", output)
	}
	if !strings.Contains(output, "Output: out/build/") {
		t.Errorf("FormatCompact missing output dir\n\nGot:\n%s", output)
	}
}

func TestFormatCompactWithSkipped(t *testing.T) {
	s := New("build").
		SetRequest([]string{"a"}, []string{"a"}).
		SetExecutionContext("local").
		SetDepmStatus(DepmStatus{Skipped: true}).
		SetDepsStatus(DepsStatus{Skipped: true})

	output := FormatCompact(s)

	if !strings.Contains(output, "Depm: ⏭️  skipped") {
		t.Errorf("FormatCompact missing skipped depm\n\nGot:\n%s", output)
	}
	if !strings.Contains(output, "Deps: ⏭️  skipped") {
		t.Errorf("FormatCompact missing skipped deps\n\nGot:\n%s", output)
	}
}

func TestFormatCompactWithIncremental(t *testing.T) {
	s := New("build").
		SetRequest([]string{"a"}, []string{"a"}).
		SetExecutionContext("local").
		SetIncremental(&IncrementalInfo{
			Enabled:       true,
			DetectionTime: 45 * time.Millisecond,
			Changed:       []string{"a"},
			UpToDate:      []string{"b", "c"},
		})

	output := FormatCompact(s)

	if !strings.Contains(output, "Incremental:") {
		t.Errorf("FormatCompact missing incremental\n\nGot:\n%s", output)
	}
	if !strings.Contains(output, "1 changed") {
		t.Errorf("FormatCompact missing changed count\n\nGot:\n%s", output)
	}
	if !strings.Contains(output, "2 up-to-date") {
		t.Errorf("FormatCompact missing up-to-date count\n\nGot:\n%s", output)
	}
}

func TestFormatCompactWithTestInfo(t *testing.T) {
	s := New("test").
		SetRequest([]string{"eac-core"}, []string{"eac-core"}).
		SetExecutionContext("local").
		SetTestInfo(&TestInfo{
			SuiteName:       "unit",
			TotalDiscovered: 100,
			Selected:        50,
		})

	output := FormatCompact(s)

	if !strings.Contains(output, "Suite: unit") {
		t.Errorf("FormatCompact missing suite name\n\nGot:\n%s", output)
	}
	if !strings.Contains(output, "Tests: 50 selected (of 100 discovered)") {
		t.Errorf("FormatCompact missing test counts\n\nGot:\n%s", output)
	}
}

func TestFormatDetailed(t *testing.T) {
	s := New("build").
		SetRequest([]string{"eac-core", "eac-commands"}, []string{"eac-core", "eac-commands", "eac-logging"}).
		SetExecutionPlan([][]string{{"eac-logging"}, {"eac-core", "eac-commands"}}).
		SetFlags(Flags{TidyFirst: true, TidyExplicit: false}).
		SetExecutionContext("local").
		SetDepmStatus(DepmStatus{
			Verified: true,
			Total:    1,
			Resolved: []string{"eac-logging"},
		}).
		SetDepsStatus(DepsStatus{
			Verified: true,
			Required: []string{"go"},
			Available: []DepsResult{
				{Name: "go", Available: true, Version: "1.21.0"},
			},
		}).
		SetOutputDir("out/build/")

	output := FormatDetailed(s)

	// Check for expected sections
	expectedSections := []string{
		"Build Initialization",
		"Execution Context: local",
		"── Modules ──",
		"Requested: 2",
		"Added depm: 1",
		"── Execution Plan ──",
		"Layers: 2",
		"Layer 1:",
		"Layer 2:",
		"── Flags ──",
		"tidy-first",
		"── Module Dependencies (depm) ──",
		"1/1 resolved",
		"── System Dependencies (deps) ──",
		"go (1.21.0)",
		"── Output ──",
		"out/build/",
	}

	for _, expected := range expectedSections {
		if !strings.Contains(output, expected) {
			t.Errorf("FormatDetailed output missing %q\n\nGot:\n%s", expected, output)
		}
	}
}

func TestFormatDetailedWithTestInfo(t *testing.T) {
	s := New("test").
		SetRequest([]string{"eac-core"}, []string{"eac-core"}).
		SetExecutionContext("local").
		SetTestInfo(&TestInfo{
			SuiteName:             "unit",
			SuiteDescription:      "Unit tests",
			TotalDiscovered:       100,
			Skipped:               10,
			NotMatchingSuite:      30,
			OSFiltered:            5,
			Selected:              55,
			InferenceRulesApplied: 3,
		}).
		SetOutputDir("out/test/unit/")

	output := FormatDetailed(s)

	expectedSections := []string{
		"Test Initialization",
		"── Suite ──",
		"Name: unit",
		"Description: Unit tests",
		"── Test Discovery ──",
		"Discovered: 100",
		"Skipped (@skip:*): 10",
		"Not matching suite: 30",
		"OS incompatible: 5",
		"Selected: 55",
		"Inference rules applied: 3",
	}

	for _, expected := range expectedSections {
		if !strings.Contains(output, expected) {
			t.Errorf("FormatDetailed output missing %q\n\nGot:\n%s", expected, output)
		}
	}
}

func TestTruncateList(t *testing.T) {
	tests := []struct {
		name   string
		items  []string
		maxLen int
		want   string
	}{
		{
			name:   "short list",
			items:  []string{"a", "b"},
			maxLen: 20,
			want:   "a, b",
		},
		{
			name:   "truncated list",
			items:  []string{"module-one", "module-two", "module-three"},
			maxLen: 25,
			want:   "module-one, ...",
		},
		{
			name:   "empty list",
			items:  []string{},
			maxLen: 20,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateList(tt.items, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateList() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatLayerSizes(t *testing.T) {
	tests := []struct {
		name  string
		sizes []int
		want  string
	}{
		{
			name:  "single layer",
			sizes: []int{3},
			want:  "3",
		},
		{
			name:  "multiple layers",
			sizes: []int{2, 1, 3},
			want:  "2 → 1 → 3",
		},
		{
			name:  "empty",
			sizes: []int{},
			want:  "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatLayerSizes(tt.sizes)
			if got != tt.want {
				t.Errorf("formatLayerSizes() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatFlagsCompact(t *testing.T) {
	tests := []struct {
		name  string
		flags Flags
		want  string
	}{
		{
			name:  "all defaults",
			flags: Flags{},
			want:  "",
		},
		{
			name:  "skip-depm only",
			flags: Flags{SkipDepm: true},
			want:  "skip-depm",
		},
		{
			name:  "multiple flags",
			flags: Flags{SkipDepm: true, DryRun: true, Version: "v1.0.0"},
			want:  "skip-depm, dry-run, version=v1.0.0",
		},
		{
			name:  "list-only",
			flags: Flags{ListOnly: true},
			want:  "list-only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatFlagsCompact(tt.flags)
			if got != tt.want {
				t.Errorf("formatFlagsCompact() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatFlagsDetailed(t *testing.T) {
	// Build with tidy-first enabled
	flags := Flags{TidyFirst: true, TidyExplicit: true, DryRun: true}
	output := formatFlagsDetailed(flags, "build")

	if !strings.Contains(output, "tidy-first: enabled") {
		t.Errorf("formatFlagsDetailed missing tidy-first\n\nGot:\n%s", output)
	}
	if !strings.Contains(output, "dry-run: enabled") {
		t.Errorf("formatFlagsDetailed missing dry-run\n\nGot:\n%s", output)
	}
}

func TestFormatFlagsDetailedDefaults(t *testing.T) {
	flags := Flags{}
	output := formatFlagsDetailed(flags, "test")

	if !strings.Contains(output, "(defaults)") {
		t.Errorf("formatFlagsDetailed should show (defaults) for empty flags\n\nGot:\n%s", output)
	}
}
