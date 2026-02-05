// Package build provides the build command implementation using cmdframework.
package build

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/builders"
	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/artifacts"
	"github.com/ready-to-release/eac/go/core/adapters"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/environment"
	"github.com/ready-to-release/eac/go/clibase/git"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/clibase/locking"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/hash"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/workunit"
)

func init() {
	// Register component-level execution support
	cmdframework.RegisterUnitProvider(cmdframework.CommandTypeBuild, ResolveUnitSpecs)
	cmdframework.RegisterUnitWorker(cmdframework.CommandTypeBuild, buildUnitWorker)
	cmdframework.SetUoWCountProvider(getBuildUoWCount)
}

// getCachedUnitWork returns cached component work specs, computing once if needed.
// This avoids duplicate calls to ResolveUnitSpecs during startup.
func getCachedUnitWork(ctx *cmdframework.ExecutionContext) []workunit.UnitSpec {
	// Check if already cached in Extra
	if ctx.Config.Extra == nil {
		ctx.Config.Extra = make(map[string]interface{})
	}
	if cached, ok := ctx.Config.Extra["componentWorkSpecs"]; ok {
		return cached.([]workunit.UnitSpec)
	}

	// Compute and cache
	specs := ResolveUnitSpecs(ctx)
	ctx.Config.Extra["componentWorkSpecs"] = specs
	return specs
}

// getBuildUoWCount returns the total number of buildable UoWs (units of work).
func getBuildUoWCount(ctx *cmdframework.ExecutionContext) int {
	specs := getCachedUnitWork(ctx)
	return CountUnits(specs)
}

// BuildConfig holds build-specific configuration.
type BuildConfig struct {
	TidyFirst       bool
	Version         string
	UseExistingDepm bool

	// Reproducible controls determinism behavior for MkDocs builds.
	// Values: "auto" (default), "true", "false"
	// - auto: CI=true (always rebuild), local=false (use cache)
	// - true: Always rebuild HTML from staging (CI default)
	// - false: Skip MkDocs if staging unchanged (local default)
	Reproducible string

	// ArtifactsMode controls the scope of artifact generation.
	// Values: "all" (CI default), "reduced" (local default)
	// - all: Build all artifacts for all platforms
	// - reduced: Build reduced artifacts for faster local builds
	// Use ArtifactsMode.AllArtifactsRequested() to check if all artifacts should be built.
	ArtifactsMode environment.ArtifactsMode

	// Set of originally requested modules (for --use-existing-depm logic)
	RequestedSet map[string]bool

	// Incremental build state
	SkippedModules []string
}

// ResolveReproducible converts the Reproducible string value to a boolean.
// "auto" resolves based on CI environment, "true"/"false" return literal values.
func (c *BuildConfig) ResolveReproducible() bool {
	switch c.Reproducible {
	case "true":
		return true
	case "false":
		return false
	case "auto", "":
		// Auto mode: use CI environment detection
		env := environment.Detect()
		return env.IsCI
	default:
		return false
	}
}

// buildContext holds build-specific state during execution.
type buildContext struct {
	cfg              *BuildConfig
	cachedModules    map[string]bool      // Modules that are up-to-date (aggregated from UoWs for TUI)
	cacheTimes       map[string]time.Time // When cached modules were last built
	componentWeights map[string]int       // Component weights: "module:component" -> weight
	tracker          *coreoutput.InMemoryTracker // UoW manifest tracker

	// UoW-level cache tracking
	cachedUoWs    map[string]bool      // UoW longname -> cached (e.g., "build:core:go:go")
	uowCacheTimes map[string]time.Time // UoW longname -> cache time

	// Pre-computed module input hashes (computed once before worker dispatch).
	// All components in a module share the same hash to avoid divergence
	// when parallel builds modify shared files (e.g., go.sum via go mod tidy).
	moduleInputHashes map[string]string // module moniker -> input hash
}

