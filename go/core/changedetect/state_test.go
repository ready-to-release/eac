package changedetect

import (
	"context"
	"errors"
	"testing"
	"time"
)

// --- Tests for ConvertToWorkspaceState ---

func TestConvertToWorkspaceState(t *testing.T) {
	tests := []struct {
		name            string
		commit          string
		uncommittedHash string
		moduleStates    map[string]string
		recordedAt      time.Time
	}{
		{
			name:            "multiple modules",
			commit:          "abc123",
			uncommittedHash: "dirty-hash",
			moduleStates: map[string]string{
				"module-a": "hash-a",
				"module-b": "hash-b",
				"module-c": "hash-c",
			},
			recordedAt: time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			name:            "empty modules",
			commit:          "def456",
			uncommittedHash: "",
			moduleStates:    map[string]string{},
			recordedAt:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:            "nil modules map",
			commit:          "ghi789",
			uncommittedHash: "",
			moduleStates:    nil,
			recordedAt:      time.Date(2025, 3, 10, 8, 30, 0, 0, time.UTC),
		},
		{
			name:            "single module clean tree",
			commit:          "xyz999",
			uncommittedHash: "",
			moduleStates: map[string]string{
				"only-module": "only-hash",
			},
			recordedAt: time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			name:            "empty commit and hash",
			commit:          "",
			uncommittedHash: "",
			moduleStates: map[string]string{
				"module-a": "hash-a",
			},
			recordedAt: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := ConvertToWorkspaceState(tt.commit, tt.uncommittedHash, tt.moduleStates, tt.recordedAt)

			if state == nil {
				t.Fatal("ConvertToWorkspaceState returned nil")
			}
			if state.Commit != tt.commit {
				t.Errorf("Commit: got %q, want %q", state.Commit, tt.commit)
			}
			if state.UncommittedHash != tt.uncommittedHash {
				t.Errorf("UncommittedHash: got %q, want %q", state.UncommittedHash, tt.uncommittedHash)
			}
			if !state.RecordedAt.Equal(tt.recordedAt) {
				t.Errorf("RecordedAt: got %v, want %v", state.RecordedAt, tt.recordedAt)
			}
			if state.Modules == nil {
				t.Fatal("Modules map should not be nil")
			}

			expectedLen := len(tt.moduleStates)
			if len(state.Modules) != expectedLen {
				t.Fatalf("Modules length: got %d, want %d", len(state.Modules), expectedLen)
			}

			for moniker, expectedHash := range tt.moduleStates {
				moduleState, exists := state.Modules[moniker]
				if !exists {
					t.Errorf("module %q not found in state", moniker)
					continue
				}
				if moduleState.SourceHash != expectedHash {
					t.Errorf("module %q SourceHash: got %q, want %q", moniker, moduleState.SourceHash, expectedHash)
				}
			}
		})
	}
}

// --- Tests for ComputeCurrentState error paths ---

func TestDetector_ComputeCurrentState_HeadCommitError(t *testing.T) {
	expectedErr := errors.New("head commit failed")
	git := &mockGitProvider{headCommitErr: expectedErr}
	hasher := &mockFileHasher{}
	resolver := &mockModuleResolver{}
	d := NewDetector(git, hasher, resolver)

	_, err := d.ComputeCurrentState(context.Background(), "/workspace", []string{"module-a"})

	if err == nil {
		t.Fatal("expected error from HeadCommit failure")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped head commit error, got: %v", err)
	}
}

func TestDetector_ComputeCurrentState_UncommittedFilesError(t *testing.T) {
	expectedErr := errors.New("uncommitted files failed")
	git := &mockGitProvider{
		headCommit:   "abc123",
		uncommittedErr: expectedErr,
	}
	hasher := &mockFileHasher{}
	resolver := &mockModuleResolver{}
	d := NewDetector(git, hasher, resolver)

	_, err := d.ComputeCurrentState(context.Background(), "/workspace", []string{"module-a"})

	if err == nil {
		t.Fatal("expected error from UncommittedFiles failure")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped uncommitted files error, got: %v", err)
	}
}

