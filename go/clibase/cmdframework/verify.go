package cmdframework

import (
	"os"
	"strings"
	"sync"

	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// detectExecutionContext returns a human-readable execution context string.
// Combines container detection (r2r vs implicit) with environment (CI vs devbox).
// Returns: "implicit-cli (devbox)", "implicit-cli (CI)", "r2r-cli (devbox)", "r2r-cli (CI)".
func detectExecutionContext() string {
	// CLI mode: r2r-cli (container) or implicit-cli (local)
	cliMode := string(logging.GetExecutionContext())

	// Environment: CI or devbox
	env := "devbox"
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		env = "CI"
	}

	return cliMode + " (" + env + ")"
}

// phaseVerify handles the verification phase:
// - Build init summary (or enhance existing one from hooks)
// - System dependency verification (via hook)
// - Module dependency validation (via hook).
func phaseVerify(ctx *ExecutionContext) error {
	// Use existing summary if set by a hook (e.g., test framework), otherwise create new one
	summary := ctx.InitSummary
	if summary == nil {
		summary = initsummary.New(string(ctx.Config.Type))
		summary.RequestedModules = ctx.Config.Monikers
		summary.CalculatedModules = ctx.GetExecutionMonikers()
		summary.AddedDepm = ctx.GetAddedDependencies()
	}

	// Always set/update execution context
	summary.ExecutionContext = detectExecutionContext()

	// Execution plan info
	if ctx.ExecutionPlan != nil {
		summary.ExecutionLayers = ctx.ExecutionPlan.Layers
		summary.LayerCount = len(ctx.ExecutionPlan.Layers)
	}
	summary.FlatExecution = !ctx.Config.Layered

	// Calculate UoW count using the registered provider (if available)
	// Only use provider if UoW count wasn't already set by a hook (e.g., test framework)
	// This allows us to show the actual UoW count, not just module count
	if summary.UoWCount == 0 && uowCountProvider != nil {
		count := uowCountProvider(ctx)
		log.Debugf("UoW count from provider: %d", count)
		summary.UoWCount = count
	} else if summary.UoWCount > 0 {
		log.Debugf("UoW count already set by hook: %d", summary.UoWCount)
	} else {
		log.Debugf("No UoW count provider registered")
	}

	// Calculate component layers using the command-type-specific provider (if available)
	// Only use provider if component layers wasn't already set by a hook
	if summary.ComponentLayerCount == 0 {
		provider := GetUnitLayersProvider(ctx.Config.Type)
		if provider != nil {
			layers := provider(ctx)
			log.Debugf("Component layers from provider: %d layers", len(layers))
			summary.ComponentExecutionLayers = layers
			summary.ComponentLayerCount = len(layers)
		}
	}

	// Set flags (merge with any existing flags from hooks)
	summary.Flags.DryRun = ctx.Config.DryRun
	summary.Flags.SkipDeps = ctx.Config.SkipDeps
	summary.Flags.SkipDepm = ctx.Config.SkipDepm
	summary.Flags.ForceRebuild = ctx.Config.ForceRebuild

	// System dependency verification (if hook provided)
	if !ctx.Config.SkipDeps && depsVerifier != nil {
		depsStatus := depsVerifier(ctx)
		if depsStatus != nil {
			summary.DepsStatus = *depsStatus
		}
	} else if ctx.Config.SkipDeps {
		summary.DepsStatus.Skipped = true
	}

	// Module dependency (build artifact) validation (test/scan only)
	// Build command creates artifacts - it doesn't need them to exist beforehand
	// Test/scan commands consume artifacts - they need to verify builds exist
	if ctx.Config.Type != CommandTypeBuild && !ctx.Config.SkipDepm && artifactValidator != nil {
		artifactInfo := artifactValidator(ctx)
		if artifactInfo != nil {
			summary.ArtifactValidation = artifactInfo
		}
	}

	ctx.InitSummary = summary
	return nil
}

// DepsVerifier is a function that verifies system dependencies.
// Commands provide their own implementation.
type DepsVerifier func(ctx *ExecutionContext) *initsummary.DepsStatus

// ArtifactValidator is a function that validates build artifacts.
// Commands provide their own implementation since artifacts package
// is internal to impl/.
type ArtifactValidator func(ctx *ExecutionContext) *initsummary.ArtifactValidationInfo

