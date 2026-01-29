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
// The Detector itself does not persist state. State persistence remains in
// operation-specific packages (buildstate, lintstate, teststate) which store
// additional operation-specific data:
//   - buildstate: Stores BuiltAt timestamp and file list for debugging
//   - lintstate: Stores Passed flag (failed modules need re-linting)
//   - teststate: Stores suite tracking, BuildID, and test-specific state
//
// Use ConvertToWorkspaceState() to convert operation-specific state to the
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
//	// Load previous state (from operation-specific package)
//	prevState, _ := buildstate.Load(workspaceRoot)
//	moduleHashes := make(map[string]string)
//	for m, s := range prevState.Modules {
//	    moduleHashes[m] = s.SourceHash
//	}
//	workspaceState := changedetect.ConvertToWorkspaceState(
//	    prevState.Commit,
//	    prevState.UncommittedHash,
//	    moduleHashes,
//	    prevState.UpdatedAt,
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
