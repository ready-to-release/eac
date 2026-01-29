// Package buildstate manages incremental build state for detecting which modules need rebuilding.
// It uses a hybrid git + file hash approach for fast and accurate change detection.
package buildstate

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

// ErrNoState is returned when no build state file exists (fresh build).
var ErrNoState = errors.New("no build state file")

// State represents the build state for incremental builds.
type State struct {
	// Git commit SHA at time of last build
	Commit string `json:"commit"`

	// Hash of uncommitted changes at build time (empty if clean)
	UncommittedHash string `json:"uncommitted_hash,omitempty"`

	// Per-module build state
	Modules map[string]ModuleState `json:"modules"`

	// Timestamp of last state update
	UpdatedAt time.Time `json:"updated_at"`
}

// ModuleState represents build state for a single module.
type ModuleState struct {
	// Hash of all source files in the module
	SourceHash string `json:"source_hash"`

	// Timestamp of successful build
	BuiltAt time.Time `json:"built_at"`

	// List of file paths included in hash (for debugging)
	Files []string `json:"files,omitempty"`
}

// ChangeResult represents the result of change detection.
type ChangeResult struct {
	// Modules that need rebuilding (changed or new)
	ChangedModules []string

	// Modules that are up-to-date
	UpToDateModules []string

	// Reason for each changed module
	ChangeReasons map[string]string

	// Whether this is a fresh build (no prior state)
	FreshBuild bool

	// Detection time
	DetectionTime time.Duration
}

const (
	stateFileName = ".build-state.json"
)

// Load loads build state from the build output directory.
// Returns ErrNoState if no state file exists (fresh build).
func Load(workspaceRoot string) (*State, error) {
	statePath := paths.BuildStatePath(workspaceRoot, stateFileName)

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoState
		}
		return nil, fmt.Errorf("failed to read build state: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		// Treat corrupted state as fresh build - return ErrNoState to trigger rebuild
		return nil, ErrNoState
	}

	return &state, nil
}

// Save saves build state to the build output directory.
func (s *State) Save(workspaceRoot string) error {
	buildDir := filepath.Join(workspaceRoot, paths.OutBuildRelPath)
	if err := os.MkdirAll(buildDir, 0o755); err != nil { //nolint:gosec // G301: Build output should be world-readable
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	s.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal build state: %w", err)
	}

	statePath := filepath.Join(buildDir, stateFileName)
	if err := os.WriteFile(statePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write build state: %w", err)
	}

	return nil
}

// DetectChanges detects which modules need rebuilding based on file changes
// moduleFiles maps moniker -> list of source file paths (relative to workspaceRoot)
//
// Detection strategy:
//  1. Fast path: If git commit + uncommitted state matches previous build exactly,
//     AND all modules were in the previous build, trust the stored hashes
//  2. Slow path: Hash source files and compare to stored hashes
//
// The module source hash is always the source of truth - git state is only used
// as an optimization to skip hashing when we know nothing could have changed.
func DetectChanges(workspaceRoot string, moduleFiles map[string][]string) (*ChangeResult, error) {
	start := time.Now()

	result := &ChangeResult{
		ChangeReasons: make(map[string]string),
	}

	// Load previous state
	prevState, err := Load(workspaceRoot)
	if err != nil {
		if errors.Is(err, ErrNoState) {
			// Fresh build - all modules need building
			result.FreshBuild = true
			for moniker := range moduleFiles {
				result.ChangedModules = append(result.ChangedModules, moniker)
				result.ChangeReasons[moniker] = "fresh build (no prior state)"
			}
			sort.Strings(result.ChangedModules)
			result.DetectionTime = time.Since(start)
			return result, nil
		}
		return nil, fmt.Errorf("failed to load build state: %w", err)
	}

	// Get current git state for fast-path optimization
	// Git state is just an optimization - errors result in empty values which trigger slow path
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
	// we can skip file hashing entirely. This covers the common case of "no changes since last build".
	gitStateMatches := currentCommit != "" &&
		currentCommit == prevState.Commit &&
		currentUncommittedHash == prevState.UncommittedHash

	// DEBUG: Log git state comparison
	log.Debugf("[TUI-CACHE] Git state: current=%s prev=%s match=%v", currentCommit, prevState.Commit, currentCommit == prevState.Commit)
	log.Debugf("[TUI-CACHE] Uncommitted hash: current=%s prev=%s match=%v", currentUncommittedHash, prevState.UncommittedHash, currentUncommittedHash == prevState.UncommittedHash)

	// Check if fast path is possible (all modules must have prior state)
	canUseFastPath := gitStateMatches
	if canUseFastPath {
		for moniker := range moduleFiles {
			if _, exists := prevState.Modules[moniker]; !exists {
				log.Debugf("[TUI-CACHE] Fast path blocked: module %s not in previous state", moniker)
				canUseFastPath = false
				break
			}
		}
	}

	log.Debugf("[TUI-CACHE] Fast path: gitStateMatches=%v canUseFastPath=%v", gitStateMatches, canUseFastPath)

	if canUseFastPath {
		// Fast path: git state unchanged, all modules have prior hashes
		// Trust that files haven't changed
		log.Debugf("[TUI-CACHE] Using fast path - all %d modules up to date", len(moduleFiles))
		for moniker := range moduleFiles {
			result.UpToDateModules = append(result.UpToDateModules, moniker)
		}
		sort.Strings(result.UpToDateModules)
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Slow path: Hash source files for each module and compare
	// This handles: branch switches, git reset, stash/unstash, cherry-pick, etc.
	log.Debugf("[TUI-CACHE] Using slow path - hashing %d modules", len(moduleFiles))
	for moniker, files := range moduleFiles {
		prevModState, exists := prevState.Modules[moniker]

		if !exists {
			result.ChangedModules = append(result.ChangedModules, moniker)
			result.ChangeReasons[moniker] = "new module (not in previous build)"
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
			log.Debugf("[TUI-CACHE] Hash mismatch for %s: current=%s prev=%s", moniker, currentHash[:16], prevModState.SourceHash[:16])
		} else {
			result.UpToDateModules = append(result.UpToDateModules, moniker)
		}
	}

	sort.Strings(result.ChangedModules)
	sort.Strings(result.UpToDateModules)
	result.DetectionTime = time.Since(start)

	return result, nil
}

// UpdateModuleState updates the build state for successfully built modules.
func UpdateModuleState(workspaceRoot string, builtModules []string, moduleFiles map[string][]string) error {
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

	// Update git state - errors are non-fatal, just use empty values
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

	// Update each built module
	for _, moniker := range builtModules {
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
			BuiltAt:    time.Now(),
			Files:      files,
		}
	}

	return state.Save(workspaceRoot)
}

