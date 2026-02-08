// Package build provides the build command implementation using cmdframework.
package build

import (
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/core/hash"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// detectUoWIncrementalChanges performs UoW-level incremental build detection.
// Instead of checking at module granularity, it checks each component:tool UoW.
// This enables partial caching - some components can be cached while others rebuild.
//
// The function:
// 1. Builds expected UoW list from resolved unit specs
// 2. Uses DiskOutputReader.DetectUoWChanges() for change detection
// 3. Populates bctx.cachedUoWs with component-level cache status
// 4. Aggregates to bctx.cachedModules for TUI compatibility
func detectUoWIncrementalChanges(ctx *cmdframework.ExecutionContext, bctx *buildContext) {
	startTime := time.Now()
	defer func() {
		ctx.SetChangeDetectionTiming(time.Since(startTime))
	}()

	// Get all expected UoWs from resolved unit specs
	specs := getCachedUnitWork(ctx)
	if len(specs) == 0 {
		return
	}

	// Build list of expected UoWs
	var expectedUoWs []workunit.UnitID
	for _, spec := range specs {
		expectedUoWs = append(expectedUoWs, spec.ID)
	}

	if len(expectedUoWs) == 0 {
		return
	}

	// Reuse pre-expanded module files from hash cache when available.
	// Falls back to expanding patterns if not pre-computed (e.g., tests).
	moduleFiles := bctx.moduleExpandedFiles
	if moduleFiles == nil {
		moduleFiles = make(map[string][]string)
		for _, id := range expectedUoWs {
			if _, ok := moduleFiles[id.Module]; ok {
				continue
			}
			if contract, ok := ctx.ModuleRegistry.Get(id.Module); ok {
				patterns := contract.GetGlobPatterns()
				files, err := hash.ExpandGlobPatterns(ctx.WorkspaceRoot, patterns)
				if err != nil {
					log.Debugf("Failed to expand patterns for %s: %v", id.Module, err)
					continue
				}
				moduleFiles[id.Module] = files
			}
		}
	}

	// Create input hash provider using pre-computed module hashes.
	// Falls back to computing from files if pre-computed hash is not available.
	getInputHash := func(id workunit.UnitID) (string, error) {
		// Use pre-computed hash if available (ensures consistency with build manifests)
		if h, ok := bctx.moduleInputHashes[id.Module]; ok && h != "" {
			return h, nil
		}
		// Fallback: compute from files
		files, ok := moduleFiles[id.Module]
		if !ok {
			return "", nil // No files = treat as changed
		}
		return hash.Files(ctx.WorkspaceRoot, files)
	}

	// Create dependency resolver from module registry for cross-module
	// build cache invalidation (e.g., ext-eac depends on eac-cli binary).
	var depResolver coreoutput.DependencyResolver
	if ctx.ModuleRegistry != nil {
		depResolver = func(module string) []string {
			if contract, ok := ctx.ModuleRegistry.Get(module); ok {
				return contract.GetDependencies()
			}
			return nil
		}
	}

	// Use DiskOutputReader for UoW-level change detection
	reader := coreoutput.NewReader(ctx.WorkspaceRoot)
	changeResult, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, depResolver)
	if err != nil {
		log.Debugf("Failed to detect UoW changes: %v", err)
		return
	}

	log.Debugf("[UOW-CACHE] DetectUoWChanges result: FreshRun=%v Changed=%d UpToDate=%d",
		changeResult.FreshRun, len(changeResult.Changed), len(changeResult.UpToDate))
	for longname, reason := range changeResult.ChangeReasons {
		log.Debugf("[UOW-CACHE] Changed: %s -> %s", longname, reason)
	}

	detectionTime := time.Since(startTime)

	if changeResult.FreshRun {
		log.Debugf("Fresh build detected (UoW mode), all components will build")
		if ctx.InitSummary != nil {
			ctx.InitSummary.SetIncremental(&initsummary.IncrementalInfo{
				Enabled:       true,
				DetectionTime: detectionTime,
				FreshBuild:    true,
			})
		}
		return
	}

	buildCfg := bctx.cfg

	// Initialize UoW-level cache maps
	bctx.cachedUoWs = make(map[string]bool)
	bctx.uowCacheTimes = make(map[string]time.Time)

	// Also maintain module-level maps for TUI compatibility
	bctx.cachedModules = make(map[string]bool)
	bctx.cacheTimes = make(map[string]time.Time)

	// Track which modules have all UoWs cached vs some changed
	moduleUoWCounts := make(map[string]int)    // Total UoWs per module
	moduleCachedCounts := make(map[string]int) // Cached UoWs per module

	// Count total UoWs per module
	for _, id := range expectedUoWs {
		moduleUoWCounts[id.Module]++
	}

	// Process up-to-date UoWs
	for _, id := range changeResult.UpToDate {
		longname := id.Longname()

		// Skip if explicitly requested (bypass cache)
		if !ctx.Config.DryRun && buildCfg.RequestedSet[id.Module] {
			continue
		}

		bctx.cachedUoWs[longname] = true
		moduleCachedCounts[id.Module]++

		// Try to load cache time from manifest
		if manifest, err := reader.GetUoW(id); err == nil {
			bctx.uowCacheTimes[longname] = manifest.ExecutedAt
		}
	}

	// Aggregate to module level: a module is cached only if ALL its UoWs are cached
	// Also collect lists for reporting
	var changedList []string
	var cachedList []string

	seenModules := make(map[string]bool)
	for _, id := range expectedUoWs {
		if seenModules[id.Module] {
			continue
		}
		seenModules[id.Module] = true

		totalUoWs := moduleUoWCounts[id.Module]
		cachedUoWs := moduleCachedCounts[id.Module]

		// Module is cached only if ALL its UoWs are cached
		if cachedUoWs == totalUoWs && cachedUoWs > 0 {
			bctx.cachedModules[id.Module] = true
			cachedList = append(cachedList, id.Module)

			// Use earliest cache time from any UoW in the module
			for _, uid := range expectedUoWs {
				if uid.Module == id.Module {
					if t, ok := bctx.uowCacheTimes[uid.Longname()]; ok {
						if existing, exists := bctx.cacheTimes[id.Module]; !exists || t.Before(existing) {
							bctx.cacheTimes[id.Module] = t
						}
					}
				}
			}
		} else {
			changedList = append(changedList, id.Module)
		}

		log.Debugf("[UOW-CACHE] Module %s: %d/%d UoWs cached -> module cached=%v",
			id.Module, cachedUoWs, totalUoWs, cachedUoWs == totalUoWs)
	}

	// Report incremental detection in init summary
	if ctx.InitSummary != nil {
		ctx.InitSummary.SetIncremental(&initsummary.IncrementalInfo{
			Enabled:       true,
			DetectionTime: detectionTime,
			Changed:       changedList,
			UpToDate:      cachedList,
			FreshBuild:    false,
		})
	}

	log.Debugf("Incremental (UoW mode): %d modules to build, %d cached, %d UoWs cached",
		len(changedList), len(cachedList), len(bctx.cachedUoWs))
}

// isUoWCached checks if a specific UoW is cached.
// Used by buildUnitWorker when UoW caching is enabled.
func isUoWCached(bctx *buildContext, unitID workunit.UnitID) bool {
	if bctx.cachedUoWs == nil {
		return false
	}
	return bctx.cachedUoWs[unitID.Longname()]
}

// getUoWCacheTime returns the cache time for a specific UoW.
func getUoWCacheTime(bctx *buildContext, unitID workunit.UnitID) time.Time {
	if bctx.uowCacheTimes == nil {
		return time.Time{}
	}
	return bctx.uowCacheTimes[unitID.Longname()]
}