// RunBuildWithFramework executes build using the cmdframework.
// This is the new implementation that will replace buildMultipleModules.
func RunBuildWithFramework(cmdCfg *cmdframework.CommandConfig, buildCfg *BuildConfig) int {
	// Store build config in Extra for access in hooks/worker
	if cmdCfg.Extra == nil {
		cmdCfg.Extra = make(map[string]interface{})
	}
	cmdCfg.Extra["buildConfig"] = buildCfg

	// Create build context
	bctx := &buildContext{
		cfg:              buildCfg,
		componentWeights: make(map[string]int),
	}
	cmdCfg.Extra["buildContext"] = bctx

	// Set up hooks
	hooks := &cmdframework.Hooks{
		AfterInit:    buildAfterInit,
		AfterResolve: buildAfterResolve,
		AfterExecute: buildAfterExecute,
	}

	// Register artifact validator
	cmdframework.SetArtifactValidator(buildArtifactValidator)

	// Register deps verifier
	cmdframework.SetDepsVerifier(buildDepsVerifier)

	return cmdframework.Run(cmdCfg, nil, hooks)
}

// buildAfterInit handles build-specific initialization after framework init.
func buildAfterInit(ctx *cmdframework.ExecutionContext) error {
	// Build init summary early so we can set incremental info in AfterResolve
	buildCfg := ctx.Config.Extra["buildConfig"].(*BuildConfig)
	summary := initsummary.New("build").
		SetRequest(ctx.Config.Monikers, ctx.GetExecutionMonikers()).
		SetExecutionContext(string(logging.GetExecutionContext())).
		SetFlags(initsummary.Flags{
			DebugMode:     ctx.Config.DebugMode,
			UseTUI:        ctx.Config.UseTUI,
			TidyFirst:     buildCfg.TidyFirst,
			ArtifactsMode: buildCfg.ArtifactsMode.String(),
		}).
		SetOutputDir(paths.OutBuildRelPath)

	ctx.InitSummary = summary

	// Initialize UoW tracker for manifest generation
	bctx := ctx.Config.Extra["buildContext"].(*buildContext)
	bctx.tracker = coreoutput.NewTracker(ctx.WorkspaceRoot, workunit.ContextBuild)

	return nil
}

// buildAfterResolve handles --use-existing-depm filtering and incremental detection.
func buildAfterResolve(ctx *cmdframework.ExecutionContext) error {
	buildCfg, ok := ctx.Config.Extra["buildConfig"].(*BuildConfig)
	if !ok {
		return fmt.Errorf("buildConfig not found or wrong type")
	}
	bctx, ok := ctx.Config.Extra["buildContext"].(*buildContext)
	if !ok {
		return fmt.Errorf("buildContext not found or wrong type")
	}

	// Handle --use-existing-depm: filter out deps that already have artifacts
	if buildCfg.UseExistingDepm && !ctx.Config.DryRun && ctx.ExecutionPlan != nil {
		var filteredOrder []string

		for _, m := range ctx.ExecutionPlan.ExecutionOrder {
			// Always include requested modules
			if buildCfg.RequestedSet[m] {
				filteredOrder = append(filteredOrder, m)
				continue
			}
			// For deps, check if artifacts exist
			if hasExistingArtifacts(m, ctx.WorkspaceRoot, buildCfg.ArtifactsMode.AllArtifactsRequested()) {
				log.Debugf("Skipping dep %s (artifacts exist)", m)
			} else {
				filteredOrder = append(filteredOrder, m)
			}
		}

		ctx.ExecutionPlan.ExecutionOrder = filteredOrder
	}

	// Pre-compute module input hashes ONCE before any workers run.
	// This ensures all components in a module get the same hash, avoiding
	// divergence when parallel builds modify shared files (e.g., go.sum via go mod tidy).
	bctx.moduleInputHashes = make(map[string]string)
	if ctx.ModuleRegistry != nil && ctx.ExecutionPlan != nil {
		for _, module := range ctx.ExecutionPlan.ExecutionOrder {
			if contract, ok := ctx.ModuleRegistry.Get(module); ok {
				bctx.moduleInputHashes[module] = computeComponentInputHash(ctx, module, contract)
			}
		}
	}

	// Incremental build detection (local only)
	// Also run in dry-run mode to show which modules would be built/skipped
	if !ctx.Config.ForceRebuild && !buildCfg.UseExistingDepm {
		// UoW-based incremental detection
		detectUoWIncrementalChanges(ctx, bctx)

		// Pass cache times to orchestrator for TUI display
		if len(bctx.cacheTimes) > 0 && ctx.Orchestrator != nil {
			ctx.Orchestrator.SetCacheTimes(bctx.cacheTimes)
		}

		// Enable early cache detection for fast TUI feedback
		// Tabs will progressively "light up" blue as cache hits are detected
		if len(bctx.cachedModules) > 0 && ctx.Orchestrator != nil {
			verifier := NewUoWBuildCacheVerifier(ctx.WorkspaceRoot, bctx.cachedUoWs, bctx.uowCacheTimes, bctx.cachedModules)
			ctx.Orchestrator.SetCacheDetection(verifier, bctx.cachedModules)
		}
	}

	// Populate componentWeights from UnitSpecs for resource scaling
	// This ensures container builds get proper CPU/memory allocation based on component type
	populateComponentWeights(ctx, bctx)

	return nil
}