// ClearState removes the build state file (for --rebuild).
func ClearState(workspaceRoot string) error {
	statePath := paths.BuildStatePath(workspaceRoot, stateFileName)
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
// Delegates to core/hash package for the actual hashing.
func hashUncommittedFiles(workspaceRoot string, files []string) string {
	return hash.UncommittedState(workspaceRoot, files)
}

// hashModuleFiles computes a hash of all source files for a module.
// Delegates to core/hash package for the actual hashing.
func hashModuleFiles(workspaceRoot string, files []string) (string, error) {
	// Debug output if enabled
	debugHash := os.Getenv("DEBUG_CACHE_HASH") != ""
	if debugHash {
		sorted := make([]string, len(files))
		copy(sorted, files)
		sort.Strings(sorted)
		fmt.Fprintf(os.Stderr, "[DEBUG hashModuleFiles] workspace=%s, files=%v\n", workspaceRoot, sorted)
	}

	result, err := hash.Files(workspaceRoot, files)
	if err != nil {
		return "", err
	}

	if debugHash {
		fmt.Fprintf(os.Stderr, "[DEBUG hashModuleFiles] result hash=%s\n", result)
	}
	return result, nil
}

// ExpandGlobPatterns expands glob patterns to actual file paths
// Returns files relative to workspaceRoot.
// Supports ** (doublestar) patterns for recursive directory matching.
func ExpandGlobPatterns(workspaceRoot string, patterns []string) ([]string, error) {
	var result []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		// Make pattern absolute
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
// This is the canonical implementation used by both build command and get changed-modules-local.
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

// DetectChangesForModules is the high-level API for detecting which modules need rebuilding.
// It combines GetModuleSourceFiles and DetectChanges into a single call.
// Returns (changedModules, upToDateModules, changeReasons, isFreshBuild, detectionTime, error).
func DetectChangesForModules(workspaceRoot string, modules map[string]ModuleFileGetter) ([]string, []string, map[string]string, bool, time.Duration, error) {
	moduleFiles, err := GetModuleSourceFiles(workspaceRoot, modules)
	if err != nil {
		return nil, nil, nil, false, 0, err
	}

	result, err := DetectChanges(workspaceRoot, moduleFiles)
	if err != nil {
		return nil, nil, nil, false, 0, err
	}

	return result.ChangedModules, result.UpToDateModules, result.ChangeReasons, result.FreshBuild, result.DetectionTime, nil
}

// HashModuleFiles computes a SHA-256 hash of all source files for a module.
// This is the public API for computing input hashes for manifest storage.
// The hash includes both filename and content to detect renames.
func HashModuleFiles(workspaceRoot string, files []string) (string, error) {
	return hashModuleFiles(workspaceRoot, files)
}

// ComputeModuleInputHash computes the input hash for a single module.
// This is a convenience function that expands glob patterns and computes the hash.
func ComputeModuleInputHash(workspaceRoot string, module ModuleFileGetter) (string, error) {
	patterns := module.GetGlobPatterns()
	files, err := ExpandGlobPatterns(workspaceRoot, patterns)
	if err != nil {
		return "", fmt.Errorf("failed to expand patterns: %w", err)
	}
	return hashModuleFiles(workspaceRoot, files)
}
