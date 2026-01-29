// Package lint provides the top-level lint command implementation using cmdframework.
package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/impl/update/lint/linters"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/git"
	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/commands/internal/locking"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/environments"
	"github.com/ready-to-release/eac/go/eac/core/hash"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/tool"
	"github.com/ready-to-release/eac/go/eac/core/workunit"
)

func init() {
	// Register component-level execution support for lint
	cmdframework.RegisterComponentProvider(cmdframework.CommandTypeLint, FlattenModulesToLintComponentWork)
	cmdframework.RegisterComponentWorker(cmdframework.CommandTypeLint, lintComponentWorker)
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
	cfg           *LintConfig
	results       map[string]*LintModuleResult
	cachedModules map[string]bool     // Modules that are up-to-date (cache hits)
	moduleFiles   map[string][]string // For state update
	mu            sync.Mutex          // Protects results map for concurrent access
}

// RunLintWithFramework executes lint using the cmdframework.
// This provides parallel execution, TUI support, and consistent output.
func RunLintWithFramework(cmdCfg *cmdframework.CommandConfig, lintCfg *LintConfig) int {
	// Store lint config in Extra for access in hooks/worker
	if cmdCfg.Extra == nil {
		cmdCfg.Extra = make(map[string]interface{})
	}
	cmdCfg.Extra["lintConfig"] = lintCfg

	// Create lint context
	lctx := &lintContext{
		cfg:         lintCfg,
		results:     make(map[string]*LintModuleResult),
		moduleFiles: make(map[string][]string),
	}
	cmdCfg.Extra["lintContext"] = lctx

	// Set up hooks
	hooks := &cmdframework.Hooks{
		AfterInit:    lintAfterInit,
		AfterResolve: lintAfterResolve,
		AfterExecute: lintAfterExecute,
	}

	// Register deps verifier
	cmdframework.SetDepsVerifier(lintDepsVerifier)

	return cmdframework.Run(cmdCfg, lintWorker, hooks)
}

// lintAfterInit handles lint-specific initialization.
func lintAfterInit(ctx *cmdframework.ExecutionContext) error {
	// Build init summary
	buildLintInitSummary(ctx)
	return nil
}

// lintAfterResolve handles incremental lint detection.
func lintAfterResolve(ctx *cmdframework.ExecutionContext) error {
	lintCfg := ctx.Config.Extra["lintConfig"].(*LintConfig)
	lctx := ctx.Config.Extra["lintContext"].(*lintContext)

	// Clear lint state if --skip-cache
	if lintCfg.ForceLint {
		stateMgr := workunit.NewStateManager(ctx.WorkspaceRoot)
		if err := stateMgr.ClearContext(workunit.ContextLint); err != nil {
			log.Warnf("Failed to clear lint state: %v", err)
		}
		return nil
	}

	// Incremental lint detection (devbox only, not CI)
	// Also run in dry-run mode to show which modules would be linted/skipped
	if !environments.IsCI() {
		detectIncrementalLintChanges(ctx, lctx)
	}

	return nil
}

