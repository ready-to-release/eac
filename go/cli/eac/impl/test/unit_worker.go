package test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/orchestrator"
	"github.com/ready-to-release/eac/go/core/execution"
	"github.com/ready-to-release/eac/go/core/hash"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// TestCacheVerifier implements execution.CacheVerifier for test commands.
// Uses UoW-level cache for fine-grained test caching.
type TestCacheVerifier struct {
	cachedUoWs    map[string]bool      // UoW longname -> cached
	uowCacheTimes map[string]time.Time // UoW longname -> cache time
	cachedModules map[string]bool      // Module-level cache (aggregated from UoWs for TUI)
}

// Verify implements execution.CacheVerifier.
func (v *TestCacheVerifier) Verify(ctx context.Context, unit workunit.UnitSpec) (execution.CacheResult, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return execution.CacheResult{}, ctx.Err()
	default:
	}

	longname := unit.ID.Longname()

	// Check UoW-level cache first
	if v.cachedUoWs != nil && v.cachedUoWs[longname] {
		return execution.CacheResult{
			Cached:    true,
			CacheTime: v.uowCacheTimes[longname],
		}, nil
	}

	// Fall back to module-level check for backwards compatibility
	if v.cachedModules != nil && v.cachedModules[unit.ID.Module] {
		return execution.CacheResult{Cached: true}, nil
	}

	return execution.CacheResult{}, nil
}

// testWorker runs tests for a module path.
func testWorker(goCtx context.Context, ctx *cmdframework.ExecutionContext, modulePath string, logWriter io.Writer) int {
	testCfg, ok := ctx.Config.TestCmdConfig.(*TestFrameworkConfig)
	if !ok || testCfg == nil {
		fmt.Fprintf(logWriter, "Error: testConfig not found or wrong type\n")
		return 1
	}

	if testCfg.ExecCtx == nil {
		return 1
	}

	tests := testCfg.ExecCtx.testsByPackage[modulePath]
	result := testCfg.ExecCtx.runPackageTests(goCtx, modulePath, tests, logWriter)

	testCfg.ExecCtx.mu.Lock()
	testCfg.ExecCtx.results[modulePath] = result
	testCfg.ExecCtx.mu.Unlock()

	if result.PackageFailed || result.TestsFailed > 0 {
		return 1
	}
	return 0
}

// testUnitWorker runs tests for a package path using component-level execution.
// This is called by the UnitScheduler for parallel test component execution.
// The orchestrator (UoW) creates the log file and output directory - worker just writes to logWriter.
func testUnitWorker(goCtx context.Context, ctx *cmdframework.ExecutionContext, spec core.UnitSpec, logWriter io.Writer) int {
	testCfg, ok := ctx.Config.TestCmdConfig.(*TestFrameworkConfig)
	if !ok || testCfg == nil {
		fmt.Fprintf(logWriter, "Error: testConfig not found or wrong type\n")
		return 1
	}

	if testCfg.ExecCtx == nil {
		fmt.Fprintf(logWriter, "Error: test execution context not initialized\n")
		return 1
	}

	// Extract identity from spec - no more string parsing
	module := spec.ID.Module
	componentType := spec.ID.ComponentType
	testType := spec.ID.Tool
	testname := spec.ID.Extra["testname"]
	component := componentType + ":" + testType + ":" + testname // For display/logging

	// Use spec's UnitID directly for cache lookup
	unitID := spec.ID

	// Check UoW-level cache via shared pipeline
	pipeline := &cmdframework.UnitPipeline{
		CachedUoWs: testCfg.CachedUoWs,
		LockStyle:  cmdframework.NoLock,
	}

	log.Debugf("[TEST-UOW-CACHE] Test worker for %s: unitID=%s", component, unitID.Longname())
	if cacheResult := pipeline.CheckCache(ctx, unitID, logWriter); cacheResult != 0 {
		return cacheResult
	}

	// Look up pkgPath from component mapping using testname
	// testname is the unique identifier within module:componentType (e.g., "impl-build")
	pkgPath, ok := testCfg.ComponentToPkgPath[testname]
	if !ok {
		fmt.Fprintf(logWriter, "Error: no pkgPath mapping for testname %s\n", testname)
		return 1
	}

	tests := testCfg.ExecCtx.testsByPackage[pkgPath]
	if len(tests) == 0 {
		fmt.Fprintf(logWriter, "No tests: Success\n")
		writeUoWTestManifest(ctx, testCfg, unitID, "", time.Now(), 0)
		return 0
	}

	// Filter tests by type if specified
	if testType != "" {
		tests = filterTestsByType(tests, testType)
		if len(tests) == 0 {
			fmt.Fprintf(logWriter, "No %s tests: Success\n", testType)
			writeUoWTestManifest(ctx, testCfg, unitID, "", time.Now(), 0)
			return 0
		}
	}

	// Use pre-computed input hash for cache consistency.
	// This ensures the hash written to the manifest matches the hash used for detection.
	var inputHash string
	if testCfg.ModuleInputHashes != nil {
		inputHash = testCfg.ModuleInputHashes[module]
	}
	if inputHash == "" {
		// Fallback: compute if not pre-computed (shouldn't happen in normal flow)
		if contract, ok := ctx.ModuleRegistry.Get(module); ok {
			inputHash, _ = computeTestInputHash(ctx, contract)
		}
	}

	// Compute the UoW output directory from spec.ID so runners can use it
	// for isolated workspaces (e.g., npm isolation). This matches the directory
	// the orchestrator creates: out/test/<module>/<dirname>
	outputDir := filepath.Join(ctx.WorkspaceRoot, spec.ID.OutDir())

	startTime := time.Now()

	// Run tests - UoW manages log file, we just write to logWriter
	// Use the testname as result key for aggregation (unique within module:componentType)
	resultKey := testname
	result := testCfg.ExecCtx.runPackageTestsDirect(goCtx, pkgPath, tests, logWriter, outputDir)

	testCfg.ExecCtx.mu.Lock()
	testCfg.ExecCtx.results[resultKey] = result
	testCfg.ExecCtx.mu.Unlock()

	// Pass test counts to orchestrator for summary display
	// Use componentType as the component name for orchestrator tracking
	ctx.Orchestrator.SetUnitExtras(module, componentType, orchestrator.UnitExtras{
		TestsTotal:   result.TestsTotal,
		TestsPassed:  result.TestsPassed,
		TestsFailed:  result.TestsFailed,
		TestsSkipped: result.TestsSkipped,
	})

	// Write UoW manifest for incremental cache
	passed := !result.PackageFailed && result.TestsFailed == 0
	exitCode := 0
	if !passed {
		exitCode = 1
	}
	writeUoWTestManifest(ctx, testCfg, unitID, inputHash, startTime, exitCode)

	if result.PackageFailed || result.TestsFailed > 0 {
		return 1
	}
	return 0
}

