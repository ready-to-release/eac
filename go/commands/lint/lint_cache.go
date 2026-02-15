package lint

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/caching"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/core/hash"
)

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
