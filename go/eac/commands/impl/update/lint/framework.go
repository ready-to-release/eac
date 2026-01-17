// Package lint provides the lint command implementation using cmdframework.
package lint

import (
	"encoding/json"
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
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	systemdeps "github.com/ready-to-release/eac/go/eac/core/system-deps"
)

// LintConfig holds lint-specific configuration
type LintConfig struct {
	Fix    bool   // Auto-fix issues where possible
	Config string // Override config file path
}

// LintModuleResult holds lint results for a single module
type LintModuleResult struct {
	Moniker      string
	Success      bool
	IssueCount   int
	ErrorMessage string
	Duration     time.Duration
}

// lintContext holds lint-specific state during execution
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

// lintAfterInit handles lint-specific initialization
func lintAfterInit(ctx *cmdframework.ExecutionContext) error {
	// Build init summary
	buildLintInitSummary(ctx)
	return nil
}

// lintAfterExecute handles post-lint tasks
func lintAfterExecute(ctx *cmdframework.ExecutionContext) error {
	// Generate lint manifest for each module
	lctx := ctx.Config.Extra["lintContext"].(*lintContext)

	for moniker, result := range lctx.results {
		if err := generateLintManifest(ctx, moniker, result); err != nil {
			log.Warnf("Failed to generate lint manifest for %s: %v", moniker, err)
		}
	}

	return nil
}

// lintWorker is the worker function that lints a single module
func lintWorker(ctx *cmdframework.ExecutionContext, moniker string, logWriter io.Writer) int {
	lintCfg := ctx.Config.Extra["lintConfig"].(*LintConfig)
	lctx := ctx.Config.Extra["lintContext"].(*lintContext)

	module, exists := ctx.ModuleRegistry.Get(moniker)
	if !exists {
		output.Writeln(logWriter, "Error: module not found: %s", moniker)
		return 1
	}

	// Get handler for this module type
	moduleType := ctx.ModuleTypes[moniker]
	handler := linters.GetHandlerForModule(moduleType)
	if handler == nil {
		output.Writeln(logWriter, "⚠️  No linter available for module type: %s", moduleType)
		// Not an error - just skip modules without linters
		return 0
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
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		output.Writeln(logWriter, "Error creating output directory: %v", err)
		return 1
	}

	// Get module root
	moduleRoot := filepath.Join(ctx.WorkspaceRoot, module.Files.Root)

	// Validate module
	if err := handler.ValidateModule(moduleRoot, ctx.WorkspaceRoot); err != nil {
		output.Writeln(logWriter, "❌ Module validation failed: %v", err)
		return 1
	}

	output.Writeln(logWriter, "🔍 Linting %s with %s handler...", moniker, handler.Name())
	lintStart := time.Now()

	// Run the linter
	opts := linters.LintOptions{
		Fix:    lintCfg.Fix,
		Config: lintCfg.Config,
	}

	exitCode := handler.Lint(moduleRoot, ctx.WorkspaceRoot, outputDir, logWriter, opts)
	duration := time.Since(lintStart)

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

// lintDepsVerifier verifies system dependencies for linting
func lintDepsVerifier(ctx *cmdframework.ExecutionContext) *initsummary.DepsStatus {
	status := &initsummary.DepsStatus{Verified: true}

	// Collect unique requirements from all handlers that will be used
	depsMap := make(map[string]bool)
	for _, moniker := range ctx.GetExecutionMonikers() {
		moduleType := ctx.ModuleTypes[moniker]
		handler := linters.GetHandlerForModule(moduleType)
		if handler != nil {
			for _, req := range handler.Requirements() {
				depsMap[req] = true
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

	// Verify all dependencies
	results := systemdeps.VerifyAll(deps)

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

// buildLintInitSummary creates the init summary for lint commands
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

// countLintIssues counts the number of issues from a golangci-lint JSON output
func countLintIssues(jsonPath string) (int, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, nil
	}

	var output struct {
		Issues []interface{} `json:"Issues"`
	}
	if err := json.Unmarshal(data, &output); err != nil {
		return 0, err
	}

	return len(output.Issues), nil
}

// generateLintManifest generates the lint manifest for a module
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

	return os.WriteFile(manifestPath, data, 0644)
}

// LintManifest represents the lint manifest structure
type LintManifest struct {
	Moniker         string    `json:"moniker"`
	ModuleType      string    `json:"module_type"`
	GitCommit       string    `json:"git_commit"`
	RunTime         time.Time `json:"run_time"`
	DurationSeconds float64   `json:"duration_seconds"`
	Success         bool      `json:"success"`
	IssueCount      int       `json:"issue_count"`
}
