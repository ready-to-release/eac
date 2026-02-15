// Package test provides the test command implementation using cmdframework.
package test

import (
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/workunit"
)

func init() {
	// Register test component-level execution support
	cmdframework.RegisterUnitProvider(core.ActionTest, ResolveTestUnitSpecs)
	cmdframework.RegisterUnitWorker(core.ActionTest, testUnitWorker)
}

// TestFrameworkConfig holds test-specific configuration for the framework.
type TestFrameworkConfig struct {
	// Input configuration
	SuiteName   string
	Coverage    bool
	ForceRetest bool
	Parallel    bool
	ListOnly    bool

	// Resolved during execution
	Suite           *testing.TestSuite
	SelectedTests   []testing.TestReference
	TestsByPackage  map[string][]testing.TestReference // Package path -> tests
	ModuleMapper    *ModuleMapper
	TestRunDir      string
	SuiteTagFilter  core.TagFilter
	OSFilteredCount int

	// Execution paths (using package paths directly)
	ParallelPaths   []string
	SequentialPaths []string

	// Stats for summary
	Stats *testSelectionStats

	// UoW-level incremental state
	CachedUoWs    map[string]bool      // UoW longname -> cached
	UoWCacheTimes map[string]time.Time // UoW longname -> cache time
	Tracker       *coreoutput.InMemoryTracker

	// Module-level incremental state (aggregated from UoWs for TUI compatibility)
	CachedModules map[string]bool      // Modules that are up-to-date (cache hits)
	CacheTimes    map[string]time.Time // Module-level cache times for TUI

	// Component mapping for clean directory names
	// Maps "cleanComponent" (e.g., "docs-drawio-cache/godog") to full pkgPath
	ComponentToPkgPath map[string]string

	// UoWTags stores pre-computed tag summaries keyed by UoW longname.
	// Populated by ResolveTestUnitSpecs, consumed by writeUoWTestManifest.
	UoWTags map[string]workunit.TagSummary

	// Pre-computed module input hashes for cache consistency.
	// Computed once before test execution, used by both detection and workers.
	// Prevents hash divergence when go generate modifies source files.
	ModuleInputHashes map[string]string

	// Execution state
	ExecCtx       *TestExecutionContext
	TestStartTime time.Time
	// NOTE: Lock acquisition moved to testUnitWorker (after cache check)
}

// testSelectionStats tracks test selection statistics.
type testSelectionStats struct {
	TotalDiscovered  int
	Skipped          int
	NotMatchingSuite int
	Selected         int
	OSFiltered       int
	ModulesRequested []string
	ModulesInScope   []string
	ModulesNoTests   []string
}

// RunTestWithFramework executes tests using the cmdframework.
func RunTestWithFramework(cmdCfg *cmdframework.CommandConfig, testCfg *TestFrameworkConfig) int {
	// Store test config in typed field for access in hooks/worker
	cmdCfg.TestCmdConfig = testCfg
	testCfg.Stats = &testSelectionStats{ModulesRequested: cmdCfg.Monikers}

	// Set up hooks
	hooks := &cmdframework.Hooks{
		AfterInit:     testAfterInit,
		AfterResolve:  testAfterResolve,
		BeforeExecute: testBeforeExecute,
		AfterExecute:  testAfterExecute,
	}

	// Don't register validators - test handles them inline with custom logic
	// cmdframework.SetArtifactValidator(nil)
	// cmdframework.SetDepsVerifier(nil)

	return cmdframework.Run(cmdCfg, testWorker, hooks)
}
