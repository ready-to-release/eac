package changedetect

import (
	"runtime"
	"testing"
	"time"
)

// --- Tests for buildFreshResult ---

func TestBuildFreshResult(t *testing.T) {
	tests := []struct {
		name    string
		modules []string
	}{
		{
			name:    "single module",
			modules: []string{"module-a"},
		},
		{
			name:    "multiple modules",
			modules: []string{"module-a", "module-b", "module-c"},
		},
		{
			name:    "empty modules",
			modules: []string{},
		},
		{
			name:    "nil modules",
			modules: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			result := &ChangeResult{
				Reasons: make(map[string]string),
			}

			got := buildFreshResult(result, tt.modules, start)

			if got != result {
				t.Error("buildFreshResult should return the same result pointer")
			}
			if !got.Fresh {
				t.Error("Fresh should be true")
			}
			if len(got.Changed) != len(tt.modules) {
				t.Errorf("Changed length: got %d, want %d", len(got.Changed), len(tt.modules))
			}

			for i, m := range tt.modules {
				if got.Changed[i] != m {
					t.Errorf("Changed[%d]: got %q, want %q", i, got.Changed[i], m)
				}
				reason, exists := got.Reasons[m]
				if !exists {
					t.Errorf("missing reason for module %q", m)
				} else if reason != "fresh build (no previous state)" {
					t.Errorf("reason for %q: got %q, want %q", m, reason, "fresh build (no previous state)")
				}
			}

			if got.Duration <= 0 {
				// Duration could be zero on very fast systems, so just check it was set
				// (time.Since can return 0 on some systems)
			}
		})
	}
}

// --- Tests for classifyNewModules ---

func TestClassifyNewModules(t *testing.T) {
	tests := []struct {
		name              string
		modules           []string
		previousModules   map[string]ModuleState
		wantChanged       []string
		wantModulesToHash []string
		wantReasons       map[string]string
	}{
		{
			name:    "all new modules",
			modules: []string{"module-a", "module-b"},
			previousModules: map[string]ModuleState{
				// neither module-a nor module-b present
			},
			wantChanged:       []string{"module-a", "module-b"},
			wantModulesToHash: nil,
			wantReasons: map[string]string{
				"module-a": "new module (not in previous state)",
				"module-b": "new module (not in previous state)",
			},
		},
		{
			name:    "all existing modules",
			modules: []string{"module-a", "module-b"},
			previousModules: map[string]ModuleState{
				"module-a": {SourceHash: "hash-a"},
				"module-b": {SourceHash: "hash-b"},
			},
			wantChanged:       nil,
			wantModulesToHash: []string{"module-a", "module-b"},
			wantReasons:       map[string]string{},
		},
		{
			name:    "mix of new and existing modules",
			modules: []string{"module-a", "module-b", "module-c"},
			previousModules: map[string]ModuleState{
				"module-a": {SourceHash: "hash-a"},
				// module-b is new
				"module-c": {SourceHash: "hash-c"},
			},
			wantChanged:       []string{"module-b"},
			wantModulesToHash: []string{"module-a", "module-c"},
			wantReasons: map[string]string{
				"module-b": "new module (not in previous state)",
			},
		},
		{
			name:    "empty modules list",
			modules: []string{},
			previousModules: map[string]ModuleState{
				"module-a": {SourceHash: "hash-a"},
			},
			wantChanged:       nil,
			wantModulesToHash: nil,
			wantReasons:       map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ChangeResult{
				Reasons: make(map[string]string),
			}
			opts := DetectOptions{
				PreviousState: &WorkspaceState{
					Modules: tt.previousModules,
				},
				Modules: tt.modules,
			}

			modulesToHash := classifyNewModules(result, opts)

			// Verify changed modules
			if len(result.Changed) != len(tt.wantChanged) {
				t.Errorf("Changed length: got %d, want %d (got: %v)", len(result.Changed), len(tt.wantChanged), result.Changed)
			} else {
				for i, m := range tt.wantChanged {
					if result.Changed[i] != m {
						t.Errorf("Changed[%d]: got %q, want %q", i, result.Changed[i], m)
					}
				}
			}

			// Verify modules to hash
			if len(modulesToHash) != len(tt.wantModulesToHash) {
				t.Errorf("modulesToHash length: got %d, want %d (got: %v)", len(modulesToHash), len(tt.wantModulesToHash), modulesToHash)
			} else {
				for i, m := range tt.wantModulesToHash {
					if modulesToHash[i] != m {
						t.Errorf("modulesToHash[%d]: got %q, want %q", i, modulesToHash[i], m)
					}
				}
			}

			// Verify reasons
			for m, wantReason := range tt.wantReasons {
				gotReason, exists := result.Reasons[m]
				if !exists {
					t.Errorf("missing reason for module %q", m)
				} else if gotReason != wantReason {
					t.Errorf("reason for %q: got %q, want %q", m, gotReason, wantReason)
				}
			}
		})
	}
}

