package lint

import (
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	coreoutput "github.com/ready-to-release/eac/go/core/output"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/environments"
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
