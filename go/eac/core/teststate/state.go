// Package teststate manages incremental test state for detecting which modules need retesting.
// It uses file hashing and test-set-aware dependency propagation:
// - Unit tests (L0/L1): Only invalidated by direct module changes
// - Integration tests (L2+): Invalidated by direct AND transitive dependency changes
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

// TestChangeResult represents the result of test change detection.
type TestChangeResult struct {
	// Test set this result applies to
	TestSet TestSet

	// Modules that need retesting
	ModulesNeedingTest []string

	// Modules that are up-to-date
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
	BuildID string
}

// testManifestFileName is the name of the per-module test manifest file.
const testManifestFileName = "test.manifest.json"

// testManifest is a minimal struct for loading test manifest files.
type testManifest struct {
	SourceHash   string                      `json:"source_hash"`
	BuildID      string                      `json:"build_id"`
	Dependencies []string                    `json:"dependencies"`
	TestSets     map[string]*testSetManifest `json:"test_sets"`
}

// testSetManifest tracks state for a specific test set (unit or integration).
type testSetManifest struct {
	SourceHash          string  `json:"source_hash"`
	TestSetHash         string  `json:"testset_hash"`
	BuildID             string  `json:"build_id"`
	Passed              bool    `json:"passed"`
	TestedAt            string  `json:"tested_at"`
	DependencyBuildHash string  `json:"dependency_build_hash,omitempty"`
	TestCount           int     `json:"test_count"`
	DurationSeconds     float64 `json:"duration_seconds"`
}

// loadModuleTestManifest loads a module's test manifest from its test output directory.
func loadModuleTestManifest(workspaceRoot, moniker string) *testManifest {
	moduleTestDir := paths.TestModuleDir(workspaceRoot, moniker)
	manifestPath := filepath.Join(moduleTestDir, testManifestFileName)

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil
	}

	var manifest testManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil
	}

	return &manifest
}

// DependencyBuildIDLoader loads the current BuildID for a dependency module.
type DependencyBuildIDLoader func(moniker string) string