// populateComponentWeights extracts weight values from UnitSpecs into the buildContext.
// This is called after component resolution so weights can be looked up during execution.
// The weights are used to scale container resources (CPU, memory) appropriately.
func populateComponentWeights(ctx *cmdframework.ExecutionContext, bctx *buildContext) {
	specs := getCachedUnitWork(ctx)
	if len(specs) == 0 {
		return
	}

	for _, spec := range specs {
		// Key format: "module:component" (without tool suffix)
		key := spec.ID.Module + ":" + spec.ID.Component
		if spec.ID.Tool != "" {
			// Include tool for unique identification when multiple tools per component
			key = spec.ID.Module + ":" + spec.ID.Component + ":" + spec.ID.Tool
		}
		bctx.componentWeights[key] = spec.Weight
		log.Debugf("[WEIGHT] Stored weight for %s: %d", key, spec.Weight)
	}

	log.Debugf("[WEIGHT] Populated %d component weights", len(bctx.componentWeights))
}


// buildAfterExecute handles post-build tasks: artifact derivations and post-build steps.
// Note: UoW manifests are written atomically during build execution via InMemoryTracker.
func buildAfterExecute(ctx *cmdframework.ExecutionContext) error {
	buildCfg, ok := ctx.Config.Extra["buildConfig"].(*BuildConfig)
	if !ok {
		return fmt.Errorf("buildConfig not found or wrong type")
	}

	// Check if any modules were actually built (vs all cached)
	// If all modules are cached (exit code -1), skip expensive post-processing
	anyBuilt := false
	for _, result := range ctx.Results {
		log.Debugf("buildAfterExecute: checking result %s, exitCode=%d", result.Moniker, result.ExitCode)
		if result.ExitCode == 0 {
			anyBuilt = true
			break
		}
	}
	log.Debugf("buildAfterExecute: anyBuilt=%v, resultCount=%d", anyBuilt, len(ctx.Results))

	// Process artifact derivations for all successfully built modules
	// This includes compression, UPX, and other post-processing
	if !ctx.Config.DryRun && anyBuilt {
		start := time.Now()
		if err := processAllArtifactDerivations(ctx, buildCfg); err != nil {
			return fmt.Errorf("artifact derivation failed: %w", err)
		}
		log.Debugf("buildAfterExecute: processAllArtifactDerivations took %v", time.Since(start))
	}

	// Assert all UoWs have valid manifests (skip in dry-run mode)
	if !ctx.Config.DryRun {
		if err := assertBuildManifestsExist(ctx); err != nil {
			return err
		}
	}

	return nil
}

// assertBuildManifestsExist verifies that all executed UoWs have valid manifests.
func assertBuildManifestsExist(ctx *cmdframework.ExecutionContext) error {
	return cmdframework.AssertManifestsExist(ctx, "build", getCachedUnitWork(ctx))
}

