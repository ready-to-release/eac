// Package lint provides the lint command implementation using cmdframework.
package lint

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	systemdeps "github.com/ready-to-release/eac/go/eac/core/system-deps"
)

// LintConfig holds lint-specific configuration.
type LintConfig struct {
	Fix    bool   // Auto-fix issues where possible
	Config string // Override config file path
}

// LintModuleResult holds lint results for a single module.
type LintModuleResult struct {
	Moniker      string
	Success      bool
	IssueCount   int
	ErrorMessage string
	Duration     time.Duration
}

// lintContext holds lint-specific state during execution.
type lintContext struct {
	cfg     *LintConfig
	results map[string]*LintModuleResult
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
		cfg:     lintCfg,
		results: make(map[string]*LintModuleResult),
	}
	cmdCfg.Extra["lintContext"] = lctx

	// Set up hooks
	hooks := &cmdframework.Hooks{
		AfterInit:    lintAfterInit,
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

// lintAfterExecute handles post-lint tasks.
func lintAfterExecute(ctx *cmdframework.ExecutionContext) error {
	// Generate lint manifest for each module
	lctx, ok := ctx.Config.Extra["lintContext"].(*lintContext)
	if !ok {
		return fmt.Errorf("lintContext not found or wrong type")
	}

	for moniker, result := range lctx.results {
		if err := generateLintManifest(ctx, moniker, result); err != nil {
			log.Warnf("Failed to generate lint manifest for %s: %v", moniker, err)
		}
	}

	return nil
}

// lintWorker is the worker function that lints a single module.
// It iterates over the module's enabled package types and runs the appropriate linter for each.
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

	// Acquire lock for this module
	lockCfg := locking.LintConfig(moniker, paths.OutLintRelPath)
	lockFile, err := locking.Acquire(ctx.WorkspaceRoot, lockCfg)
	if err != nil {
		output.Writeln(logWriter, "Error: %v", err)
		return 1
	}
	defer locking.Release(lockFile)

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

	// Use package-based linting: iterate over enabled packages
	if module.Components != nil && len(module.Components.GetEnabled()) > 0 {
		exitCode = lintByPackages(ctx, module, moduleRoot, outputDir, logWriter, lintCfg, &lintersRun)
	}
	// No packages = nothing to lint

	duration := time.Since(lintStart)

	if lintersRun == 0 {
		output.Writeln(logWriter, "⚠️  No linters available for module: %s", moniker)
		return 0
	}

	// Record result
	result := &LintModuleResult{
		Moniker:  moniker,
		Success:  exitCode == 0,
		Duration: duration,
	}

	// Try to get issue count from lint.json if it exists
	jsonPath := filepath.Join(outputDir, "lint.json")
	if issueCount, err := countLintIssues(jsonPath); err == nil {
		result.IssueCount = issueCount
	}

	lctx.results[moniker] = result

	return exitCode
}

// lintByPackages lints a module by finding applicable lint providers for its components.
func lintByPackages(ctx *cmdframework.ExecutionContext, module *modules.ModuleContract,
	moduleRoot, outputDir string, logWriter io.Writer, lintCfg *LintConfig, lintersRun *int,
) int {
	cfg := config.Global()
	if cfg == nil || cfg.LintProviders == nil {
		output.Writeln(logWriter, "Warning: lint providers config not loaded")
		return 0
	}

	// Check if linting is disabled for this module
	if module.Linting != nil && containsString(module.Linting.Disabled, "all") {
		output.Writeln(logWriter, "Linting disabled for module: %s", module.Moniker)
		return 0
	}

	overallExitCode := 0

	// Find which lint providers apply to this module's component types
	for _, compName := range module.Components.GetEnabled() {
		// Get the actual component type (may differ from name for named components)
		compType := module.Components.GetComponentType(compName)

		// Find providers that apply to this component type
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

			output.Writeln(logWriter, "🔍 Linting %s with %s...", compType, providerName)
			*lintersRun++

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

	return overallExitCode
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
	status.Required = deps

	// Verify only lint-phase dependencies
	results := systemdeps.VerifyAllForPhase(deps, "lint")

	for _, result := range results {
		status.Available = append(status.Available, initsummary.DepsResult{
			Name:      result.Name,
			Available: result.Available,
			Version:   result.Version,
		})
		if !result.Available {
			status.Missing = append(status.Missing, result.Moniker)
		}
	}

	return status
}

// buildLintInitSummary creates the init summary for lint commands.
func buildLintInitSummary(ctx *cmdframework.ExecutionContext) {
	summary := initsummary.New("update lint").
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
	manifestPath := filepath.Join(outputDir, "lint.manifest.json")

	manifest := LintManifest{
		Moniker:         moniker,
		ModuleType:      ctx.ModuleTypes[moniker],
		GitCommit:       git.GetCommitSHA(ctx.WorkspaceRoot),
		RunTime:         time.Now(),
		DurationSeconds: result.Duration.Seconds(),
		Success:         result.Success,
		IssueCount:      result.IssueCount,
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
}
