// Package caching provides shared incremental change detection for commands.
//
// This package extracts the common UoW-level incremental detection pattern
// used by lint, scan, and test commands. The pattern involves:
//  1. Resolving unit specs into expected UoW IDs
//  2. Expanding glob patterns to collect module files
//  3. Pre-computing module input hashes for cache consistency
//  4. Calling AggregateUoWChanges for UoW-level change detection
//  5. Logging results and reporting to InitSummary
//
// Build command has unique logic (pre-expanded files, RequestedSet bypass)
// and is intentionally not included here.
package caching

import (
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/core/hash"
	"github.com/ready-to-release/eac/go/core/logging"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// IncrementalResult holds the results of incremental change detection.
// Callers copy the fields they need into their command-specific context structs.
type IncrementalResult struct {
	// UoW-level cache state
	CachedUoWs    map[string]bool      // UoW longname -> cached
	UoWCacheTimes map[string]time.Time // UoW longname -> cache time

	// Module-level aggregated state (derived from UoW results)
	CachedModules    map[string]bool      // Module -> all UoWs cached
	ModuleCacheTimes map[string]time.Time // Module -> earliest UoW cache time

	// Pre-computed input hashes for workers to reuse
	ModuleInputHashes map[string]string // Module -> hash

	// Module lists for init summary reporting
	ChangedModules  []string
	UpToDateModules []string

	// FreshRun indicates no prior cache exists (all modules should execute)
	FreshRun bool
}

// DetectIncrementalChanges performs UoW-level incremental change detection.
// It returns nil if there are no specs or on error (callers should treat nil as "run everything").
//
// The function:
//  1. Extracts UnitIDs from the provided specs
//  2. Expands glob patterns for each module to collect files
//  3. Pre-computes module input hashes ONCE for consistency (workers reuse these)
//  4. Calls AggregateUoWChanges for UoW-level change detection
//  5. Logs change detection results with the provided logPrefix
//  6. Reports FreshRun or Changed/UpToDate to InitSummary
//  7. Logs module-level aggregation stats
func DetectIncrementalChanges(
	ctx *cmdframework.ExecutionContext,
	wuContext core.ActionType,
	specs []workunit.UnitSpec,
	logPrefix string,
) *IncrementalResult {
	log := logging.C()
	startTime := time.Now()
	defer func() {
		ctx.SetChangeDetectionTiming(time.Since(startTime))
	}()

	if len(specs) == 0 {
		return nil
	}

	// Build list of expected UoWs from resolved specs
	expectedUoWs := make([]workunit.UnitID, len(specs))
	for i, spec := range specs {
		expectedUoWs[i] = spec.ID
	}

	// Collect module files for hash computation (deduplicated by module)
	moduleFiles := make(map[string][]string)
	for _, id := range expectedUoWs {
		if _, ok := moduleFiles[id.Module]; ok {
			continue // Already collected
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

	// Pre-compute module input hashes ONCE for consistency.
	// Workers will reuse these instead of recomputing independently.
	moduleInputHashes := make(map[string]string)
	for module, files := range moduleFiles {
		h, err := hash.Files(ctx.WorkspaceRoot, files)
		if err != nil {
			log.Debugf("Failed to compute input hash for %s: %v", module, err)
			continue
		}
		moduleInputHashes[module] = h
	}

	// Use pre-computed hashes for detection (same hashes workers will use)
	reader := coreoutput.NewReader(ctx.WorkspaceRoot)
	getInputHash := coreoutput.InputHashProvider(func(id workunit.UnitID) (string, error) {
		h, ok := moduleInputHashes[id.Module]
		if !ok {
			return "", nil
		}
		return h, nil
	})

	aggResult, err := coreoutput.AggregateUoWChanges(reader, wuContext, expectedUoWs, getInputHash)
	if err != nil {
		log.Debugf("Failed to detect %s UoW changes: %v", logPrefix, err)
		return nil
	}

	// Log change detection results
	log.Debugf("[%s-UOW-CACHE] DetectUoWChanges result: FreshRun=%v Changed=%d UpToDate=%d",
		logPrefix, aggResult.UoWResult.FreshRun, len(aggResult.UoWResult.Changed), len(aggResult.UoWResult.UpToDate))
	for longname, reason := range aggResult.UoWResult.ChangeReasons {
		log.Debugf("[%s-UOW-CACHE] Changed: %s -> %s", logPrefix, longname, reason)
	}

	detectionTime := time.Since(startTime)

	// Handle fresh run - no prior cache exists
	if aggResult.UoWResult.FreshRun {
		log.Debugf("Fresh %s detected (UoW mode), all components will execute", logPrefix)
		if ctx.InitSummary != nil {
			ctx.InitSummary.SetIncremental(&initsummary.IncrementalInfo{
				Enabled:       true,
				DetectionTime: detectionTime,
				FreshBuild:    true,
			})
		}
		return &IncrementalResult{
			ModuleInputHashes: moduleInputHashes,
			FreshRun:          true,
		}
	}

	// Log module-level aggregation stats
	agg := workunit.NewUoWAggregator(expectedUoWs)
	for _, id := range aggResult.UoWResult.UpToDate {
		agg.MarkCached(id)
	}
	for module := range aggResult.CachedModules {
		total, cached := agg.Stats(module)
		log.Debugf("[%s-UOW-CACHE] Module %s: %d/%d UoWs cached -> module cached=%v",
			logPrefix, module, cached, total, true)
	}
	for _, module := range aggResult.ChangedModules {
		total, cached := agg.Stats(module)
		log.Debugf("[%s-UOW-CACHE] Module %s: %d/%d UoWs cached -> module cached=%v",
			logPrefix, module, cached, total, false)
	}

	// Report incremental detection in init summary
	if ctx.InitSummary != nil {
		ctx.InitSummary.SetIncremental(&initsummary.IncrementalInfo{
			Enabled:       true,
			DetectionTime: detectionTime,
			Changed:       aggResult.ChangedModules,
			UpToDate:      aggResult.UpToDateModules,
			FreshBuild:    false,
		})
	}

	log.Debugf("Incremental %s (UoW mode): %d modules to execute, %d cached, %d UoWs cached",
		logPrefix, len(aggResult.ChangedModules), len(aggResult.UpToDateModules), len(aggResult.CachedUoWs))

	return &IncrementalResult{
		CachedUoWs:        aggResult.CachedUoWs,
		UoWCacheTimes:     aggResult.UoWCacheTimes,
		CachedModules:     aggResult.CachedModules,
		ModuleCacheTimes:  aggResult.ModuleCacheTimes,
		ModuleInputHashes: moduleInputHashes,
		ChangedModules:    aggResult.ChangedModules,
		UpToDateModules:   aggResult.UpToDateModules,
		FreshRun:          false,
	}
}
