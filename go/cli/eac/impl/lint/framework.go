// Package lint provides the top-level lint command implementation using cmdframework.
package lint

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/caching"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/clibase/locking"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/hash"
	"github.com/ready-to-release/eac/go/core/logging"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workunit"
)

func init() {
	// Register component-level execution support for lint
	cmdframework.RegisterUnitProvider(core.ActionLint, ResolveLintUnitSpecs)
	cmdframework.RegisterUnitWorker(core.ActionLint, lintUnitWorker)
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

// lintAfterInit handles lint-specific initialization.
func lintAfterInit(ctx *cmdframework.ExecutionContext) error {
	// Build init summary
	buildLintInitSummary(ctx)

	// Initialize UoW tracker for manifest generation
	lctx := ctx.Config.LintCmdContext.(*lintContext)
	lctx.tracker = coreoutput.NewTracker(ctx.WorkspaceRoot, core.ActionLint)

	return nil
}

// lintAfterResolve handles incremental lint detection.
func lintAfterResolve(ctx *cmdframework.ExecutionContext) error {
	lintCfg := ctx.Config.LintCmdConfig.(*LintConfig)
	lctx := ctx.Config.LintCmdContext.(*lintContext)

	// Clear lint output if --skip-cache (force relint)
	if lintCfg.ForceLint {
		lintOutDir := filepath.Join(ctx.WorkspaceRoot, "out", "lint")
		if err := os.RemoveAll(lintOutDir); err != nil {
			log.Warnf("Failed to clear lint output: %v", err)
		}
		return nil
	}

	// Incremental lint detection (devbox only, not CI)
	// Also run in dry-run mode to show which modules would be linted/skipped
	if !environments.IsCI() {
		detectUoWIncrementalLintChanges(ctx, lctx)

		// Pass cache times to orchestrator for TUI display
		if len(lctx.cacheTimes) > 0 && ctx.Orchestrator != nil {
			ctx.Orchestrator.SetCacheTimes(lctx.cacheTimes)
		}
	}

	return nil
}

// detectUoWIncrementalLintChanges performs UoW-level incremental lint detection.
// Instead of checking at module granularity, it checks each component:provider UoW.
// This enables partial caching - some components can be cached while others relint.
func detectUoWIncrementalLintChanges(ctx *cmdframework.ExecutionContext, lctx *lintContext) {
	specs := ResolveLintUnitSpecs(ctx)
	result := caching.DetectIncrementalChanges(ctx, core.ActionLint, specs, "LINT")
	if result == nil {
		return
	}

	// Always store pre-computed hashes for worker reuse
	lctx.moduleInputHashes = result.ModuleInputHashes

	if result.FreshRun {
		return
	}

	// Copy aggregated results to lint context
	lctx.cachedUoWs = result.CachedUoWs
	lctx.uowCacheTimes = result.UoWCacheTimes
	lctx.cachedModules = result.CachedModules
	lctx.cacheTimes = result.ModuleCacheTimes
}

// lintAfterExecute handles post-lint tasks.
// Note: UoW manifests are written atomically during execution via InMemoryTracker.
func lintAfterExecute(ctx *cmdframework.ExecutionContext) error {
	// Assert all UoWs have valid manifests (skip in dry-run mode)
	if !ctx.Config.DryRun {
		if err := assertLintManifestsExist(ctx); err != nil {
			return err
		}
	}

	return nil
}

// assertLintManifestsExist verifies that all executed UoWs have valid manifests.
func assertLintManifestsExist(ctx *cmdframework.ExecutionContext) error {
	return cmdframework.AssertManifestsExist(ctx, "lint", ResolveLintUnitSpecs(ctx))
}


// lintUnitWorker lints a single component with a specific provider.
// This is called by the UnitScheduler for parallel component execution.
// The component parameter is in "compName:providerName" format (e.g., "go:go-lint").
func lintUnitWorker(goCtx context.Context, ctx *cmdframework.ExecutionContext, module, component string, logWriter io.Writer) int {
	lintCfg, ok := ctx.Config.LintCmdConfig.(*LintConfig)
	if !ok {
		output.Writeln(logWriter, "Error: lintConfig not found or wrong type")
		return 1
	}
	lctx, ok := ctx.Config.LintCmdContext.(*lintContext)
	if !ok {
		output.Writeln(logWriter, "Error: lintContext not found or wrong type")
		return 1
	}

	// Parse component parameter to build UnitID for UoW cache lookup
	parts := strings.SplitN(component, ":", 2)
	compName := parts[0]
	providerName := ""
	if len(parts) == 2 {
		providerName = parts[1]
	}

	// Build UnitID for UoW-level cache lookup
	unitID := workunit.UnitID{
		Action:    core.ActionLint,
		Module:    module,
		Component: compName,
		Tool:      providerName,
	}

	// Check UoW-level cache
	isCached := lctx.cachedUoWs != nil && lctx.cachedUoWs[unitID.Longname()]
	log.Debugf("[LINT-UOW-CACHE] Component worker for %s: unitID=%s, isCached=%v",
		component, unitID.Longname(), isCached)

	if isCached {
		if ctx.Config.DryRun {
			output.Writeln(logWriter, "⏭️  %s is up-to-date (would be skipped)", module)
		} else {
			// Verify UoW manifest artifacts are intact before declaring cache hit
			reader := coreoutput.NewReader(ctx.WorkspaceRoot)
			uowID := workunit.UnitID{
				Action:    core.ActionLint,
				Module:    module,
				Component: compName,
				Tool:      providerName,
			}
			validationResult := reader.ValidateUoW(uowID)
			if !validationResult.ManifestExists {
				// No manifest - trust source hash, allow cache hit
				log.Debugf("Lint UoW cache hit (no manifest to verify)")
				output.Writeln(logWriter, "⏭️  Cached (unchanged)")
				return -1
			}
			if !validationResult.Valid {
				output.Writeln(logWriter, "UoW cache miss: artifacts invalid")
				// Fall through to lint - clear cache flag
				delete(lctx.cachedUoWs, unitID.Longname())
			} else {
				output.Writeln(logWriter, "⏭️  Cached (verified)")
				return -1
			}
		}
		if ctx.Config.DryRun {
			return -1 // -1 = skipped/cached = blue in TUI
		}
	}

	cfg := config.Global()
	if cfg == nil || cfg.LintProviders == nil {
		output.Writeln(logWriter, "Error: lint providers config not loaded")
		return 1
	}

	moduleContract, exists := ctx.ModuleRegistry.Get(module)
	if !exists {
		output.Writeln(logWriter, "Error: module not found: %s", module)
		return 1
	}

	// Validate component format (already parsed above for cache check)
	if providerName == "" {
		output.Writeln(logWriter, "Error: invalid component format: %s (expected compName:providerName)", component)
		return 1
	}

	provider := cfg.LintProviders.Get(providerName)
	if provider == nil {
		output.Writeln(logWriter, "Error: lint provider not found: %s", providerName)
		return 1
	}

	handler := tool.GlobalLintBridge().GetHandlerForProvider(providerName)
	if handler == nil {
		output.Writeln(logWriter, "Error: no handler for provider: %s", providerName)
		return 1
	}

	// Acquire component-level lock with wait (use dash separator to match UnitID.DirName())
	componentDir := compName + "-" + providerName
	if !ctx.Config.DryRun {
		lockCfg := locking.UnitLintConfig(module, componentDir, paths.OutLintRelPath)
		lockFile, err := locking.AcquireWithWait(context.Background(), ctx.WorkspaceRoot, lockCfg,
			ctx.Orchestrator.GetRegistry(), locking.DefaultWaitConfig())
		if err != nil {
			output.Writeln(logWriter, "Error: %v", err)
			return 1
		}
		defer locking.ReleaseTracked(lockFile)
	}

	// Create output directory for this module+component+provider
	outputDir := paths.ComponentLintOutputPath(ctx.WorkspaceRoot, module, componentDir)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		output.Writeln(logWriter, "Error creating output directory: %v", err)
		return 1
	}

	// Get component root for the specific component
	compRoot := moduleContract.GetComponentRoot(compName)
	if compRoot == "" {
		output.Writeln(logWriter, "Warning: no root found for component %s in module %s", compName, module)
		return 0
	}
	moduleRoot := filepath.Join(ctx.WorkspaceRoot, compRoot)

	// Validate module for this handler
	if err := handler.ValidateModule(moduleRoot, ctx.WorkspaceRoot); err != nil {
		output.Writeln(logWriter, "Module validation failed for %s: %v", handler.Name(), err)
		return 1
	}

	// In dry-run mode, show what would happen without actually linting
	if ctx.Config.DryRun {
		output.Writeln(logWriter, "🔍 %s would be linted (changed)", module)
		output.Writeln(logWriter, "   Component: %s, Provider: %s", compName, providerName)
		return 0
	}

	output.Writeln(logWriter, "━━━ Linting %s with %s ━━━", compName, providerName)

	lintStart := time.Now()

	opts := tool.LintOptions{
		Fix:       lintCfg.Fix,
		Config:    lintCfg.Config,
		InputMode: provider.GetInputMode(),
	}

	exitCode := handler.Lint(moduleRoot, ctx.WorkspaceRoot, outputDir, logWriter, opts)

	duration := time.Since(lintStart)

	// Record result in lintContext (keyed by module:component for aggregation)
	resultKey := module + ":" + compName
	lctx.mu.Lock()
	if existing, ok := lctx.results[resultKey]; ok {
		// Update existing result
		existing.Providers = append(existing.Providers, providerName)
		if exitCode != 0 {
			existing.Success = false
		}
		existing.Duration += duration
	} else {
		// Create new result
		result := &LintModuleResult{
			Moniker:   resultKey,
			Success:   exitCode == 0,
			Duration:  duration,
			Providers: []string{providerName},
		}
		lctx.results[resultKey] = result
	}
	lctx.mu.Unlock()

	// Write UoW manifest via tracker
	if lctx.tracker != nil {
		inputHash := computeLintInputHash(ctx, module)
		manifest := &coreoutput.UoWManifest{
			ExitCode:   exitCode,
			InputHash:  inputHash,
			ExecutedAt: time.Now().UTC(),
			Duration:   duration,
			Artifacts:  nil, // Lint doesn't produce artifacts tracked in manifest
			OutputHash: "",
			Version:    "1.0.0",
		}
		if err := lctx.tracker.RecordComplete(unitID, manifest); err != nil {
			log.Debugf("Failed to record lint UoW completion for %s: %v", unitID.Longname(), err)
		}
	}

	if exitCode != 0 {
		output.Writeln(logWriter, "❌ Lint failed for %s:%s with %s", module, compName, providerName)
	} else {
		output.Writeln(logWriter, "✅ Lint passed for %s:%s with %s", module, compName, providerName)
	}

	return exitCode
}

