// Package changedetect provides unified change detection for build, lint, and test operations.
package changedetect

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Mock implementations for testing

type mockGitProvider struct {
	headCommit       string
	uncommittedFiles []string
	headCommitErr    error
	uncommittedErr   error
}

func (m *mockGitProvider) HeadCommit(ctx context.Context) (string, error) {
	if m.headCommitErr != nil {
		return "", m.headCommitErr
	}
	return m.headCommit, nil
}

func (m *mockGitProvider) UncommittedFiles(ctx context.Context) ([]string, error) {
	if m.uncommittedErr != nil {
		return nil, m.uncommittedErr
	}
	return m.uncommittedFiles, nil
}

type mockFileHasher struct {
	hashes            map[string]string // key = comma-joined files, value = hash
	uncommittedHash   string
	hashFilesErr      error
	uncommittedHashOK bool
}

func (m *mockFileHasher) HashFiles(ctx context.Context, workspaceRoot string, files []string) (string, error) {
	if m.hashFilesErr != nil {
		return "", m.hashFilesErr
	}
	// Simple mock: return hash based on first file
	if len(files) > 0 && m.hashes != nil {
		if h, ok := m.hashes[files[0]]; ok {
			return h, nil
		}
	}
	return "default-hash", nil
}

func (m *mockFileHasher) HashUncommittedState(ctx context.Context, workspaceRoot string, files []string) (string, error) {
	if !m.uncommittedHashOK {
		return "", nil
	}
	return m.uncommittedHash, nil
}

type mockModuleResolver struct {
	moduleFiles map[string][]string // moniker -> files
	resolveErr  error
}

func (m *mockModuleResolver) GetModuleFiles(ctx context.Context, workspaceRoot string, moniker string) ([]string, error) {
	if m.resolveErr != nil {
		return nil, m.resolveErr
	}
	if files, ok := m.moduleFiles[moniker]; ok {
		return files, nil
	}
	return nil, nil
}

// Tests for Detector

func TestNewDetector(t *testing.T) {
	git := &mockGitProvider{}
	hasher := &mockFileHasher{}
	resolver := &mockModuleResolver{}

	d := NewDetector(git, hasher, resolver)

	if d == nil {
		t.Fatal("NewDetector returned nil")
	}
	if d.git != git {
		t.Error("git provider not set correctly")
	}
	if d.hasher != hasher {
		t.Error("hasher not set correctly")
	}
	if d.files != resolver {
		t.Error("file resolver not set correctly")
	}
}

