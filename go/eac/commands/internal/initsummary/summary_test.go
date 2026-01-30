//go:build L1

package initsummary

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{"build command", "build"},
		{"test command", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.command)

			if s.Command != tt.command {
				t.Errorf("Command = %q, want %q", s.Command, tt.command)
			}
			if s.RequestedModules == nil {
				t.Error("RequestedModules should be initialized")
			}
			if s.CalculatedModules == nil {
				t.Error("CalculatedModules should be initialized")
			}
			if s.AddedDepm == nil {
				t.Error("AddedDepm should be initialized")
			}
			if s.ExecutionLayers == nil {
				t.Error("ExecutionLayers should be initialized")
			}
		})
	}
}

func TestSetRequest(t *testing.T) {
	tests := []struct {
		name          string
		requested     []string
		calculated    []string
		wantAddedDepm []string
		wantHasAdded  bool
	}{
		{
			name:          "no dependencies added",
			requested:     []string{"a", "b"},
			calculated:    []string{"a", "b"},
			wantAddedDepm: []string{},
			wantHasAdded:  false,
		},
		{
			name:          "dependencies added",
			requested:     []string{"a"},
			calculated:    []string{"a", "b", "c"},
			wantAddedDepm: []string{"b", "c"},
			wantHasAdded:  true,
		},
		{
			name:          "all requested modules",
			requested:     []string{"a", "b", "c"},
			calculated:    []string{"a", "b", "c"},
			wantAddedDepm: []string{},
			wantHasAdded:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New("build").SetRequest(tt.requested, tt.calculated)

			if len(s.RequestedModules) != len(tt.requested) {
				t.Errorf("RequestedModules length = %d, want %d",
					len(s.RequestedModules), len(tt.requested))
			}
			if len(s.CalculatedModules) != len(tt.calculated) {
				t.Errorf("CalculatedModules length = %d, want %d",
					len(s.CalculatedModules), len(tt.calculated))
			}
			if len(s.AddedDepm) != len(tt.wantAddedDepm) {
				t.Errorf("AddedDepm length = %d, want %d",
					len(s.AddedDepm), len(tt.wantAddedDepm))
			}
			if s.HasDepmAdded() != tt.wantHasAdded {
				t.Errorf("HasDepmAdded() = %v, want %v",
					s.HasDepmAdded(), tt.wantHasAdded)
			}
		})
	}
}

func TestSetExecutionPlan(t *testing.T) {
	tests := []struct {
		name      string
		layers    [][]string
		wantCount int
		wantSizes []int
	}{
		{
			name:      "single layer",
			layers:    [][]string{{"a", "b"}},
			wantCount: 1,
			wantSizes: []int{2},
		},
		{
			name:      "multiple layers",
			layers:    [][]string{{"a"}, {"b", "c"}, {"d"}},
			wantCount: 3,
			wantSizes: []int{1, 2, 1},
		},
		{
			name:      "empty layers",
			layers:    [][]string{},
			wantCount: 0,
			wantSizes: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New("build").SetExecutionPlan(tt.layers)

			if s.LayerCount != tt.wantCount {
				t.Errorf("LayerCount = %d, want %d", s.LayerCount, tt.wantCount)
			}

			sizes := s.LayerSizes()
			if len(sizes) != len(tt.wantSizes) {
				t.Errorf("LayerSizes length = %d, want %d", len(sizes), len(tt.wantSizes))
			}
			for i, size := range sizes {
				if size != tt.wantSizes[i] {
					t.Errorf("LayerSizes[%d] = %d, want %d", i, size, tt.wantSizes[i])
				}
			}
		})
	}
}

func TestTotalModules(t *testing.T) {
	s := New("build").SetRequest(
		[]string{"a"},
		[]string{"a", "b", "c"},
	)

	if got := s.TotalModules(); got != 3 {
		t.Errorf("TotalModules() = %d, want 3", got)
	}
}

