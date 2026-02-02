// Package changedetect provides unified change detection for build, lint, and test operations.
// It encapsulates git state, file hashing, and module mapping to answer:
// "Which modules have changed since the last recorded state?"
//
// This package provides the core detection algorithm. State persistence is handled
// by workunit.StateManager which stores per-module state in out/<context>/<module>/state.json.
package changedetect

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

// GitStateProvider abstracts git operations for testing and flexibility.
type GitStateProvider interface {
	// HeadCommit returns the full SHA of HEAD.
	HeadCommit(ctx context.Context) (string, error)

	// UncommittedFiles returns paths of files with uncommitted changes.
	// Uses git status --porcelain format parsing.
	UncommittedFiles(ctx context.Context) ([]string, error)
}

// FileHasher abstracts file content hashing for testing.
type FileHasher interface {
	// HashFiles computes a deterministic hash of file contents.
	// Files are sorted, and both path and content are included.
	HashFiles(ctx context.Context, workspaceRoot string, files []string) (string, error)

	// HashUncommittedState computes a hash representing dirty working tree state.
	HashUncommittedState(ctx context.Context, workspaceRoot string, files []string) (string, error)
}

// ModuleFileResolver provides file lists for modules.
type ModuleFileResolver interface {
	// GetModuleFiles expands glob patterns and returns files for a module.
	GetModuleFiles(ctx context.Context, workspaceRoot string, moniker string) ([]string, error)
}

// WorkspaceState represents the recorded state at a point in time.
type WorkspaceState struct {
	// Git commit SHA at time of recording
	Commit string `json:"commit"`

	// Hash of uncommitted changes (empty if working tree clean)
	UncommittedHash string `json:"uncommitted_hash,omitempty"`

	// Per-module state
	Modules map[string]ModuleState `json:"modules"`

	// When this state was recorded
	RecordedAt time.Time `json:"recorded_at"`
}

// Clone creates a deep copy of the workspace state.
func (s *WorkspaceState) Clone() *WorkspaceState {
	clone := &WorkspaceState{
		Commit:          s.Commit,
		UncommittedHash: s.UncommittedHash,
		Modules:         make(map[string]ModuleState, len(s.Modules)),
		RecordedAt:      s.RecordedAt,
	}
	for k, v := range s.Modules {
		// Clone the Extra map
		var extraClone map[string]interface{}
		if v.Extra != nil {
			extraClone = make(map[string]interface{}, len(v.Extra))
			for ek, ev := range v.Extra {
				extraClone[ek] = ev
			}
		}
		clone.Modules[k] = ModuleState{
			SourceHash: v.SourceHash,
			Extra:      extraClone,
		}
	}
	return clone
}