// processAllArtifactDerivations runs artifact derivations for all successfully built modules.
func processAllArtifactDerivations(ctx *cmdframework.ExecutionContext, buildCfg *BuildConfig) error {
	cfg := config.Global()
	if cfg == nil {
		return nil
	}

	for _, result := range ctx.Results {
		if result.ExitCode != 0 {
			continue // Skip failed modules
		}

		moniker := result.Moniker
		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		// Get merged artifacts (module-level takes priority over type-level)
		mergedArtifacts := cfg.GetBuildArtifacts(moniker, buildCfg.ArtifactsMode.AllArtifactsRequested())
		if len(mergedArtifacts) == 0 {
			continue
		}

		// Determine which artifacts were requested
		requestedArtifacts := determineRequestedArtifactsForBuild(module, buildCfg.ArtifactsMode, ctx.WorkspaceRoot)

		// Build output directory
		moduleOutputDir := paths.BuildOutputPath(ctx.WorkspaceRoot, moniker)

		// Process derivations (compression, UPX, etc.)
		if err := ProcessArtifactDerivations(moniker, mergedArtifacts, moduleOutputDir, requestedArtifacts, module.Metadata, io.Discard); err != nil {
			log.Warnf("Artifact derivation warning for %s: %v", moniker, err)
			// Continue with other modules - derivation failure is not fatal
		}
	}

	// Execute post-build steps for each successfully built component
	for _, compResult := range ctx.UnitResults {
		if compResult.ExitCode != 0 {
			continue // Skip failed components
		}

		// Reconstruct component directory path matching build output structure.
		// Build uses "component-handler" format (e.g., "typescript-npm-build") for the output path.
		// Must match UnitID.DirName() which uses dash separator.
		componentDir := compResult.Component
		if compResult.Handler != "" {
			componentDir = compResult.Component + "-" + compResult.Handler
		}
		componentOutputDir := paths.ComponentBuildOutputPath(ctx.WorkspaceRoot, compResult.Module, componentDir)
		if exitCode := builders.ExecutePostBuildSteps(compResult.Module, compResult.Component, ctx.WorkspaceRoot, componentOutputDir, io.Discard); exitCode != 0 {
			log.Warnf("Post-build steps warning for %s/%s: exit code %d", compResult.Module, compResult.Component, exitCode)
			// Continue with other components
		}
	}

	return nil
}


