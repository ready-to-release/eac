// Package changedetect provides unified change detection for incremental operations.
//
// # Overview
//
// The changedetect package provides a common change detection mechanism used by
// build, lint, and test operations to determine which modules need processing.
// It implements a hybrid git + file hash approach for fast and accurate detection.
//
// # Architecture
//
// The package uses a Detector that composes three interfaces:
//   - GitStateProvider: Provides git commit SHA and uncommitted file list
//   - FileHasher: Computes deterministic hashes of file contents
//   - ModuleFileResolver: Maps module monikers to their source files
//
// This composition allows:
//   - Easy testing with mock implementations
//   - Flexible integration with different module registry implementations
//   - Separation of concerns between git state, hashing, and module resolution
//
// # Detection Strategy
//
// Change detection uses a two-level approach:
//
//  1. Fast Path: If git commit SHA and uncommitted file hash match the previous
//     state exactly, AND all requested modules have stored hashes, skip file
//     hashing entirely. This handles the common "no changes" case efficiently.
//
//  2. Slow Path: Hash source files for each module and compare to stored hashes.
//     This handles branch switches, git reset, stash/unstash, cherry-pick, etc.
//
// The module source hash is always the source of truth - git state is only used
// as an optimization to skip hashing when nothing could have changed.
//
// # Module-Level Detection
//
// Detection operates at the module level, not the component level. Each module's
// GetGlobPatterns() aggregates patterns from all its components (go, specs, etc.).
// This means a change to any component within a module marks the entire module
// as needing processing.
//
// # Dependency Propagation
//
// For test operations, the Detector supports dependency propagation via the
// CheckDependencies and DependencyGraph options. When enabled, a module is
// marked as changed if any of its transitive dependencies changed.
//
// # State Persistence
//
// The Detector itself does not persist state. State persistence is handled by
// workunit.StateManager which stores per-module state files with:
//   - SourceHash: Hash of source files for change detection
//   - Passed: Whether the last operation succeeded (failed units need re-execution)
//   - BuildID: For test/scan - the build this was tested against
//   - DependencyHash: For integration tests - hash of dependency build IDs
//   - ExecutedAt: Timestamp of last execution
//
// Use ConvertToWorkspaceState() to convert workunit state to the
// WorkspaceState format expected by Detector.DetectChanges().
//
// # Usage Example
//
// Basic usage with module registry:
//
//	// Open git repository
//	gitMgr := git.NewManager(nil)
//	gitRepo, _ := gitMgr.Open(workspaceRoot)
//
//	// Create detector with registry-based file resolution
//	detector := changedetect.NewDetectorWithRegistry(
//	    changedetect.NewGitRepositoryAdapter(gitRepo),
//	    func(m string) (interface{ GetGlobPatterns() []string }, bool) {
//	        return registry.Get(m)
//	    },
//	)
//
//	// Load previous state using workunit.StateManager
//	stateMgr := workunit.NewStateManager(workspaceRoot)
//	moduleHashes := make(map[string]string)
//	for _, m := range []string{"module-a", "module-b"} {
//	    unitID := workunit.UnitID{
//	        Context:   workunit.ContextBuild,
//	        Module:    m,
//	        Component: "_module",
//	        Tool:      "_",
//	    }
//	    if state, err := stateMgr.Load(unitID); err == nil {
//	        moduleHashes[m] = state.SourceHash
//	    }
//	}
//	workspaceState := changedetect.ConvertToWorkspaceState(
//	    "", // git commit tracked separately if needed
//	    "",
//	    moduleHashes,
//	    time.Now(),
//	)
//
//	// Detect changes
//	result, _ := detector.DetectChanges(ctx, changedetect.DetectOptions{
//	    WorkspaceRoot: workspaceRoot,
//	    PreviousState: workspaceState,
//	    Modules:       []string{"module-a", "module-b"},
//	})
//
//	// Use result
//	for _, m := range result.Changed {
//	    // Process changed module
//	}
//
// # Testing
//
// The package provides mock implementations for all interfaces, making it easy
// to test detection logic without real git repositories or file systems.
package changedetect
