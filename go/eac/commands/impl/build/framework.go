// Package build provides the build command implementation using cmdframework.
package build

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/impl/build/builders"
	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/artifacts"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
	"github.com/ready-to-release/eac/go/eac/commands/internal/git"
	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/commands/internal/locking"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/core/buildstate"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

func init() {
	// Register component-level execution support
	cmdframework.RegisterComponentProvider(cmdframework.CommandTypeBuild, FlattenModulesToComponentWork)
	cmdframework.RegisterComponentWorker(cmdframework.CommandTypeBuild, buildComponentWorker)
	cmdframework.SetComponentCountProvider(getBuildComponentCount)
	cmdframework.SetComponentLayersProvider(getBuildComponentLayers)
}

// getBuildComponentCount returns the total number of buildable components.
func getBuildComponentCount(ctx *cmdframework.ExecutionContext) int {
	layers := FlattenModulesToComponentWork(ctx)
	return CountComponents(layers)
}

// getBuildComponentLayers returns the component execution layers as string slices.
// Each layer contains component identifiers in "module:component" format.
func getBuildComponentLayers(ctx *cmdframework.ExecutionContext) [][]string {
	layers := FlattenModulesToComponentWork(ctx)
	result := make([][]string, len(layers))
	for i, layer := range layers {
		result[i] = make([]string, len(layer))
		for j, work := range layer {
			result[i][j] = fmt.Sprintf("%s:%s", work.Module, work.Component)
		}
	}
	return result
}