// UoWCountProvider is a function that returns the total UoW count.
// Commands provide their own implementation based on their work items.
type UoWCountProvider func(ctx *ExecutionContext) int

// UnitLayersProvider is a function that returns the component execution layers.
// Used to compute component layer info for the init summary display.
type UnitLayersProvider func(ctx *ExecutionContext) [][]string

var (
	depsVerifier              DepsVerifier
	artifactValidator         ArtifactValidator
	uowCountProvider          UoWCountProvider
	unitLayersProviders       = make(map[CommandType]UnitLayersProvider)
	unitLayersProvidersMu     sync.RWMutex
)

// SetDepsVerifier sets the global system dependency verifier function.
func SetDepsVerifier(v DepsVerifier) {
	depsVerifier = v
}

// SetArtifactValidator sets the global artifact validator function.
func SetArtifactValidator(v ArtifactValidator) {
	artifactValidator = v
}

// SetUoWCountProvider sets the global UoW count provider function.
func SetUoWCountProvider(p UoWCountProvider) {
	uowCountProvider = p
}

// SetUnitLayersProvider is deprecated. Use RegisterUnitLayersProvider instead.
// This function is kept for backward compatibility but should be removed.
func SetUnitLayersProvider(p UnitLayersProvider) {
	// No-op - callers should use RegisterUnitLayersProvider
}

// RegisterUnitLayersProvider registers a component layers provider for a specific command type.
// This ensures each command type (build, test, scan, lint) uses its own provider.
func RegisterUnitLayersProvider(cmdType CommandType, p UnitLayersProvider) {
	unitLayersProvidersMu.Lock()
	defer unitLayersProvidersMu.Unlock()
	unitLayersProviders[cmdType] = p
}

// GetUnitLayersProvider returns the component layers provider for a command type.
func GetUnitLayersProvider(cmdType CommandType) UnitLayersProvider {
	unitLayersProvidersMu.RLock()
	defer unitLayersProvidersMu.RUnlock()
	return unitLayersProviders[cmdType]
}

// displayInitSummary outputs the initialization summary to console.
// TUI receives the summary separately after TUI is started.
func displayInitSummary(ctx *ExecutionContext) {
	if ctx.InitSummary == nil {
		return
	}

	// Output text lines to console (TUI not started yet)
	var formatted string
	if ctx.Config.UseTUI {
		formatted = initsummary.FormatCompact(ctx.InitSummary)
	} else {
		formatted = initsummary.FormatDetailed(ctx.InitSummary)
	}

	for _, line := range strings.Split(strings.TrimSpace(formatted), "\n") {
		if line != "" {
			ctx.WriteInit("%s", line)
		}
	}
}

