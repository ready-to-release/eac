// Package scanstate manages incremental scan state for detecting which modules need rescanning.
// It uses a hybrid git + file hash approach for fast and accurate change detection.
//
// State is stored in out/.scan-state.json (global file, similar to lintstate).
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
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

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

	// Per-scanner state (tracks which scanners passed)
	Scanners map[string]ScannerState `json:"scanners"`

	// Timestamp of successful scan
	ScannedAt time.Time `json:"scanned_at"`

	// List of file paths included in hash (for debugging)
	Files []string `json:"files,omitempty"`
}

// ScannerState represents state for a specific scanner type.
type ScannerState struct {
	// Whether the last scan run passed (no findings/errors)
	Passed bool `json:"passed"`

	// Timestamp of last run
	RunAt time.Time `json:"run_at"`

	// Evidence file SHA256 for verification
	EvidenceSHA256 string `json:"evidence_sha256,omitempty"`
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
		return nil, fmt.Errorf("failed to parse scan state: %w", err)
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
// scannerTypes is the list of scanner types being requested
//
// Detection strategy:
//  1. Fast path: If git commit + uncommitted state matches previous scan exactly,
//     AND all modules were in the previous scan with all requested scanners passed,
//     trust the stored hashes
//  2. Slow path: Hash source files and compare to stored hashes
//
// A module needs rescanning if:
// 1. Its source files changed
// 2. It's a new module (not in previous state)
// 3. A requested scanner was not run previously
// 4. A requested scanner failed (Passed=false).
func DetectChanges(workspaceRoot string, moduleFiles map[string][]string, scannerTypes []string) (*ChangeResult, error) {
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
				result.ChangeReasons[moniker] = "fresh run (no prior state)"
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

	// Fast path: If git state matches exactly AND all requested modules have stored hashes
	// with all requested scanners passed, we can skip file hashing entirely.
	gitStateMatches := currentCommit != "" &&
		currentCommit == prevState.Commit &&
		currentUncommittedHash == prevState.UncommittedHash

	// Check if fast path is possible
	canUseFastPath := gitStateMatches
	if canUseFastPath {
		for moniker := range moduleFiles {
			modState, exists := prevState.Modules[moniker]
			if !exists {
				canUseFastPath = false
				break
			}
			// Check all requested scanners passed
			for _, scannerType := range scannerTypes {
				scannerState, scannerExists := modState.Scanners[scannerType]
				if !scannerExists || !scannerState.Passed {
					canUseFastPath = false
					break
				}
			}
			if !canUseFastPath {
				break
			}
		}
	}

	if canUseFastPath {
		// Fast path: git state unchanged, all modules have prior hashes and all scanners passed
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
			result.ChangeReasons[moniker] = "new module (not previously scanned)"
			continue
		}

		// Check if any requested scanner was not run or failed
		var missingScanners []string
		var failedScanners []string
		for _, scannerType := range scannerTypes {
			scannerState, scannerExists := prevModState.Scanners[scannerType]
			if !scannerExists {
				missingScanners = append(missingScanners, scannerType)
			} else if !scannerState.Passed {
				failedScanners = append(failedScanners, scannerType)
			}
		}

		if len(failedScanners) > 0 {
			result.ChangedModules = append(result.ChangedModules, moniker)
			result.ChangeReasons[moniker] = fmt.Sprintf("previous scan failed for scanner: %s", failedScanners[0])
			continue
		}

		if len(missingScanners) > 0 {
			result.ChangedModules = append(result.ChangedModules, moniker)
			result.ChangeReasons[moniker] = fmt.Sprintf("scanner not previously run: %s", missingScanners[0])
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

// UpdateModuleState updates the scan state for scanned modules.
// scannedModules maps moniker -> map of scanner -> passed (true = no issues).
func UpdateModuleState(workspaceRoot string, scannedModules map[string]map[string]bool, moduleFiles map[string][]string) error {
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

	// Update each scanned module
	for moniker, scannerResults := range scannedModules {
		files, ok := moduleFiles[moniker]
		if !ok {
			continue
		}

		sourceHash, err := hashModuleFiles(workspaceRoot, files)
		if err != nil {
			continue // Skip modules we can't hash
		}

		// Get or create module state
		modState, exists := state.Modules[moniker]
		if !exists {
			modState = ModuleState{
				Scanners: make(map[string]ScannerState),
			}
		}

		modState.SourceHash = sourceHash
		modState.ScannedAt = time.Now()
		modState.Files = files

		// Update scanner states
		if modState.Scanners == nil {
			modState.Scanners = make(map[string]ScannerState)
		}
		for scannerType, passed := range scannerResults {
			modState.Scanners[scannerType] = ScannerState{
				Passed: passed,
				RunAt:  time.Now(),
			}
		}

		state.Modules[moniker] = modState
	}

	return state.Save(workspaceRoot)
}

// ClearState removes the scan state file (for --skip-cache/force rescan).
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
// Delegates to core/hash package for the actual hashing.
func hashUncommittedFiles(workspaceRoot string, files []string) string {
	return hash.UncommittedState(workspaceRoot, files)
}

// hashModuleFiles computes a hash of all source files for a module.
// Delegates to core/hash package for the actual hashing.
func hashModuleFiles(workspaceRoot string, files []string) (string, error) {
	return hash.Files(workspaceRoot, files)
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
// It combines GetModuleSourceFiles and DetectChanges into a single call.
func DetectChangesForModules(workspaceRoot string, modules map[string]ModuleFileGetter, scannerTypes []string) (*ChangeResult, error) {
	moduleFiles, err := GetModuleSourceFiles(workspaceRoot, modules)
	if err != nil {
		return nil, err
	}

	return DetectChanges(workspaceRoot, moduleFiles, scannerTypes)
}

// HashModuleFiles computes a SHA-256 hash of all source files for a module.
// This is the public API for computing input hashes for manifest storage.
func HashModuleFiles(workspaceRoot string, files []string) (string, error) {
	return hashModuleFiles(workspaceRoot, files)
}

// GetCacheStatus returns a human-readable summary of cache state.
func GetCacheStatus(result *ChangeResult) string {
	if result.FreshRun {
		return "fresh scan (no prior state)"
	}

	total := len(result.ChangedModules) + len(result.UpToDateModules)
	if len(result.UpToDateModules) == total {
		return fmt.Sprintf("all %d modules cached", total)
	}

	return fmt.Sprintf("%d/%d modules need rescanning", len(result.ChangedModules), total)
}

// filterUncommittedByModule filters uncommitted files to only those in the module.
func filterUncommittedByModule(uncommittedFiles []string, moduleFiles []string) []string {
	moduleFileSet := make(map[string]bool)
	for _, f := range moduleFiles {
		moduleFileSet[f] = true
	}

	var filtered []string
	for _, f := range uncommittedFiles {
		if moduleFileSet[f] {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// isPathInModule checks if a file path is within a module's files.
func isPathInModule(path string, moduleFiles []string) bool {
	for _, f := range moduleFiles {
		if strings.HasPrefix(path, f) || path == f {
			return true
		}
	}
	return false
}
