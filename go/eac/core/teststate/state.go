// Package teststate manages incremental test state for detecting which modules need retesting.
// It uses file hashing and dependency propagation - if a dependency changes, all dependents need retesting.
package teststate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// State represents the test state for incremental testing
type State struct {
	// Git commit SHA at time of last test
	Commit string `json:"commit"`

	// Hash of uncommitted changes at test time (empty if clean)
	UncommittedHash string `json:"uncommitted_hash,omitempty"`

	// Per-module test state
	Modules map[string]ModuleTestState `json:"modules"`

	// Timestamp of last state update
	UpdatedAt time.Time `json:"updated_at"`
}

// ModuleTestState represents test state for a single module
type ModuleTestState struct {
	// Hash of all source files in the module
	SourceHash string `json:"source_hash"`

	// Hash of all test files in the module
	TestHash string `json:"test_hash"`

	// Build ID from manifest at time of test (links tests to specific build)
	BuildID string `json:"build_id,omitempty"`

	// Whether the last test run passed
	Passed bool `json:"passed"`

	// Timestamp of last test run
	TestedAt time.Time `json:"tested_at"`

	// Dependencies at time of last test (for debugging)
	Dependencies []string `json:"dependencies,omitempty"`
}

// TestChangeResult represents the result of test change detection
type TestChangeResult struct {
	// Modules that need retesting (directly changed or affected by dependency changes)
	ModulesNeedingTest []string

	// Modules that are up-to-date (no changes, previous tests passed)
	UpToDateModules []string

	// Reason for each module needing test
	ChangeReasons map[string]string

	// Whether this is a fresh test run (no prior state)
	FreshRun bool

	// Detection time
	DetectionTime time.Duration
}

// ModuleTestFiles contains the file information for a module's tests
type ModuleTestFiles struct {
	// Source files (the code being tested)
	SourceFiles []string

	// Test files (the tests themselves)
	TestFiles []string

	// Direct dependencies (module monikers this module depends on)
	Dependencies []string

	// BuildID from the module's build manifest (empty if no manifest)
	// Used to detect when a module was rebuilt and needs retesting
	BuildID string
}

const (
	stateFileName = "test-suite.state.json"
)

// Load loads test state from the test output directory
func Load(workspaceRoot string) (*State, error) {
	statePath := paths.TestStatePath(workspaceRoot, stateFileName)

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No state file = fresh test run
		}
		return nil, fmt.Errorf("failed to read test state: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse test state: %w", err)
	}

	return &state, nil
}

// Save saves test state to the test output directory
func (s *State) Save(workspaceRoot string) error {
	testDir := filepath.Join(workspaceRoot, paths.OutTestRelPath)
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	s.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test state: %w", err)
	}

	statePath := filepath.Join(testDir, stateFileName)
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write test state: %w", err)
	}

	return nil
}

// DependencyBuildIDLoader is a function that loads the current BuildID for a dependency module.
// This allows the detection logic to check if dependencies were rebuilt even if they're not
// in the current test run's moduleInfo.
type DependencyBuildIDLoader func(moniker string) string

// DetectChanges detects which modules need retesting based on file changes and dependency propagation.
// moduleInfo maps moniker -> ModuleTestFiles (source files, test files, dependencies)
//
// A module needs retesting if:
// 1. Its source files changed
// 2. Its test files changed
// 3. Any of its dependencies (transitively) need retesting
// 4. Its previous test run failed
// 5. It's a new module (not in previous state)
// 6. Any dependency was rebuilt (BuildID changed) - even if not in current test run
func DetectChanges(workspaceRoot string, moduleInfo map[string]ModuleTestFiles) (*TestChangeResult, error) {
	return DetectChangesWithLoader(workspaceRoot, moduleInfo, nil)
}

