// Package lintstate manages incremental lint state for detecting which modules need relinting.
// It uses a hybrid git + file hash approach for fast and accurate change detection.
//
// State is stored in out/.lint-state.json (global file, similar to buildstate).
package lintstate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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

// ErrNoState is returned when no lint state file exists (fresh lint).
var ErrNoState = errors.New("no lint state file")

// State represents the lint state for incremental linting.
type State struct {
	// Git commit SHA at time of last lint
	Commit string `json:"commit"`

	// Hash of uncommitted changes at lint time (empty if clean)
	UncommittedHash string `json:"uncommitted_hash,omitempty"`

	// Per-module lint state
	Modules map[string]ModuleState `json:"modules"`

	// Timestamp of last state update
	UpdatedAt time.Time `json:"updated_at"`
}

// ModuleState represents lint state for a single module.
type ModuleState struct {
	// Hash of all source files in the module
	SourceHash string `json:"source_hash"`

	// Whether the last lint run passed (no issues)
	Passed bool `json:"passed"`

	// Timestamp of successful lint
	LintedAt time.Time `json:"linted_at"`

	// List of file paths included in hash (for debugging)
	Files []string `json:"files,omitempty"`
}

// ChangeResult represents the result of change detection.
type ChangeResult struct {
	// Modules that need relinting (changed or new)
	ChangedModules []string

	// Modules that are up-to-date
	UpToDateModules []string

	// Reason for each changed module
	ChangeReasons map[string]string

	// Whether this is a fresh lint (no prior state)
	FreshRun bool

	// Detection time
	DetectionTime time.Duration
}

const (
	stateFileName = ".lint-state.json"
)

// Load loads lint state from the out directory.
// Returns ErrNoState if no state file exists (fresh lint).
func Load(workspaceRoot string) (*State, error) {
	statePath := filepath.Join(workspaceRoot, paths.OutDir, stateFileName)

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoState
		}
		return nil, fmt.Errorf("failed to read lint state: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse lint state: %w", err)
	}

	return &state, nil
}

// Save saves lint state to the out directory.
func (s *State) Save(workspaceRoot string) error {
	outDir := filepath.Join(workspaceRoot, paths.OutDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil { //nolint:gosec // G301: Output should be world-readable
		return fmt.Errorf("failed to create out directory: %w", err)
	}

	s.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lint state: %w", err)
	}

	statePath := filepath.Join(outDir, stateFileName)
	if err := os.WriteFile(statePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write lint state: %w", err)
	}

	return nil
}

