// Package scanstate manages incremental scan state for detecting which modules need rescanning.
// It uses a hybrid git + file hash approach for fast and accurate change detection.
//
// State is stored in out/.scan-state.json (global file, similar to buildstate).
package scanstate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/ready-to-release/eac/go/eac/core/hash"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

var log = logging.C()

// ErrNoState is returned when no scan state file exists (fresh scan).
var ErrNoState = errors.New("no scan state file")

// State represents the scan state for incremental scanning.
type State struct {
	// Git commit SHA at time of last scan
	Commit string `json:"commit"`

	// Hash of uncommitted changes at scan time (empty if clean)
	UncommittedHash string `json:"uncommitted_hash,omitempty"`

	// Per-module scan state
	Modules map[string]ModuleState `json:"modules"`

	// Timestamp of last state update
	UpdatedAt time.Time `json:"updated_at"`
}

// ModuleState represents scan state for a single module.
type ModuleState struct {
	// Hash of all source files in the module
	SourceHash string `json:"source_hash"`

	// Whether the last scan passed (all scanners passed)
	Passed bool `json:"passed"`

	// Timestamp of successful scan
	ScannedAt time.Time `json:"scanned_at"`

	// List of file paths included in hash (for debugging)
	Files []string `json:"files,omitempty"`
}

// ChangeResult represents the result of change detection.
type ChangeResult struct {
	// Modules that need rescanning (changed or new)
	ChangedModules []string

	// Modules that are up-to-date
	UpToDateModules []string

	// Reason for each changed module
	ChangeReasons map[string]string

	// Whether this is a fresh scan (no prior state)
	FreshRun bool

	// Detection time
	DetectionTime time.Duration
}

const (
	stateFileName = ".scan-state.json"
)

// Load loads scan state from the out directory.
// Returns ErrNoState if no state file exists (fresh scan).
func Load(workspaceRoot string) (*State, error) {
	statePath := filepath.Join(workspaceRoot, paths.OutDir, stateFileName)

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoState
		}
		return nil, fmt.Errorf("failed to read scan state: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		// Treat corrupted state as fresh scan
		return nil, ErrNoState
	}

	return &state, nil
}

// Save saves scan state to the out directory.
func (s *State) Save(workspaceRoot string) error {
	outDir := filepath.Join(workspaceRoot, paths.OutDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil { //nolint:gosec // G301: Output should be world-readable
		return fmt.Errorf("failed to create out directory: %w", err)
	}

	s.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scan state: %w", err)
	}

	statePath := filepath.Join(outDir, stateFileName)
	if err := os.WriteFile(statePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write scan state: %w", err)
	}

	return nil
}