// detectIncrementalLintChanges detects which modules need relinting.
// Instead of filtering modules from the execution plan, it stores which modules
// are cached so the component worker can skip them with -1 (blue in TUI).
// This keeps all modules visible and clickable in the TUI.
func detectIncrementalLintChanges(ctx *cmdframework.ExecutionContext, lctx *lintContext) {
	startTime := time.Now()
	defer func() {
		ctx.SetChangeDetectionTiming(time.Since(startTime))
	}()

	// Collect modules for change detection
	monikers := ctx.GetExecutionMonikers()
	if len(monikers) == 0 {
		return
	}

	// Build module files map for later state update
	moduleFiles := make(map[string][]string)
	for _, moniker := range monikers {
		if contract, ok := ctx.ModuleRegistry.Get(moniker); ok {
			patterns := contract.GetGlobPatterns()
			files, err := hash.ExpandGlobPatterns(ctx.WorkspaceRoot, patterns)
			if err != nil {
				log.Debugf("Failed to expand patterns for %s: %v", moniker, err)
				continue
			}
			moduleFiles[moniker] = files
		}
	}

	// Store for later state update
	lctx.moduleFiles = moduleFiles

	// Use StateManager for change detection
	stateMgr := workunit.NewStateManager(ctx.WorkspaceRoot)
	rule := workunit.DefaultRules[workunit.ContextLint]

	// Create hash provider that computes hash from expanded files
	hashProvider := func(module string) (string, error) {
		files, ok := moduleFiles[module]
		if !ok {
			return "", fmt.Errorf("no files for module %s", module)
		}
		return hash.Files(ctx.WorkspaceRoot, files)
	}

	changeResult, err := stateMgr.DetectModuleChanges(workunit.ContextLint, monikers, rule, hashProvider)
	if err != nil {
		log.Debugf("Failed to detect lint changes: %v", err)
		return
	}

	detectionTime := time.Since(startTime)

	if changeResult.FreshRun {
		if ctx.InitSummary != nil {
			ctx.InitSummary.SetIncremental(&initsummary.IncrementalInfo{
				Enabled:       true,
				DetectionTime: detectionTime,
				FreshBuild:    true,
			})
		}
		return
	}

	// Build set of changed modules
	changedSet := make(map[string]bool)
	for _, m := range changeResult.ChangedModules {
		changedSet[m] = true
	}

	// Build set of cached modules (modules that are up-to-date)
	// These will be skipped at the component worker level, not filtered from the plan.
	lctx.cachedModules = make(map[string]bool)
	var changedList []string
	var cachedList []string

	for _, moniker := range monikers {
		if changedSet[moniker] {
			changedList = append(changedList, moniker)
		} else {
			lctx.cachedModules[moniker] = true
			cachedList = append(cachedList, moniker)
		}
	}

	if ctx.InitSummary != nil {
		ctx.InitSummary.SetIncremental(&initsummary.IncrementalInfo{
			Enabled:       true,
			DetectionTime: detectionTime,
			Changed:       changedList,
			UpToDate:      cachedList,
			FreshBuild:    false,
		})
	}

	log.Debugf("Incremental lint: %d modules to lint, %d cached (will show blue in TUI)",
		len(changedList), len(cachedList))
}

// lintAfterExecute handles post-lint tasks.
func lintAfterExecute(ctx *cmdframework.ExecutionContext) error {
	lctx, ok := ctx.Config.Extra["lintContext"].(*lintContext)
	if !ok {
		return fmt.Errorf("lintContext not found or wrong type")
	}

	// Generate lint manifest for each module
	for moniker, result := range lctx.results {
		if err := generateLintManifest(ctx, moniker, result); err != nil {
			log.Warnf("Failed to generate lint manifest for %s: %v", moniker, err)
		}
	}

	// Update incremental lint state (devbox only)
	if !environments.IsCI() && !ctx.Config.DryRun {
		updateLintState(ctx, lctx)
	}

	return nil
}

// updateLintState updates the lint state after execution.
func updateLintState(ctx *cmdframework.ExecutionContext, lctx *lintContext) {
	if len(lctx.results) == 0 {
		return
	}

	stateMgr := workunit.NewStateManager(ctx.WorkspaceRoot)

	// Update state for each linted module
	for moniker, result := range lctx.results {
		files, ok := lctx.moduleFiles[moniker]
		if !ok {
			continue
		}

		sourceHash, err := hash.Files(ctx.WorkspaceRoot, files)
		if err != nil {
			log.Debugf("Failed to hash files for %s: %v", moniker, err)
			continue
		}

		if err := stateMgr.SaveModuleResult(workunit.ContextLint, moniker, result.Success, sourceHash); err != nil {
			log.Warnf("Failed to update lint state for %s: %v", moniker, err)
		}
	}
}