func TestHasIncremental(t *testing.T) {
	tests := []struct {
		name        string
		incremental *IncrementalInfo
		want        bool
	}{
		{
			name:        "nil incremental",
			incremental: nil,
			want:        false,
		},
		{
			name:        "with incremental",
			incremental: &IncrementalInfo{Enabled: true},
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New("build").SetIncremental(tt.incremental)

			if got := s.HasIncremental(); got != tt.want {
				t.Errorf("HasIncremental() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasTestInfo(t *testing.T) {
	tests := []struct {
		name     string
		testInfo *TestInfo
		want     bool
	}{
		{
			name:     "nil test info",
			testInfo: nil,
			want:     false,
		},
		{
			name:     "with test info",
			testInfo: &TestInfo{SuiteName: "unit"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New("test").SetTestInfo(tt.testInfo)

			if got := s.HasTestInfo(); got != tt.want {
				t.Errorf("HasTestInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasArtifactValidation(t *testing.T) {
	tests := []struct {
		name       string
		validation *ArtifactValidationInfo
		want       bool
	}{
		{
			name:       "nil validation",
			validation: nil,
			want:       false,
		},
		{
			name:       "with validation",
			validation: &ArtifactValidationInfo{Validated: true},
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New("test").SetArtifactValidation(tt.validation)

			if got := s.HasArtifactValidation(); got != tt.want {
				t.Errorf("HasArtifactValidation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBuildAndIsTest(t *testing.T) {
	buildSummary := New("build")
	testSummary := New("test")

	if !buildSummary.IsBuild() {
		t.Error("build summary should return true for IsBuild()")
	}
	if buildSummary.IsTest() {
		t.Error("build summary should return false for IsTest()")
	}
	if !testSummary.IsTest() {
		t.Error("test summary should return true for IsTest()")
	}
	if testSummary.IsBuild() {
		t.Error("test summary should return false for IsBuild()")
	}
}

func TestBuilderChaining(t *testing.T) {
	// Test that all builder methods return *Summary for chaining
	s := New("build").
		SetRequest([]string{"a"}, []string{"a", "b"}).
		SetExecutionPlan([][]string{{"a"}, {"b"}}).
		SetFlags(Flags{TidyFirst: true}).
		SetExecutionContext("local").
		SetDepmStatus(DepmStatus{Verified: true, Total: 1, Resolved: []string{"b"}}).
		SetDepsStatus(DepsStatus{Verified: true, Required: []string{"go"}}).
		SetIncremental(&IncrementalInfo{Enabled: true, DetectionTime: 50 * time.Millisecond}).
		SetOutputDir("out/build/")

	// Verify chain worked
	if s.Command != "build" {
		t.Error("Chain broken: Command not set")
	}
	if len(s.RequestedModules) != 1 {
		t.Error("Chain broken: RequestedModules not set")
	}
	if s.LayerCount != 2 {
		t.Error("Chain broken: ExecutionPlan not set")
	}
	if !s.Flags.TidyFirst {
		t.Error("Chain broken: Flags not set")
	}
	if s.ExecutionContext != "local" {
		t.Error("Chain broken: ExecutionContext not set")
	}
	if !s.DepmStatus.Verified {
		t.Error("Chain broken: DepmStatus not set")
	}
	if !s.DepsStatus.Verified {
		t.Error("Chain broken: DepsStatus not set")
	}
	if !s.HasIncremental() {
		t.Error("Chain broken: Incremental not set")
	}
	if s.OutputDir != "out/build/" {
		t.Error("Chain broken: OutputDir not set")
	}
}

func TestTestSummaryChaining(t *testing.T) {
	s := New("test").
		SetRequest([]string{"eac-core"}, []string{"eac-core"}).
		SetExecutionContext("local").
		SetTestInfo(&TestInfo{
			SuiteName:        "unit",
			TotalDiscovered:  100,
			Selected:         50,
			Skipped:          10,
			NotMatchingSuite: 40,
		}).
		SetArtifactValidation(&ArtifactValidationInfo{
			Validated:      true,
			ModulesChecked: []string{"eac-core"},
			AllPresent:     true,
		}).
		SetOutputDir("out/test/unit/")

	if !s.IsTest() {
		t.Error("Should be test command")
	}
	if !s.HasTestInfo() {
		t.Error("Should have test info")
	}
	if s.Test.SuiteName != "unit" {
		t.Errorf("SuiteName = %q, want %q", s.Test.SuiteName, "unit")
	}
	if !s.HasArtifactValidation() {
		t.Error("Should have artifact validation")
	}
	if !s.ArtifactValidation.AllPresent {
		t.Error("Artifacts should all be present")
	}
}

func TestSetParallelism(t *testing.T) {
	tests := []struct {
		name     string
		parallel *ParallelismInfo
		hasInfo  bool
	}{
		{
			name:     "nil parallelism",
			parallel: nil,
			hasInfo:  false,
		},
		{
			name: "devbox mode",
			parallel: &ParallelismInfo{
				Mode:             "devbox",
				BaseWorkers:      8,
				TurboBoost:       2,
				EffectiveWorkers: 10,
			},
			hasInfo: true,
		},
		{
			name: "ci mode",
			parallel: &ParallelismInfo{
				Mode:             "ci",
				BaseWorkers:      4,
				TurboBoost:       0,
				EffectiveWorkers: 4,
			},
			hasInfo: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New("build").SetParallelism(tt.parallel)

			if s.HasParallelism() != tt.hasInfo {
				t.Errorf("HasParallelism() = %v, want %v", s.HasParallelism(), tt.hasInfo)
			}

			if tt.hasInfo {
				if s.Parallelism.Mode != tt.parallel.Mode {
					t.Errorf("Mode = %q, want %q", s.Parallelism.Mode, tt.parallel.Mode)
				}
				if s.Parallelism.EffectiveWorkers != tt.parallel.EffectiveWorkers {
					t.Errorf("EffectiveWorkers = %d, want %d",
						s.Parallelism.EffectiveWorkers, tt.parallel.EffectiveWorkers)
				}
			}
		})
	}
}

func TestParallelismInfo_FullConfig(t *testing.T) {
	s := New("build").
		SetParallelism(&ParallelismInfo{
			Mode:             "devbox",
			BaseWorkers:      8,
			TurboBoost:       2,
			EffectiveWorkers: 10,
			ScannerCapacity:  3,
			PDFCapacity:      4,
			WeightedCapacity: 10,
		})

	if !s.HasParallelism() {
		t.Fatal("should have parallelism info")
	}

	p := s.Parallelism
	if p.ScannerCapacity != 3 {
		t.Errorf("ScannerCapacity = %d, want 3", p.ScannerCapacity)
	}
	if p.PDFCapacity != 4 {
		t.Errorf("PDFCapacity = %d, want 4", p.PDFCapacity)
	}
	if p.WeightedCapacity != 10 {
		t.Errorf("WeightedCapacity = %d, want 10", p.WeightedCapacity)
	}
}

func TestFormatInitLine(t *testing.T) {
	tests := []struct {
		name     string
		summary  *Summary
		contains []string
	}{
		{
			name: "build with parallelism",
			summary: New("build").
				SetExecutionContext("devbox").
				SetRequest([]string{"a", "b"}, []string{"a", "b", "c"}).
				SetExecutionPlan([][]string{{"a", "b"}, {"c"}}).
				SetUoWCount(45).
				SetParallelism(&ParallelismInfo{
					Mode:             "devbox",
					BaseWorkers:      8,
					TurboBoost:       2,
					EffectiveWorkers: 10,
				}),
			contains: []string{"build", "devbox", "3 modules", "2 layers", "workers:10"},
		},
		{
			name: "build without turbo",
			summary: New("build").
				SetExecutionContext("ci").
				SetRequest([]string{"a"}, []string{"a"}).
				SetParallelism(&ParallelismInfo{
					Mode:             "ci",
					BaseWorkers:      4,
					TurboBoost:       0,
					EffectiveWorkers: 4,
				}),
			contains: []string{"build", "ci", "workers:4"},
		},
		{
			name: "build flat execution",
			summary: New("build").
				SetExecutionContext("local").
				SetRequest([]string{"a", "b"}, []string{"a", "b"}).
				SetExecutionPlan([][]string{{"a"}, {"b"}}).
				SetFlatExecution(true),
			contains: []string{"parallel"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.summary.FormatInitLine()
			for _, want := range tt.contains {
				if !containsSubstring(result, want) {
					t.Errorf("FormatInitLine() = %q, want to contain %q", result, want)
				}
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