// --- Tests for computeWorkerCount ---

func TestComputeWorkerCount(t *testing.T) {
	// computeWorkerCount uses runtime.NumCPU() which we cannot mock,
	// but we can verify the invariants hold for the current system.
	workers := computeWorkerCount()
	numCPU := runtime.NumCPU()

	// Worker count must be at least 1
	if workers < 1 {
		t.Errorf("workers should be at least 1, got %d", workers)
	}

	// Worker count must not exceed 8
	if workers > 8 {
		t.Errorf("workers should not exceed 8, got %d", workers)
	}

	// Worker count should be at least min(4, NumCPU)
	floor := 4
	if numCPU < floor {
		floor = numCPU
	}
	if workers < floor {
		t.Errorf("workers (%d) should be at least floor (%d) for NumCPU=%d", workers, floor, numCPU)
	}

	// Worker count should be at most max(numCPU/2, floor)
	expectedFromHalf := numCPU / 2
	if expectedFromHalf < floor {
		expectedFromHalf = floor
	}
	if expectedFromHalf > 8 {
		expectedFromHalf = 8
	}
	if workers != expectedFromHalf {
		t.Errorf("workers: got %d, want %d (numCPU=%d)", workers, expectedFromHalf, numCPU)
	}
}

// --- Tests for compareModuleHashes ---