// buildUnitWorker builds a single component within a module.
// This is called by the UnitScheduler for parallel component execution.
// The component parameter is in "compName:builderName" format (e.g., "go:go", "docs:mkdocs").
func buildUnitWorker(ctx *cmdframework.ExecutionContext, module, component string, logWriter io.Writer) int {
	buildCfg, ok := ctx.Config.Extra["buildConfig"].(*BuildConfig)
	if !ok {
		output.Writeln(logWriter, "Error: buildConfig not found or wrong type")
		return 1
	}
	bctx, ok := ctx.Config.Extra["buildContext"].(*buildContext)
	if !ok {
		output.Writeln(logWriter, "Error: buildContext not found or wrong type")
		return 1
	}

	// Check incremental cache first - if module/UoW is cached, verify artifacts and skip (blue in TUI)
	// Parse component to build unitID for UoW cache lookup
	parts := strings.SplitN(component, ":", 2)
	compName := parts[0]
	toolName := ""
	if len(parts) == 2 {
		toolName = parts[1]
	}

	// Build UnitID for UoW-level cache lookup
	unitID := workunit.UnitID{
		Context:   workunit.ContextBuild,
		Module:    module,
		Component: compName,
		Tool:      toolName,
	}

	// Check UoW-level cache
	isCached := isUoWCached(bctx, unitID)
	log.Debugf("[UOW-CACHE] Component worker for %s: unitID=%s, isCached=%v",
		component, unitID.Longname(), isCached)

	if isCached {
		// In dry-run mode, just report what would happen
		if ctx.Config.DryRun {
			output.Writeln(logWriter, "⏭️  %s is up-to-date (would be skipped)", module)
			return -1 // -1 = skipped/cached = blue in TUI
		}

		// Verify UoW manifest artifacts are intact before declaring cache hit
		// If no manifests exist, trust the source hash check (which already passed)
		// Only fail cache if manifests exist AND integrity check fails (actual tampering)
		reader := coreoutput.NewReader(ctx.WorkspaceRoot)

		uowID := workunit.UnitID{
			Context:   workunit.ContextBuild,
			Module:    module,
			Component: compName,
			Tool:      toolName,
		}
		validationResult := reader.ValidateUoW(uowID)
		if !validationResult.ManifestExists {
			// No manifest - trust source hash, allow cache hit
			log.Debugf("UoW cache hit (no manifest to verify)")
			output.Writeln(logWriter, "⏭️  Cached (unchanged)")
			return -1
		}
		if !validationResult.Valid {
			output.Writeln(logWriter, "UoW cache miss: artifacts invalid")
			// Fall through to build - clear cache flag
			delete(bctx.cachedUoWs, unitID.Longname())
		} else {
			output.Writeln(logWriter, "⏭️  Cached (verified)")
			return -1
		}
	}

	// Get module contract
	moduleContract, exists := ctx.ModuleRegistry.Get(module)
	if !exists {
		output.Writeln(logWriter, "Error: module not found: %s", module)
		return 1
	}

	// Handle placeholder "none" component (modules with no buildable components)
	if component == "none" {
		output.Writeln(logWriter, "ℹ️  No buildable components for module: %s", module)
		// Write a NoOp manifest for the "none" placeholder to satisfy manifest assertions
		// NoOp manifests are always considered up-to-date in cache detection
		if bctx.tracker != nil {
			noneID := workunit.UnitID{
				Context:   workunit.ContextBuild,
				Module:    module,
				Component: "none",
				Tool:      "",
			}
			manifest := coreoutput.NewNoOpManifest(
				workunit.ContextBuild,
				module,
				"none",
				"",
				"no buildable components",
			)
			if err := bctx.tracker.RecordStart(noneID); err != nil {
				log.Debugf("Failed to record UoW start for %s: %v", noneID.Longname(), err)
			}
			if err := bctx.tracker.RecordComplete(noneID, manifest); err != nil {
				log.Debugf("Failed to record UoW completion for %s: %v", noneID.Longname(), err)
			}
		}
		return 0
	}

	// Component was already parsed above for cache check
	// compName and toolName are already set, use toolName as builderName
	builderName := toolName

	// Skip if dependency and artifacts exist (--use-existing-depm)
	if buildCfg.UseExistingDepm && !ctx.Config.DryRun && !buildCfg.RequestedSet[module] {
		if hasExistingArtifacts(module, ctx.WorkspaceRoot, buildCfg.ArtifactsMode.AllArtifactsRequested()) {
			output.Writeln(logWriter, "⏭️  Skipping %s:%s (module dependency artifacts exist)", module, component)
			return 0
		}
	}

	// Acquire component-level lock with wait (skip in dry-run)
	// Use component-level locking to allow parallel builds of different components within the same module
	// Use dash separator to match UnitID.DirName() for consistent manifest paths
	componentDir := compName
	if builderName != "" {
		componentDir = compName + "-" + builderName
	}
	if !ctx.Config.DryRun {
		lockCfg := locking.UnitBuildConfig(module, componentDir, paths.OutBuildRelPath)
		lockFile, err := locking.AcquireWithWait(context.Background(), ctx.WorkspaceRoot, lockCfg,
			ctx.Orchestrator.GetRegistry(), locking.DefaultWaitConfig())
		if err != nil {
			output.Writeln(logWriter, "Error: %v", err)
			return 1
		}
		defer locking.ReleaseTracked(lockFile)
	}

	// Get the handler for this component/tool
	// For tool chain components, builderName is the tool name (e.g., "pdf-oci")
	// and we need to get the handler directly by name
	var handler builders.Handler
	if builderName != "" {
		// Try to get handler directly by tool name (for tool chain components)
		handler = builders.GetHandler(builderName)
	}
	if handler == nil {
		// Fall back to component-based handler lookup
		handler = getHandlerForComponent(moduleContract, compName, builderName)
	}
	if handler == nil {
		output.Writeln(logWriter, "Error: no handler found for component %s in module %s", component, module)
		return 1
	}

	// Determine which artifacts to build
	requestedArtifacts := determineRequestedArtifactsForBuild(moduleContract, buildCfg.ArtifactsMode, ctx.WorkspaceRoot)

	// Look up weight for this component from pre-computed weights map
	// Key format matches what was stored in populateComponentWeights
	weightKey := module + ":" + component // "module:component:tool" format
	weight := bctx.componentWeights[weightKey]
	if weight <= 0 {
		// Try without tool suffix if not found
		weight = bctx.componentWeights[module+":"+compName]
	}
	if weight <= 0 {
		weight = 1 // Default to 1 if not found
	}
	log.Debugf("[WEIGHT] buildUnitWorker %s: weight=%d (key=%s)", component, weight, weightKey)

	opts := builders.BuildOptions{
		TidyFirst:          buildCfg.TidyFirst,
		Version:            buildCfg.Version,
		DryRun:             ctx.Config.DryRun,
		RequestedArtifacts: requestedArtifacts,
		Component:          compName, // Pass original component name for component-level parallelism
		Reproducible:       buildCfg.ResolveReproducible(),
		ForceRebuild:       ctx.Config.ForceRebuild,
		Weight:             weight,                   // Resource multiplier for container builds
		CacheConfig:        ctx.Config.CacheConfig,   // Fine-grained cache control
		ArtifactsMode:      buildCfg.ArtifactsMode,   // Artifact scope mode
	}

	// In dry-run mode, simulate a successful build
	if ctx.Config.DryRun {
		output.Writeln(logWriter, "🔨 %s would be built (changed)", module)
		output.Writeln(logWriter, "   Component: %s, Handler: %s", component, handler.Name())
		return 0
	}

	// Log which component we're building
	output.Writeln(logWriter, "━━━ Building component: %s (handler: %s) ━━━", compName, handler.Name())

	// Adapt module to port interface for handler methods
	modulePort := adapters.AdaptModule(moduleContract)

	// Validate module before building
	if err := handler.ValidateModule(modulePort, ctx.WorkspaceRoot, compName); err != nil {
		output.Writeln(logWriter, "❌ Module validation failed for %s: %v", compName, err)
		return 1
	}

	// Update UoW ID with actual handler name (may differ from parsed tool name)
	// and record start
	unitID = workunit.UnitID{
		Context:   workunit.ContextBuild,
		Module:    module,
		Component: compName,
		Tool:      handler.Name(),
	}
	if bctx.tracker != nil {
		if err := bctx.tracker.RecordStart(unitID); err != nil {
			log.Debugf("Failed to record UoW start for %s: %v", unitID.Longname(), err)
		}
	}

	// Use pre-computed module input hash to ensure all components in a module
	// share the same hash. This prevents divergence when parallel builds modify
	// shared files (e.g., go.sum via go mod tidy).
	inputHash := bctx.moduleInputHashes[module]
	if inputHash == "" {
		// Fallback: compute if not pre-computed (shouldn't happen in normal flow)
		inputHash = computeComponentInputHash(ctx, module, moduleContract)
	}

	// Build the component
	// Use component-level output directory: out/build/<module>/<component>/<builder>
	componentOutputDir := paths.ComponentBuildOutputPath(ctx.WorkspaceRoot, module, componentDir)
	exitCode := handler.Build(modulePort, ctx.WorkspaceRoot, componentOutputDir, logWriter, opts)

	// Handle exit codes:
	// -1 = skipped (cached), 0 = success, >0 = failure
	if exitCode > 0 {
		output.Writeln(logWriter, "❌ Build failed for component: %s", compName)
		// Record failure manifest with pre-computed input hash for debugging
		if bctx.tracker != nil {
			manifest := &coreoutput.UoWManifest{
				ExitCode:   exitCode,
				InputHash:  inputHash, // Use pre-computed hash from before build
				ExecutedAt: time.Now().UTC(),
				Artifacts:  nil,
				OutputHash: "",
				Version:    "1.0.0",
			}
			if err := bctx.tracker.RecordComplete(unitID, manifest); err != nil {
				log.Debugf("Failed to record UoW completion for %s: %v", unitID.Longname(), err)
			}
		}
		return exitCode
	}
	if exitCode == -1 {
		output.Writeln(logWriter, "⏭️  Component %s unchanged (cached)", compName)
		// Record cache hit - validate and return existing manifest
		if bctx.tracker != nil {
			if _, err := bctx.tracker.RecordCacheHit(unitID); err != nil {
				log.Debugf("Cache validation for %s: %v", unitID.Longname(), err)
			}
		}
		return exitCode // Pass through -1 so TUI shows blue
	}

	output.Writeln(logWriter, "✅ Component %s built successfully", compName)

	// Record successful completion with artifacts
	if bctx.tracker != nil {
		artifacts := collectUoWArtifacts(componentOutputDir, handler, modulePort, ctx.WorkspaceRoot)

		// Compute OutputHash from artifact hashes
		outputHash := computeOutputHash(artifacts)

		manifest := &coreoutput.UoWManifest{
			ExitCode:   exitCode,
			InputHash:  inputHash, // Use pre-computed hash from before build
			ExecutedAt: time.Now().UTC(),
			Artifacts:  artifacts,
			OutputHash: outputHash,
			Version:    "1.0.0",
		}
		if err := bctx.tracker.RecordComplete(unitID, manifest); err != nil {
			log.Debugf("Failed to record UoW completion for %s: %v", unitID.Longname(), err)
		}
	}

	// NOTE: UoW manifests are written atomically via tracker.RecordComplete above
	// No separate module-level state update needed

	return 0
}