// DetectChanges detects which modules need relinting based on file changes.
// moduleFiles maps moniker -> list of source file paths (relative to workspaceRoot)
//
// Detection strategy:
//  1. Fast path: If git commit + uncommitted state matches previous lint exactly,
//     AND all modules were in the previous lint, trust the stored hashes
//  2. Slow path: Hash source files and compare to stored hashes
//
// A module needs relinting if:
// 1. Its source files changed
// 2. It's a new module (not in previous state)
// 3. Its previous lint run failed (had issues).
func DetectChanges(workspaceRoot string, moduleFiles map[string][]string) (*ChangeResult, error) {
	start := time.Now()

	result := &ChangeResult{
		ChangeReasons: make(map[string]string),
	}

	// Load previous state
	prevState, err := Load(workspaceRoot)
	if err != nil {
		if errors.Is(err, ErrNoState) {
			// Fresh lint - all modules need linting
			result.FreshRun = true
			for moniker := range moduleFiles {
				result.ChangedModules = append(result.ChangedModules, moniker)
				result.ChangeReasons[moniker] = "fresh run (no prior state)"
			}
			sort.Strings(result.ChangedModules)
			result.DetectionTime = time.Since(start)
			return result, nil
		}
		return nil, fmt.Errorf("failed to load lint state: %w", err)
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

	// Fast path: If git state matches exactly AND all requested modules have stored hashes,
	// we can skip file hashing entirely.
	gitStateMatches := currentCommit != "" &&
		currentCommit == prevState.Commit &&
		currentUncommittedHash == prevState.UncommittedHash

	// Check if fast path is possible (all modules must have prior state AND passed)
	canUseFastPath := gitStateMatches
	if canUseFastPath {
		for moniker := range moduleFiles {
			modState, exists := prevState.Modules[moniker]
			if !exists || !modState.Passed {
				canUseFastPath = false
				break
			}
		}
	}

	if canUseFastPath {
		// Fast path: git state unchanged, all modules have prior hashes and passed
		for moniker := range moduleFiles {
			result.UpToDateModules = append(result.UpToDateModules, moniker)
		}
		sort.Strings(result.UpToDateModules)
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Slow path: Hash source files for each module and compare
	for moniker, files := range moduleFiles {
		prevModState, exists := prevState.Modules[moniker]

		if !exists {
			result.ChangedModules = append(result.ChangedModules, moniker)
			result.ChangeReasons[moniker] = "new module (not previously linted)"
			continue
		}

		// If previous lint failed, needs relinting
		if !prevModState.Passed {
			result.ChangedModules = append(result.ChangedModules, moniker)
			result.ChangeReasons[moniker] = "previous lint had issues"
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
		} else {
			result.UpToDateModules = append(result.UpToDateModules, moniker)
		}
	}

	sort.Strings(result.ChangedModules)
	sort.Strings(result.UpToDateModules)
	result.DetectionTime = time.Since(start)

	return result, nil
}

// UpdateModuleState updates the lint state for linted modules.
// lintedModules maps moniker -> whether lint passed (true = no issues).
func UpdateModuleState(workspaceRoot string, lintedModules map[string]bool, moduleFiles map[string][]string) error {
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

	// Update git state - errors are non-fatal
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

	// Update each linted module
	for moniker, passed := range lintedModules {
		files, ok := moduleFiles[moniker]
		if !ok {
			continue
		}

		hash, err := hashModuleFiles(workspaceRoot, files)
		if err != nil {
			continue // Skip modules we can't hash
		}

		state.Modules[moniker] = ModuleState{
			SourceHash: hash,
			Passed:     passed,
			LintedAt:   time.Now(),
			Files:      files,
		}
	}

	return state.Save(workspaceRoot)
}

// ClearState removes the lint state file (for --relint).
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
	h := sha256.New()

	// Sort for deterministic hash
	sorted := make([]string, len(files))
	copy(sorted, files)
	sort.Strings(sorted)

	for _, file := range sorted {
		path := filepath.Join(workspaceRoot, file)
		f, err := os.Open(path)
		if err != nil {
			// File might be deleted - include that in hash
			h.Write([]byte(file + ":deleted\n"))
			continue
		}
		if _, err := io.Copy(h, f); err != nil {
			h.Write([]byte(file + ":error\n"))
		}
		f.Close()
	}

	return hex.EncodeToString(h.Sum(nil))[:16] // Short hash is sufficient
}

// hashModuleFiles computes a hash of all source files for a module.
func hashModuleFiles(workspaceRoot string, files []string) (string, error) {
	h := sha256.New()

	// Sort for deterministic hash
	sorted := make([]string, len(files))
	copy(sorted, files)
	sort.Strings(sorted)

	for _, file := range sorted {
		path := filepath.Join(workspaceRoot, file)
		f, err := os.Open(path)
		if err != nil {
			return "", fmt.Errorf("failed to open %s: %w", file, err)
		}

		// Include filename in hash (so renames are detected)
		h.Write([]byte(file + "\n"))
		if _, err := io.Copy(h, f); err != nil {
			f.Close()
			return "", fmt.Errorf("failed to read %s: %w", file, err)
		}
		f.Close()
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// ExpandGlobPatterns expands glob patterns to actual file paths.
// Returns files relative to workspaceRoot.
func ExpandGlobPatterns(workspaceRoot string, patterns []string) ([]string, error) {
	var result []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		// Make pattern absolute
		absPattern := pattern
		if !filepath.IsAbs(pattern) {
			absPattern = filepath.Join(workspaceRoot, pattern)
		}

		// Expand glob
		matches, err := filepath.Glob(absPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %s: %w", pattern, err)
		}

		for _, match := range matches {
			// Convert back to relative path
			rel, err := filepath.Rel(workspaceRoot, match)
			if err != nil {
				continue
			}
			rel = filepath.ToSlash(rel)

			// Skip directories, only include files
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

// DetectChangesForModules is the high-level API for detecting which modules need relinting.
// It combines GetModuleSourceFiles and DetectChanges into a single call.
func DetectChangesForModules(workspaceRoot string, modules map[string]ModuleFileGetter) (*ChangeResult, error) {
	moduleFiles, err := GetModuleSourceFiles(workspaceRoot, modules)
	if err != nil {
		return nil, err
	}

	return DetectChanges(workspaceRoot, moduleFiles)
}

// HashModuleFiles computes a SHA-256 hash of all source files for a module.
// This is the public API for computing input hashes for manifest storage.
func HashModuleFiles(workspaceRoot string, files []string) (string, error) {
	return hashModuleFiles(workspaceRoot, files)
}
