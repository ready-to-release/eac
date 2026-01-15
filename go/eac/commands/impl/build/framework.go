// Package build provides the build command implementation using cmdframework.
package build

import (
	"fmt"
	"io"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/artifacts"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/git"
	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/commands/internal/locking"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/core/buildstate"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// BuildConfig holds build-specific configuration
type BuildConfig struct {
	TidyFirst       bool
	Version         string
	BuildAll        bool
	UseExistingDepm bool
	LayeredBuild    bool

	// Set of originally requested modules (for --use-existing-depm logic)
	RequestedSet map[string]bool

	// Incremental build state
	SkippedModules []string
}

// buildContext holds build-specific state during execution
type buildContext struct {
	cfg            *BuildConfig
	skippedModules []string
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
		cfg: buildCfg,
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

	return cmdframework.Run(cmdCfg, buildWorker, hooks)
}

// buildAfterInit handles build-specific initialization after framework init
func buildAfterInit(ctx *cmdframework.ExecutionContext) error {
	// Nothing special needed for now
	return nil
}

// buildAfterResolve handles --use-existing-depm filtering and incremental detection
func buildAfterResolve(ctx *cmdframework.ExecutionContext) error {
	buildCfg := ctx.Config.Extra["buildConfig"].(*BuildConfig)
	bctx := ctx.Config.Extra["buildContext"].(*buildContext)

	// Handle --use-existing-depm: filter out deps that already have artifacts
	if buildCfg.UseExistingDepm && !ctx.Config.DryRun && ctx.ExecutionPlan != nil {
		var filteredOrder []string
		var filteredLayers [][]string

		for _, layer := range ctx.ExecutionPlan.Layers {
			var filteredLayer []string
			for _, m := range layer {
				// Always include requested modules
				if buildCfg.RequestedSet[m] {
					filteredLayer = append(filteredLayer, m)
					filteredOrder = append(filteredOrder, m)
					continue
				}
				// For deps, check if artifacts exist
				moduleType := ctx.ModuleTypes[m]
				if hasExistingArtifacts(m, moduleType, ctx.WorkspaceRoot, buildCfg.BuildAll) {
					log.Debugf("Skipping dep %s (artifacts exist)", m)
				} else {
					filteredLayer = append(filteredLayer, m)
					filteredOrder = append(filteredOrder, m)
				}
			}
			if len(filteredLayer) > 0 {
				filteredLayers = append(filteredLayers, filteredLayer)
			}
		}

		ctx.ExecutionPlan.ExecutionOrder = filteredOrder
		ctx.ExecutionPlan.Layers = filteredLayers
	}

	// Incremental build detection (devbox only)
	if !ctx.Config.ForceRebuild && !ctx.Config.DryRun && !buildCfg.UseExistingDepm {
		detectIncrementalChanges(ctx, bctx)
	}

	return nil
}

// detectIncrementalChanges performs incremental build detection
func detectIncrementalChanges(ctx *cmdframework.ExecutionContext, bctx *buildContext) {
	startTime := time.Now()

	// Build modules map for change detection
	modulesMap := make(map[string]buildstate.ModuleFileGetter)
	for _, moniker := range ctx.GetExecutionMonikers() {
		if contract, ok := ctx.ModuleRegistry.Get(moniker); ok {
			modulesMap[moniker] = contract
		}
	}

	moduleFiles, err := buildstate.GetModuleSourceFiles(ctx.WorkspaceRoot, modulesMap)
	if err != nil {
		log.Debugf("Failed to get module source files: %v", err)
		return
	}

	changeResult, err := buildstate.DetectChanges(ctx.WorkspaceRoot, moduleFiles)
	if err != nil {
		log.Debugf("Failed to detect changes: %v", err)
		return
	}

	detectionTime := time.Since(startTime)

	if changeResult.FreshBuild {
		log.Debugf("Fresh build detected, no incremental filtering")
		// Report fresh build in init summary
		if ctx.InitSummary != nil {
			ctx.InitSummary.SetIncremental(&initsummary.IncrementalInfo{
				Enabled:       true,
				DetectionTime: detectionTime,
				FreshBuild:    true,
			})
		}
		return
	}

	// Filter execution plan to only changed modules
	changedSet := make(map[string]bool)
	for _, m := range changeResult.ChangedModules {
		changedSet[m] = true
	}

	buildCfg := bctx.cfg
	var filteredOrder []string
	var filteredLayers [][]string

	for _, layer := range ctx.ExecutionPlan.Layers {
		var filteredLayer []string
		for _, m := range layer {
			// Include if changed or explicitly requested
			if changedSet[m] || buildCfg.RequestedSet[m] {
				filteredLayer = append(filteredLayer, m)
				filteredOrder = append(filteredOrder, m)
			} else {
				bctx.skippedModules = append(bctx.skippedModules, m)
			}
		}
		if len(filteredLayer) > 0 {
			filteredLayers = append(filteredLayers, filteredLayer)
		}
	}

	ctx.ExecutionPlan.ExecutionOrder = filteredOrder
	ctx.ExecutionPlan.Layers = filteredLayers

	// Report incremental detection in init summary
	if ctx.InitSummary != nil {
		ctx.InitSummary.SetIncremental(&initsummary.IncrementalInfo{
			Enabled:       true,
			DetectionTime: detectionTime,
			Changed:       filteredOrder,
			UpToDate:      bctx.skippedModules,
			FreshBuild:    false,
		})
	}

	log.Debugf("Incremental: %d modules to build, %d skipped",
		len(filteredOrder), len(bctx.skippedModules))
}

// buildAfterExecute handles post-build tasks: manifest generation, state updates
func buildAfterExecute(ctx *cmdframework.ExecutionContext) error {
	buildCfg := ctx.Config.Extra["buildConfig"].(*BuildConfig)
	bctx := ctx.Config.Extra["buildContext"].(*buildContext)

	// Generate build manifest
	if err := generateBuildManifest(ctx.WorkspaceRoot, ctx.Results, ctx.ModuleTypes,
		ctx.GetExecutionMonikers(), buildCfg.BuildAll); err != nil {
		return fmt.Errorf("failed to generate build manifest: %w", err)
	}

	// Update skipped module manifests
	if !ctx.Config.DryRun && len(bctx.skippedModules) > 0 {
		gitCommit := getGitCommit(ctx.WorkspaceRoot)
		updateSkippedModuleManifests(ctx.WorkspaceRoot, bctx.skippedModules, gitCommit)
	}

	// Update incremental build state
	if !ctx.Config.DryRun {
		updateIncrementalState(ctx, bctx)
	}

	return nil
}

// updateIncrementalState updates the build state for incremental detection
func updateIncrementalState(ctx *cmdframework.ExecutionContext, bctx *buildContext) {
	// Collect successfully built modules
	var successfulModules []string
	for _, result := range ctx.Results {
		if result.ExitCode == 0 {
			successfulModules = append(successfulModules, result.Moniker)
		}
	}

	// Include skipped modules (they were already up-to-date)
	allSuccessful := append(successfulModules, bctx.skippedModules...)

	// Build modules map for state update
	modulesMap := make(map[string]buildstate.ModuleFileGetter)
	for _, moniker := range allSuccessful {
		if contract, ok := ctx.ModuleRegistry.Get(moniker); ok {
			modulesMap[moniker] = contract
		}
	}

	moduleFiles, err := buildstate.GetModuleSourceFiles(ctx.WorkspaceRoot, modulesMap)
	if err != nil {
		log.Warnf("Failed to get module files for state update: %v", err)
		return
	}

	if err := buildstate.UpdateModuleState(ctx.WorkspaceRoot, allSuccessful, moduleFiles); err != nil {
		log.Warnf("Failed to update build state: %v", err)
	}
}

// buildWorker is the worker function that builds a single module
func buildWorker(ctx *cmdframework.ExecutionContext, moniker string, logWriter io.Writer) int {
	buildCfg := ctx.Config.Extra["buildConfig"].(*BuildConfig)

	module, exists := ctx.ModuleRegistry.Get(moniker)
	if !exists {
		output.Writeln(logWriter, "Error: module not found: %s", moniker)
		return 1
	}

	// Skip if dependency and artifacts exist (--use-existing-depm)
	if buildCfg.UseExistingDepm && !ctx.Config.DryRun && !buildCfg.RequestedSet[moniker] {
		if hasExistingArtifacts(moniker, ctx.ModuleTypes[moniker], ctx.WorkspaceRoot, buildCfg.BuildAll) {
			output.Writeln(logWriter, "⏭️  Skipping %s (module dependency artifacts exist)", moniker)
			return 0
		}
	}

	// Acquire lock (skip in dry-run)
	if !ctx.Config.DryRun {
		lockCfg := locking.BuildConfig(moniker, paths.OutBuildRelPath)
		lockFile, err := locking.Acquire(ctx.WorkspaceRoot, lockCfg)
		if err != nil {
			output.Writeln(logWriter, "Error: %v", err)
			return 1
		}
		defer locking.Release(lockFile)
	}

	// Run the build
	moduleOutputDir := paths.BuildOutputPath(ctx.WorkspaceRoot, moniker)
	exitCode := runModuleBuild(module, ctx.WorkspaceRoot, moduleOutputDir, logWriter,
		buildCfg.TidyFirst, buildCfg.Version, ctx.Config.DryRun, buildCfg.BuildAll)

	// Validate artifacts if build succeeded
	if exitCode == 0 && !ctx.Config.DryRun {
		if err := validateModuleBuildOutputs(moniker, ctx.ModuleTypes[moniker],
			ctx.WorkspaceRoot, logWriter, buildCfg.BuildAll); err != nil {
			output.Writeln(logWriter, "\n❌ Build artifact validation failed: %v", err)
			return 1
		}
	}

	return exitCode
}

// buildArtifactValidator validates build artifacts for dependencies
func buildArtifactValidator(ctx *cmdframework.ExecutionContext) *initsummary.ArtifactValidationInfo {
	return artifacts.ValidateBuildArtifacts(
		ctx.GetExecutionMonikers(),
		ctx.EACConfig,
		ctx.WorkspaceRoot,
		ctx.ModuleRegistry,
	)
}

// buildDepsVerifier verifies system dependencies for build
func buildDepsVerifier(ctx *cmdframework.ExecutionContext) *initsummary.DepsStatus {
	// Get unique module types
	moduleTypes := make(map[string]bool)
	for _, moniker := range ctx.GetExecutionMonikers() {
		if t, ok := ctx.ModuleTypes[moniker]; ok {
			moduleTypes[t] = true
		}
	}

	// Convert to slice for verification
	var types []string
	for t := range moduleTypes {
		types = append(types, t)
	}

	_, status := verifyBuildDependenciesQuiet(ctx.GetExecutionMonikers(), ctx.ModuleReport)
	return &status
}

// getGitCommit retrieves git commit SHA
func getGitCommit(workspaceRoot string) string {
	return git.GetCommitSHA(workspaceRoot)
}
