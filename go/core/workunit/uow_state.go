package workunit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/core/paths"
)

// StateManager handles persistence of unit states.
type StateManager struct {
	workspaceRoot string
}

// NewStateManager creates a new state manager rooted at the given workspace.
func NewStateManager(root string) *StateManager {
	return &StateManager{workspaceRoot: root}
}

// Load reads the state for a unit from disk.
// Returns an error if the state file doesn't exist or contains invalid JSON.
func (m *StateManager) Load(id UnitID) (*UnitState, error) {
	path := filepath.Join(m.workspaceRoot, id.StateFile())
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var state UnitState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// Save writes the state for a unit to disk.
// Creates the state cache directory if it doesn't exist.
func (m *StateManager) Save(state *UnitState) error {
	dir := filepath.Join(m.workspaceRoot, state.ID.StateCacheDir())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(m.workspaceRoot, state.ID.StateFile())
	return os.WriteFile(path, data, 0644)
}

// NeedsExecution determines if a unit needs to be executed based on its
// cached state and the provided invalidation rule.
// Returns (true, reason) if execution is needed, (false, "") otherwise.
// If cacheConfig is provided and state cache is skipped, always returns true.
func (m *StateManager) NeedsExecution(spec UnitSpec, rule InvalidationRule, currentHash string, cacheConfig *cache.Config) (bool, string) {
	// If state cache is skipped, always execute
	if cacheConfig != nil && cacheConfig.ShouldSkipState() {
		return true, "cache skipped (--skip-cache=state)"
	}

	state, err := m.Load(spec.ID)
	if err != nil {
		return true, "no prior state"
	}

	// Check failure invalidation first (if enabled)
	if rule.OnFailure && !state.Passed {
		return true, "previous failure"
	}

	// Check source change (if enabled)
	if rule.OnSourceChange && state.SourceHash != currentHash {
		return true, "source changed"
	}

	// Check build change (if enabled) - for tests that depend on build output
	if rule.OnBuildChange && spec.Metadata != nil {
		if currentBuildID, ok := spec.Metadata["build_id"].(string); ok {
			if state.BuildID != currentBuildID {
				return true, "build changed"
			}
		}
	}

	// Check dependency change (if enabled) - for integration tests
	if rule.OnDependencyChange && spec.Metadata != nil {
		if currentDepHash, ok := spec.Metadata["dependency_hash"].(string); ok {
			if state.DependencyHash != currentDepHash {
				return true, "dependency changed"
			}
		}
	}

	return false, ""
}

// ChangeResult represents the result of batch change detection.
type ChangeResult struct {
	// Units that need execution (changed or new)
	Changed []UnitSpec
	// Units that are up-to-date (cached)
	UpToDate []UnitSpec
	// Reason for each changed unit (key is Longname)
	ChangeReasons map[string]string
	// Whether this is a fresh run (no prior state for any unit)
	FreshRun bool
	// Time taken for detection
	DetectionTime time.Duration
}

// HashProvider is a function that computes the current source hash for a unit.
type HashProvider func(spec UnitSpec) (string, error)

// DetectChanges performs batch change detection for multiple units.
// Uses the provided HashProvider to compute current hashes.
// Returns which units need execution and which are up-to-date.
// If cacheConfig is provided and state cache is skipped, all units are marked as changed.
func (m *StateManager) DetectChanges(specs []UnitSpec, rule InvalidationRule, hashProvider HashProvider, cacheConfig *cache.Config) (*ChangeResult, error) {
	start := time.Now()
	result := &ChangeResult{
		ChangeReasons: make(map[string]string),
	}

	if len(specs) == 0 {
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// If state cache is skipped, all units need execution
	if cacheConfig != nil && cacheConfig.ShouldSkipState() {
		for _, spec := range specs {
			result.Changed = append(result.Changed, spec)
			result.ChangeReasons[spec.ID.Longname()] = "cache skipped (--skip-cache=state)"
		}
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Check if any prior state exists
	hasAnyState := false
	for _, spec := range specs {
		if _, err := m.Load(spec.ID); err == nil {
			hasAnyState = true
			break
		}
	}

	if !hasAnyState {
		// Fresh run - all units need execution
		result.FreshRun = true
		for _, spec := range specs {
			result.Changed = append(result.Changed, spec)
			result.ChangeReasons[spec.ID.Longname()] = "fresh run (no prior state)"
		}
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Check each unit
	for _, spec := range specs {
		currentHash := ""
		if hashProvider != nil {
			var err error
			currentHash, err = hashProvider(spec)
			if err != nil {
				result.Changed = append(result.Changed, spec)
				result.ChangeReasons[spec.ID.Longname()] = "hash error: " + err.Error()
				continue
			}
		}

		needsExec, reason := m.NeedsExecution(spec, rule, currentHash, cacheConfig)
		if needsExec {
			result.Changed = append(result.Changed, spec)
			result.ChangeReasons[spec.ID.Longname()] = reason
		} else {
			result.UpToDate = append(result.UpToDate, spec)
		}
	}

	// Sort for deterministic output
	sort.Slice(result.Changed, func(i, j int) bool {
		return result.Changed[i].ID.Longname() < result.Changed[j].ID.Longname()
	})
	sort.Slice(result.UpToDate, func(i, j int) bool {
		return result.UpToDate[i].ID.Longname() < result.UpToDate[j].ID.Longname()
	})

	result.DetectionTime = time.Since(start)
	return result, nil
}

// SaveResult saves the state for a completed unit execution.
// Creates a UnitState from the result and persists it.
func (m *StateManager) SaveResult(spec UnitSpec, exitCode int, sourceHash string) error {
	state := &UnitState{
		ID:         spec.ID,
		SourceHash: sourceHash,
		Passed:     exitCode == 0,
		ExecutedAt: time.Now(),
	}

	// Copy build/dependency hashes from metadata if present
	if spec.Metadata != nil {
		if buildID, ok := spec.Metadata["build_id"].(string); ok {
			state.BuildID = buildID
		}
		if depHash, ok := spec.Metadata["dependency_hash"].(string); ok {
			state.DependencyHash = depHash
		}
	}

	return m.Save(state)
}

// Exists checks if state exists for a unit without loading it.
func (m *StateManager) Exists(id UnitID) bool {
	path := filepath.Join(m.workspaceRoot, id.StateFile())
	_, err := os.Stat(path)
	return err == nil
}

// Delete removes the state file for a unit.
func (m *StateManager) Delete(id UnitID) error {
	path := filepath.Join(m.workspaceRoot, id.StateFile())
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil // Already doesn't exist
	}
	return err
}

// ClearContext removes all state files for a given action type (build, test, lint, scan).
func (m *StateManager) ClearContext(ctx core.ActionType) error {
	dir := filepath.Join(paths.IncrementalCachePath(m.workspaceRoot), string(ctx))
	return os.RemoveAll(dir)
}