// lintWorker runs linting for a single module.
func lintWorker(ctx *cmdframework.ExecutionContext, moniker string, logWriter io.Writer) int {
	lintCfg, ok := ctx.Config.Extra["lintConfig"].(*LintConfig)
	if !ok {
		output.Writeln(logWriter, "Error: lintConfig not found or wrong type")
		return 1
	}
	lctx, ok := ctx.Config.Extra["lintContext"].(*lintContext)
	if !ok {
		output.Writeln(logWriter, "Error: lintContext not found or wrong type")
		return 1
	}

	module, exists := ctx.ModuleRegistry.Get(moniker)
	if !exists {
		output.Writeln(logWriter, "Error: module not found: %s", moniker)
		return 1
	}

	// Acquire lock for this module with wait
	lockCfg := locking.LintConfig(moniker, paths.OutLintRelPath)
	lockFile, err := locking.AcquireWithWait(context.Background(), ctx.WorkspaceRoot, lockCfg,
		ctx.Orchestrator.GetRegistry(), locking.DefaultWaitConfig())
	if err != nil {
		output.Writeln(logWriter, "Error: %v", err)
		return 1
	}
	defer locking.ReleaseTracked(lockFile)

	// Create output directory
	outputDir := paths.LintOutputPath(ctx.WorkspaceRoot, moniker)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		output.Writeln(logWriter, "Error creating output directory: %v", err)
		return 1
	}

	// Get module root from first available package
	var moduleRoot string
	for _, pkgRoot := range module.GetComponentRoots() {
		moduleRoot = filepath.Join(ctx.WorkspaceRoot, pkgRoot)
		break
	}
	if moduleRoot == "" {
		output.Writeln(logWriter, "Warning: no package roots found for module %s", moniker)
		return 0
	}

	lintStart := time.Now()
	exitCode := 0
	lintersRun := 0
	var providersRun []string

	// Use package-based linting: iterate over enabled packages
	if module.Components != nil && len(module.Components.GetEnabled()) > 0 {
		exitCode, providersRun = lintByPackages(ctx, module, moduleRoot, outputDir, logWriter, lintCfg, &lintersRun)
	}

	duration := time.Since(lintStart)

	if lintersRun == 0 {
		output.Writeln(logWriter, "No linters available for module: %s", moniker)
		return 0
	}

	// Record result
	result := &LintModuleResult{
		Moniker:   moniker,
		Success:   exitCode == 0,
		Duration:  duration,
		Providers: providersRun,
	}

	// Try to get issue count from lint.json if it exists
	jsonPath := filepath.Join(outputDir, "lint.json")
	if issueCount, err := countLintIssues(jsonPath); err == nil {
		result.IssueCount = issueCount
	}

	lctx.results[moniker] = result

	return exitCode
}

