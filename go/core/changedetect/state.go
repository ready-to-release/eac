package changedetect

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

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