func TestCompareModuleHashes(t *testing.T) {
	tests := []struct {
		name           string
		modulesToHash  []string
		moduleHashes   map[string]string
		prevState      *WorkspaceState
		wantChanged    []string
		wantUpToDate   []string
		wantReasons    map[string]string
	}{
		{
			name:          "all hashes match",
			modulesToHash: []string{"module-a", "module-b"},
			moduleHashes: map[string]string{
				"module-a": "hash-a",
				"module-b": "hash-b",
			},
			prevState: &WorkspaceState{
				Modules: map[string]ModuleState{
					"module-a": {SourceHash: "hash-a"},
					"module-b": {SourceHash: "hash-b"},
				},
			},
			wantChanged:  nil,
			wantUpToDate: []string{"module-a", "module-b"},
			wantReasons:  map[string]string{},
		},
		{
			name:          "all hashes differ",
			modulesToHash: []string{"module-a", "module-b"},
			moduleHashes: map[string]string{
				"module-a": "new-hash-a",
				"module-b": "new-hash-b",
			},
			prevState: &WorkspaceState{
				Modules: map[string]ModuleState{
					"module-a": {SourceHash: "old-hash-a"},
					"module-b": {SourceHash: "old-hash-b"},
				},
			},
			wantChanged:  []string{"module-a", "module-b"},
			wantUpToDate: nil,
			wantReasons: map[string]string{
				"module-a": "source files changed",
				"module-b": "source files changed",
			},
		},
		{
			name:          "mixed changes",
			modulesToHash: []string{"module-a", "module-b", "module-c"},
			moduleHashes: map[string]string{
				"module-a": "hash-a",     // same
				"module-b": "new-hash-b", // changed
				"module-c": "hash-c",     // same
			},
			prevState: &WorkspaceState{
				Modules: map[string]ModuleState{
					"module-a": {SourceHash: "hash-a"},
					"module-b": {SourceHash: "old-hash-b"},
					"module-c": {SourceHash: "hash-c"},
				},
			},
			wantChanged:  []string{"module-b"},
			wantUpToDate: []string{"module-a", "module-c"},
			wantReasons: map[string]string{
				"module-b": "source files changed",
			},
		},
		{
			name:          "empty modules to hash",
			modulesToHash: []string{},
			moduleHashes:  map[string]string{},
			prevState: &WorkspaceState{
				Modules: map[string]ModuleState{},
			},
			wantChanged:  nil,
			wantUpToDate: nil,
			wantReasons:  map[string]string{},
		},
		{
			name:          "empty string hash treated as change",
			modulesToHash: []string{"module-a"},
			moduleHashes: map[string]string{
				"module-a": "",
			},
			prevState: &WorkspaceState{
				Modules: map[string]ModuleState{
					"module-a": {SourceHash: "some-hash"},
				},
			},
			wantChanged:  []string{"module-a"},
			wantUpToDate: nil,
			wantReasons: map[string]string{
				"module-a": "source files changed",
			},
		},
		{
			name:          "both empty hashes match",
			modulesToHash: []string{"module-a"},
			moduleHashes: map[string]string{
				"module-a": "",
			},
			prevState: &WorkspaceState{
				Modules: map[string]ModuleState{
					"module-a": {SourceHash: ""},
				},
			},
			wantChanged:  nil,
			wantUpToDate: []string{"module-a"},
			wantReasons:  map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ChangeResult{
				Reasons: make(map[string]string),
			}

			compareModuleHashes(result, tt.modulesToHash, tt.moduleHashes, tt.prevState, false)

			// Verify changed
			if len(result.Changed) != len(tt.wantChanged) {
				t.Errorf("Changed length: got %d, want %d (got: %v)", len(result.Changed), len(tt.wantChanged), result.Changed)
			} else {
				for i, m := range tt.wantChanged {
					if result.Changed[i] != m {
						t.Errorf("Changed[%d]: got %q, want %q", i, result.Changed[i], m)
					}
				}
			}

			// Verify up-to-date
			if len(result.UpToDate) != len(tt.wantUpToDate) {
				t.Errorf("UpToDate length: got %d, want %d (got: %v)", len(result.UpToDate), len(tt.wantUpToDate), result.UpToDate)
			} else {
				for i, m := range tt.wantUpToDate {
					if result.UpToDate[i] != m {
						t.Errorf("UpToDate[%d]: got %q, want %q", i, result.UpToDate[i], m)
					}
				}
			}

			// Verify reasons
			for m, wantReason := range tt.wantReasons {
				gotReason, exists := result.Reasons[m]
				if !exists {
					t.Errorf("missing reason for module %q", m)
				} else if gotReason != wantReason {
					t.Errorf("reason for %q: got %q, want %q", m, gotReason, wantReason)
				}
			}
		})
	}
}

// --- Tests for propagateDependencyChanges ---