// DetectChanges detects which modules need rescanning based on file changes.
// moduleFiles maps moniker -> list of source file paths (relative to workspaceRoot)
//
// Detection strategy (same as buildstate):
//  1. Fast path: If git commit + uncommitted state matches previous scan exactly,
//     AND all modules were in the previous scan and passed, trust the stored hashes
//  2. Slow path: Hash source files and compare to stored hashes
//
// A module needs rescanning if:
// 1. Its source files changed
// 2. It's a new module (not in previous state)
// 3. Its previous scan failed (Passed=false).
func DetectChanges(workspaceRoot string, moduleFiles map[string][]string) (*ChangeResult, error) {
	start := time.Now()

	result := &ChangeResult{
		ChangeReasons: make(map[string]string),
	}

	// Load previous state
	prevState, err := Load(workspaceRoot)
	if err != nil {
		if errors.Is(err, ErrNoState) {
			// Fresh scan - all modules need scanning
			result.FreshRun = true
			for moniker := range moduleFiles {
				result.ChangedModules = append(result.ChangedModules, moniker)
				result.ChangeReasons[moniker] = "fresh scan (no prior state)"
			}
			sort.Strings(result.ChangedModules)
			result.DetectionTime = time.Since(start)
			return result, nil
		}
		return nil, fmt.Errorf("failed to load scan state: %w", err)
	}

	// Get current git state for fast-path optimization
	currentCommit, gitErr := getGitCommit(workspaceRoot)
	if gitErr != nil {
		currentCommit = ""
	}
	uncommittedFiles, gitErr := getUncommittedFiles(workspaceRoot)
	if gitErr != nil {
		uncommittedFiles = nil
	}

	// Calculate current uncommitted hash
	currentUncommittedHash := ""
	if len(uncommittedFiles) > 0 {
		currentUncommittedHash = hashUncommittedFiles(workspaceRoot, uncommittedFiles)
	}

	// Fast path check
	gitStateMatches := currentCommit != "" &&
		currentCommit == prevState.Commit &&
		currentUncommittedHash == prevState.UncommittedHash

	log.Debugf("[SCAN-CACHE] Git state: current=%s prev=%s match=%v",
		currentCommit, prevState.Commit, currentCommit == prevState.Commit)
	log.Debugf("[SCAN-CACHE] Uncommitted hash: current=%s prev=%s match=%v",
		currentUncommittedHash, prevState.UncommittedHash, currentUncommittedHash == prevState.UncommittedHash)

	// Check if fast path is possible (all modules must have prior state and passed)
	canUseFastPath := gitStateMatches
	if canUseFastPath {
		for moniker := range moduleFiles {
			modState, exists := prevState.Modules[moniker]
			if !exists || !modState.Passed {
				log.Debugf("[SCAN-CACHE] Fast path blocked: module %s not in state or failed", moniker)
				canUseFastPath = false
				break
			}
		}
	}

	log.Debugf("[SCAN-CACHE] Fast path: gitStateMatches=%v canUseFastPath=%v", gitStateMatches, canUseFastPath)

	if canUseFastPath {
		// Fast path: git state unchanged, all modules have prior hashes and passed
		log.Debugf("[SCAN-CACHE] Using fast path - all %d modules up to date", len(moduleFiles))
		for moniker := range moduleFiles {
			result.UpToDateModules = append(result.UpToDateModules, moniker)
		}
		sort.Strings(result.UpToDateModules)
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Slow path: Hash source files for each module and compare
	log.Debugf("[SCAN-CACHE] Using slow path - hashing %d modules", len(moduleFiles))
	for moniker, files := range moduleFiles {
		prevModState, exists := prevState.Modules[moniker]

		if !exists {
			result.ChangedModules = append(result.ChangedModules, moniker)
			result.ChangeReasons[moniker] = "new module (not in previous scan)"
			continue
		}

		// If previous scan failed, needs rescanning
		if !prevModState.Passed {
			result.ChangedModules = append(result.ChangedModules, moniker)
			result.ChangeReasons[moniker] = "previous scan had issues"
			continue
		}

		// Calculate current source hash
		currentHash, err := hashModuleFiles(workspaceRoot, files)
		if err != nil {
			result.ChangedModules = append(result.ChangedModules, moniker)
			result.ChangeReasons[moniker] = fmt.Sprintf("hash error: %v", err)
			continue
		}

		if currentHash != prevModState.SourceHash {
			result.ChangedModules = append(result.ChangedModules, moniker)
			result.ChangeReasons[moniker] = "source files changed"
			log.Debugf("[SCAN-CACHE] Hash mismatch for %s: current=%s prev=%s",
				moniker, currentHash[:16], prevModState.SourceHash[:16])
		} else {
			result.UpToDateModules = append(result.UpToDateModules, moniker)
		}
	}

	sort.Strings(result.ChangedModules)
	sort.Strings(result.UpToDateModules)
	result.DetectionTime = time.Since(start)

	return result, nil
}

// UpdateModuleState updates the scan state for scanned modules.
// scannedModules maps moniker -> passed (true = scan passed).
func UpdateModuleState(workspaceRoot string, scannedModules map[string]bool, moduleFiles map[string][]string) error {
	// Load or create state
	state, err := Load(workspaceRoot)
	if err != nil {
		if !errors.Is(err, ErrNoState) {
			return err
		}
		state = &State{
			Modules: make(map[string]ModuleState),
		}
	}

	// Update git state
	commit, gitErr := getGitCommit(workspaceRoot)
	if gitErr == nil {
		state.Commit = commit
	}
	uncommittedFiles, gitErr := getUncommittedFiles(workspaceRoot)
	if gitErr != nil {
		uncommittedFiles = nil
	}
	if len(uncommittedFiles) > 0 {
		state.UncommittedHash = hashUncommittedFiles(workspaceRoot, uncommittedFiles)
	} else {
		state.UncommittedHash = ""
	}

	// Update each scanned module
	for moniker, passed := range scannedModules {
		files, ok := moduleFiles[moniker]
		if !ok {
			continue
		}

		sourceHash, err := hashModuleFiles(workspaceRoot, files)
		if err != nil {
			log.Debugf("[SCAN-CACHE] Failed to hash module %s: %v", moniker, err)
			continue
		}

		state.Modules[moniker] = ModuleState{
			SourceHash: sourceHash,
			Passed:     passed,
			ScannedAt:  time.Now(),
			Files:      files,
		}

		log.Debugf("[SCAN-CACHE] Updated state for %s: hash=%s passed=%v",
			moniker, sourceHash[:16], passed)
	}

	return state.Save(workspaceRoot)
}

// ClearState removes the scan state file (for --skip-cache).
func ClearState(workspaceRoot string) error {
	statePath := filepath.Join(workspaceRoot, paths.OutDir, stateFileName)
	err := os.Remove(statePath)
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

// getUncommittedFiles returns list of uncommitted file paths (relative to repo root).
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
		// Format: XY filename or XY "filename" for paths with spaces
		path := strings.TrimSpace(line[3:])
		path = strings.Trim(path, "\"")
		files = append(files, path)
	}

	return files, nil
}

// hashUncommittedFiles creates a hash representing the uncommitted state.
func hashUncommittedFiles(workspaceRoot string, files []string) string {
	return hash.UncommittedState(workspaceRoot, files)
}

// hashModuleFiles computes a hash of all source files for a module.
func hashModuleFiles(workspaceRoot string, files []string) (string, error) {
	return hash.Files(workspaceRoot, files)
}

// ExpandGlobPatterns expands glob patterns to actual file paths.
// Returns files relative to workspaceRoot.
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

// ModuleFileGetter is an interface for getting module glob patterns.
type ModuleFileGetter interface {
	GetGlobPatterns() []string
}

// GetModuleSourceFiles returns source files for each module for change detection.
func GetModuleSourceFiles(workspaceRoot string, modules map[string]ModuleFileGetter) (map[string][]string, error) {
	result := make(map[string][]string)

	for moniker, module := range modules {
		patterns := module.GetGlobPatterns()
		files, err := ExpandGlobPatterns(workspaceRoot, patterns)
		if err != nil {
			return nil, fmt.Errorf("failed to expand patterns for %s: %w", moniker, err)
		}
		result[moniker] = files
	}

	return result, nil
}

// DetectChangesForModules is the high-level API for detecting which modules need rescanning.
func DetectChangesForModules(workspaceRoot string, modules map[string]ModuleFileGetter) (*ChangeResult, error) {
	moduleFiles, err := GetModuleSourceFiles(workspaceRoot, modules)
	if err != nil {
		return nil, err
	}

	return DetectChanges(workspaceRoot, moduleFiles)
}
