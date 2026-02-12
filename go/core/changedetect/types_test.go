package changedetect

import (
	"testing"
	"time"
)

func TestWorkspaceState_Clone_DeepCopiesModules(t *testing.T) {
	original := &WorkspaceState{
		Commit:          "abc123",
		UncommittedHash: "dirty-hash",
		Modules: map[string]ModuleState{
			"module-a": {SourceHash: "hash-a"},
			"module-b": {SourceHash: "hash-b"},
		},
		RecordedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	clone := original.Clone()

	// Verify all fields are copied
	if clone.Commit != original.Commit {
		t.Errorf("Commit: got %q, want %q", clone.Commit, original.Commit)
	}
	if clone.UncommittedHash != original.UncommittedHash {
		t.Errorf("UncommittedHash: got %q, want %q", clone.UncommittedHash, original.UncommittedHash)
	}
	if !clone.RecordedAt.Equal(original.RecordedAt) {
		t.Errorf("RecordedAt: got %v, want %v", clone.RecordedAt, original.RecordedAt)
	}
	if len(clone.Modules) != len(original.Modules) {
		t.Fatalf("Modules length: got %d, want %d", len(clone.Modules), len(original.Modules))
	}

	// Mutate the clone's module map and verify the original is unaffected
	clone.Modules["module-a"] = ModuleState{SourceHash: "mutated"}
	clone.Modules["module-c"] = ModuleState{SourceHash: "new-module"}

	if original.Modules["module-a"].SourceHash != "hash-a" {
		t.Error("mutating clone's module value affected original")
	}
	if _, exists := original.Modules["module-c"]; exists {
		t.Error("adding module to clone affected original")
	}
}

func TestWorkspaceState_Clone_WithExtraMap(t *testing.T) {
	original := &WorkspaceState{
		Commit: "abc123",
		Modules: map[string]ModuleState{
			"module-a": {
				SourceHash: "hash-a",
				Extra: map[string]interface{}{
					"passed":   true,
					"build_id": "build-42",
					"count":    float64(7),
				},
			},
			"module-b": {
				SourceHash: "hash-b",
				Extra:      nil, // No extra data
			},
		},
		RecordedAt: time.Now(),
	}

	clone := original.Clone()

	// Verify Extra map is deep-copied
	modA := clone.Modules["module-a"]
	if len(modA.Extra) != 3 {
		t.Fatalf("clone Extra length: got %d, want 3", len(modA.Extra))
	}
	if modA.Extra["passed"] != true {
		t.Errorf("Extra[passed]: got %v, want true", modA.Extra["passed"])
	}
	if modA.Extra["build_id"] != "build-42" {
		t.Errorf("Extra[build_id]: got %v, want build-42", modA.Extra["build_id"])
	}

	// Mutate clone's Extra map
	modA.Extra["passed"] = false
	modA.Extra["new_key"] = "new_value"
	clone.Modules["module-a"] = modA

	// Original should be unaffected
	origExtra := original.Modules["module-a"].Extra
	if origExtra["passed"] != true {
		t.Error("mutating clone's Extra affected original")
	}
	if _, exists := origExtra["new_key"]; exists {
		t.Error("adding key to clone's Extra affected original")
	}

	// module-b with nil Extra should clone as nil
	modB := clone.Modules["module-b"]
	if modB.Extra != nil {
		t.Errorf("expected nil Extra for module-b clone, got %v", modB.Extra)
	}
}

func TestWorkspaceState_Clone_EmptyModules(t *testing.T) {
	original := &WorkspaceState{
		Commit:     "abc123",
		Modules:    map[string]ModuleState{},
		RecordedAt: time.Now(),
	}

	clone := original.Clone()

	if clone.Modules == nil {
		t.Fatal("clone Modules should not be nil")
	}
	if len(clone.Modules) != 0 {
		t.Errorf("expected empty Modules map, got %d entries", len(clone.Modules))
	}

	// Adding to clone should not affect original
	clone.Modules["new"] = ModuleState{SourceHash: "x"}
	if len(original.Modules) != 0 {
		t.Error("adding to clone's Modules affected original")
	}
}

func TestChangeResult_ZeroValue(t *testing.T) {
	// Verify zero-value ChangeResult behaves safely
	var result ChangeResult

	if result.Fresh {
		t.Error("zero-value Fresh should be false")
	}
	if len(result.Changed) != 0 {
		t.Error("zero-value Changed should be empty")
	}
	if len(result.UpToDate) != 0 {
		t.Error("zero-value UpToDate should be empty")
	}
	if result.Reasons != nil {
		t.Error("zero-value Reasons should be nil")
	}
	if result.Duration != 0 {
		t.Error("zero-value Duration should be zero")
	}
}

func TestDetectOptions_ZeroValue(t *testing.T) {
	// Verify zero-value DetectOptions behaves safely
	var opts DetectOptions

	if opts.WorkspaceRoot != "" {
		t.Error("zero-value WorkspaceRoot should be empty")
	}
	if opts.PreviousState != nil {
		t.Error("zero-value PreviousState should be nil")
	}
	if len(opts.Modules) != 0 {
		t.Error("zero-value Modules should be empty")
	}
	if opts.CheckDependencies {
		t.Error("zero-value CheckDependencies should be false")
	}
	if opts.DependencyGraph != nil {
		t.Error("zero-value DependencyGraph should be nil")
	}
}
