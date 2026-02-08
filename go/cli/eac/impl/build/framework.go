// Package build provides the build command implementation using cmdframework.
package build

import (
	"fmt"
	"io"
	"sync"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/builders"
	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/artifacts"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/environment"
	"github.com/ready-to-release/eac/go/clibase/git"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/hash"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/workunit"
)

func init() {
	// Register component-level execution support
	cmdframework.RegisterUnitProvider(core.ActionBuild, ResolveUnitSpecs)
	cmdframework.RegisterUnitWorker(core.ActionBuild, buildUnitWorker)
	cmdframework.SetUoWCountProvider(getBuildUoWCount)
}

// cachedUnitWorkMu protects concurrent access to the BuildWorkSpecs field
// in ctx.Config, preventing races when getCachedUnitWork is called from
// multiple goroutines.
var cachedUnitWorkMu sync.Mutex

// getCachedUnitWork returns cached component work specs, computing once if needed.
// This avoids duplicate calls to ResolveUnitSpecs during startup.
// It is safe for concurrent use from multiple goroutines.
func getCachedUnitWork(ctx *cmdframework.ExecutionContext) []workunit.UnitSpec {
	cachedUnitWorkMu.Lock()
	defer cachedUnitWorkMu.Unlock()

	if len(ctx.Config.BuildWorkSpecs) > 0 {
		return ctx.Config.BuildWorkSpecs
	}

	// Compute and cache
	specs := ResolveUnitSpecs(ctx)
	ctx.Config.BuildWorkSpecs = specs
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
	ArtifactsMode environments.ArtifactsMode

	// Components filters which components to build within each module.
	// When empty, all components are built. When set, only matching components are built.
	Components []string

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

	// Pre-expanded module files (reused by detectUoWIncrementalChanges to avoid duplicate glob expansion).
	moduleExpandedFiles map[string][]string // module moniker -> expanded file list
}

// RunBuildWithFramework executes build using the cmdframework.
// This is the new implementation that will replace buildMultipleModules.
func RunBuildWithFramework(cmdCfg *cmdframework.CommandConfig, buildCfg *BuildConfig) int {
	// Store build config in typed fields for access in hooks/worker
	cmdCfg.BuildCmdConfig = buildCfg

	// Create build context
	bctx := &buildContext{
		cfg:              buildCfg,
		componentWeights: make(map[string]int),
	}
	cmdCfg.BuildCmdContext = bctx

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
	buildCfg := ctx.Config.BuildCmdConfig.(*BuildConfig)
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
	bctx := ctx.Config.BuildCmdContext.(*buildContext)
	bctx.tracker = coreoutput.NewTracker(ctx.WorkspaceRoot, core.ActionBuild)

	return nil
}

// buildAfterResolve handles --use-existing-depm filtering and incremental detection.
func buildAfterResolve(ctx *cmdframework.ExecutionContext) error {
	buildCfg, ok := ctx.Config.BuildCmdConfig.(*BuildConfig)
	if !ok {
		return fmt.Errorf("BuildCmdConfig not found or wrong type")
	}
	bctx, ok := ctx.Config.BuildCmdContext.(*buildContext)
	if !ok {
		return fmt.Errorf("BuildCmdContext not found or wrong type")
	}

	// Handle --use-existing-depm: filter out deps that already have artifacts
	if buildCfg.UseExistingDepm && !ctx.Config.DryRun && len(ctx.ScopeMonikers) > 0 {
		cfg, err := config.Load(config.DefaultLoadOptions())
		if err != nil {
			return fmt.Errorf("failed to load config for --use-existing-depm: %w", err)
		}

		var filteredScope []string

		for _, m := range ctx.ScopeMonikers {
			// Always include requested modules
			if buildCfg.RequestedSet[m] {
				filteredScope = append(filteredScope, m)
				continue
			}
			// For deps, check if artifacts exist
			if hasExistingArtifacts(m, ctx.WorkspaceRoot, buildCfg.ArtifactsMode.AllArtifactsRequested(), cfg) {
				log.Debugf("Skipping dep %s (artifacts exist)", m)
			} else {
				filteredScope = append(filteredScope, m)
			}
		}

		ctx.ScopeMonikers = filteredScope
	}

	// Pre-compute module input hashes ONCE before any workers run.
	// This ensures all components in a module get the same hash, avoiding
	// divergence when parallel builds modify shared files (e.g., go.sum via go mod tidy).
	// Uses mtime-based caching to skip re-hashing unchanged files.
	bctx.moduleInputHashes = make(map[string]string)
	bctx.moduleExpandedFiles = make(map[string][]string)
	if ctx.ModuleRegistry != nil && len(ctx.ScopeMonikers) > 0 {
		hashCache := hash.LoadCache(paths.InputHashCachePath(ctx.WorkspaceRoot))
		for _, module := range ctx.ScopeMonikers {
			if contract, ok := ctx.ModuleRegistry.Get(module); ok {
				h, files, err := hashCache.GetOrCompute(module, contract.GetGlobPatterns(), ctx.WorkspaceRoot)
				if err != nil {
					log.Debugf("Failed to compute input hash for %s: %v", module, err)
					continue
				}
				bctx.moduleInputHashes[module] = h
				bctx.moduleExpandedFiles[module] = files
			}
		}
		if err := hashCache.Save(); err != nil {
			log.Debugf("Failed to save hash cache: %v", err)
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
	buildCfg, ok := ctx.Config.BuildCmdConfig.(*BuildConfig)
	if !ok {
		return fmt.Errorf("BuildCmdConfig not found or wrong type")
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