// getHandlerForComponent finds the build handler for a specific component.
// It matches by component name and optionally by builder name.
func getHandlerForComponent(module *modules.ModuleContract, compName, builderName string) builders.Handler {
	compHandlers := builders.GetHandlersForModule(module)
	for _, ch := range compHandlers {
		if ch.Component == compName {
			// If builder name specified, verify it matches
			if builderName != "" && ch.Handler.Name() != builderName {
				continue
			}
			return ch.Handler
		}
	}
	return nil
}

// buildArtifactValidator validates build artifacts for dependencies.
func buildArtifactValidator(ctx *cmdframework.ExecutionContext) *initsummary.ArtifactValidationInfo {
	return artifacts.ValidateBuildArtifacts(
		ctx.GetExecutionMonikers(),
		ctx.EACConfig,
		ctx.WorkspaceRoot,
		ctx.ModuleRegistry,
	)
}

// buildDepsVerifier verifies system dependencies for build.
func buildDepsVerifier(ctx *cmdframework.ExecutionContext) *initsummary.DepsStatus {
	_, status := verifyBuildDependenciesQuiet(ctx.GetExecutionMonikers(), ctx.ModuleReport)
	return &status
}

// getGitCommit retrieves git commit SHA.
func getGitCommit(workspaceRoot string) string {
	return git.GetCommitSHA(workspaceRoot)
}