// BuildConfig holds build-specific configuration.
type BuildConfig struct {
	TidyFirst       bool
	Version         string
	BuildAll        bool
	UseExistingDepm bool
	LayeredBuild    bool

	// Reproducible controls determinism behavior for MkDocs builds.
	// Values: "auto" (default), "true", "false"
	// - auto: CI=true (always rebuild), local=false (use cache)
	// - true: Always rebuild HTML from staging (CI default)
	// - false: Skip MkDocs if staging unchanged (local default)
	Reproducible string

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

// buildAfterInit handles build-specific initialization after framework init.
func buildAfterInit(ctx *cmdframework.ExecutionContext) error {
	// Nothing special needed for now
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

// detectIncrementalChanges performs incremental build detection.
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

// buildAfterExecute handles post-build tasks: artifact derivations, manifest generation, state updates.
func buildAfterExecute(ctx *cmdframework.ExecutionContext) error {
	buildCfg, ok := ctx.Config.Extra["buildConfig"].(*BuildConfig)
	if !ok {
		return fmt.Errorf("buildConfig not found or wrong type")
	}
	bctx, ok := ctx.Config.Extra["buildContext"].(*buildContext)
	if !ok {
		return fmt.Errorf("buildContext not found or wrong type")
	}

	// Process artifact derivations for all successfully built modules
	// This includes compression, UPX, and other post-processing
	if !ctx.Config.DryRun {
		if err := processAllArtifactDerivations(ctx, buildCfg); err != nil {
			return fmt.Errorf("artifact derivation failed: %w", err)
		}
	}

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
		mergedArtifacts := cfg.GetBuildArtifacts(moniker, buildCfg.BuildAll)
		if len(mergedArtifacts) == 0 {
			continue
		}

		// Determine which artifacts were requested
		requestedArtifacts := determineRequestedArtifactsForBuild(module, buildCfg.BuildAll, ctx.WorkspaceRoot)

		// Build output directory
		moduleOutputDir := paths.BuildOutputPath(ctx.WorkspaceRoot, moniker)

		// Process derivations (compression, UPX, etc.)
		if err := ProcessArtifactDerivations(moniker, mergedArtifacts, moduleOutputDir, requestedArtifacts, module.Metadata, io.Discard); err != nil {
			log.Warnf("Artifact derivation warning for %s: %v", moniker, err)
			// Continue with other modules - derivation failure is not fatal
		}
	}

	// Execute post-build steps for each successfully built component
	for _, compResult := range ctx.ComponentResults {
		if compResult.ExitCode != 0 {
			continue // Skip failed components
		}

		componentOutputDir := paths.ComponentBuildOutputPath(ctx.WorkspaceRoot, compResult.Module, compResult.Component)
		if exitCode := builders.ExecutePostBuildSteps(compResult.Module, compResult.Component, ctx.WorkspaceRoot, componentOutputDir, io.Discard); exitCode != 0 {
			log.Warnf("Post-build steps warning for %s/%s: exit code %d", compResult.Module, compResult.Component, exitCode)
			// Continue with other components
		}
	}

	return nil
}

// updateIncrementalState updates the build state for incremental detection.
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

// buildWorker is the worker function that builds a single module.
func buildWorker(ctx *cmdframework.ExecutionContext, moniker string, logWriter io.Writer) int {
	buildCfg, ok := ctx.Config.Extra["buildConfig"].(*BuildConfig)
	if !ok {
		output.Writeln(logWriter, "Error: buildConfig not found or wrong type")
		return 1
	}

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
		lockFile, err := locking.AcquireTracked(ctx.WorkspaceRoot, lockCfg, ctx.Orchestrator.GetRegistry())
		if err != nil {
			output.Writeln(logWriter, "Error: %v", err)
			return 1
		}
		defer locking.ReleaseTracked(lockFile)
	}

	// Run the build
	moduleOutputDir := paths.BuildOutputPath(ctx.WorkspaceRoot, moniker)
	exitCode := runModuleBuild(module, ctx.WorkspaceRoot, moduleOutputDir, logWriter,
		buildCfg.TidyFirst, buildCfg.Version, ctx.Config.DryRun, buildCfg.BuildAll, buildCfg.ResolveReproducible())

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

// buildComponentWorker builds a single component within a module.
// This is called by the ComponentScheduler for parallel component execution.
// The component parameter is in "compName:builderName" format (e.g., "go:go", "docs:mkdocs").
func buildComponentWorker(ctx *cmdframework.ExecutionContext, module, component string, logWriter io.Writer) int {
	buildCfg, ok := ctx.Config.Extra["buildConfig"].(*BuildConfig)
	if !ok {
		output.Writeln(logWriter, "Error: buildConfig not found or wrong type")
		return 1
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
		return 0
	}

	// Parse component parameter: "compName:builderName" (e.g., "go:go", "docs:mkdocs")
	parts := strings.SplitN(component, ":", 2)
	compName := parts[0]
	builderName := ""
	if len(parts) == 2 {
		builderName = parts[1]
	}

	// Skip if dependency and artifacts exist (--use-existing-depm)
	if buildCfg.UseExistingDepm && !ctx.Config.DryRun && !buildCfg.RequestedSet[module] {
		if hasExistingArtifacts(module, ctx.ModuleTypes[module], ctx.WorkspaceRoot, buildCfg.BuildAll) {
			output.Writeln(logWriter, "⏭️  Skipping %s:%s (module dependency artifacts exist)", module, component)
			return 0
		}
	}

	// Acquire component-level lock (skip in dry-run)
	// Use component-level locking to allow parallel builds of different components within the same module
	// Use underscore separator for Windows compatibility (same as lint)
	componentDir := compName
	if builderName != "" {
		componentDir = compName + "_" + builderName
	}
	if !ctx.Config.DryRun {
		lockCfg := locking.ComponentBuildConfig(module, componentDir, paths.OutBuildRelPath)
		lockFile, err := locking.AcquireTracked(ctx.WorkspaceRoot, lockCfg, ctx.Orchestrator.GetRegistry())
		if err != nil {
			output.Writeln(logWriter, "Error: %v", err)
			return 1
		}
		defer locking.ReleaseTracked(lockFile)
	}

	// Get the handler for this component
	handler := getHandlerForComponent(moduleContract, compName, builderName)
	if handler == nil {
		output.Writeln(logWriter, "Error: no handler found for component %s in module %s", component, module)
		return 1
	}

	// Determine which artifacts to build
	requestedArtifacts := determineRequestedArtifactsForBuild(moduleContract, buildCfg.BuildAll, ctx.WorkspaceRoot)

	opts := builders.BuildOptions{
		TidyFirst:          buildCfg.TidyFirst,
		Version:            buildCfg.Version,
		DryRun:             ctx.Config.DryRun,
		RequestedArtifacts: requestedArtifacts,
		Component:          compName, // Pass original component name for component-level parallelism
		Reproducible:       buildCfg.ResolveReproducible(),
		ForceRebuild:       ctx.Config.ForceRebuild,
	}

	// In dry-run mode, simulate a successful build
	if ctx.Config.DryRun {
		output.Writeln(logWriter, "Build: %s:%s (dry-run)", module, component)
		output.Writeln(logWriter, "Handler: %s", handler.Name())
		output.Writeln(logWriter, "Dry-run mode: skipping actual build")
		return 0
	}

	// Log which component we're building
	output.Writeln(logWriter, "━━━ Building component: %s (handler: %s) ━━━", compName, handler.Name())

	// Validate module before building
	if err := handler.ValidateModule(moduleContract, ctx.WorkspaceRoot, compName); err != nil {
		output.Writeln(logWriter, "❌ Module validation failed for %s: %v", compName, err)
		return 1
	}

	// Build the component
	// Use component-level output directory: out/build/<module>/<component>/<builder>
	componentOutputDir := paths.ComponentBuildOutputPath(ctx.WorkspaceRoot, module, componentDir)
	exitCode := handler.Build(moduleContract, ctx.WorkspaceRoot, componentOutputDir, logWriter, opts)

	// Handle exit codes:
	// -1 = skipped (cached), 0 = success, >0 = failure
	if exitCode > 0 {
		output.Writeln(logWriter, "❌ Build failed for component: %s", compName)
		return exitCode
	}
	if exitCode == -1 {
		output.Writeln(logWriter, "⏭️  Component %s unchanged (cached)", compName)
		return exitCode // Pass through -1 so TUI shows blue
	}

	output.Writeln(logWriter, "✅ Component %s built successfully", compName)
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

// getGitCommit retrieves git commit SHA.
func getGitCommit(workspaceRoot string) string {
	return git.GetCommitSHA(workspaceRoot)
}