// DetectChangesWithLoader is like DetectChanges but accepts an optional loader function
// to check BuildIDs of dependencies that aren't in the current test run.
func DetectChangesWithLoader(workspaceRoot string, moduleInfo map[string]ModuleTestFiles, loadDepBuildID DependencyBuildIDLoader) (*TestChangeResult, error) {
	start := time.Now()

	result := &TestChangeResult{
		ChangeReasons: make(map[string]string),
	}

	// Load previous state
	prevState, err := Load(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load test state: %w", err)
	}

	if prevState == nil {
		// Fresh test run - all modules need testing
		result.FreshRun = true
		for moniker := range moduleInfo {
			result.ModulesNeedingTest = append(result.ModulesNeedingTest, moniker)
			result.ChangeReasons[moniker] = "fresh run (no prior state)"
		}
		sort.Strings(result.ModulesNeedingTest)
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// First pass: detect directly changed modules (source or test files changed, or build changed)
	directlyChanged := make(map[string]bool)
	for moniker, info := range moduleInfo {
		prevModState, exists := prevState.Modules[moniker]

		if !exists {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "new module (not in previous test)"
			continue
		}

		// Check if previous test failed
		if !prevModState.Passed {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "previous test failed"
			continue
		}

		// Check if build changed (rebuild detected via manifest BuildID)
		// This ensures `build --rebuild` triggers `test --retest`
		if info.BuildID != "" && prevModState.BuildID != "" && info.BuildID != prevModState.BuildID {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "build changed (rebuild detected)"
			continue
		}

		// Check if module was tested without a build but now has one
		if info.BuildID != "" && prevModState.BuildID == "" {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "new build detected"
			continue
		}

		// Calculate current source hash
		currentSourceHash, err := hashFiles(workspaceRoot, info.SourceFiles)
		if err != nil {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = fmt.Sprintf("source hash error: %v", err)
			continue
		}

		// Check source files changed
		if currentSourceHash != prevModState.SourceHash {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "source files changed"
			continue
		}

		// Calculate current test hash
		currentTestHash, err := hashFiles(workspaceRoot, info.TestFiles)
		if err != nil {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = fmt.Sprintf("test hash error: %v", err)
			continue
		}

		// Check test files changed
		if currentTestHash != prevModState.TestHash {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "test files changed"
			continue
		}
	}

	// Second pass: propagate changes through dependency chain
	// Build dependency graph and memoization cache
	needsTest := make(map[string]bool)
	checked := make(map[string]bool)
	depBuildIDChanged := make(map[string]bool) // Cache for dependency BuildID checks

	// Helper to check if a dependency's BuildID changed (even if not in moduleInfo)
	checkDepBuildIDChanged := func(dep string) bool {
		if changed, ok := depBuildIDChanged[dep]; ok {
			return changed
		}

		// If dependency is in moduleInfo, we already checked it in first pass
		if _, inModuleInfo := moduleInfo[dep]; inModuleInfo {
			depBuildIDChanged[dep] = directlyChanged[dep]
			return directlyChanged[dep]
		}

		// Dependency not in current test run - check its BuildID against stored state
		if loadDepBuildID != nil && prevState != nil {
			currentBuildID := loadDepBuildID(dep)
			if prevDepState, exists := prevState.Modules[dep]; exists {
				if currentBuildID != "" && prevDepState.BuildID != "" && currentBuildID != prevDepState.BuildID {
					depBuildIDChanged[dep] = true
					return true
				}
			}
		}

		depBuildIDChanged[dep] = false
		return false
	}

	var checkNeedsTest func(moniker string) bool
	checkNeedsTest = func(moniker string) bool {
		// Already computed
		if result, ok := needsTest[moniker]; ok {
			return result
		}

		// Prevent infinite recursion
		if checked[moniker] {
			return false
		}
		checked[moniker] = true

		// Directly changed modules need testing
		if directlyChanged[moniker] {
			needsTest[moniker] = true
			return true
		}

		// Check if any dependency needs testing (recursively)
		info, exists := moduleInfo[moniker]
		if !exists {
			needsTest[moniker] = false
			return false
		}

		for _, dep := range info.Dependencies {
			// First check if dependency's BuildID changed (even if not in moduleInfo)
			if checkDepBuildIDChanged(dep) {
				needsTest[moniker] = true
				if result.ChangeReasons[moniker] == "" {
					result.ChangeReasons[moniker] = fmt.Sprintf("dependency rebuilt: %s", dep)
				}
				return true
			}

			// Then check if dependency needs testing (recursive)
			if checkNeedsTest(dep) {
				needsTest[moniker] = true
				if result.ChangeReasons[moniker] == "" {
					result.ChangeReasons[moniker] = fmt.Sprintf("dependency changed: %s", dep)
				}
				return true
			}
		}

		needsTest[moniker] = false
		return false
	}

	// Check all modules
	for moniker := range moduleInfo {
		if checkNeedsTest(moniker) {
			result.ModulesNeedingTest = append(result.ModulesNeedingTest, moniker)
		} else {
			result.UpToDateModules = append(result.UpToDateModules, moniker)
		}
	}

	sort.Strings(result.ModulesNeedingTest)
	sort.Strings(result.UpToDateModules)
	result.DetectionTime = time.Since(start)

	return result, nil
}

// UpdateModuleState updates the test state for modules that were tested
func UpdateModuleState(workspaceRoot string, testedModules map[string]bool, moduleInfo map[string]ModuleTestFiles) error {
	// Load or create state
	state, err := Load(workspaceRoot)
	if err != nil {
		return err
	}
	if state == nil {
		state = &State{
			Modules: make(map[string]ModuleTestState),
		}
	}

	// Update git state
	state.Commit, _ = getGitCommit(workspaceRoot)
	uncommittedFiles, _ := getUncommittedFiles(workspaceRoot)
	if len(uncommittedFiles) > 0 {
		state.UncommittedHash = hashUncommittedFiles(workspaceRoot, uncommittedFiles)
	} else {
		state.UncommittedHash = ""
	}

	// Update each tested module
	for moniker, passed := range testedModules {
		info, ok := moduleInfo[moniker]
		if !ok {
			continue
		}

		sourceHash, _ := hashFiles(workspaceRoot, info.SourceFiles)
		testHash, _ := hashFiles(workspaceRoot, info.TestFiles)

		state.Modules[moniker] = ModuleTestState{
			SourceHash:   sourceHash,
			TestHash:     testHash,
			BuildID:      info.BuildID, // Record which build was tested
			Passed:       passed,
			TestedAt:     time.Now(),
			Dependencies: info.Dependencies,
		}
	}

	return state.Save(workspaceRoot)
}

// ClearState removes the test state file (for --retest)
func ClearState(workspaceRoot string) error {
	statePath := paths.TestStatePath(workspaceRoot, stateFileName)
	err := os.Remove(statePath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// getGitCommit returns the current HEAD commit SHA
func getGitCommit(workspaceRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = workspaceRoot
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// getUncommittedFiles returns list of uncommitted file paths (relative to repo root)
func getUncommittedFiles(workspaceRoot string) ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = workspaceRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var files []string
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		path = strings.Trim(path, "\"")
		files = append(files, path)
	}

	return files, nil
}

// hashUncommittedFiles creates a hash representing the uncommitted state
func hashUncommittedFiles(workspaceRoot string, files []string) string {
	h := sha256.New()

	sorted := make([]string, len(files))
	copy(sorted, files)
	sort.Strings(sorted)

	for _, file := range sorted {
		path := filepath.Join(workspaceRoot, file)
		f, err := os.Open(path)
		if err != nil {
			h.Write([]byte(file + ":deleted\n"))
			continue
		}
		io.Copy(h, f)
		f.Close()
	}

	return hex.EncodeToString(h.Sum(nil))[:16]
}

// hashFiles computes a hash of all files
func hashFiles(workspaceRoot string, files []string) (string, error) {
	if len(files) == 0 {
		return "", nil
	}

	h := sha256.New()

	sorted := make([]string, len(files))
	copy(sorted, files)
	sort.Strings(sorted)

	for _, file := range sorted {
		path := filepath.Join(workspaceRoot, file)
		f, err := os.Open(path)
		if err != nil {
			return "", fmt.Errorf("failed to open %s: %w", file, err)
		}

		h.Write([]byte(file + "\n"))
		io.Copy(h, f)
		f.Close()
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// ExpandGlobPatterns expands glob patterns to actual file paths
// Returns files relative to workspaceRoot
func ExpandGlobPatterns(workspaceRoot string, patterns []string) ([]string, error) {
	var result []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		absPattern := pattern
		if !filepath.IsAbs(pattern) {
			absPattern = filepath.Join(workspaceRoot, pattern)
		}

		matches, err := filepath.Glob(absPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %s: %w", pattern, err)
		}

		for _, match := range matches {
			rel, err := filepath.Rel(workspaceRoot, match)
			if err != nil {
				continue
			}
			rel = filepath.ToSlash(rel)

			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}

			if !seen[rel] {
				seen[rel] = true
				result = append(result, rel)
			}
		}
	}

	sort.Strings(result)
	return result, nil
}
