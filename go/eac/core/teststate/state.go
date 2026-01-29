// Package teststate manages incremental test state for detecting which modules need retesting.
// It uses file hashing and dependency propagation - if a dependency changes, all dependents need retesting.
//
// State is stored per-module in test.manifest.json files (out/test/<module>/test.manifest.json).
package teststate

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/ready-to-release/eac/go/eac/core/hash"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// ModuleTestState represents test state for a single module.
// This is now stored in each module's test.manifest.json.
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

// TestChangeResult represents the result of test change detection.
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

// ModuleTestFiles contains the file information for a module's tests.
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

// testManifestFileName is the name of the per-module test manifest file.
const testManifestFileName = "test.manifest.json"

// testManifest is a minimal struct for loading test manifest files.
// We only need the fields relevant for incremental test detection.
type testManifest struct {
	SourceHash   string                    `json:"source_hash"`
	BuildID      string                    `json:"build_id"`
	Dependencies []string                  `json:"dependencies"`
	Summary      testSummary               `json:"summary"`
	Suites       map[string]suiteRunResult `json:"suites"` // Track which suites have been run
}

// suiteRunResult tracks when a suite was run.
type suiteRunResult struct {
	RunTime string      `json:"run_time"`
	Tests   testSummary `json:"tests"`
}

type testSummary struct {
	Failed    int `json:"failed"`
	Undefined int `json:"undefined"`
	Pending   int `json:"pending"`
}

// hasSuite returns true if the specified suite has been run.
func (m *testManifest) hasSuite(suiteName string) bool {
	if m.Suites == nil {
		return false
	}
	_, exists := m.Suites[suiteName]
	return exists
}

// suitePassed returns true if the specified suite passed (no failures).
func (m *testManifest) suitePassed(suiteName string) bool {
	if m.Suites == nil {
		return false
	}
	suite, exists := m.Suites[suiteName]
	if !exists {
		return false
	}
	return suite.Tests.Failed == 0 && suite.Tests.Undefined == 0 && suite.Tests.Pending == 0
}

// loadModuleTestManifest loads a module's test manifest from its test output directory.
// Returns nil if no manifest exists (module not yet tested).
func loadModuleTestManifest(workspaceRoot, moniker string) *testManifest {
	moduleTestDir := paths.TestModuleDir(workspaceRoot, moniker)
	manifestPath := filepath.Join(moduleTestDir, testManifestFileName)

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil // No manifest = module not yet tested
	}

	var manifest testManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil // Invalid manifest = treat as not tested
	}

	return &manifest
}

// DependencyBuildIDLoader is a function that loads the current BuildID for a dependency module.
// This allows the detection logic to check if dependencies were rebuilt even if they're not
// in the current test run's moduleInfo.
type DependencyBuildIDLoader func(moniker string) string

// DetectChanges detects which modules need retesting based on file changes and dependency propagation.
// moduleInfo maps moniker -> ModuleTestFiles (source files, test files, dependencies)
// suiteNames is the list of test suites being run - detection checks ALL suites
// For composite suites (e.g., "unit+integration"), pass constituent suites: ["unit", "integration"]
//
// A module needs retesting if:
// 1. Its source files changed
// 2. Any of its dependencies (transitively) need retesting
// 3. Any of the specified suites hasn't been run before
// 4. Any suite's previous run failed
// 5. It's a new module (not in previous state)
// 6. Any dependency was rebuilt (BuildID changed) - even if not in current test run.
func DetectChanges(workspaceRoot string, moduleInfo map[string]ModuleTestFiles, suiteNames []string) (*TestChangeResult, error) {
	return DetectChangesWithLoader(workspaceRoot, moduleInfo, suiteNames, nil)
}

