// Package lint provides the top-level lint command implementation using cmdframework.
package lint

import (
	"slices"
	"sync"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/core/config"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
)

func init() {
	// Register component-level execution support for lint
	cmdframework.RegisterUnitProvider(core.ActionLint, ResolveLintUnitSpecs)
	cmdframework.RegisterUnitWorker(core.ActionLint, lintUnitWorker)
	cmdframework.SetUoWCountProvider(getLintUoWCount)
}

// getLintUoWCount returns the total number of lintable UoWs (units of work).
func getLintUoWCount(ctx *cmdframework.ExecutionContext) int {
	specs := ResolveLintUnitSpecs(ctx)
	return CountLintComponents(specs)
}

// LintConfig holds lint-specific configuration.
type LintConfig struct {
	Fix       bool   // Auto-fix issues where possible
	Config    string // Override config file path
	ForceLint bool   // Bypass incremental detection
}

// LintModuleResult holds lint results for a single module.
type LintModuleResult struct {
	Moniker      string
	Success      bool
	IssueCount   int
	FixedCount   int
	ErrorMessage string
	Duration     time.Duration
	Providers    []string
}

// lintContext holds lint-specific state during execution.
type lintContext struct {
	cfg               *LintConfig
	results           map[string]*LintModuleResult
	cachedModules     map[string]bool      // Modules that are up-to-date (aggregated from UoWs for TUI)
	cacheTimes        map[string]time.Time // When cached modules were last linted
	moduleFiles       map[string][]string  // For input hash computation
	moduleInputHashes map[string]string    // Pre-computed hashes for cache consistency
	tracker           *coreoutput.InMemoryTracker // UoW manifest tracker
	mu                sync.Mutex           // Protects results map for concurrent access

	// UoW-level cache tracking
	cachedUoWs    map[string]bool      // UoW longname -> cached
	uowCacheTimes map[string]time.Time // UoW longname -> cache time
}

// RunLintWithFramework executes lint using the cmdframework.
// This provides parallel execution, TUI support, and consistent output.
func RunLintWithFramework(cmdCfg *cmdframework.CommandConfig, lintCfg *LintConfig) int {
	// Store lint config in typed fields for access in hooks/worker
	cmdCfg.LintCmdConfig = lintCfg

	// Create lint context
	lctx := &lintContext{
		cfg:         lintCfg,
		results:     make(map[string]*LintModuleResult),
		moduleFiles: make(map[string][]string),
	}
	cmdCfg.LintCmdContext = lctx

	// Set up hooks
	hooks := &cmdframework.Hooks{
		AfterInit:    lintAfterInit,
		AfterResolve: lintAfterResolve,
		AfterExecute: lintAfterExecute,
	}

	// Register deps verifier
	cmdframework.SetDepsVerifier(lintDepsVerifier)

	return cmdframework.Run(cmdCfg, nil, hooks)
}

// isProviderEnabledForModule checks if a lint provider should run for a module.
func isProviderEnabledForModule(providerName string, linting *config.ModuleLinting) bool {
	if linting == nil {
		return true // No overrides, use default behavior
	}

	// Check if provider is explicitly disabled
	if slices.Contains(linting.Disabled, providerName) {
		return false
	}

	// If enabled list is specified, provider must be in it
	if len(linting.Enabled) > 0 {
		return slices.Contains(linting.Enabled, providerName)
	}

	return true
}