// collectUoWArtifacts collects artifact information from the build output directory.
// It lists files in the output directory and computes their hashes.
// Directory paths (ending with "/") are enumerated to collect all files within.
func collectUoWArtifacts(outputDir string, handler builders.Handler, modulePort interfaces.ModuleContractPort, workspaceRoot string) []coreoutput.Artifact {
	// Get expected artifacts from handler
	expectedPaths := handler.ListArtifacts(modulePort, workspaceRoot)
	if len(expectedPaths) == 0 {
		return nil
	}

	var artifacts []coreoutput.Artifact
	for _, relPath := range expectedPaths {
		fullPath := filepath.Join(outputDir, relPath)

		// Check if path exists
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		// Handle directories: enumerate all files within
		if info.IsDir() {
			dirArtifacts := collectDirectoryArtifacts(outputDir, relPath)
			artifacts = append(artifacts, dirArtifacts...)
			continue
		}

		// Handle regular files
		size, hashHex, err := coreoutput.HashFile(fullPath)
		if err != nil {
			log.Debugf("Failed to hash artifact %s: %v", relPath, err)
			continue
		}

		artifacts = append(artifacts, coreoutput.Artifact{
			ID:     filepath.Base(relPath),
			Path:   relPath,
			SHA256: hashHex, // HashFile already includes "sha256:" prefix
			Size:   size,
			Type:   inferArtifactType(relPath),
		})
	}

	return artifacts
}

