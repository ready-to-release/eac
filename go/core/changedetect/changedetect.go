// Package changedetect provides unified change detection for build, lint, and test operations.
// It encapsulates git state, file hashing, and module mapping to answer:
// "Which modules have changed since the last recorded state?"
//
// This package provides the core detection algorithm. State persistence is handled
// by workunit.StateManager which stores per-module state in .cache/eac/incremental/<context>/<module>/state.json.
package changedetect

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/core/environments"
)

// DetectChanges compares current state against recorded state.
// Returns which modules need processing.
func (d *Detector) DetectChanges(ctx context.Context, opts DetectOptions) (*ChangeResult, error) {
	start := time.Now()
	result := &ChangeResult{
		Reasons: make(map[string]string),
	}

	debugDetect := os.Getenv(environments.EnvDebugChangeDetect) != ""
	logDetectInput(debugDetect, opts)

	// Handle fresh build (no previous state)
	if opts.PreviousState == nil {
		return buildFreshResult(result, opts.Modules, start), nil
	}

	// Get current git state in parallel
	currentCommit, uncommittedHash, err := d.fetchGitState(ctx, opts.WorkspaceRoot)
	if err != nil {
		return nil, err
	}

	// Identify new modules vs modules needing hash comparison
	modulesToHash := classifyNewModules(result, opts)

	// Compute current hashes for existing modules
	moduleHashes, err := d.computeModuleHashes(ctx, opts.WorkspaceRoot, modulesToHash)
	if err != nil {
		return nil, err
	}

	// Compare current hashes against previous state
	compareModuleHashes(result, modulesToHash, moduleHashes, opts.PreviousState, debugDetect)

	// Propagate changes through dependency graph
	propagateDependencyChanges(result, opts)

	// Store uncommitted hash for comparison (used by callers)
	_ = uncommittedHash
	_ = currentCommit

	result.Duration = time.Since(start)
	return result, nil
}

// logDetectInput emits debug output for the detection inputs when debug mode is enabled.
func logDetectInput(debugDetect bool, opts DetectOptions) {
	if !debugDetect {
		return
	}
	fmt.Fprintf(os.Stderr, "[DEBUG changedetect] Modules: %v\n", opts.Modules)
	fmt.Fprintf(os.Stderr, "[DEBUG changedetect] PreviousState nil: %v\n", opts.PreviousState == nil)
	if opts.PreviousState != nil {
		for m, s := range opts.PreviousState.Modules {
			fmt.Fprintf(os.Stderr, "[DEBUG changedetect] PrevState module=%s hash=%s\n", m, s.SourceHash)
		}
	}
}

// buildFreshResult populates a ChangeResult for a fresh build where no previous state exists.
func buildFreshResult(result *ChangeResult, modules []string, start time.Time) *ChangeResult {
	result.Fresh = true
	result.Changed = modules
	for _, m := range modules {
		result.Reasons[m] = "fresh build (no previous state)"
	}
	result.Duration = time.Since(start)
	return result
}

// fetchGitState retrieves the current HEAD commit and uncommitted file hash in parallel.
func (d *Detector) fetchGitState(ctx context.Context, workspaceRoot string) (currentCommit string, uncommittedHash string, err error) {
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
		return "", "", fmt.Errorf("failed to get HEAD commit: %w", commitErr)
	}
	if uncommittedErr != nil {
		return "", "", fmt.Errorf("failed to get uncommitted files: %w", uncommittedErr)
	}

	if len(uncommittedFiles) > 0 {
		uncommittedHash, err = d.hasher.HashUncommittedState(ctx, workspaceRoot, uncommittedFiles)
		if err != nil {
			return "", "", fmt.Errorf("failed to hash uncommitted state: %w", err)
		}
	}

	return currentCommit, uncommittedHash, nil
}

// classifyNewModules separates modules into those that are new (no previous state)
// and those that need hash comparison. New modules are added directly to result.Changed.
func classifyNewModules(result *ChangeResult, opts DetectOptions) []string {
	var modulesToHash []string
	for _, moniker := range opts.Modules {
		if _, hasPrevious := opts.PreviousState.Modules[moniker]; !hasPrevious {
			result.Changed = append(result.Changed, moniker)
			result.Reasons[moniker] = "new module (not in previous state)"
		} else {
			modulesToHash = append(modulesToHash, moniker)
		}
	}
	return modulesToHash
}

// moduleHashResult holds the outcome of a single module hash computation.
type moduleHashResult struct {
	moniker string
	hash    string
	err     error
}

// computeModuleHashes computes source hashes for the given modules in parallel
// with bounded concurrency (half CPU, floor min(4,NumCPU), cap 8).
func (d *Detector) computeModuleHashes(ctx context.Context, workspaceRoot string, modules []string) (map[string]string, error) {
	moduleHashes := make(map[string]string)
	if len(modules) == 0 {
		return moduleHashes, nil
	}

	results := make(chan moduleHashResult, len(modules))
	workers := computeWorkerCount()
	sem := make(chan struct{}, workers)

	var hashWg sync.WaitGroup
	for _, moniker := range modules {
		hashWg.Add(1)
		go func(m string) {
			defer hashWg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			hash, err := d.ComputeModuleHash(ctx, workspaceRoot, m)
			results <- moduleHashResult{moniker: m, hash: hash, err: err}
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

	return moduleHashes, nil
}

// computeWorkerCount determines the number of parallel workers for hash computation.
// Uses half CPU count with a floor of min(4, NumCPU) and a cap of 8.
func computeWorkerCount() int {
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
	return workers
}

// compareModuleHashes compares computed hashes against previous state and populates
// the result with changed and up-to-date modules.
func compareModuleHashes(result *ChangeResult, modulesToHash []string, moduleHashes map[string]string, prevState *WorkspaceState, debugDetect bool) {
	for _, moniker := range modulesToHash {
		prevModuleState := prevState.Modules[moniker]
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
}

// propagateDependencyChanges marks up-to-date modules as changed if any of their
// dependencies have changed. Only runs when dependency checking is enabled.
func propagateDependencyChanges(result *ChangeResult, opts DetectOptions) {
	if !opts.CheckDependencies || len(result.Changed) == 0 || opts.DependencyGraph == nil {
		return
	}

	changedSet := make(map[string]bool)
	for _, m := range result.Changed {
		changedSet[m] = true
	}

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
