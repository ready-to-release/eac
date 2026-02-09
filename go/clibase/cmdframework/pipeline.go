package cmdframework

import (
	"context"
	"io"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/locking"
	"github.com/ready-to-release/eac/go/clibase/output"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// LockStyle controls when unit locks are acquired during pipeline execution.
type LockStyle int

const (
	// NoLock never acquires locks. Used by test workers.
	NoLock LockStyle = iota
	// LockUnlessDryRun acquires locks except in dry-run mode. Used by build and lint workers.
	LockUnlessDryRun
	// AlwaysLock always acquires locks regardless of dry-run. Used by scan workers.
	AlwaysLock
)

// UnitPipeline provides shared orchestration for unit worker execution.
// It handles the common steps shared by build, test, lint, and scan unit workers:
//   - Cache checking with optional artifact integrity validation
//   - Lock acquisition with configurable style (none, skip-in-dryrun, always)
//   - UoW manifest recording via tracker
//
// Each unit worker creates a pipeline configured for its action type, then calls
// CheckCache, AcquireLock, and RecordManifest at the appropriate points in its flow.
type UnitPipeline struct {
	// CachedUoWs maps UoW longnames to cached status.
	// Nil disables caching entirely.
	CachedUoWs map[string]bool

	// ValidateCacheArtifacts enables manifest artifact integrity validation on cache hits.
	// When true, a cache hit is additionally verified by checking that manifest artifacts
	// exist and are intact. If validation fails, the cache entry is invalidated and
	// execution proceeds normally. Used by build and lint; test and scan skip this.
	ValidateCacheArtifacts bool

	// OnCacheInvalidated is called when a cached entry fails artifact validation.
	// Typically used to remove the entry from the caller's CachedUoWs map.
	OnCacheInvalidated func(longname string)

	// LockStyle controls when locks are acquired.
	LockStyle LockStyle

	// LockConfigFn creates a locking.Config for the given module and component directory.
	// Required when LockStyle is not NoLock.
	LockConfigFn func(module, componentDir string) locking.Config

	// Tracker writes UoW manifests. May be nil to skip manifest recording.
	Tracker *coreoutput.InMemoryTracker
}

// CheckCache checks if a unit is cached and handles the cache hit flow.
// Returns:
//
//	-1: unit is cached (caller should return -1 for blue TUI status)
//	 0: unit is not cached (caller should continue with execution)
//
// In dry-run mode, cached units log a "would be skipped" message.
// When ValidateCacheArtifacts is true, cache hits are verified by checking
// that manifest artifacts are intact. Invalid artifacts cause a cache miss (returns 0).
func (p *UnitPipeline) CheckCache(ctx *ExecutionContext, unitID workunit.UnitID, logWriter io.Writer) int {
	longname := unitID.Longname()
	isCached := p.CachedUoWs != nil && p.CachedUoWs[longname]
	if !isCached {
		return 0
	}

	if ctx.Config.DryRun {
		output.Writeln(logWriter, "⏭️  %s is up-to-date (would be skipped)", unitID.Module)
		return -1
	}

	if p.ValidateCacheArtifacts {
		reader := coreoutput.NewReader(ctx.WorkspaceRoot)
		validation := reader.ValidateUoW(unitID)
		if !validation.ManifestExists {
			// No manifest to verify — trust source hash, allow cache hit
			log.Debugf("UoW cache hit (no manifest to verify)")
			output.Writeln(logWriter, "⏭️  Cached (unchanged)")
			return -1
		}
		if !validation.Valid {
			output.Writeln(logWriter, "UoW cache miss: artifacts invalid")
			if p.OnCacheInvalidated != nil {
				p.OnCacheInvalidated(longname)
			}
			return 0 // Fall through to execution
		}
		output.Writeln(logWriter, "⏭️  Cached (verified)")
		return -1
	}

	output.Writeln(logWriter, "⏭️  Cached (unchanged)")
	return -1
}

// AcquireLock acquires a unit-level lock based on the pipeline's LockStyle.
// Returns a release function that the caller must defer.
// If no lock is needed (NoLock or DryRun with LockUnlessDryRun), release is a no-op.
func (p *UnitPipeline) AcquireLock(ctx *ExecutionContext, module, componentDir string) (release func(), err error) {
	noop := func() {}

	switch p.LockStyle {
	case NoLock:
		return noop, nil
	case LockUnlessDryRun:
		if ctx.Config.DryRun {
			return noop, nil
		}
	case AlwaysLock:
		// Always acquire
	}

	cfg := p.LockConfigFn(module, componentDir)
	lockFile, err := locking.AcquireWithWait(context.Background(), ctx.WorkspaceRoot, cfg,
		ctx.Orchestrator.GetRegistry(), locking.DefaultWaitConfig())
	if err != nil {
		return noop, err
	}
	return func() { locking.ReleaseTracked(lockFile) }, nil
}

// RecordManifest writes a UoW manifest via the tracker.
// No-op if the tracker or manifest is nil.
func (p *UnitPipeline) RecordManifest(unitID workunit.UnitID, manifest *coreoutput.UoWManifest) {
	if p.Tracker == nil || manifest == nil {
		return
	}
	if err := p.Tracker.RecordComplete(unitID, manifest); err != nil {
		log.Debugf("Failed to record UoW completion for %s: %v", unitID.Longname(), err)
	}
}

// ParseUnit splits a "compName:toolName" unit string into its parts.
// Returns the component name and tool/provider name (empty if no colon separator).
func ParseUnit(unit string) (compName, toolName string) {
	parts := strings.SplitN(unit, ":", 2)
	compName = parts[0]
	if len(parts) == 2 {
		toolName = parts[1]
	}
	return
}

// UnitDir returns the directory name for a component+tool unit combination.
// Uses dash separator to match UnitID.DirName() for consistent manifest paths.
func UnitDir(compName, toolName string) string {
	if toolName == "" {
		return compName
	}
	return compName + "-" + toolName
}