// collectDirectoryArtifacts enumerates files in a directory and collects artifact info.
// Only collects top-level files to avoid excessive tracking of nested content.
func collectDirectoryArtifacts(baseDir, relDir string) []coreoutput.Artifact {
	dirPath := filepath.Join(baseDir, relDir)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Debugf("Failed to read directory %s: %v", dirPath, err)
		return nil
	}

	var artifacts []coreoutput.Artifact
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Only collect files, not subdirectories
		}

		relPath := filepath.Join(relDir, entry.Name())
		fullPath := filepath.Join(baseDir, relPath)

		size, hashHex, err := coreoutput.HashFile(fullPath)
		if err != nil {
			log.Debugf("Failed to hash artifact %s: %v", relPath, err)
			continue
		}

		artifacts = append(artifacts, coreoutput.Artifact{
			ID:     entry.Name(),
			Path:   relPath,
			SHA256: hashHex,
			Size:   size,
			Type:   inferArtifactType(relPath),
		})
	}

	return artifacts
}

// inferArtifactType infers the artifact type from its file path.
func inferArtifactType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".exe", "":
		if strings.Contains(path, "windows") || strings.HasSuffix(path, ".exe") {
			return "executable"
		}
		// Check if it's likely a binary (no extension on linux/darwin)
		return "binary"
	case ".tar", ".gz", ".tgz", ".zip":
		return "archive"
	case ".pdf":
		return "document"
	case ".html":
		return "html"
	case ".json":
		return "data"
	case ".log", ".txt":
		return "text"
	default:
		return "file"
	}
}

// computeComponentInputHash computes a hash of the input sources for a component.
// This is used to detect when source files have changed.
func computeComponentInputHash(ctx *cmdframework.ExecutionContext, module string, moduleContract *modules.ModuleContract) string {
	if moduleContract == nil {
		return ""
	}

	// Get source file patterns from module contract
	patterns := moduleContract.GetGlobPatterns()
	if len(patterns) == 0 {
		return ""
	}

	// Expand patterns and compute hash
	files, err := hash.ExpandGlobPatterns(ctx.WorkspaceRoot, patterns)
	if err != nil {
		log.Debugf("Failed to expand patterns for input hash of %s: %v", module, err)
		return ""
	}

	inputHash, err := hash.Files(ctx.WorkspaceRoot, files)
	if err != nil {
		log.Debugf("Failed to compute input hash for %s: %v", module, err)
		return ""
	}

	return inputHash
}

// computeOutputHash computes a combined hash from all artifact hashes.
// This provides a single value representing all outputs.
func computeOutputHash(artifacts []coreoutput.Artifact) string {
	if len(artifacts) == 0 {
		return ""
	}

	// Combine all artifact hashes (sorted by path for determinism)
	var hashes []string
	for _, a := range artifacts {
		hashes = append(hashes, a.SHA256)
	}

	// Sort for determinism
	sort.Strings(hashes)
	combined := strings.Join(hashes, ":")

	// Hash the combined string using SHA256
	h := sha256.Sum256([]byte(combined))
	return "sha256:" + hex.EncodeToString(h[:])
}