// DetectChanges detects which modules need retesting based on test set invalidation rules.
// testSet determines invalidation behavior:
// - TestSetUnit: Only direct module changes trigger retest (isolated tests)
// - TestSetIntegration: Direct AND transitive dependency changes trigger retest
func DetectChanges(workspaceRoot string, moduleInfo map[string]ModuleTestFiles, testSet TestSet, loadDepBuildID DependencyBuildIDLoader) (*TestChangeResult, error) {
	start := time.Now()
	rule := DefaultInvalidationRules()[testSet]

	result := &TestChangeResult{
		TestSet:       testSet,
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

	// Check if this is a fresh run
	anyManifestExists := false
	for moniker := range moduleInfo {
		if getManifest(moniker) != nil {
			anyManifestExists = true
			break
		}
	}

	if !anyManifestExists {
		result.FreshRun = true
		for moniker := range moduleInfo {
			result.ModulesNeedingTest = append(result.ModulesNeedingTest, moniker)
			result.ChangeReasons[moniker] = "fresh run (no prior state)"
		}
		sort.Strings(result.ModulesNeedingTest)
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Phase 1: Detect directly changed modules
	directlyChanged := make(map[string]bool)
	for moniker, info := range moduleInfo {
		manifest := getManifest(moniker)

		if manifest == nil {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "new module (not previously tested)"
			continue
		}

		// Check test set state
		testSetState := manifest.TestSets[string(testSet)]
		if testSetState == nil {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = fmt.Sprintf("test set '%s' not previously run", testSet)
			continue
		}

		// Check if previous run failed
		if !testSetState.Passed {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "previous test failed"
			continue
		}

		// Check build change
		if rule.OnBuildChange && info.BuildID != "" && testSetState.BuildID != "" && info.BuildID != testSetState.BuildID {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "build changed (rebuild detected)"
			continue
		}

		// Check source files changed
		currentSourceHash, err := hashFiles(workspaceRoot, info.SourceFiles)
		if err != nil {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = fmt.Sprintf("source hash error: %v", err)
			continue
		}

		if rule.OnModuleChange && currentSourceHash != testSetState.SourceHash {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "source files changed"
			continue
		}
	}

	// Phase 2: Apply invalidation rules
	if !rule.OnTransitiveChange {
		// Unit tests: only directly changed modules
		for moniker := range moduleInfo {
			if directlyChanged[moniker] {
				result.ModulesNeedingTest = append(result.ModulesNeedingTest, moniker)
			} else {
				result.UpToDateModules = append(result.UpToDateModules, moniker)
			}
		}
	} else {
		// Integration tests: propagate through dependency chain
		needsTest := make(map[string]bool)
		checked := make(map[string]bool)

		var checkNeedsTest func(moniker string) bool
		checkNeedsTest = func(moniker string) bool {
			if result, ok := needsTest[moniker]; ok {
				return result
			}
			if checked[moniker] {
				return false
			}
			checked[moniker] = true

			if directlyChanged[moniker] {
				needsTest[moniker] = true
				return true
			}

			info, exists := moduleInfo[moniker]
			if !exists {
				needsTest[moniker] = false
				return false
			}

			// Check dependencies
			for _, dep := range info.Dependencies {
				// Check if dependency's BuildID changed
				if rule.OnDependencyBuildChange && loadDepBuildID != nil {
					currentBuildID := loadDepBuildID(dep)
					depManifest := getManifest(dep)
					if depManifest != nil && currentBuildID != "" && depManifest.BuildID != "" && currentBuildID != depManifest.BuildID {
						needsTest[moniker] = true
						if result.ChangeReasons[moniker] == "" {
							result.ChangeReasons[moniker] = fmt.Sprintf("dependency rebuilt: %s", dep)
						}
						return true
					}
				}

				// Recursively check dependency
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

		for moniker := range moduleInfo {
			if checkNeedsTest(moniker) {
				result.ModulesNeedingTest = append(result.ModulesNeedingTest, moniker)
			} else {
				result.UpToDateModules = append(result.UpToDateModules, moniker)
			}
		}
	}

	sort.Strings(result.ModulesNeedingTest)
	sort.Strings(result.UpToDateModules)
	result.DetectionTime = time.Since(start)

	return result, nil
}

// UpdateModuleTestSetState updates the test state for a specific test set.
func UpdateModuleTestSetState(workspaceRoot string, testedModules map[string]bool, moduleInfo map[string]ModuleTestFiles, testSet TestSet, dependencyBuildIDs map[string]map[string]string) error {
	gitCommit, err := getGitCommit(workspaceRoot)
	if err != nil {
		gitCommit = ""
	}

	for moniker, passed := range testedModules {
		info, ok := moduleInfo[moniker]
		if !ok {
			continue
		}

		moduleTestDir := paths.TestModuleDir(workspaceRoot, moniker)
		manifestPath := filepath.Join(moduleTestDir, testManifestFileName)

		if err := os.MkdirAll(moduleTestDir, 0o755); err != nil {
			continue
		}

		// Load or create manifest
		var manifest map[string]interface{}
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			manifest = map[string]interface{}{
				"moniker":   moniker,
				"test_time": time.Now().Format(time.RFC3339),
			}
		} else {
			if err := json.Unmarshal(data, &manifest); err != nil {
				manifest = map[string]interface{}{
					"moniker":   moniker,
					"test_time": time.Now().Format(time.RFC3339),
				}
			}
		}

		// Calculate hashes
		sourceHash, _ := hashFiles(workspaceRoot, info.SourceFiles)
		testHash, _ := hashFiles(workspaceRoot, info.TestFiles)

		// Update top-level fields
		manifest["source_hash"] = sourceHash
		manifest["build_id"] = info.BuildID
		manifest["dependencies"] = info.Dependencies
		manifest["git_commit"] = gitCommit
		manifest["test_time"] = time.Now().Format(time.RFC3339)

		// Update test_sets
		testSets, ok := manifest["test_sets"].(map[string]interface{})
		if !ok {
			testSets = make(map[string]interface{})
		}

		// Compute dependency build hash for integration tests
		depBuildHash := ""
		if testSet == TestSetIntegration {
			if depIDs, ok := dependencyBuildIDs[moniker]; ok {
				depBuildHash = ComputeDependencyBuildHash(depIDs)
			}
		}

		testSets[string(testSet)] = map[string]interface{}{
			"source_hash":           sourceHash,
			"testset_hash":          testHash,
			"build_id":              info.BuildID,
			"passed":                passed,
			"tested_at":             time.Now().Format(time.RFC3339),
			"dependency_build_hash": depBuildHash,
		}
		manifest["test_sets"] = testSets

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

// ClearState removes test manifests for all modules under out/test/.
func ClearState(workspaceRoot string) error {
	testDir := filepath.Join(workspaceRoot, paths.OutTestRelPath)

	err := filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == testManifestFileName {
			os.Remove(path)
		}
		return nil
	})

	if os.IsNotExist(err) {
		return nil
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
func hashFiles(workspaceRoot string, files []string) (string, error) {
	return hash.Files(workspaceRoot, files)
}

// ExpandGlobPatterns expands glob patterns to actual file paths.
func ExpandGlobPatterns(workspaceRoot string, patterns []string) ([]string, error) {
	var result []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		absPattern := pattern
		if !filepath.IsAbs(pattern) {
			absPattern = filepath.Join(workspaceRoot, pattern)
		}

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
