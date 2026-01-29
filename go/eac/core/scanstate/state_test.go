package scanstate

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestDir creates a temporary directory with a mock workspace structure.
func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create out directory
	outDir := filepath.Join(dir, "out")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatalf("Failed to create out dir: %v", err)
	}

	return dir
}

// writeTestFile creates a test file with content.
func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create dir %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}

func TestLoad_NoStateFile(t *testing.T) {
	dir := setupTestDir(t)

	_, err := Load(dir)
	if err != ErrNoState {
		t.Errorf("Expected ErrNoState, got: %v", err)
	}
}

func TestLoad_ValidState(t *testing.T) {
	dir := setupTestDir(t)

	// Create a valid state file
	state := &State{
		Commit:          "abc123",
		UncommittedHash: "def456",
		Modules: map[string]ModuleState{
			"test-module": {
				SourceHash: "source-hash-1",
				Passed:     true,
				ScannedAt:  time.Now(),
				Files:      []string{"file1.go", "file2.go"},
			},
		},
		UpdatedAt: time.Now(),
	}

	if err := state.Save(dir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Load and verify
	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	if loaded.Commit != "abc123" {
		t.Errorf("Expected commit 'abc123', got: %s", loaded.Commit)
	}

	if loaded.UncommittedHash != "def456" {
		t.Errorf("Expected uncommitted hash 'def456', got: %s", loaded.UncommittedHash)
	}

	modState, ok := loaded.Modules["test-module"]
	if !ok {
		t.Fatal("Expected test-module in state")
	}

	if modState.SourceHash != "source-hash-1" {
		t.Errorf("Expected source hash 'source-hash-1', got: %s", modState.SourceHash)
	}

	if !modState.Passed {
		t.Error("Expected module to have passed")
	}
}

func TestSave_CreatesOutDir(t *testing.T) {
	dir := t.TempDir() // No out dir created yet

	state := &State{
		Commit:  "abc123",
		Modules: make(map[string]ModuleState),
	}

	if err := state.Save(dir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Verify file exists
	statePath := filepath.Join(dir, "out", ".scan-state.json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("Expected state file to be created")
	}
}

func TestClearState_RemovesFile(t *testing.T) {
	dir := setupTestDir(t)

	// Create state file
	state := &State{Commit: "abc123", Modules: make(map[string]ModuleState)}
	if err := state.Save(dir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Clear state
	if err := ClearState(dir); err != nil {
		t.Fatalf("Failed to clear state: %v", err)
	}

	// Verify file is gone
	_, err := Load(dir)
	if err != ErrNoState {
		t.Errorf("Expected ErrNoState after clear, got: %v", err)
	}
}

func TestClearState_NoFile(t *testing.T) {
	dir := setupTestDir(t)

	// Clear non-existent state should not error
	if err := ClearState(dir); err != nil {
		t.Errorf("ClearState should not error on missing file, got: %v", err)
	}
}

func TestDetectChanges_FreshRun(t *testing.T) {
	dir := setupTestDir(t)

	// Create test source files
	writeTestFile(t, filepath.Join(dir, "mod1", "main.go"), "package main")
	writeTestFile(t, filepath.Join(dir, "mod2", "main.go"), "package main")

	moduleFiles := map[string][]string{
		"mod1": {"mod1/main.go"},
		"mod2": {"mod2/main.go"},
	}

	result, err := DetectChanges(dir, moduleFiles)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	if !result.FreshRun {
		t.Error("Expected FreshRun to be true")
	}

	if len(result.ChangedModules) != 2 {
		t.Errorf("Expected 2 changed modules, got: %d", len(result.ChangedModules))
	}

	for _, mod := range result.ChangedModules {
		reason, ok := result.ChangeReasons[mod]
		if !ok {
			t.Errorf("Missing change reason for %s", mod)
		}
		if reason != "fresh scan (no prior state)" {
			t.Errorf("Unexpected change reason for %s: %s", mod, reason)
		}
	}
}

func TestDetectChanges_AllCached(t *testing.T) {
	dir := setupTestDir(t)

	// Create source files
	writeTestFile(t, filepath.Join(dir, "mod1", "main.go"), "package main")

	moduleFiles := map[string][]string{
		"mod1": {"mod1/main.go"},
	}

	// First, update state with current files (simulating a previous successful scan)
	scannedModules := map[string]bool{
		"mod1": true,
	}
	if err := UpdateModuleState(dir, scannedModules, moduleFiles); err != nil {
		t.Fatalf("UpdateModuleState failed: %v", err)
	}

	// Now detect changes - should be cached
	result, err := DetectChanges(dir, moduleFiles)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	if result.FreshRun {
		t.Error("Expected FreshRun to be false")
	}

	if len(result.ChangedModules) != 0 {
		t.Errorf("Expected 0 changed modules (all cached), got: %d", len(result.ChangedModules))
	}

	if len(result.UpToDateModules) != 1 {
		t.Errorf("Expected 1 up-to-date module, got: %d", len(result.UpToDateModules))
	}
}

func TestDetectChanges_SourceFileChanged(t *testing.T) {
	dir := setupTestDir(t)

	// Create source files
	writeTestFile(t, filepath.Join(dir, "mod1", "main.go"), "package main")

	moduleFiles := map[string][]string{
		"mod1": {"mod1/main.go"},
	}

	// Save initial state
	scannedModules := map[string]bool{
		"mod1": true,
	}
	if err := UpdateModuleState(dir, scannedModules, moduleFiles); err != nil {
		t.Fatalf("UpdateModuleState failed: %v", err)
	}

	// Modify source file
	writeTestFile(t, filepath.Join(dir, "mod1", "main.go"), "package main\n\nfunc foo() {}")

	// Detect changes - should detect modification
	result, err := DetectChanges(dir, moduleFiles)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	if len(result.ChangedModules) != 1 {
		t.Errorf("Expected 1 changed module, got: %d", len(result.ChangedModules))
	}

	reason := result.ChangeReasons["mod1"]
	if reason != "source files changed" {
		t.Errorf("Expected reason 'source files changed', got: %s", reason)
	}
}

func TestDetectChanges_NewModule(t *testing.T) {
	dir := setupTestDir(t)

	// Create source files for two modules
	writeTestFile(t, filepath.Join(dir, "mod1", "main.go"), "package main")
	writeTestFile(t, filepath.Join(dir, "mod2", "main.go"), "package main")

	// Save state for only mod1
	scannedModules := map[string]bool{
		"mod1": true,
	}
	moduleFilesPartial := map[string][]string{
		"mod1": {"mod1/main.go"},
	}
	if err := UpdateModuleState(dir, scannedModules, moduleFilesPartial); err != nil {
		t.Fatalf("UpdateModuleState failed: %v", err)
	}

	// Now request both modules - mod2 is new
	moduleFilesFull := map[string][]string{
		"mod1": {"mod1/main.go"},
		"mod2": {"mod2/main.go"},
	}
	result, err := DetectChanges(dir, moduleFilesFull)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	if len(result.ChangedModules) != 1 {
		t.Errorf("Expected 1 changed module (new one), got: %d", len(result.ChangedModules))
	}

	if result.ChangedModules[0] != "mod2" {
		t.Errorf("Expected mod2 to be changed, got: %s", result.ChangedModules[0])
	}

	reason := result.ChangeReasons["mod2"]
	if reason != "new module (not in previous scan)" {
		t.Errorf("Expected reason about new module, got: %s", reason)
	}
}

func TestDetectChanges_FailedScanRerun(t *testing.T) {
	dir := setupTestDir(t)

	// Create source files
	writeTestFile(t, filepath.Join(dir, "mod1", "main.go"), "package main")

	moduleFiles := map[string][]string{
		"mod1": {"mod1/main.go"},
	}

	// Save state with failed scan
	scannedModules := map[string]bool{
		"mod1": false, // Failed
	}
	if err := UpdateModuleState(dir, scannedModules, moduleFiles); err != nil {
		t.Fatalf("UpdateModuleState failed: %v", err)
	}

	// Detect changes - should re-run because previous scan failed
	result, err := DetectChanges(dir, moduleFiles)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	if len(result.ChangedModules) != 1 {
		t.Errorf("Expected 1 changed module (failed scan), got: %d", len(result.ChangedModules))
	}

	reason := result.ChangeReasons["mod1"]
	if reason != "previous scan had issues" {
		t.Errorf("Expected reason about failed scan, got: %s", reason)
	}
}

func TestUpdateModuleState_CreatesNewState(t *testing.T) {
	dir := setupTestDir(t)

	// Create source files
	writeTestFile(t, filepath.Join(dir, "mod1", "main.go"), "package main")

	moduleFiles := map[string][]string{
		"mod1": {"mod1/main.go"},
	}

	scannedModules := map[string]bool{
		"mod1": true,
	}

	if err := UpdateModuleState(dir, scannedModules, moduleFiles); err != nil {
		t.Fatalf("UpdateModuleState failed: %v", err)
	}

	// Load and verify
	state, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	modState, ok := state.Modules["mod1"]
	if !ok {
		t.Fatal("Expected mod1 in state")
	}

	if !modState.Passed {
		t.Error("Expected mod1 to have passed")
	}

	if modState.SourceHash == "" {
		t.Error("Expected mod1 to have a source hash")
	}
}

func TestUpdateModuleState_UpdatesExistingState(t *testing.T) {
	dir := setupTestDir(t)

	// Create source files
	writeTestFile(t, filepath.Join(dir, "mod1", "main.go"), "package main")

	moduleFiles := map[string][]string{
		"mod1": {"mod1/main.go"},
	}

	// First update: failed
	scannedModules1 := map[string]bool{
		"mod1": false,
	}
	if err := UpdateModuleState(dir, scannedModules1, moduleFiles); err != nil {
		t.Fatalf("UpdateModuleState failed: %v", err)
	}

	// Second update: passed now
	scannedModules2 := map[string]bool{
		"mod1": true,
	}
	if err := UpdateModuleState(dir, scannedModules2, moduleFiles); err != nil {
		t.Fatalf("UpdateModuleState failed: %v", err)
	}

	// Load and verify
	state, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !state.Modules["mod1"].Passed {
		t.Error("Expected mod1 to be updated to passed")
	}
}

func TestExpandGlobPatterns(t *testing.T) {
	dir := setupTestDir(t)

	// Create test files
	writeTestFile(t, filepath.Join(dir, "src", "main.go"), "package main")
	writeTestFile(t, filepath.Join(dir, "src", "util.go"), "package main")
	writeTestFile(t, filepath.Join(dir, "src", "sub", "nested.go"), "package sub")

	patterns := []string{
		filepath.Join(dir, "src", "**", "*.go"),
	}

	files, err := ExpandGlobPatterns(dir, patterns)
	if err != nil {
		t.Fatalf("ExpandGlobPatterns failed: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Expected 3 files, got: %d", len(files))
	}
}
