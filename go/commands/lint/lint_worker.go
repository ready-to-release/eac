package lint

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/locking"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/core/config"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

// lintWorkerParams holds resolved parameters for a single lint unit worker invocation.
// This avoids threading many individual values through the helper functions.
type lintWorkerParams struct {
	lintCfg      *LintConfig
	lctx         *lintContext
	module       string
	compName     string
	providerName string
	component    string // compName:providerName display string
	pipeline     *cmdframework.UnitPipeline
}

// lintUnitWorker lints a single component with a specific provider.
// This is called by the UnitScheduler for parallel component execution.
func lintUnitWorker(goCtx context.Context, ctx *cmdframework.ExecutionContext, spec core.UnitSpec, logWriter io.Writer) int {
	params, code := extractLintWorkerParams(ctx, spec, logWriter)
	if code != 0 {
		return code
	}

	unitID := spec.ID

	log.Debugf("[LINT-UOW-CACHE] Component worker for %s: unitID=%s", params.component, unitID.Longname())
	if cacheResult := params.pipeline.CheckCache(ctx, unitID, logWriter); cacheResult != 0 {
		return cacheResult
	}

	handler, provider, moduleContract, code := resolveLintHandler(ctx, params, logWriter)
	if code != 0 {
		return code
	}

	release, outputDir, moduleRoot, code := prepareLintWorkspace(ctx, params, unitID.DirName(), handler, moduleContract, logWriter)
	if code != 0 {
		return code
	}
	defer release()

	// prepareLintWorkspace returns empty moduleRoot when no component root is found (non-error skip)
	if moduleRoot == "" {
		return 0
	}

	// In dry-run mode, show what would happen without actually linting
	if ctx.Config.DryRun {
		output.Writeln(logWriter, "🔍 %s would be linted (changed)", params.module)
		output.Writeln(logWriter, "   Component: %s, Provider: %s", params.compName, params.providerName)
		return 0
	}

	exitCode, duration := executeLintProvider(params, handler, provider, moduleRoot, outputDir, ctx.WorkspaceRoot, logWriter)

	recordLintResult(params, exitCode, duration)

	recordLintManifest(ctx, params, unitID, exitCode, duration, logWriter)

	return exitCode
}

// extractLintWorkerParams extracts and validates the lint configuration, context,
// and identity fields from the execution context and unit spec.
// Returns the populated params and 0, or nil and a non-zero exit code on error.
func extractLintWorkerParams(ctx *cmdframework.ExecutionContext, spec core.UnitSpec, logWriter io.Writer) (*lintWorkerParams, int) {
	lintCfg, ok := ctx.Config.LintCmdConfig.(*LintConfig)
	if !ok {
		output.Writeln(logWriter, "Error: lintConfig not found or wrong type")
		return nil, 1
	}
	lctx, ok := ctx.Config.LintCmdContext.(*lintContext)
	if !ok {
		output.Writeln(logWriter, "Error: lintContext not found or wrong type")
		return nil, 1
	}

	module := spec.ID.Module
	compName := spec.ID.ComponentName
	providerName := spec.ID.Tool

	params := &lintWorkerParams{
		lintCfg:      lintCfg,
		lctx:         lctx,
		module:       module,
		compName:     compName,
		providerName: providerName,
		component:    compName + ":" + providerName,
		pipeline: &cmdframework.UnitPipeline{
			CachedUoWs:             lctx.cachedUoWs,
			ValidateCacheArtifacts: true,
			OnCacheInvalidated:     func(ln string) { delete(lctx.cachedUoWs, ln) },
			LockStyle:              cmdframework.LockUnlessDryRun,
			LockConfigFn:           func(m, cd string) locking.Config { return locking.UnitLintConfig(m, cd, paths.OutLintRelPath) },
			Tracker:                lctx.tracker,
		},
	}

	return params, 0
}

// resolveLintHandler looks up the global config, module contract, lint provider,
// and lint handler for the given worker params. Returns all resolved objects and 0,
// or nils and a non-zero exit code on error.
func resolveLintHandler(ctx *cmdframework.ExecutionContext, params *lintWorkerParams, logWriter io.Writer) (tool.LintHandler, *config.LintProvider, core.ModuleContractPort, int) {
	cfg := config.Global()
	if cfg == nil || cfg.LintProviders == nil {
		output.Writeln(logWriter, "Error: lint providers config not loaded")
		return nil, nil, nil, 1
	}

	moduleContract, exists := ctx.ModuleRegistry.Get(params.module)
	if !exists {
		output.Writeln(logWriter, "Error: module not found: %s", params.module)
		return nil, nil, nil, 1
	}

	if params.providerName == "" {
		output.Writeln(logWriter, "Error: invalid component format: %s (expected compName:providerName)", params.component)
		return nil, nil, nil, 1
	}

	provider := cfg.LintProviders.Get(params.providerName)
	if provider == nil {
		output.Writeln(logWriter, "Error: lint provider not found: %s", params.providerName)
		return nil, nil, nil, 1
	}

	handler := tool.GlobalLintBridge().GetHandlerForProvider(params.providerName)
	if handler == nil {
		output.Writeln(logWriter, "Error: no handler for provider: %s", params.providerName)
		return nil, nil, nil, 1
	}

	return handler, provider, moduleContract, 0
}