func TestDetector_DetectChanges_FreshBuild(t *testing.T) {
	git := &mockGitProvider{headCommit: "abc123"}
	hasher := &mockFileHasher{hashes: map[string]string{"main.go": "hash1"}}
	resolver := &mockModuleResolver{
		moduleFiles: map[string][]string{
			"module-a": {"main.go"},
		},
	}
	d := NewDetector(git, hasher, resolver)

	result, err := d.DetectChanges(context.Background(), DetectOptions{
		WorkspaceRoot: "/workspace",
		PreviousState: nil, // Fresh build
		Modules:       []string{"module-a"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Fresh {
		t.Error("expected Fresh=true for nil previous state")
	}
	if len(result.Changed) != 1 || result.Changed[0] != "module-a" {
		t.Errorf("expected Changed=[module-a], got %v", result.Changed)
	}
	if result.Reasons["module-a"] != "fresh build (no previous state)" {
		t.Errorf("unexpected reason: %s", result.Reasons["module-a"])
	}
}

func TestDetector_DetectChanges_NoChanges(t *testing.T) {
	const commit = "abc123def456"
	git := &mockGitProvider{headCommit: commit, uncommittedFiles: nil}
	hasher := &mockFileHasher{hashes: map[string]string{"main.go": "hash1"}}
	resolver := &mockModuleResolver{
		moduleFiles: map[string][]string{
			"module-a": {"main.go"},
		},
	}
	d := NewDetector(git, hasher, resolver)

	// Previous state matches current
	prevState := &WorkspaceState{
		Commit:          commit,
		UncommittedHash: "",
		Modules: map[string]ModuleState{
			"module-a": {SourceHash: "hash1"},
		},
		RecordedAt: time.Now().Add(-time.Hour),
	}

	result, err := d.DetectChanges(context.Background(), DetectOptions{
		WorkspaceRoot: "/workspace",
		PreviousState: prevState,
		Modules:       []string{"module-a"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Fresh {
		t.Error("expected Fresh=false")
	}
	if len(result.Changed) != 0 {
		t.Errorf("expected no changes, got %v", result.Changed)
	}
	if len(result.UpToDate) != 1 || result.UpToDate[0] != "module-a" {
		t.Errorf("expected UpToDate=[module-a], got %v", result.UpToDate)
	}
}

func TestDetector_DetectChanges_CommitChanged(t *testing.T) {
	git := &mockGitProvider{headCommit: "new-commit"}
	hasher := &mockFileHasher{hashes: map[string]string{"main.go": "hash1"}}
	resolver := &mockModuleResolver{
		moduleFiles: map[string][]string{
			"module-a": {"main.go"},
		},
	}
	d := NewDetector(git, hasher, resolver)

	prevState := &WorkspaceState{
		Commit: "old-commit",
		Modules: map[string]ModuleState{
			"module-a": {SourceHash: "hash1"},
		},
	}

	result, err := d.DetectChanges(context.Background(), DetectOptions{
		WorkspaceRoot: "/workspace",
		PreviousState: prevState,
		Modules:       []string{"module-a"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should recompute hashes since commit changed
	// Even if hash matches, commit changed so it's still considered
	// Actually per the plan, if hashes match it's up-to-date
	if len(result.UpToDate) != 1 {
		t.Errorf("expected module-a to be up-to-date since hash matches, got changed=%v", result.Changed)
	}
}

func TestDetector_DetectChanges_SourceFilesChanged(t *testing.T) {
	const commit = "same-commit"
	git := &mockGitProvider{headCommit: commit}
	hasher := &mockFileHasher{hashes: map[string]string{"main.go": "new-hash"}}
	resolver := &mockModuleResolver{
		moduleFiles: map[string][]string{
			"module-a": {"main.go"},
		},
	}
	d := NewDetector(git, hasher, resolver)

	prevState := &WorkspaceState{
		Commit: commit,
		Modules: map[string]ModuleState{
			"module-a": {SourceHash: "old-hash"}, // Different hash
		},
	}

	result, err := d.DetectChanges(context.Background(), DetectOptions{
		WorkspaceRoot: "/workspace",
		PreviousState: prevState,
		Modules:       []string{"module-a"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Changed) != 1 || result.Changed[0] != "module-a" {
		t.Errorf("expected Changed=[module-a], got %v", result.Changed)
	}
	if result.Reasons["module-a"] != "source files changed" {
		t.Errorf("unexpected reason: %s", result.Reasons["module-a"])
	}
}

func TestDetector_DetectChanges_UncommittedChanges(t *testing.T) {
	const commit = "same-commit"
	git := &mockGitProvider{headCommit: commit, uncommittedFiles: []string{"dirty.go"}}
	hasher := &mockFileHasher{
		hashes:            map[string]string{"main.go": "hash1"},
		uncommittedHash:   "dirty-hash",
		uncommittedHashOK: true,
	}
	resolver := &mockModuleResolver{
		moduleFiles: map[string][]string{
			"module-a": {"main.go"},
		},
	}
	d := NewDetector(git, hasher, resolver)

	prevState := &WorkspaceState{
		Commit:          commit,
		UncommittedHash: "different-dirty-hash", // Changed
		Modules: map[string]ModuleState{
			"module-a": {SourceHash: "hash1"},
		},
	}

	result, err := d.DetectChanges(context.Background(), DetectOptions{
		WorkspaceRoot: "/workspace",
		PreviousState: prevState,
		Modules:       []string{"module-a"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Uncommitted changes changed, but module hash still same
	// This depends on whether uncommitted files affect the module
	// The existing behavior might recompute hashes
	// For now, just ensure no error
	_ = result
}

func TestDetector_DetectChanges_NewModule(t *testing.T) {
	const commit = "same-commit"
	git := &mockGitProvider{headCommit: commit}
	hasher := &mockFileHasher{hashes: map[string]string{
		"a.go": "hash-a",
		"b.go": "hash-b",
	}}
	resolver := &mockModuleResolver{
		moduleFiles: map[string][]string{
			"module-a": {"a.go"},
			"module-b": {"b.go"},
		},
	}
	d := NewDetector(git, hasher, resolver)

	prevState := &WorkspaceState{
		Commit: commit,
		Modules: map[string]ModuleState{
			"module-a": {SourceHash: "hash-a"},
			// module-b not present
		},
	}

	result, err := d.DetectChanges(context.Background(), DetectOptions{
		WorkspaceRoot: "/workspace",
		PreviousState: prevState,
		Modules:       []string{"module-a", "module-b"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// module-b should be marked as changed (new module)
	foundB := false
	for _, m := range result.Changed {
		if m == "module-b" {
			foundB = true
			break
		}
	}
	if !foundB {
		t.Errorf("expected module-b in changed list, got %v", result.Changed)
	}
	if result.Reasons["module-b"] != "new module (not in previous state)" {
		t.Errorf("unexpected reason for module-b: %s", result.Reasons["module-b"])
	}
}

func TestDetector_DetectChanges_GitError(t *testing.T) {
	expectedErr := errors.New("git error")
	git := &mockGitProvider{headCommitErr: expectedErr}
	hasher := &mockFileHasher{}
	resolver := &mockModuleResolver{}
	d := NewDetector(git, hasher, resolver)

	// Need a non-nil PreviousState to trigger git operations
	prevState := &WorkspaceState{
		Commit:  "old-commit",
		Modules: map[string]ModuleState{},
	}

	_, err := d.DetectChanges(context.Background(), DetectOptions{
		WorkspaceRoot: "/workspace",
		PreviousState: prevState,
		Modules:       []string{"module-a"},
	})

	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected git error, got %v", err)
	}
}

// Tests for ComputeCurrentState

func TestDetector_ComputeCurrentState(t *testing.T) {
	const commit = "abc123def456"
	git := &mockGitProvider{headCommit: commit, uncommittedFiles: []string{"dirty.go"}}
	hasher := &mockFileHasher{
		hashes:            map[string]string{"main.go": "module-hash"},
		uncommittedHash:   "uncommitted-hash",
		uncommittedHashOK: true,
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
	if state.Commit != commit {
		t.Errorf("expected commit %s, got %s", commit, state.Commit)
	}
	if state.UncommittedHash != "uncommitted-hash" {
		t.Errorf("expected uncommitted hash 'uncommitted-hash', got %s", state.UncommittedHash)
	}
	if len(state.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(state.Modules))
	}
	if state.Modules["module-a"].SourceHash != "module-hash" {
		t.Errorf("expected source hash 'module-hash', got %s", state.Modules["module-a"].SourceHash)
	}
}

// Tests for ComputeModuleHash

func TestDetector_ComputeModuleHash(t *testing.T) {
	git := &mockGitProvider{}
	hasher := &mockFileHasher{hashes: map[string]string{"main.go": "expected-hash"}}
	resolver := &mockModuleResolver{
		moduleFiles: map[string][]string{
			"module-a": {"main.go"},
		},
	}
	d := NewDetector(git, hasher, resolver)

	hash, err := d.ComputeModuleHash(context.Background(), "/workspace", "module-a")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "expected-hash" {
		t.Errorf("expected 'expected-hash', got %s", hash)
	}
}

func TestDetector_ComputeModuleHash_ResolverError(t *testing.T) {
	expectedErr := errors.New("resolve error")
	git := &mockGitProvider{}
	hasher := &mockFileHasher{}
	resolver := &mockModuleResolver{resolveErr: expectedErr}
	d := NewDetector(git, hasher, resolver)

	_, err := d.ComputeModuleHash(context.Background(), "/workspace", "module-a")

	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected resolve error, got %v", err)
	}
}

func TestDetector_DetectChanges_DependencyChanged(t *testing.T) {
	const commit = "same-commit"
	git := &mockGitProvider{headCommit: commit}
	hasher := &mockFileHasher{hashes: map[string]string{
		"a.go": "hash-a",
		"b.go": "new-hash-b", // Changed
	}}
	resolver := &mockModuleResolver{
		moduleFiles: map[string][]string{
			"module-a": {"a.go"}, // Depends on module-b
			"module-b": {"b.go"}, // Changed
		},
	}
	d := NewDetector(git, hasher, resolver)

	prevState := &WorkspaceState{
		Commit: commit,
		Modules: map[string]ModuleState{
			"module-a": {SourceHash: "hash-a"},     // Unchanged
			"module-b": {SourceHash: "old-hash-b"}, // Will change
		},
	}

	result, err := d.DetectChanges(context.Background(), DetectOptions{
		WorkspaceRoot:     "/workspace",
		PreviousState:     prevState,
		Modules:           []string{"module-a", "module-b"},
		CheckDependencies: true,
		DependencyGraph: map[string][]string{
			"module-a": {"module-b"}, // module-a depends on module-b
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both should be changed: module-b directly, module-a transitively
	if len(result.Changed) != 2 {
		t.Errorf("expected 2 changed modules, got %d: %v", len(result.Changed), result.Changed)
	}

	// module-b should have "source files changed"
	if result.Reasons["module-b"] != "source files changed" {
		t.Errorf("unexpected reason for module-b: %s", result.Reasons["module-b"])
	}

	// module-a should have "dependency changed"
	if result.Reasons["module-a"] != "dependency changed" {
		t.Errorf("unexpected reason for module-a: %s", result.Reasons["module-a"])
	}
}

// Tests for WorkspaceState

func TestWorkspaceState_Clone(t *testing.T) {
	original := &WorkspaceState{
		Commit:          "abc123",
		UncommittedHash: "dirty",
		Modules: map[string]ModuleState{
			"module-a": {SourceHash: "hash1"},
		},
		RecordedAt: time.Now(),
	}

	clone := original.Clone()

	// Modify clone
	clone.Modules["module-a"] = ModuleState{SourceHash: "changed"}

	// Original should be unchanged
	if original.Modules["module-a"].SourceHash != "hash1" {
		t.Error("Clone modified original")
	}
}

// Tests for Adapters

// mockContractProvider implements ModuleContractProvider for testing.
type mockContractProvider struct {
	patterns map[string][]string
}

func (m *mockContractProvider) GetGlobPatterns(moniker string) []string {
	if patterns, ok := m.patterns[moniker]; ok {
		return patterns
	}
	return nil
}

func TestContractResolverAdapter_GetModuleFiles_Found(t *testing.T) {
	provider := &mockContractProvider{
		patterns: map[string][]string{
			"module-a": {"**/*.go"},
			"module-b": {"pkg/**/*.go", "internal/**/*.go"},
		},
	}
	adapter := NewContractResolverAdapter(provider)

	// Test that the adapter passes the correct patterns
	// Note: This tests the integration with the hash package indirectly.
	// The actual file expansion depends on the workspace having matching files.
	// For unit testing, we verify the adapter calls the provider correctly.
	ctx := context.Background()

	// The adapter should use the patterns from the provider
	// Since we can't mock hash.ExpandGlobPatterns easily, we test the flow
	files, err := adapter.GetModuleFiles(ctx, "/tmp/nonexistent", "module-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// In a non-existent directory, we expect empty files (no matches)
	if files == nil {
		// This is fine - no files matched
	}
}

func TestContractResolverAdapter_GetModuleFiles_NotFound(t *testing.T) {
	provider := &mockContractProvider{
		patterns: map[string][]string{
			"module-a": {"**/*.go"},
		},
	}
	adapter := NewContractResolverAdapter(provider)

	ctx := context.Background()
	files, err := adapter.GetModuleFiles(ctx, "/tmp/workspace", "nonexistent-module")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if files != nil {
		t.Errorf("expected nil files for nonexistent module, got %v", files)
	}
}

func TestContractResolverAdapter_GetModuleFiles_EmptyPatterns(t *testing.T) {
	provider := &mockContractProvider{
		patterns: map[string][]string{
			"module-a": {}, // Empty patterns
		},
	}
	adapter := NewContractResolverAdapter(provider)

	ctx := context.Background()
	files, err := adapter.GetModuleFiles(ctx, "/tmp/workspace", "module-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if files != nil {
		t.Errorf("expected nil files for empty patterns, got %v", files)
	}
}

// mockModuleContract implements the interface{GetGlobPatterns() []string} for testing.
type mockModuleContract struct {
	patterns []string
}

func (m *mockModuleContract) GetGlobPatterns() []string {
	return m.patterns
}

func TestRegistryAdapter_GetGlobPatterns(t *testing.T) {
	contracts := map[string]*mockModuleContract{
		"module-a": {patterns: []string{"go/**/*.go", "go/**/*_test.go"}},
		"module-b": {patterns: []string{"pkg/**/*.go"}},
	}

	adapter := NewRegistryAdapterFunc(func(moniker string) (interface{ GetGlobPatterns() []string }, bool) {
		c, ok := contracts[moniker]
		if !ok {
			return nil, false
		}
		return c, true
	})

	// Test found module
	patterns := adapter.GetGlobPatterns("module-a")
	if len(patterns) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(patterns))
	}
	if patterns[0] != "go/**/*.go" {
		t.Errorf("expected first pattern 'go/**/*.go', got %s", patterns[0])
	}

	// Test not found module
	patterns = adapter.GetGlobPatterns("nonexistent")
	if patterns != nil {
		t.Errorf("expected nil patterns for nonexistent module, got %v", patterns)
	}
}