// DetectChangesWithLoader is like DetectChanges but accepts an optional loader function
// to check BuildIDs of dependencies that aren't in the current test run.
// State is now read from per-module test.manifest.json files instead of a global state file.
// suiteNames specifies which suites are being run - detection checks if ALL specified suites need to run.
func DetectChangesWithLoader(workspaceRoot string, moduleInfo map[string]ModuleTestFiles, suiteNames []string, loadDepBuildID DependencyBuildIDLoader) (*TestChangeResult, error) {
	start := time.Now()

	result := &TestChangeResult{
		ChangeReasons: make(map[string]string),
	}

	// Cache for loaded manifests
	manifestCache := make(map[string]*testManifest)
	getManifest := func(moniker string) *testManifest {
		if m, ok := manifestCache[moniker]; ok {
			return m
		}
		m := loadModuleTestManifest(workspaceRoot, moniker)
		manifestCache[moniker] = m
		return m
	}

	// Check if this is a fresh run (no manifests exist for any module)
	anyManifestExists := false
	for moniker := range moduleInfo {
		if getManifest(moniker) != nil {
			anyManifestExists = true
			break
		}
	}

	if !anyManifestExists {
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
		prevManifest := getManifest(moniker)

		if prevManifest == nil {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "new module (not previously tested)"
			continue
		}

		// Check if manifest has incremental state fields (SourceHash must be set)
		// If not, treat as if not tested (legacy manifest without incremental fields)
		if prevManifest.SourceHash == "" {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "no incremental state in manifest"
			continue
		}

		// Check if ALL specified suites have been run before and passed
		// For composite suites (e.g., "unit+integration"), suiteNames contains constituent suites
		missingSuites := []string{}
		failedSuites := []string{}
		for _, suiteName := range suiteNames {
			if !prevManifest.hasSuite(suiteName) {
				missingSuites = append(missingSuites, suiteName)
			} else if !prevManifest.suitePassed(suiteName) {
				failedSuites = append(failedSuites, suiteName)
			}
		}
		if len(missingSuites) > 0 {
			directlyChanged[moniker] = true
			if len(missingSuites) == 1 {
				result.ChangeReasons[moniker] = fmt.Sprintf("suite '%s' not previously run", missingSuites[0])
			} else {
				result.ChangeReasons[moniker] = fmt.Sprintf("suites not previously run: %s", strings.Join(missingSuites, ", "))
			}
			continue
		}
		if len(failedSuites) > 0 {
			directlyChanged[moniker] = true
			if len(failedSuites) == 1 {
				result.ChangeReasons[moniker] = fmt.Sprintf("suite '%s' previously failed", failedSuites[0])
			} else {
				result.ChangeReasons[moniker] = fmt.Sprintf("suites previously failed: %s", strings.Join(failedSuites, ", "))
			}
			continue
		}

		// Check if build changed (rebuild detected via manifest BuildID)
		// This ensures `build --rebuild` triggers `test --retest`
		if info.BuildID != "" && prevManifest.BuildID != "" && info.BuildID != prevManifest.BuildID {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "build changed (rebuild detected)"
			continue
		}

		// Check if module was tested without a build but now has one
		if info.BuildID != "" && prevManifest.BuildID == "" {
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
		if currentSourceHash != prevManifest.SourceHash {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "source files changed"
			continue
		}

		// Note: We don't check test_hash anymore since it was suite-specific and caused
		// false positives when switching suites. Suite tracking handles this instead.
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

		// Dependency not in current test run - check its BuildID against stored manifest
		if loadDepBuildID != nil {
			currentBuildID := loadDepBuildID(dep)
			prevManifest := getManifest(dep)
			if prevManifest != nil && currentBuildID != "" && prevManifest.BuildID != "" && currentBuildID != prevManifest.BuildID {
				depBuildIDChanged[dep] = true
				return true
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

// UpdateModuleState updates the incremental test state in each module's test.manifest.json.
// This should be called after tests complete to record state for change detection.
// It updates the SourceHash, TestHash, BuildID, and Dependencies fields in each manifest.
func UpdateModuleState(workspaceRoot string, testedModules map[string]bool, moduleInfo map[string]ModuleTestFiles) error {
	// Use default empty suite name for backwards compatibility
	return UpdateModuleStateForSuite(workspaceRoot, testedModules, moduleInfo, "")
}

// UpdateModuleStateForSuite updates the incremental test state with suite information.
// suiteName specifies which suite was run - this is critical for change detection.
// Pass empty string for backwards compatibility (won't record suite info).
func UpdateModuleStateForSuite(workspaceRoot string, testedModules map[string]bool, moduleInfo map[string]ModuleTestFiles, suiteName string) error {
	gitCommit, err := getGitCommit(workspaceRoot)
	if err != nil {
		gitCommit = "" // Use empty if git is unavailable
	}

	// Update each tested module's manifest
	for moniker, passed := range testedModules {
		info, ok := moduleInfo[moniker]
		if !ok {
			continue
		}

		moduleTestDir := paths.TestModuleDir(workspaceRoot, moniker)
		manifestPath := filepath.Join(moduleTestDir, testManifestFileName)

		// Ensure directory exists
		if err := os.MkdirAll(moduleTestDir, 0o755); err != nil {
			continue
		}

		// Load existing manifest or create minimal one
		var manifest map[string]interface{}
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			// No manifest - create a minimal one for incremental state
			manifest = map[string]interface{}{
				"moniker":   moniker,
				"test_time": time.Now().Format(time.RFC3339),
			}
		} else {
			// Parse existing manifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				// Invalid manifest - create fresh
				manifest = map[string]interface{}{
					"moniker":   moniker,
					"test_time": time.Now().Format(time.RFC3339),
				}
			}
		}

		// Calculate and update incremental state fields
		sourceHash, hashErr := hashFiles(workspaceRoot, info.SourceFiles)
		if hashErr != nil {
			sourceHash = "" // Skip incremental detection if hashing fails
		}
		testHash, hashErr := hashFiles(workspaceRoot, info.TestFiles)
		if hashErr != nil {
			testHash = "" // Skip incremental detection if hashing fails
		}

		manifest["source_hash"] = sourceHash
		manifest["test_hash"] = testHash
		manifest["build_id"] = info.BuildID
		manifest["dependencies"] = info.Dependencies
		manifest["git_commit"] = gitCommit
		manifest["test_time"] = time.Now().Format(time.RFC3339)

		// Update suite results if suite name provided
		if suiteName != "" {
			suites, ok := manifest["suites"].(map[string]interface{})
			if !ok {
				suites = make(map[string]interface{})
			}

			// Calculate test summary based on pass/fail
			testSummary := map[string]interface{}{
				"total":   0,
				"passed":  0,
				"failed":  0,
				"skipped": 0,
			}
			if passed {
				testSummary["passed"] = 1
				testSummary["total"] = 1
			} else {
				testSummary["failed"] = 1
				testSummary["total"] = 1
			}

			// Handle composite suites (e.g., "unit+integration")
			// Record each constituent suite separately for proper change detection
			suiteNamesToRecord := []string{suiteName}
			if strings.Contains(suiteName, "+") {
				suiteNamesToRecord = strings.Split(suiteName, "+")
			}

			for _, name := range suiteNamesToRecord {
				suites[name] = map[string]interface{}{
					"run_time":         time.Now().Format(time.RFC3339),
					"duration_seconds": 0,
					"tests":            testSummary,
				}
			}
			manifest["suites"] = suites
		}

		// Write updated manifest
		updatedData, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			continue
		}

		if err := os.WriteFile(manifestPath, updatedData, 0o644); err != nil {
			return fmt.Errorf("failed to update manifest for %s: %w", moniker, err)
		}
	}

	return nil
}

// ClearState removes test manifests for all modules under out/test/ (for --retest).
// This forces a full retest by removing the incremental state.
func ClearState(workspaceRoot string) error {
	testDir := filepath.Join(workspaceRoot, paths.OutTestRelPath)

	// Walk the test output directory and remove all test.manifest.json files
	err := filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() && info.Name() == testManifestFileName {
			os.Remove(path) // Ignore errors on individual files
		}
		return nil
	})

	if os.IsNotExist(err) {
		return nil // No test directory = nothing to clear
	}
	return err
}

// getGitCommit returns the current HEAD commit SHA.
func getGitCommit(workspaceRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = workspaceRoot
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// hashFiles computes a hash of all files.
// Delegates to core/hash package for the actual hashing.
func hashFiles(workspaceRoot string, files []string) (string, error) {
	return hash.Files(workspaceRoot, files)
}

// ExpandGlobPatterns expands glob patterns to actual file paths
// Returns files relative to workspaceRoot.
func ExpandGlobPatterns(workspaceRoot string, patterns []string) ([]string, error) {
	var result []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		absPattern := pattern
		if !filepath.IsAbs(pattern) {
			absPattern = filepath.Join(workspaceRoot, pattern)
		}

		// Use doublestar for glob expansion to support ** patterns
		matches, err := doublestar.FilepathGlob(absPattern)
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
