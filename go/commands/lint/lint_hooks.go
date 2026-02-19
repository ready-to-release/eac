package lint

import (
	"os"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/hash"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
)

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

	// Incremental lint detection (devbox only, not CI — unless dry-run)
	// Dry-run always checks cache to correctly report skip/execute status
	if !environments.IsCI() || ctx.Config.DryRun {
		detectUoWIncrementalLintChanges(ctx, lctx)

		// Pass cache times to orchestrator for TUI display
		if len(lctx.cacheTimes) > 0 && ctx.Orchestrator != nil {
			ctx.Orchestrator.SetCacheTimes(lctx.cacheTimes)
		}

		// Enable early cache detection for fast TUI feedback
		// Tabs will progressively "light up" blue as cache hits are detected
		if (len(lctx.cachedUoWs) > 0 || len(lctx.cachedModules) > 0) && ctx.Orchestrator != nil {
			verifier := &LintCacheVerifier{
				cachedUoWs:    lctx.cachedUoWs,
				uowCacheTimes: lctx.uowCacheTimes,
				cachedModules: lctx.cachedModules,
			}
			ctx.Orchestrator.SetCacheDetection(verifier, lctx.cachedModules)
		}
	}

	// Pre-compute module input hashes if not already set by incremental detection.
	// In CI mode, incremental detection is skipped, so hashes are never computed.
	// Pre-computing ensures all workers for the same module get a consistent hash,
	// preventing divergence when parallel workers modify shared files.
	preComputeModuleInputHashes(ctx, lctx)

	return nil
}

// preComputeModuleInputHashes pre-computes input hashes for all execution modules.
// Skips modules that already have hashes from incremental detection.
func preComputeModuleInputHashes(ctx *cmdframework.ExecutionContext, lctx *lintContext) {
	if ctx.ModuleRegistry == nil {
		return
	}

	if lctx.moduleInputHashes == nil {
		lctx.moduleInputHashes = make(map[string]string)
	}

	for _, moniker := range ctx.GetExecutionMonikers() {
		// Skip if already pre-computed by incremental detection
		if _, ok := lctx.moduleInputHashes[moniker]; ok {
			continue
		}

		contract, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		patterns := contract.GetGlobPatterns()
		files, err := hash.ExpandGlobPatterns(ctx.WorkspaceRoot, patterns)
		if err != nil {
			log.Debugf("Failed to expand patterns for input hash of %s: %v", moniker, err)
			continue
		}

		h, err := hash.Files(ctx.WorkspaceRoot, files)
		if err != nil {
			log.Debugf("Failed to compute input hash for %s: %v", moniker, err)
			continue
		}

		lctx.moduleInputHashes[moniker] = h
	}
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
