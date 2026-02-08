package build

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/locking"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/hash"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// buildUnitWorker builds a single component within a module.
// This is called by the UnitScheduler for parallel component execution.
// The component parameter is in "compName:builderName" format (e.g., "go:go", "docs:mkdocs").
func buildUnitWorker(goCtx context.Context, ctx *cmdframework.ExecutionContext, module, component string, logWriter io.Writer) int {
	buildCfg, ok := ctx.Config.BuildCmdConfig.(*BuildConfig)
	if !ok {
		output.Writeln(logWriter, "Error: BuildCmdConfig not found or wrong type")
		return 1
	}
	bctx, ok := ctx.Config.BuildCmdContext.(*buildContext)
	if !ok {
		output.Writeln(logWriter, "Error: BuildCmdContext not found or wrong type")
		return 1
	}

	// Check incremental cache first - if module/UoW is cached, verify artifacts and skip (blue in TUI)
	// Parse component to build unitID for UoW cache lookup
	compName, toolName := cmdframework.ParseComponent(component)

	// Build UnitID for UoW-level cache lookup
	unitID := workunit.UnitID{
		Action:    core.ActionBuild,
		Module:    module,
		Component: compName,
		Tool:      toolName,
	}

	// Create pipeline for shared cache/lock/manifest orchestration
	pipeline := &cmdframework.UnitPipeline{
		CachedUoWs:             bctx.cachedUoWs,
		ValidateCacheArtifacts: true,
		OnCacheInvalidated:     func(ln string) { delete(bctx.cachedUoWs, ln) },
		LockStyle:              cmdframework.LockUnlessDryRun,
		LockConfigFn:           func(m, cd string) locking.Config { return locking.UnitBuildConfig(m, cd, paths.OutBuildRelPath) },
		Tracker:                bctx.tracker,
	}

	// Check UoW-level cache
	log.Debugf("[UOW-CACHE] Component worker for %s: unitID=%s", component, unitID.Longname())
	if cacheResult := pipeline.CheckCache(ctx, unitID, logWriter); cacheResult != 0 {
		return cacheResult
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
				Action:    core.ActionBuild,
				Module:    module,
				Component: "none",
				Tool:      "",
			}
			manifest := coreoutput.NewNoOpManifest(
				core.ActionBuild,
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
		cfg, err := config.Load(config.DefaultLoadOptions())
		if err == nil {
			if hasExistingArtifacts(module, ctx.WorkspaceRoot, buildCfg.ArtifactsMode.AllArtifactsRequested(), cfg) {
				output.Writeln(logWriter, "⏭️  Skipping %s:%s (module dependency artifacts exist)", module, component)
				return 0
			}
		}
	}

	// Acquire component-level lock with wait (skip in dry-run)
	componentDir := cmdframework.ComponentDir(compName, builderName)
	release, err := pipeline.AcquireLock(ctx, module, componentDir)
	if err != nil {
		output.Writeln(logWriter, "Error: %v", err)
		return 1
	}
	defer release()

	// Get the handler for this component/tool
	// For tool chain components, builderName is the tool name (e.g., "pdf-oci")
	// and we need to get the handler directly by name
	var handler tool.BuildHandler
	if builderName != "" {
		// Try to get handler directly by tool name (for tool chain components)
		handler = tool.GlobalBuildBridge().GetHandler(builderName)
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

	opts := tool.BuildOptions{
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
		Action:    core.ActionBuild,
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
		pipeline.RecordManifest(unitID, &coreoutput.UoWManifest{
			ExitCode:   exitCode,
			InputHash:  inputHash,
			ExecutedAt: time.Now().UTC(),
			Artifacts:  nil,
			OutputHash: "",
			Version:    "1.0.0",
		})
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
	artifacts := collectUoWArtifacts(componentOutputDir, handler, modulePort, ctx.WorkspaceRoot)
	outputHash := computeOutputHash(artifacts)
	pipeline.RecordManifest(unitID, &coreoutput.UoWManifest{
		ExitCode:   exitCode,
		InputHash:  inputHash,
		ExecutedAt: time.Now().UTC(),
		Artifacts:  artifacts,
		OutputHash: outputHash,
		Version:    "1.0.0",
	})

	// NOTE: UoW manifests are written atomically via tracker.RecordComplete above
	// No separate module-level state update needed

	return 0
}

// getHandlerForComponent finds the build handler for a specific component.
// It matches by component name and optionally by builder name.
func getHandlerForComponent(module *modules.ModuleContract, compName, builderName string) tool.BuildHandler {
	compHandlers := tool.GlobalBuildBridge().GetHandlersForModule(module)
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

// collectUoWArtifacts collects artifact information from the build output directory.
// It lists files in the output directory and computes their hashes.
// Directory paths (ending with "/") are enumerated to collect all files within.
func collectUoWArtifacts(outputDir string, handler tool.BuildHandler, modulePort core.ModuleContractPort, workspaceRoot string) []coreoutput.Artifact {
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