// filterTestsByType returns only tests matching the specified type.
func filterTestsByType(tests []testing.TestReference, testType string) []testing.TestReference {
	var filtered []testing.TestReference
	for i := range tests {
		if tests[i].Type == testType {
			filtered = append(filtered, tests[i])
		}
	}
	return filtered
}

// computeTestInputHash computes the input hash for a module's tests.
func computeTestInputHash(ctx *cmdframework.ExecutionContext, contract interface{ GetGlobPatterns() []string }) (string, error) {
	patterns := contract.GetGlobPatterns()
	files, err := hash.ExpandGlobPatterns(ctx.WorkspaceRoot, patterns)
	if err != nil {
		return "", err
	}
	return hash.Files(ctx.WorkspaceRoot, files)
}

// collectTestArtifacts scans the UoW output directory for test output files
// and returns them as artifacts for the UoW manifest.
func collectTestArtifacts(uowDir string) []coreoutput.Artifact {
	var artifacts []coreoutput.Artifact

	type artifactSpec struct {
		id      string
		artType string
	}

	knownFiles := map[string]artifactSpec{
		// test.log is excluded — it is owned by the orchestrator which appends
		// [memory] after: instrumentation after the worker has already hashed it,
		// causing every cache check to see a hash mismatch and re-execute.
		"cucumber.json": {id: "cucumber-report", artType: "cucumber-report"},
		"unit.json":     {id: "ctrf-report", artType: "ctrf-report"},
		"coverage.out":  {id: "coverage", artType: "coverage"},
	}

	entries, err := os.ReadDir(uowDir)
	if err != nil {
		return artifacts
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() == "uow.manifest.json" {
			continue
		}

		spec, known := knownFiles[entry.Name()]
		if !known {
			if strings.HasSuffix(entry.Name(), ".cucumber.json") {
				spec = artifactSpec{id: "cucumber-report", artType: "cucumber-report"}
			} else {
				continue
			}
		}

		fullPath := filepath.Join(uowDir, entry.Name())
		size, hash, err := coreoutput.HashFile(fullPath)
		if err != nil {
			continue
		}

		artifacts = append(artifacts, coreoutput.Artifact{
			ID:     spec.id,
			Path:   entry.Name(),
			SHA256: hash,
			Size:   size,
			Type:   spec.artType,
		})
	}

	return artifacts
}

// writeUoWTestManifest writes a UoW manifest for a completed test.
func writeUoWTestManifest(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig, unitID workunit.UnitID, inputHash string, startTime time.Time, exitCode int) {
	// Initialize tracker if needed (thread-safe)
	testCfg.ExecCtx.mu.Lock()
	if testCfg.Tracker == nil {
		testCfg.Tracker = coreoutput.NewTracker(ctx.WorkspaceRoot, core.ActionTest)
	}
	tracker := testCfg.Tracker
	testCfg.ExecCtx.mu.Unlock()

	// Compute UoW directory path to collect artifacts
	uowDir := filepath.Join(ctx.WorkspaceRoot, "out", "test", unitID.Module, unitID.DirName())

	// Collect artifacts from UoW output directory
	artifacts := collectTestArtifacts(uowDir)

	// Compute output hash from artifact hashes
	outputHash := coreoutput.ComputeOutputHash(artifacts)

	// Look up pre-computed tag summary for this UoW
	var tags workunit.TagSummary
	if testCfg.UoWTags != nil {
		tags = testCfg.UoWTags[unitID.Longname()]
	}

	// Create and record the manifest
	// Include Extra field for testname which ensures unique directory names
	manifest := &coreoutput.UoWManifest{
		Action:     core.ActionTest,
		Module:     unitID.Module,
		Component:  unitID.ComponentName,
		Tool:       unitID.Tool,
		Extra:      unitID.Extra, // Include testname for unique directory path
		InputHash:  inputHash,
		ExecutedAt: startTime,
		ExitCode:   exitCode,
		Duration:   time.Since(startTime),
		Artifacts:  artifacts,
		OutputHash: outputHash,
		Tags:       tags,
	}

	if err := tracker.RecordComplete(unitID, manifest); err != nil {
		log.Debugf("[TEST-UOW-CACHE] Failed to write UoW manifest for %s: %v", unitID.Longname(), err)
	} else {
		log.Debugf("[TEST-UOW-CACHE] Wrote UoW manifest for %s (exitCode=%d, artifacts=%d)",
			unitID.Longname(), exitCode, len(artifacts))
	}
}