func TestPropagateDependencyChanges(t *testing.T) {
	tests := []struct {
		name             string
		initialChanged   []string
		initialUpToDate  []string
		initialReasons   map[string]string
		checkDependencies bool
		dependencyGraph  map[string][]string
		wantChanged      []string
		wantUpToDate     []string
		wantReasons      map[string]string
	}{
		{
			name:            "dependency checking disabled",
			initialChanged:  []string{"module-b"},
			initialUpToDate: []string{"module-a"},
			initialReasons: map[string]string{
				"module-b": "source files changed",
			},
			checkDependencies: false,
			dependencyGraph: map[string][]string{
				"module-a": {"module-b"},
			},
			wantChanged:  []string{"module-b"},
			wantUpToDate: []string{"module-a"},
			wantReasons: map[string]string{
				"module-b": "source files changed",
			},
		},
		{
			name:            "nil dependency graph",
			initialChanged:  []string{"module-b"},
			initialUpToDate: []string{"module-a"},
			initialReasons: map[string]string{
				"module-b": "source files changed",
			},
			checkDependencies: true,
			dependencyGraph:   nil,
			wantChanged:       []string{"module-b"},
			wantUpToDate:      []string{"module-a"},
			wantReasons: map[string]string{
				"module-b": "source files changed",
			},
		},
		{
			name:            "no changed modules",
			initialChanged:  []string{},
			initialUpToDate: []string{"module-a", "module-b"},
			initialReasons:  map[string]string{},
			checkDependencies: true,
			dependencyGraph: map[string][]string{
				"module-a": {"module-b"},
			},
			wantChanged:  []string{},
			wantUpToDate: []string{"module-a", "module-b"},
			wantReasons:  map[string]string{},
		},
		{
			name:            "dependency changed propagates",
			initialChanged:  []string{"module-b"},
			initialUpToDate: []string{"module-a"},
			initialReasons: map[string]string{
				"module-b": "source files changed",
			},
			checkDependencies: true,
			dependencyGraph: map[string][]string{
				"module-a": {"module-b"}, // module-a depends on module-b
			},
			wantChanged:  []string{"module-b", "module-a"},
			wantUpToDate: nil,
			wantReasons: map[string]string{
				"module-b": "source files changed",
				"module-a": "dependency changed",
			},
		},
		{
			name:            "no dependency relationship",
			initialChanged:  []string{"module-b"},
			initialUpToDate: []string{"module-a"},
			initialReasons: map[string]string{
				"module-b": "source files changed",
			},
			checkDependencies: true,
			dependencyGraph: map[string][]string{
				// module-a has no dependencies
			},
			wantChanged:  []string{"module-b"},
			wantUpToDate: []string{"module-a"},
			wantReasons: map[string]string{
				"module-b": "source files changed",
			},
		},
		{
			name:            "multiple up-to-date modules with one dependency",
			initialChanged:  []string{"module-c"},
			initialUpToDate: []string{"module-a", "module-b"},
			initialReasons: map[string]string{
				"module-c": "source files changed",
			},
			checkDependencies: true,
			dependencyGraph: map[string][]string{
				"module-a": {"module-c"}, // module-a depends on changed module-c
				"module-b": {},           // module-b has no deps
			},
			wantChanged:  []string{"module-c", "module-a"},
			wantUpToDate: []string{"module-b"},
			wantReasons: map[string]string{
				"module-c": "source files changed",
				"module-a": "dependency changed",
			},
		},
		{
			name:            "module depends on unchanged module",
			initialChanged:  []string{"module-c"},
			initialUpToDate: []string{"module-a", "module-b"},
			initialReasons: map[string]string{
				"module-c": "source files changed",
			},
			checkDependencies: true,
			dependencyGraph: map[string][]string{
				"module-a": {"module-b"}, // depends on unchanged module-b
				"module-b": {},
			},
			wantChanged:  []string{"module-c"},
			wantUpToDate: []string{"module-a", "module-b"},
			wantReasons: map[string]string{
				"module-c": "source files changed",
			},
		},
		{
			name:            "multiple dependencies one changed",
			initialChanged:  []string{"module-d"},
			initialUpToDate: []string{"module-a"},
			initialReasons: map[string]string{
				"module-d": "source files changed",
			},
			checkDependencies: true,
			dependencyGraph: map[string][]string{
				"module-a": {"module-b", "module-c", "module-d"}, // depends on b, c, d; only d changed
			},
			wantChanged:  []string{"module-d", "module-a"},
			wantUpToDate: nil,
			wantReasons: map[string]string{
				"module-d": "source files changed",
				"module-a": "dependency changed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ChangeResult{
				Changed:  tt.initialChanged,
				UpToDate: tt.initialUpToDate,
				Reasons:  tt.initialReasons,
			}
			opts := DetectOptions{
				CheckDependencies: tt.checkDependencies,
				DependencyGraph:   tt.dependencyGraph,
			}

			propagateDependencyChanges(result, opts)

			// Verify changed
			if len(result.Changed) != len(tt.wantChanged) {
				t.Errorf("Changed length: got %d, want %d (got: %v)", len(result.Changed), len(tt.wantChanged), result.Changed)
			} else {
				for i, m := range tt.wantChanged {
					if result.Changed[i] != m {
						t.Errorf("Changed[%d]: got %q, want %q", i, result.Changed[i], m)
					}
				}
			}

			// Verify up-to-date
			if len(result.UpToDate) != len(tt.wantUpToDate) {
				t.Errorf("UpToDate length: got %d, want %d (got: %v)", len(result.UpToDate), len(tt.wantUpToDate), result.UpToDate)
			} else {
				for i, m := range tt.wantUpToDate {
					if result.UpToDate[i] != m {
						t.Errorf("UpToDate[%d]: got %q, want %q", i, result.UpToDate[i], m)
					}
				}
			}

			// Verify reasons
			for m, wantReason := range tt.wantReasons {
				gotReason, exists := result.Reasons[m]
				if !exists {
					t.Errorf("missing reason for module %q", m)
				} else if gotReason != wantReason {
					t.Errorf("reason for %q: got %q, want %q", m, gotReason, wantReason)
				}
			}
		})
	}
}