// convertToTUIInitSummary converts initsummary.Summary to tui.InitSummary.
// It uses the ExecutionContext to access the UnitProvider for proper ID generation.
func convertToTUIInitSummary(ctx *ExecutionContext) *tui.InitSummary {
	s := ctx.InitSummary
	ts := &tui.InitSummary{
		Command:           s.Command,
		ExecutionContext:  s.ExecutionContext,
		RequestedModules:  len(s.RequestedModules),
		CalculatedModules: len(s.CalculatedModules),
		AddedDepm:         len(s.AddedDepm),
		UoWCount:          s.UoWCount,
		FlatExecution:     s.FlatExecution,
		OutputDir:         s.OutputDir,
	}

	// Build execution tree: layers → modules → UoWs
	// Use UnitProvider to get proper IDs with Longname()
	ts.ExecutionTree = buildExecutionTreeFromUnits(ctx, s)

	// Compute layer sizes from execution tree
	ts.LayerCount = len(s.ExecutionLayers)
	ts.LayerSizes = make([]int, len(s.ExecutionLayers))
	ts.ComponentsPerModLayer = make([]int, len(s.ExecutionLayers))

	for i, layer := range s.ExecutionLayers {
		ts.LayerSizes[i] = len(layer)
	}

	// Count components per module layer from ComponentExecutionLayers
	if len(s.ComponentExecutionLayers) > 0 {
		for i, compLayer := range s.ComponentExecutionLayers {
			if i < len(ts.ComponentsPerModLayer) {
				ts.ComponentsPerModLayer[i] = len(compLayer)
			}
		}
	}

	// Pass component layers to TUI for visualization
	ts.ComponentLayers = s.ComponentExecutionLayers
	ts.ComponentLayerCount = s.ComponentLayerCount

	// Parallelism info
	if s.Parallelism != nil {
		ts.ParallelismMode = s.Parallelism.Mode
		ts.EffectiveWorkers = s.Parallelism.EffectiveWorkers
		ts.TurboBoost = s.Parallelism.TurboBoost
		ts.WeightedCapacity = s.Parallelism.WeightedCapacity
	}

	// Flags
	ts.Flags = tui.InitSummaryFlags{
		TidyFirst:    s.Flags.TidyFirst,
		ForceRebuild: s.Flags.ForceRebuild,
		DryRun:       s.Flags.DryRun,
		UseTUI:       s.Flags.UseTUI,
		SkipDeps:     s.Flags.SkipDeps,
		SkipDepm:     s.Flags.SkipDepm,
	}

	// Deps status
	ts.DepsVerified = s.DepsStatus.Verified
	ts.DepsSkipped = s.DepsStatus.Skipped
	if len(s.DepsStatus.Available) > 0 {
		for _, dep := range s.DepsStatus.Available {
			if dep.Available {
				ts.DepsAvailable = append(ts.DepsAvailable, dep.Name)
			}
		}
	}
	ts.DepsMissing = s.DepsStatus.Missing

	// Depm status
	ts.DepmVerified = s.DepmStatus.Verified
	ts.DepmSkipped = s.DepmStatus.Skipped
	ts.DepmResolved = len(s.DepmStatus.Resolved)
	ts.DepmExisting = len(s.DepmStatus.Existing)
	ts.DepmTotal = s.DepmStatus.Total
	ts.DepmMissing = s.DepmStatus.Missing

	// Incremental info
	if s.Incremental != nil {
		ts.IncrementalEnabled = s.Incremental.Enabled
		ts.IncrementalChanged = len(s.Incremental.Changed)
		ts.IncrementalUpToDate = len(s.Incremental.UpToDate)
		ts.IncrementalFresh = s.Incremental.FreshBuild
	}

	// Test info
	if s.Test != nil {
		ts.TestSuiteName = s.Test.SuiteName
		ts.TestSelected = s.Test.Selected
		ts.TestDiscovered = s.Test.TotalDiscovered
		ts.TestOSFiltered = s.Test.OSFiltered
	}

	return ts
}

// ExtractPlannedTools extracts unique tools from component work layers.
// Returns a list of tools with their IsContainer status for TUI display.
func ExtractPlannedTools(ctx *ExecutionContext) []tui.PlannedTool {
	provider := GetUnitProvider(ctx.Config.Type)
	if provider == nil {
		return nil
	}

	layers := provider(ctx)
	if len(layers) == 0 {
		return nil
	}

	// Track unique tools (tool name -> isContainer)
	toolMap := make(map[string]bool)
	for _, layer := range layers {
		for _, unit := range layer {
			if unit.ID.Tool != "" {
				// Use the full tool key to ensure uniqueness
				toolMap[unit.ID.Tool] = unit.Container
			}
		}
	}

	// Convert to slice
	tools := make([]tui.PlannedTool, 0, len(toolMap))
	for name, isContainer := range toolMap {
		tools = append(tools, tui.PlannedTool{
			Name:        name,
			IsContainer: isContainer,
		})
	}

	return tools
}