// prepareLintWorkspace acquires the unit-level lock, creates the output directory,
// resolves the component root, and validates the module for the handler.
// Returns the lock release function, output directory, module root, and exit code.
// The caller must defer the release function when exit code is 0.
func prepareLintWorkspace(ctx *cmdframework.ExecutionContext, params *lintWorkerParams, componentDir string, handler tool.LintHandler, moduleContract core.ModuleContractPort, logWriter io.Writer) (func(), string, string, int) {

	release, err := params.pipeline.AcquireLock(ctx, params.module, componentDir)
	if err != nil {
		output.Writeln(logWriter, "Error: %v", err)
		return nil, "", "", 1
	}

	outputDir := paths.UnitLintOutputPath(ctx.WorkspaceRoot, params.module, componentDir)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		release()
		output.Writeln(logWriter, "Error creating output directory: %v", err)
		return nil, "", "", 1
	}

	compRoot := moduleContract.GetComponentRoot(params.compName)
	if compRoot == "" {
		release()
		output.Writeln(logWriter, "Warning: no root found for component %s in module %s", params.compName, params.module)
		return func() {}, "", "", 0
	}
	moduleRoot := filepath.Join(ctx.WorkspaceRoot, compRoot)

	if err := handler.ValidateModule(moduleRoot, ctx.WorkspaceRoot); err != nil {
		release()
		output.Writeln(logWriter, "Module validation failed for %s: %v", handler.Name(), err)
		return nil, "", "", 1
	}

	return release, outputDir, moduleRoot, 0
}

// executeLintProvider runs the lint handler against the module and returns
// the exit code and elapsed duration.
func executeLintProvider(params *lintWorkerParams, handler tool.LintHandler, provider *config.LintProvider, moduleRoot, outputDir, workspaceRoot string, logWriter io.Writer) (int, time.Duration) {
	output.Writeln(logWriter, "━━━ Linting %s with %s ━━━", params.compName, params.providerName)

	lintStart := time.Now()

	opts := tool.LintOptions{
		Fix:       params.lintCfg.Fix,
		Config:    params.lintCfg.Config,
		InputMode: provider.GetInputMode(),
	}

	exitCode := handler.Lint(moduleRoot, workspaceRoot, outputDir, logWriter, opts)
	duration := time.Since(lintStart)

	return exitCode, duration
}

// recordLintResult records the lint outcome in the shared lintContext results map,
// aggregating by module:component key to support multi-provider results.
func recordLintResult(params *lintWorkerParams, exitCode int, duration time.Duration) {
	resultKey := params.module + ":" + params.compName
	params.lctx.mu.Lock()
	defer params.lctx.mu.Unlock()

	if existing, ok := params.lctx.results[resultKey]; ok {
		existing.Providers = append(existing.Providers, params.providerName)
		if exitCode != 0 {
			existing.Success = false
		}
		existing.Duration += duration
	} else {
		params.lctx.results[resultKey] = &LintModuleResult{
			Moniker:   resultKey,
			Success:   exitCode == 0,
			Duration:  duration,
			Providers: []string{params.providerName},
		}
	}
}

// recordLintManifest writes the UoW manifest via the pipeline and logs the
// final pass/fail status for the lint operation.
func recordLintManifest(ctx *cmdframework.ExecutionContext, params *lintWorkerParams, unitID core.UnitID, exitCode int, duration time.Duration, logWriter io.Writer) {
	inputHash := computeLintInputHash(ctx, params.module)
	params.pipeline.RecordManifest(unitID, &coreoutput.UoWManifest{
		ExitCode:   exitCode,
		InputHash:  inputHash,
		ExecutedAt: time.Now().UTC(),
		Duration:   duration,
		Version:    "1.0.0",
	})

	if exitCode != 0 {
		output.Writeln(logWriter, "❌ Lint failed for %s:%s with %s", params.module, params.compName, params.providerName)
	} else {
		output.Writeln(logWriter, "✅ Lint passed for %s:%s with %s", params.module, params.compName, params.providerName)
	}
}