// computeLintInputHash computes a hash of the input sources for a lint operation.
// Uses pre-computed hashes when available for cache consistency.
func computeLintInputHash(ctx *cmdframework.ExecutionContext, module string) string {
	lctx, ok := ctx.Config.LintCmdContext.(*lintContext)
	if !ok {
		return ""
	}

	// Use pre-computed hash if available
	if lctx.moduleInputHashes != nil {
		if h, ok := lctx.moduleInputHashes[module]; ok {
			return h
		}
	}

	// Fallback: compute from cached files or fresh expansion
	files, ok := lctx.moduleFiles[module]
	if !ok {
		contract, exists := ctx.ModuleRegistry.Get(module)
		if !exists {
			return ""
		}
		patterns := contract.GetGlobPatterns()
		var err error
		files, err = hash.ExpandGlobPatterns(ctx.WorkspaceRoot, patterns)
		if err != nil {
			log.Debugf("Failed to expand patterns for input hash of %s: %v", module, err)
			return ""
		}
	}

	inputHash, err := hash.Files(ctx.WorkspaceRoot, files)
	if err != nil {
		log.Debugf("Failed to compute input hash for %s: %v", module, err)
		return ""
	}

	return inputHash
}

// isProviderEnabledForModule checks if a lint provider should run for a module.
func isProviderEnabledForModule(providerName string, linting *domain.ModuleLinting) bool {
	if linting == nil {
		return true // No overrides, use default behavior
	}

	// Check if provider is explicitly disabled
	if containsString(linting.Disabled, providerName) {
		return false
	}

	// If enabled list is specified, provider must be in it
	if len(linting.Enabled) > 0 {
		return containsString(linting.Enabled, providerName)
	}

	return true
}