// buildExecutionTreeFromUnits builds the hierarchical tree: layers → modules → UoWs.
// Uses UnitProvider to get proper IDs with Longname() for globally unique identification.
// Falls back to string-based parsing if UnitProvider is not available.
func buildExecutionTreeFromUnits(ctx *ExecutionContext, s *initsummary.Summary) []tui.ExecutionLayer {
	if len(s.ExecutionLayers) == 0 {
		return nil
	}

	// Initialize Extra map if needed
	if ctx.Config.Extra == nil {
		ctx.Config.Extra = make(map[string]interface{})
	}

	// First check if UnitSpecs are cached (populated by UnitLayersProvider during phaseVerify)
	// This ensures we use the same data that was used for component layers
	// Check both the command-specific cache key (e.g., "componentWorkLayers" for build)
	// and the generic cache key used by this function
	if cached, ok := ctx.Config.Extra["componentWorkLayers"].([][]workunit.UnitSpec); ok && len(cached) > 0 {
		return buildTreeFromUnitSpecs(s.ExecutionLayers, cached)
	}
	if cached, ok := ctx.Config.Extra["unitSpecsCache"].([][]workunit.UnitSpec); ok && len(cached) > 0 {
		return buildTreeFromUnitSpecs(s.ExecutionLayers, cached)
	}

	// Try to get UoWs from UnitProvider for proper ID generation
	provider := GetUnitProvider(ctx.Config.Type)
	if provider != nil {
		unitLayers := provider(ctx)
		if len(unitLayers) > 0 {
			// Cache for subsequent calls (e.g., from scheduler if called again)
			ctx.Config.Extra["unitSpecsCache"] = unitLayers
			return buildTreeFromUnitSpecs(s.ExecutionLayers, unitLayers)
		}
	}

	// Fallback: build from ComponentExecutionLayers strings (legacy path)
	// Note: This path uses hardcoded Weight=1 and simplified ID format
	return buildTreeFromStrings(s)
}

// buildTreeFromUnitSpecs builds the execution tree using UnitSpec data.
// Uses Longname() for globally unique IDs and DisplayName() for display.
func buildTreeFromUnitSpecs(moduleLayers [][]string, unitLayers [][]workunit.UnitSpec) []tui.ExecutionLayer {
	// Build a map of module -> UoWEntries from unit specs
	moduleUoWs := make(map[string][]tui.UoWEntry)
	for _, layer := range unitLayers {
		for _, spec := range layer {
			module := spec.ID.Module
			entry := tui.UoWEntry{
				ID:          spec.ID.Longname(),
				DisplayName: spec.ID.Shortname(),
				Weight:      spec.Weight,
			}
			moduleUoWs[module] = append(moduleUoWs[module], entry)
		}
	}

	// Build the tree structure following module layer order
	tree := make([]tui.ExecutionLayer, len(moduleLayers))
	for i, moduleNames := range moduleLayers {
		layer := tui.ExecutionLayer{
			Modules: make([]tui.ExecutionModule, len(moduleNames)),
		}
		for j, moduleName := range moduleNames {
			layer.Modules[j] = tui.ExecutionModule{
				Name: moduleName,
				UoWs: moduleUoWs[moduleName],
			}
		}
		tree[i] = layer
	}

	return tree
}

// buildTreeFromStrings builds the execution tree from string-based ComponentExecutionLayers.
// This is the legacy fallback path when UnitProvider is not available.
func buildTreeFromStrings(s *initsummary.Summary) []tui.ExecutionLayer {
	// Build a map of module -> UoW entries from ComponentExecutionLayers
	// Components are named "module:component" or "module:component:handler"
	// We store the full ID and extract display name (part after first colon)
	moduleUoWs := make(map[string][]tui.UoWEntry)
	for _, layer := range s.ComponentExecutionLayers {
		for _, comp := range layer {
			// Extract module name and rest from "module:component[:handler]" format
			if idx := strings.Index(comp, ":"); idx > 0 {
				module := comp[:idx]
				// Store full ID and display name (everything after first colon)
				displayName := comp[idx+1:]
				moduleUoWs[module] = append(moduleUoWs[module], tui.UoWEntry{
					ID:          comp,
					DisplayName: displayName,
					Weight:      1,
				})
			} else {
				// Component without colon - use as-is (module is the component)
				moduleUoWs[comp] = append(moduleUoWs[comp], tui.UoWEntry{
					ID:          comp,
					DisplayName: comp,
					Weight:      1,
				})
			}
		}
	}

	// Build the tree structure
	tree := make([]tui.ExecutionLayer, len(s.ExecutionLayers))
	for i, moduleNames := range s.ExecutionLayers {
		layer := tui.ExecutionLayer{
			Modules: make([]tui.ExecutionModule, len(moduleNames)),
		}
		for j, moduleName := range moduleNames {
			layer.Modules[j] = tui.ExecutionModule{
				Name: moduleName,
				UoWs: moduleUoWs[moduleName],
			}
		}
		tree[i] = layer
	}

	return tree
}