// lintComponentWorker lints a single component with a specific provider.
// This is called by the ComponentScheduler for parallel component execution.
// The component parameter is in "compName:providerName" format (e.g., "go:go-lint").
func lintComponentWorker(ctx *cmdframework.ExecutionContext, module, component string, logWriter io.Writer) int {
	lintCfg, ok := ctx.Config.Extra["lintConfig"].(*LintConfig)
	if !ok {
		output.Writeln(logWriter, "Error: lintConfig not found or wrong type")
		return 1
	}
	lctx, ok := ctx.Config.Extra["lintContext"].(*lintContext)
	if !ok {
		output.Writeln(logWriter, "Error: lintContext not found or wrong type")
		return 1
	}

	// Check incremental cache first - if module is cached, skip immediately (blue in TUI)
	if lctx.cachedModules != nil && lctx.cachedModules[module] {
		if ctx.Config.DryRun {
			output.Writeln(logWriter, "⏭️  %s is up-to-date (would be skipped)", module)
		} else {
			output.Writeln(logWriter, "⏭️  Cached (unchanged)")
		}
		return -1 // -1 = skipped/cached = blue in TUI
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

	// Parse component parameter: "compName:providerName"
	parts := strings.SplitN(component, ":", 2)
	if len(parts) != 2 {
		output.Writeln(logWriter, "Error: invalid component format: %s (expected compName:providerName)", component)
		return 1
	}
	compName := parts[0]
	providerName := parts[1]

	provider := cfg.LintProviders.Get(providerName)
	if provider == nil {
		output.Writeln(logWriter, "Error: lint provider not found: %s", providerName)
		return 1
	}

	handler := linters.GetHandlerForProvider(providerName)
	if handler == nil {
		output.Writeln(logWriter, "Error: no handler for provider: %s", providerName)
		return 1
	}

	// Acquire component-level lock with wait (use underscore separator for Windows compatibility)
	componentDir := compName + "_" + providerName
	if !ctx.Config.DryRun {
		lockCfg := locking.ComponentLintConfig(module, componentDir, paths.OutLintRelPath)
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

	opts := linters.LintOptions{
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

	if exitCode != 0 {
		output.Writeln(logWriter, "❌ Lint failed for %s:%s with %s", module, compName, providerName)
	} else {
		output.Writeln(logWriter, "✅ Lint passed for %s:%s with %s", module, compName, providerName)

		// Atomically update lint state for this module immediately after success
		// This ensures interrupted lint runs preserve cache for completed modules
		if err := updateModuleLintStateAtomic(ctx, lctx, module); err != nil {
			log.Debugf("Failed to update lint state for %s: %v", module, err)
		}
	}

	return exitCode
}

// updateModuleLintStateAtomic updates lint state for a single module immediately after completion.
// This ensures interrupted lint runs preserve cache for completed modules (atomic caching).
func updateModuleLintStateAtomic(ctx *cmdframework.ExecutionContext, lctx *lintContext, module string) error {
	contract, exists := ctx.ModuleRegistry.Get(module)
	if !exists {
		return nil
	}

	// Get source files for this module
	patterns := contract.GetGlobPatterns()
	files, err := hash.ExpandGlobPatterns(ctx.WorkspaceRoot, patterns)
	if err != nil {
		return err
	}

	// Compute source hash
	sourceHash, err := hash.Files(ctx.WorkspaceRoot, files)
	if err != nil {
		return err
	}

	// Save module state
	stateMgr := workunit.NewStateManager(ctx.WorkspaceRoot)
	return stateMgr.SaveModuleResult(workunit.ContextLint, module, true, sourceHash)
}

// lintByPackages lints a module by finding applicable lint providers for its components.
func lintByPackages(ctx *cmdframework.ExecutionContext, module *modules.ModuleContract,
	moduleRoot, outputDir string, logWriter io.Writer, lintCfg *LintConfig, lintersRun *int,
) (int, []string) {
	cfg := config.Global()
	if cfg == nil || cfg.LintProviders == nil {
		output.Writeln(logWriter, "Warning: lint providers config not loaded")
		return 0, nil
	}

	// Check if linting is disabled for this module
	if module.Linting != nil && containsString(module.Linting.Disabled, "all") {
		output.Writeln(logWriter, "Linting disabled for module: %s", module.Moniker)
		return 0, nil
	}

	overallExitCode := 0
	var providersRun []string

	// Find which lint providers apply to this module's component types
	for _, compName := range module.Components.GetEnabled() {
		compType := module.Components.GetComponentType(compName)

		providerNames := cfg.LintProviders.GetProvidersForComponentType(compType)
		for _, providerName := range providerNames {
			// Check module-level linting overrides
			if !isProviderEnabledForModule(providerName, module.Linting) {
				log.Debugf("Provider %s disabled for module %s", providerName, module.Moniker)
				continue
			}

			provider := cfg.LintProviders.Get(providerName)
			if provider == nil {
				continue
			}

			handler := linters.GetHandlerForProvider(providerName)
			if handler == nil {
				log.Debugf("No handler registered for provider: %s (component type: %s)", providerName, compType)
				continue
			}

			// Validate module for this handler
			if err := handler.ValidateModule(moduleRoot, ctx.WorkspaceRoot); err != nil {
				log.Debugf("Module validation failed for %s handler: %v", handler.Name(), err)
				continue
			}

			output.Writeln(logWriter, "Linting %s with %s...", compType, providerName)
			*lintersRun++
			providersRun = append(providersRun, providerName)

			opts := linters.LintOptions{
				Fix:       lintCfg.Fix,
				Config:    lintCfg.Config,
				InputMode: provider.GetInputMode(),
			}

			exitCode := handler.Lint(moduleRoot, ctx.WorkspaceRoot, outputDir, logWriter, opts)
			if exitCode != 0 {
				overallExitCode = exitCode
			}
		}
	}

	return overallExitCode, providersRun
}

// isProviderEnabledForModule checks if a lint provider should run for a module.
func isProviderEnabledForModule(providerName string, linting *contracts.ModuleLinting) bool {
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
		SetOutputDir(paths.OutLintRelPath).
		SetFlatExecution(true) // Lint runs in parallel (no layered execution)

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

// generateLintManifest generates the lint manifest for a module.
func generateLintManifest(ctx *cmdframework.ExecutionContext, moniker string, result *LintModuleResult) error {
	outputDir := paths.LintOutputPath(ctx.WorkspaceRoot, moniker)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	manifestPath := filepath.Join(outputDir, "lint.manifest.json")

	manifest := LintManifest{
		Moniker:         moniker,
		ModuleType:      ctx.ModuleTypes[moniker],
		GitCommit:       git.GetCommitSHA(ctx.WorkspaceRoot),
		RunTime:         time.Now(),
		DurationSeconds: result.Duration.Seconds(),
		Success:         result.Success,
		IssueCount:      result.IssueCount,
		Providers:       result.Providers,
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(manifestPath, data, 0o644)
}

// LintManifest represents the lint manifest structure.
type LintManifest struct {
	Moniker         string    `json:"moniker"`
	ModuleType      string    `json:"module_type"`
	GitCommit       string    `json:"git_commit"`
	RunTime         time.Time `json:"run_time"`
	DurationSeconds float64   `json:"duration_seconds"`
	Success         bool      `json:"success"`
	IssueCount      int       `json:"issue_count"`
	FixedCount      int       `json:"fixed_count,omitempty"`
	Providers       []string  `json:"providers,omitempty"`
}