// containsString checks if a slice contains a string.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// lintDepsVerifier verifies system dependencies for linting.
func lintDepsVerifier(ctx *cmdframework.ExecutionContext) *initsummary.DepsStatus {
	status := &initsummary.DepsStatus{Verified: true}

	// ModuleRegistry may not be populated yet (async deps check starts before phaseResolve).
	// If unavailable, skip provider-based dep detection — return verified with no requirements.
	if ctx.ModuleRegistry == nil {
		return status
	}

	// Collect unique requirements from all lint providers that will be used
	depsMap := make(map[string]bool)
	cfg := config.Global()

	for _, moniker := range ctx.GetExecutionMonikers() {
		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		// Get requirements from lint providers for each component type
		if module.Components != nil && len(module.Components.GetEnabled()) > 0 && cfg != nil && cfg.LintProviders != nil {
			for _, compName := range module.Components.GetEnabled() {
				compType := module.Components.GetComponentType(compName)
				providerNames := cfg.LintProviders.GetProvidersForComponentType(compType)
				for _, providerName := range providerNames {
					provider := cfg.LintProviders.Get(providerName)
					if provider != nil && provider.SystemDependency != "" {
						depsMap[provider.SystemDependency] = true
					}
				}
			}
		}
	}

	// No dependencies to verify
	if len(depsMap) == 0 {
		return status
	}

	// Convert to sorted slice for consistent output
	deps := make([]string, 0, len(depsMap))
	for dep := range depsMap {
		deps = append(deps, dep)
	}
	sort.Strings(deps)

	// Filter out platform-incompatible tools before verification
	deps = tool.FilterPlatformSupported(deps)
	status.Required = deps

	// Verify dependencies using tool registry
	registry := tool.GlobalRegistry()
	results := registry.VerifyAll(deps)

	for _, result := range results {
		status.Available = append(status.Available, initsummary.DepsResult{
			Name:      result.ToolID,
			Available: result.Available,
			Version:   result.Version,
		})
		if !result.Available {
			status.Missing = append(status.Missing, result.ToolID)
		}
	}

	return status
}

// buildLintInitSummary creates the init summary for lint commands.
func buildLintInitSummary(ctx *cmdframework.ExecutionContext) {
	summary := initsummary.New("lint").
		SetRequest(ctx.Config.Monikers, ctx.GetExecutionMonikers()).
		SetExecutionContext(string(logging.GetExecutionContext())).
		SetFlags(initsummary.Flags{
			DebugMode: ctx.Config.DebugMode,
			UseTUI:    ctx.Config.UseTUI,
		}).
		SetOutputDir(paths.OutLintRelPath)

	ctx.InitSummary = summary
}

// countLintIssues counts the number of issues from a golangci-lint JSON output.
func countLintIssues(jsonPath string) (int, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, nil
	}

	var jsonOutput struct {
		Issues []interface{} `json:"Issues"`
	}
	if err := json.Unmarshal(data, &jsonOutput); err != nil {
		return 0, err
	}

	return len(jsonOutput.Issues), nil
}