// ModuleState represents recorded state for a single module.
type ModuleState struct {
	// Hash of source files
	SourceHash string `json:"source_hash"`

	// Operation-specific fields (stored but not interpreted by changedetect)
	// Build: BuiltAt timestamp
	// Lint: Passed bool, LintedAt timestamp
	// Test: TestHash, BuildID, Passed bool, Dependencies
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// ChangeResult represents detected changes.
type ChangeResult struct {
	// Modules that changed
	Changed []string

	// Modules that are up-to-date
	UpToDate []string

	// Reason for each changed module
	Reasons map[string]string

	// True if no prior state exists
	Fresh bool

	// Detection duration
	Duration time.Duration
}

// DetectOptions configures change detection behavior.
type DetectOptions struct {
	// WorkspaceRoot is the repository root directory
	WorkspaceRoot string

	// PreviousState is the recorded state to compare against (nil = fresh)
	PreviousState *WorkspaceState

	// Modules to check (monikers)
	Modules []string

	// CheckDependencies enables transitive dependency checking (for tests)
	CheckDependencies bool

	// DependencyGraph maps module -> dependencies (for tests)
	DependencyGraph map[string][]string
}

// Detector detects module changes against recorded state.
type Detector struct {
	git    GitStateProvider
	hasher FileHasher
	files  ModuleFileResolver
}

// NewDetector creates a change detector with the given providers.
func NewDetector(git GitStateProvider, hasher FileHasher, files ModuleFileResolver) *Detector {
	return &Detector{
		git:    git,
		hasher: hasher,
		files:  files,
	}
}

// DetectChanges compares current state against recorded state.
// Returns which modules need processing.
func (d *Detector) DetectChanges(ctx context.Context, opts DetectOptions) (*ChangeResult, error) {
	start := time.Now()
	result := &ChangeResult{
		Reasons: make(map[string]string),
	}

	// Debug output
	debugDetect := os.Getenv("DEBUG_CHANGEDETECT") != ""
	if debugDetect {
		fmt.Fprintf(os.Stderr, "[DEBUG changedetect] Modules: %v\n", opts.Modules)
		fmt.Fprintf(os.Stderr, "[DEBUG changedetect] PreviousState nil: %v\n", opts.PreviousState == nil)
		if opts.PreviousState != nil {
			for m, s := range opts.PreviousState.Modules {
				fmt.Fprintf(os.Stderr, "[DEBUG changedetect] PrevState module=%s hash=%s\n", m, s.SourceHash)
			}
		}
	}

	// Handle fresh build (no previous state)
	if opts.PreviousState == nil {
		result.Fresh = true
		result.Changed = opts.Modules
		for _, m := range opts.Modules {
			result.Reasons[m] = "fresh build (no previous state)"
		}
		result.Duration = time.Since(start)
		return result, nil
	}

	// Get current git state in parallel
	var currentCommit string
	var uncommittedFiles []string
	var commitErr, uncommittedErr error

	var gitWg sync.WaitGroup
	gitWg.Add(2)
	go func() {
		defer gitWg.Done()
		currentCommit, commitErr = d.git.HeadCommit(ctx)
	}()
	go func() {
		defer gitWg.Done()
		uncommittedFiles, uncommittedErr = d.git.UncommittedFiles(ctx)
	}()
	gitWg.Wait()

	if commitErr != nil {
		return nil, fmt.Errorf("failed to get HEAD commit: %w", commitErr)
	}
	if uncommittedErr != nil {
		return nil, fmt.Errorf("failed to get uncommitted files: %w", uncommittedErr)
	}

	var uncommittedHash string
	if len(uncommittedFiles) > 0 {
		var hashErr error
		uncommittedHash, hashErr = d.hasher.HashUncommittedState(ctx, opts.WorkspaceRoot, uncommittedFiles)
		if hashErr != nil {
			return nil, fmt.Errorf("failed to hash uncommitted state: %w", hashErr)
		}
	}

	// Compute module hashes in parallel
	type moduleResult struct {
		moniker string
		hash    string
		err     error
	}

	// Identify modules that need hash computation
	var modulesToHash []string
	for _, moniker := range opts.Modules {
		if _, hasPrevious := opts.PreviousState.Modules[moniker]; !hasPrevious {
			result.Changed = append(result.Changed, moniker)
			result.Reasons[moniker] = "new module (not in previous state)"
		} else {
			modulesToHash = append(modulesToHash, moniker)
		}
	}

	// Compute hashes in parallel with bounded concurrency (half CPU, floor min(4,NumCPU), cap 8)
	moduleHashes := make(map[string]string)
	if len(modulesToHash) > 0 {
		results := make(chan moduleResult, len(modulesToHash))
		numCPU := runtime.NumCPU()
		workers := numCPU / 2
		floor := 4
		if numCPU < floor {
			floor = numCPU
		}
		if workers < floor {
			workers = floor
		}
		if workers > 8 {
			workers = 8
		}
		sem := make(chan struct{}, workers)

		var hashWg sync.WaitGroup
		for _, moniker := range modulesToHash {
			hashWg.Add(1)
			go func(m string) {
				defer hashWg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				hash, err := d.ComputeModuleHash(ctx, opts.WorkspaceRoot, m)
				results <- moduleResult{moniker: m, hash: hash, err: err}
			}(moniker)
		}

		// Close results channel when all goroutines complete
		go func() {
			hashWg.Wait()
			close(results)
		}()

		// Collect results
		for r := range results {
			if r.err != nil {
				return nil, fmt.Errorf("failed to compute hash for module %s: %w", r.moniker, r.err)
			}
			moduleHashes[r.moniker] = r.hash
		}
	}

	// Compare hashes for modules we computed
	for _, moniker := range modulesToHash {
		prevModuleState := opts.PreviousState.Modules[moniker]
		currentHash := moduleHashes[moniker]

		if debugDetect {
			fmt.Fprintf(os.Stderr, "[DEBUG changedetect] Module %s: prev=%s curr=%s match=%v\n",
				moniker, prevModuleState.SourceHash, currentHash, currentHash == prevModuleState.SourceHash)
		}

		if currentHash != prevModuleState.SourceHash {
			result.Changed = append(result.Changed, moniker)
			result.Reasons[moniker] = "source files changed"
		} else {
			result.UpToDate = append(result.UpToDate, moniker)
		}
	}

	// Handle dependency propagation if enabled
	if opts.CheckDependencies && len(result.Changed) > 0 && opts.DependencyGraph != nil {
		// Mark modules whose dependencies changed
		changedSet := make(map[string]bool)
		for _, m := range result.Changed {
			changedSet[m] = true
		}

		// Check each up-to-date module for changed dependencies
		var stillUpToDate []string
		for _, moniker := range result.UpToDate {
			deps := opts.DependencyGraph[moniker]
			depChanged := false
			for _, dep := range deps {
				if changedSet[dep] {
					depChanged = true
					break
				}
			}

			if depChanged {
				result.Changed = append(result.Changed, moniker)
				result.Reasons[moniker] = "dependency changed"
			} else {
				stillUpToDate = append(stillUpToDate, moniker)
			}
		}
		result.UpToDate = stillUpToDate
	}

	// Store uncommitted hash for comparison (used by callers)
	_ = uncommittedHash
	_ = currentCommit

	result.Duration = time.Since(start)
	return result, nil
}

// ComputeCurrentState captures the current workspace state for all modules.
func (d *Detector) ComputeCurrentState(ctx context.Context, workspaceRoot string, modules []string) (*WorkspaceState, error) {
	state := &WorkspaceState{
		Modules:    make(map[string]ModuleState, len(modules)),
		RecordedAt: time.Now(),
	}

	// Get git state in parallel
	var commit string
	var uncommittedFiles []string
	var commitErr, uncommittedErr error

	var gitWg sync.WaitGroup
	gitWg.Add(2)
	go func() {
		defer gitWg.Done()
		commit, commitErr = d.git.HeadCommit(ctx)
	}()
	go func() {
		defer gitWg.Done()
		uncommittedFiles, uncommittedErr = d.git.UncommittedFiles(ctx)
	}()
	gitWg.Wait()

	if commitErr != nil {
		return nil, fmt.Errorf("failed to get HEAD commit: %w", commitErr)
	}
	if uncommittedErr != nil {
		return nil, fmt.Errorf("failed to get uncommitted files: %w", uncommittedErr)
	}

	state.Commit = commit

	if len(uncommittedFiles) > 0 {
		uncommittedHash, err := d.hasher.HashUncommittedState(ctx, workspaceRoot, uncommittedFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to hash uncommitted state: %w", err)
		}
		state.UncommittedHash = uncommittedHash
	}

	// Compute hashes for all modules in parallel (half CPU, floor min(4,NumCPU), cap 8)
	if len(modules) > 0 {
		type moduleResult struct {
			moniker string
			hash    string
			err     error
		}

		results := make(chan moduleResult, len(modules))
		numCPU := runtime.NumCPU()
		workers := numCPU / 2
		floor := 4
		if numCPU < floor {
			floor = numCPU
		}
		if workers < floor {
			workers = floor
		}
		if workers > 8 {
			workers = 8
		}
		sem := make(chan struct{}, workers)

		var hashWg sync.WaitGroup
		for _, moniker := range modules {
			hashWg.Add(1)
			go func(m string) {
				defer hashWg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				hash, err := d.ComputeModuleHash(ctx, workspaceRoot, m)
				results <- moduleResult{moniker: m, hash: hash, err: err}
			}(moniker)
		}

		go func() {
			hashWg.Wait()
			close(results)
		}()

		for r := range results {
			if r.err != nil {
				return nil, fmt.Errorf("failed to compute hash for module %s: %w", r.moniker, r.err)
			}
			state.Modules[r.moniker] = ModuleState{
				SourceHash: r.hash,
			}
		}
	}

	return state, nil
}

// ComputeModuleHash computes the current hash for a single module.
func (d *Detector) ComputeModuleHash(ctx context.Context, workspaceRoot string, moniker string) (string, error) {
	files, err := d.files.GetModuleFiles(ctx, workspaceRoot, moniker)
	if err != nil {
		return "", fmt.Errorf("failed to get module files: %w", err)
	}
	return d.hasher.HashFiles(ctx, workspaceRoot, files)
}

// NewDetectorWithRegistry creates a Detector using a module registry for file resolution.
// This is the recommended way to create a Detector for production use.
//
// The gitRepo parameter should implement the GitRepository interface from core/git.
// The getContract function should return the ModuleContract for a given moniker.
//
// Example:
//
//	gitMgr := git.NewManager(logger)
//	repo, _ := gitMgr.Open(workspaceRoot)
//	detector := changedetect.NewDetectorWithRegistry(repo, func(m string) (interface{ GetGlobPatterns() []string }, bool) {
//	    return registry.Get(m)
//	})
func NewDetectorWithRegistry(
	gitRepo GitStateProvider,
	getContract func(moniker string) (interface{ GetGlobPatterns() []string }, bool),
) *Detector {
	registryAdapter := NewRegistryAdapterFunc(getContract)
	return NewDetector(
		gitRepo,
		NewHashAdapter(),
		NewContractResolverAdapter(registryAdapter),
	)
}

// ConvertToWorkspaceState converts module state to WorkspaceState format.
// This is a helper for integrating with workunit.StateManager.
//
// The moduleStates parameter should map moniker to source hash.
func ConvertToWorkspaceState(
	commit string,
	uncommittedHash string,
	moduleStates map[string]string, // moniker -> sourceHash
	recordedAt time.Time,
) *WorkspaceState {
	state := &WorkspaceState{
		Commit:          commit,
		UncommittedHash: uncommittedHash,
		Modules:         make(map[string]ModuleState, len(moduleStates)),
		RecordedAt:      recordedAt,
	}
	for moniker, sourceHash := range moduleStates {
		state.Modules[moniker] = ModuleState{
			SourceHash: sourceHash,
		}
	}
	return state
}