func TestDetector_ComputeCurrentState_NoModules(t *testing.T) {
	git := &mockGitProvider{headCommit: "abc123"}
	hasher := &mockFileHasher{}
	resolver := &mockModuleResolver{}
	d := NewDetector(git, hasher, resolver)

	state, err := d.ComputeCurrentState(context.Background(), "/workspace", []string{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state == nil {
		t.Fatal("state should not be nil")
	}
	if state.Commit != "abc123" {
		t.Errorf("Commit: got %q, want %q", state.Commit, "abc123")
	}
	if len(state.Modules) != 0 {
		t.Errorf("expected 0 modules, got %d", len(state.Modules))
	}
}

func TestDetector_ComputeCurrentState_NilModules(t *testing.T) {
	git := &mockGitProvider{headCommit: "abc123"}
	hasher := &mockFileHasher{}
	resolver := &mockModuleResolver{}
	d := NewDetector(git, hasher, resolver)

	state, err := d.ComputeCurrentState(context.Background(), "/workspace", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state == nil {
		t.Fatal("state should not be nil")
	}
	if len(state.Modules) != 0 {
		t.Errorf("expected 0 modules, got %d", len(state.Modules))
	}
}

func TestDetector_ComputeCurrentState_CleanWorkingTree(t *testing.T) {
	git := &mockGitProvider{
		headCommit:       "abc123",
		uncommittedFiles: nil, // Clean working tree
	}
	hasher := &mockFileHasher{
		hashes: map[string]string{"main.go": "module-hash"},
	}
	resolver := &mockModuleResolver{
		moduleFiles: map[string][]string{
			"module-a": {"main.go"},
		},
	}
	d := NewDetector(git, hasher, resolver)

	state, err := d.ComputeCurrentState(context.Background(), "/workspace", []string{"module-a"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.UncommittedHash != "" {
		t.Errorf("expected empty UncommittedHash for clean tree, got %q", state.UncommittedHash)
	}
	if state.Commit != "abc123" {
		t.Errorf("Commit: got %q, want %q", state.Commit, "abc123")
	}
	if len(state.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(state.Modules))
	}
	if state.Modules["module-a"].SourceHash != "module-hash" {
		t.Errorf("SourceHash: got %q, want %q", state.Modules["module-a"].SourceHash, "module-hash")
	}
}

func TestDetector_ComputeCurrentState_MultipleModules(t *testing.T) {
	git := &mockGitProvider{headCommit: "multi-commit"}
	hasher := &mockFileHasher{
		hashes: map[string]string{
			"a.go": "hash-a",
			"b.go": "hash-b",
			"c.go": "hash-c",
		},
	}
	resolver := &mockModuleResolver{
		moduleFiles: map[string][]string{
			"module-a": {"a.go"},
			"module-b": {"b.go"},
			"module-c": {"c.go"},
		},
	}
	d := NewDetector(git, hasher, resolver)

	state, err := d.ComputeCurrentState(context.Background(), "/workspace", []string{"module-a", "module-b", "module-c"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(state.Modules) != 3 {
		t.Fatalf("expected 3 modules, got %d", len(state.Modules))
	}

	expected := map[string]string{
		"module-a": "hash-a",
		"module-b": "hash-b",
		"module-c": "hash-c",
	}
	for moniker, expectedHash := range expected {
		ms, exists := state.Modules[moniker]
		if !exists {
			t.Errorf("module %q not found in state", moniker)
			continue
		}
		if ms.SourceHash != expectedHash {
			t.Errorf("module %q SourceHash: got %q, want %q", moniker, ms.SourceHash, expectedHash)
		}
	}
}

func TestDetector_ComputeCurrentState_ModuleHashError(t *testing.T) {
	expectedErr := errors.New("hash computation failed")
	git := &mockGitProvider{headCommit: "abc123"}
	hasher := &mockFileHasher{hashFilesErr: expectedErr}
	resolver := &mockModuleResolver{
		moduleFiles: map[string][]string{
			"module-a": {"main.go"},
		},
	}
	d := NewDetector(git, hasher, resolver)

	_, err := d.ComputeCurrentState(context.Background(), "/workspace", []string{"module-a"})

	if err == nil {
		t.Fatal("expected error from hash computation failure")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped hash error, got: %v", err)
	}
}

func TestDetector_ComputeCurrentState_ModuleResolverError(t *testing.T) {
	expectedErr := errors.New("resolver failed")
	git := &mockGitProvider{headCommit: "abc123"}
	hasher := &mockFileHasher{}
	resolver := &mockModuleResolver{resolveErr: expectedErr}
	d := NewDetector(git, hasher, resolver)

	_, err := d.ComputeCurrentState(context.Background(), "/workspace", []string{"module-a"})

	if err == nil {
		t.Fatal("expected error from module resolver failure")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped resolver error, got: %v", err)
	}
}

func TestDetector_ComputeCurrentState_RecordedAtIsRecent(t *testing.T) {
	git := &mockGitProvider{headCommit: "abc123"}
	hasher := &mockFileHasher{}
	resolver := &mockModuleResolver{}
	d := NewDetector(git, hasher, resolver)

	before := time.Now()
	state, err := d.ComputeCurrentState(context.Background(), "/workspace", []string{})
	after := time.Now()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.RecordedAt.Before(before) || state.RecordedAt.After(after) {
		t.Errorf("RecordedAt %v not between %v and %v", state.RecordedAt, before, after)
	}
}
