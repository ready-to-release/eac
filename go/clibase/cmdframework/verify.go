package cmdframework

import (
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/display"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// detectExecutionContext returns a human-readable execution context string.
// Combines container detection (clie vs implicit) with environment (CI vs devbox).
// Returns: "implicit-cli (devbox)", "implicit-cli (CI)", "clie-cli (devbox)", "clie-cli (CI)".
func detectExecutionContext() string {
	// CLI mode: clie-cli (container) or implicit-cli (local)
	cliMode := string(logging.GetExecutionContext())

	// Environment: CI or devbox
	env := "devbox"
	if os.Getenv(environments.EnvCI) != "" || os.Getenv(environments.EnvGitHubActions) != "" {
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

	// Set flags (merge with any existing flags from hooks)
	summary.Flags.DryRun = ctx.Config.DryRun
	summary.Flags.SkipDeps = ctx.Config.SkipDeps
	summary.Flags.SkipDepm = ctx.Config.SkipDepm
	summary.Flags.ForceRebuild = ctx.Config.ForceRebuild

	// System dependency verification - collect from async check started in phaseInitDeferred.
	// This overlaps the Docker check (~100-500ms) with config loading and module resolution.
	if ctx.asyncDepsResult != nil {
		depsStatus := ctx.asyncDepsResult.wait()
		if depsStatus != nil {
			summary.DepsStatus = *depsStatus
		}
	} else if ctx.Config.SkipDeps {
		summary.DepsStatus.Skipped = true
	}

	// Module dependency (build artifact) validation (test/scan only)
	// Build command creates artifacts - it doesn't need them to exist beforehand
	// Test/scan commands consume artifacts - they need to verify builds exist
	if ctx.Config.Type != core.ActionBuild && !ctx.Config.SkipDepm && artifactValidator != nil {
		artifactInfo := artifactValidator(ctx)
		if artifactInfo != nil {
			summary.ArtifactValidation = artifactInfo
		}
	}

	ctx.InitSummary = summary
	return nil
}

// asyncDepsCheck holds the result of a background dependency verification.
// Started in phaseInitDeferred, collected in phaseVerify.
type asyncDepsCheck struct {
	done   chan struct{}
	result *initsummary.DepsStatus
}

// startAsyncDepsCheck begins dependency verification in a goroutine.
// Returns nil if deps are skipped or no verifier is registered.
func startAsyncDepsCheck(ctx *ExecutionContext) *asyncDepsCheck {
	if ctx.Config.SkipDeps || depsVerifier == nil {
		return nil
	}
	check := &asyncDepsCheck{done: make(chan struct{})}
	go func() {
		defer close(check.done)
		check.result = depsVerifier(ctx)
	}()
	return check
}

// wait blocks until the background check completes and returns the result.
func (c *asyncDepsCheck) wait() *initsummary.DepsStatus {
	<-c.done
	return c.result
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

var (
	depsVerifier      DepsVerifier
	artifactValidator ArtifactValidator
	uowCountProvider  UoWCountProvider
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

// convertToTUIInitSummary converts initsummary.Summary to display.InitSummary.
// It uses the ExecutionContext to access the UnitProvider for proper ID generation.
func convertToTUIInitSummary(ctx *ExecutionContext) *display.InitSummary {
	s := ctx.InitSummary
	ts := &display.InitSummary{
		Command:           s.Command,
		ExecutionContext:  s.ExecutionContext,
		RequestedModules:  len(s.RequestedModules),
		CalculatedModules: len(s.CalculatedModules),
		AddedDepm:         len(s.AddedDepm),
		UoWCount:          s.UoWCount,
		OutputDir:         s.OutputDir,
	}

	// Build execution tree: modules → UoWs
	// Use UnitProvider to get proper IDs with Longname()
	ts.ExecutionTree = buildExecutionTreeFromUnits(ctx)

	// Parallelism info
	if s.Parallelism != nil {
		ts.ParallelismMode = s.Parallelism.Mode
		ts.EffectiveWorkers = s.Parallelism.EffectiveWorkers
		ts.TurboBoost = s.Parallelism.TurboBoost
		ts.WeightedCapacity = s.Parallelism.WeightedCapacity
	}

	// Flags
	ts.Flags = display.InitSummaryFlags{
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

// ExtractPlannedTools extracts unique tools from component work units.
// Returns a list of tools with their IsContainer status for TUI display.
func ExtractPlannedTools(ctx *ExecutionContext) []display.PlannedTool {
	provider := GetUnitProvider(ctx.Config.Type)
	if provider == nil {
		return nil
	}

	units := provider(ctx)
	if len(units) == 0 {
		return nil
	}

	// Track unique tools (tool name -> isContainer)
	toolMap := make(map[string]bool)
	for _, unit := range units {
		if unit.ID.Tool != "" {
			toolMap[unit.ID.Tool] = unit.Container
		}
	}

	// Convert to slice
	tools := make([]display.PlannedTool, 0, len(toolMap))
	for name, isContainer := range toolMap {
		tools = append(tools, display.PlannedTool{
			Name:        name,
			IsContainer: isContainer,
		})
	}

	return tools
}

// buildExecutionTreeFromUnits builds a flat list of modules with their UoWs.
// Uses UnitProvider to get proper IDs with Longname() for globally unique identification.
func buildExecutionTreeFromUnits(ctx *ExecutionContext) []display.ExecutionModule {
	// Check if UnitSpecs are cached
	if len(ctx.Config.UnitSpecsCache) > 0 {
		return buildModulesFromUnitSpecs(ctx.Config.UnitSpecsCache)
	}

	// Try to get UoWs from UnitProvider for proper ID generation
	provider := GetUnitProvider(ctx.Config.Type)
	if provider != nil {
		units := provider(ctx)
		if len(units) > 0 {
			// Cache for subsequent calls
			ctx.Config.UnitSpecsCache = units
			return buildModulesFromUnitSpecs(units)
		}
	}

	return nil
}

// buildModulesFromUnitSpecs builds module list from unit specs.
func buildModulesFromUnitSpecs(units []workunit.UnitSpec) []display.ExecutionModule {
	// Build a map of module -> UoWEntries from unit specs
	moduleUoWs := make(map[string][]display.UoWEntry)
	moduleOrder := []string{} // Track module order
	seenModules := make(map[string]bool)

	for _, spec := range units {
		module := spec.ID.Module
		entry := display.UoWEntry{
			ID:            spec.ID.Longname(),
			DisplayName:   spec.ID.DisplayName(),
			Weight:        spec.Weight,
			Module:        spec.ID.Module,
			Component:     spec.ID.Component,
			Tool:          spec.ID.Tool,
			ComponentType: spec.ComponentType,
			Container:     spec.Container,
		}
		moduleUoWs[module] = append(moduleUoWs[module], entry)

		if !seenModules[module] {
			seenModules[module] = true
			moduleOrder = append(moduleOrder, module)
		}
	}

	// Build the flat module list maintaining order
	modules := make([]display.ExecutionModule, len(moduleOrder))
	for i, moduleName := range moduleOrder {
		modules[i] = display.ExecutionModule{
			Name: moduleName,
			UoWs: moduleUoWs[moduleName],
		}
	}

	return modules
}
