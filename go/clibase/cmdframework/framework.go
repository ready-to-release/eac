package cmdframework

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/display"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/workunit"
)

var log = logging.C()

// Run executes the command framework with the given configuration and worker.
// It orchestrates the standard phases: init, resolve, verify, execute, summary.
//
// Usage:
//
//	return cmdframework.Run(&cmdframework.CommandConfig{
//	    Type:      core.ActionBuild,
//	    OutputDir: "out/build",
//	    // ... other config
//	}, myWorkerFunc, nil)
func Run(cfg *CommandConfig, worker CommandWorkerFunc, hooks *Hooks) int {
	if failed := checkTestIsolation(cfg); failed {
		return 1
	}

	if hooks == nil {
		hooks = &Hooks{}
	}

	ctx := &ExecutionContext{
		Config:                cfg,
		ComponentTypesDisplay: make(map[string]string),
	}

	// Phase 1: Initialize (early TUI + deferred config/tools)
	if code := runInitPhase(ctx); code >= 0 {
		return code
	}
	defer ctx.Cleanup()

	// Hook: AfterInit
	if err := runHook("AfterInit", hooks.AfterInit, ctx); err != nil {
		return 1
	}

	// Phase 2: Resolve modules + AfterResolve hook
	if code := runResolvePhase(ctx, hooks); code >= 0 {
		return code
	}

	// Phase 3: Verify dependencies
	if code := runVerifyPhase(ctx); code >= 0 {
		return code
	}

	// Publish init summary to console and TUI
	publishInitSummary(ctx)

	// Gate: check system dependencies (docker, etc.)
	if code := checkSystemDependencies(ctx); code >= 0 {
		return code
	}

	// Gate: check required build artifacts
	if code := checkRequiredArtifacts(ctx); code >= 0 {
		return code
	}

	// Hook: BeforeExecute
	if err := runHook("BeforeExecute", hooks.BeforeExecute, ctx); err != nil {
		return 1
	}

	// Phase 4: Execute workers
	if err := phaseExecute(ctx, worker); err != nil {
		log.Errorf("Execution failed: %v", err)
		return 1
	}

	// Phase 5: AfterExecute (async) + Summary + wait
	return runPostExecuteAndSummary(ctx, hooks)
}

// RunSimple is a convenience wrapper for commands that don't need hooks.
func RunSimple(cfg *CommandConfig, worker CommandWorkerFunc) int {
	return Run(cfg, worker, nil)
}

// ---------------------------------------------------------------------------
// Run helper functions: each encapsulates a logical phase of the Run pipeline.
// These are unexported and exist solely to keep Run readable.
// ---------------------------------------------------------------------------

// checkTestIsolation fails fast if a build/test/scan/lint command runs inside a
// test scope without --dry-run. Returns true when the caller should exit with 1.
func checkTestIsolation(cfg *CommandConfig) bool {
	if os.Getenv(environments.EnvCLIETestScope) == "" || cfg.DryRun {
		return false
	}
	switch cfg.Type {
	case core.ActionBuild, core.ActionTest, core.ActionScan, core.ActionLint, core.ActionDeploy:
		log.Errorf("ISOLATION VIOLATION: %s command cannot run within test scope without --dry-run", cfg.Type)
		log.Errorf("Tests should not actively execute %s - use --dry-run for validation", cfg.Type)
		return true
	}
	return false
}

// runInitPhase runs Phase 1a (early init / TUI) and Phase 1b (deferred config/tools),
// then emits boot-status lines and sends config metadata to the TUI.
// Returns a non-negative exit code on failure; returns -1 to continue.
func runInitPhase(ctx *ExecutionContext) int {
	cfg := ctx.Config

	// Phase 1a: Early init - TUI starts immediately (shows loading animation)
	if err := phaseInitEarly(ctx); err != nil {
		log.Errorf("Early initialization failed: %v", err)
		return 1
	}

	// Phase 1b: Deferred init - config, tools, etc. (TUI shows loading dots)
	if err := phaseInitDeferred(ctx); err != nil {
		log.Errorf("Initialization failed: %v", err)
		if cfg.UseTUI && ctx.Orchestrator != nil {
			ctx.Orchestrator.SendInitLine("ERROR: " + err.Error())
			ctx.Orchestrator.SendSummary(&display.SummaryData{
				Success:   false,
				TotalTime: time.Since(ctx.StartTime),
				Details:   []string{"Initialization failed: " + err.Error()},
			})
			ctx.Orchestrator.WaitTUI()
			ctx.Orchestrator.StopTUI()
		}
		return 1
	}

	// Output boot status for init phase
	ctx.WriteStatus(true, "Loading configuration... %s", formatTiming(ctx.initTimings.ConfigLoad))
	if ctx.initTimings.ToolInit > 0 {
		ctx.WriteStatus(true, "Initializing tool system... %s", formatTiming(ctx.initTimings.ToolInit))
	}

	// Send early config metadata to TUI for progressive display.
	if cfg.UseTUI && ctx.Orchestrator != nil {
		sendConfigReady(ctx)
	}

	return -1 // continue
}

// runHook executes a single optional PhaseHook. If the hook is nil, it is a no-op.
// Returns nil on success (or nil hook), or the hook error after logging.
func runHook(name string, hook PhaseHook, ctx *ExecutionContext) error {
	if hook == nil {
		return nil
	}
	if err := hook(ctx); err != nil {
		log.Errorf("%s hook failed: %v", name, err)
		return err
	}
	return nil
}

// runResolvePhase runs Phase 2 (module resolution) and the AfterResolve hook,
// including informational-exit handling and timing output.
// Returns a non-negative exit code on failure; returns -1 to continue.
func runResolvePhase(ctx *ExecutionContext, hooks *Hooks) int {
	cfg := ctx.Config

	// Phase 2: Resolve Modules (skipped if SkipResolve is set)
	if !cfg.SkipResolve {
		if err := phaseResolve(ctx); err != nil {
			log.Errorf("Module resolution failed: %v", err)
			return 1
		}
		moduleCount := len(ctx.ModuleRegistry.All())
		ctx.WriteStatus(true, "Discovering modules... %d found %s", moduleCount, formatTiming(ctx.initTimings.ModuleDiscovery))
		if ctx.initTimings.ExecutionOrder > 0 {
			ctx.WriteStatus(true, "Calculating execution order... %s", formatTiming(ctx.initTimings.ExecutionOrder))
		}
	}

	// Hook: AfterResolve (always called - allows command to set up execution plan if SkipResolve)
	if hooks.AfterResolve != nil {
		if code := runAfterResolveHook(ctx, hooks.AfterResolve); code >= 0 {
			return code
		}
	}

	return -1 // continue
}

// runAfterResolveHook executes the AfterResolve hook with informational-exit and
// timing support. Returns a non-negative exit code on failure; returns -1 to continue.
func runAfterResolveHook(ctx *ExecutionContext, hook PhaseHook) int {
	cfg := ctx.Config
	hookStart := time.Now()

	if err := hook(ctx); err != nil {
		// Check for informational exit (user-facing output already written)
		var infoErr ErrInformationalExit
		if errors.As(err, &infoErr) {
			if cfg.UseTUI && ctx.Orchestrator != nil {
				ctx.Orchestrator.SendSummary(&display.SummaryData{
					Success:   false,
					TotalTime: time.Since(ctx.StartTime),
					Details:   []string{infoErr.Reason},
				})
				ctx.Orchestrator.WaitTUI()
				ctx.Orchestrator.StopTUI()
			}
			return 1
		}
		log.Errorf("AfterResolve hook failed: %v", err)
		return 1
	}

	// Emit timing for change detection or generic hook duration
	if ctx.initTimings.ChangeDetection > 0 {
		ctx.WriteStatus(true, "Detecting changes... %s", formatTiming(ctx.initTimings.ChangeDetection))
	} else if time.Since(hookStart) > 10*time.Millisecond {
		ctx.WriteStatus(true, "Preparing execution... %s", formatTiming(time.Since(hookStart)))
	}

	return -1 // continue
}

// runVerifyPhase runs Phase 3 (dependency verification) and emits timing output.
// Returns a non-negative exit code on failure; returns -1 to continue.
func runVerifyPhase(ctx *ExecutionContext) int {
	verifyStart := time.Now()
	if err := phaseVerify(ctx); err != nil {
		log.Errorf("Verification failed: %v", err)
		return 1
	}
	ctx.initTimings.DepsVerify = time.Since(verifyStart)
	if !ctx.Config.SkipDeps && ctx.initTimings.DepsVerify > 5*time.Millisecond {
		ctx.WriteStatus(true, "Verifying dependencies... %s", formatTiming(ctx.initTimings.DepsVerify))
	}
	return -1 // continue
}

// publishInitSummary sends the init summary to both the console and the TUI,
// including UoW data for visualization.
func publishInitSummary(ctx *ExecutionContext) {
	cfg := ctx.Config

	// Display init summary to console
	displayInitSummary(ctx)

	// Send init summary to TUI (replaces loading dots with compact init + tabs)
	if cfg.UseTUI && ctx.InitSummary != nil {
		tuiSummary := convertToTUIInitSummary(ctx)
		tuiSummary.PlannedTools = ExtractPlannedTools(ctx)
		ctx.Orchestrator.SetInitSummary(tuiSummary)

		// TUI Hook: Send UoW data for visualization (after full command is known)
		if ctx.TUIHooks != nil {
			uowData := buildUoWData(ctx)
			ctx.TUIHooks.ReceiveUoWs(uowData)
		}
	}
}

// checkSystemDependencies verifies that required system dependencies are available.
// Returns a non-negative exit code on failure; returns -1 to continue.
func checkSystemDependencies(ctx *ExecutionContext) int {
	if ctx.InitSummary == nil || ctx.Config.SkipDeps {
		return -1
	}
	ds := ctx.InitSummary.DepsStatus
	if ds.Skipped || len(ds.Missing) == 0 {
		return -1
	}

	ctx.WriteInit("")
	ctx.WriteInit("❌ Required system dependencies are not available: %s", strings.Join(ds.Missing, ", "))
	ctx.WriteInit("   Use --skip-deps to proceed anyway")
	if ctx.Orchestrator != nil && ctx.Config.UseTUI {
		ctx.Orchestrator.SignalAllWorkDone()
		ctx.Orchestrator.SendSummary(&display.SummaryData{
			Success:   false,
			TotalTime: time.Since(ctx.StartTime),
			Details:   []string{"Missing system dependencies: " + strings.Join(ds.Missing, ", ")},
		})
		ctx.Orchestrator.WaitTUI()
		ctx.Orchestrator.StopTUI()
	}
	return 1
}

// checkRequiredArtifacts verifies that build artifacts required by test/scan
// commands are present. Returns a non-negative exit code on failure; returns -1
// to continue.
func checkRequiredArtifacts(ctx *ExecutionContext) int {
	cfg := ctx.Config
	if cfg.Type == core.ActionBuild {
		return -1 // build creates artifacts, it doesn't require them
	}
	if ctx.InitSummary == nil || ctx.InitSummary.ArtifactValidation == nil {
		return -1
	}
	av := ctx.InitSummary.ArtifactValidation
	if av.AllPresent || len(av.MissingFrom) == 0 {
		return -1
	}

	ctx.WriteInit("")
	ctx.WriteInit("❌ Missing required artifacts from: %v", av.MissingFrom)
	ctx.WriteInit("   Run 'build' command first, or use --skip-depm to skip validation")
	if ctx.Orchestrator != nil && cfg.UseTUI {
		ctx.Orchestrator.SignalAllWorkDone()
		ctx.Orchestrator.SendSummary(&display.SummaryData{
			Success:   false,
			TotalTime: time.Since(ctx.StartTime),
			Details:   []string{fmt.Sprintf("Missing required artifacts from: %v", av.MissingFrom)},
		})
		ctx.Orchestrator.WaitTUI()
		ctx.Orchestrator.StopTUI()
	}
	return 1
}

// afterExecResult captures the outcome of an asynchronous AfterExecute hook.
type afterExecResult struct {
	err error
}

// runPostExecuteAndSummary handles the AfterExecute hook (async), the summary
// phase, and waits for AfterExecute to finish. Returns the final exit code.
func runPostExecuteAndSummary(ctx *ExecutionContext, hooks *Hooks) int {
	afterExecDone := startAfterExecuteHook(ctx, hooks.AfterExecute)

	// Phase 5: Summary - handles TUI exit (user timer if active, else immediate)
	exitCode := phaseSummary(ctx, hooks.CustomSummary)

	// Wait for AfterExecute to complete and propagate errors
	exitCode = awaitAfterExecute(afterExecDone, exitCode)

	return exitCode
}

// startAfterExecuteHook launches the AfterExecute hook asynchronously if present.
// When the hook (or immediate signal) completes, it calls SignalAllWorkDone on the
// orchestrator. Returns a channel to await the result (nil if no hook).
func startAfterExecuteHook(ctx *ExecutionContext, hook PhaseHook) <-chan afterExecResult {
	if hook == nil {
		// No AfterExecute hook - signal immediately
		ctx.Orchestrator.SignalAllWorkDone()
		return nil
	}

	done := make(chan afterExecResult, 1)
	go func() {
		afterExecStart := time.Now()
		log.Debugf("AfterExecute: starting")
		err := hook(ctx)
		if err != nil {
			log.Errorf("AfterExecute hook failed: %v", err)
		}
		log.Debugf("AfterExecute: completed in %v", time.Since(afterExecStart))
		done <- afterExecResult{err: err}
		close(done)
		// Signal TUI that all work is done (including AfterExecute)
		ctx.Orchestrator.SignalAllWorkDone()
	}()
	return done
}

// awaitAfterExecute blocks until the AfterExecute hook completes (if running)
// and propagates errors to the exit code. Only overrides exitCode when it was 0.
func awaitAfterExecute(done <-chan afterExecResult, exitCode int) int {
	if done == nil {
		return exitCode
	}
	log.Debugf("Waiting for AfterExecute to complete...")
	result := <-done
	log.Debugf("AfterExecute complete, exiting")
	if result.err != nil && exitCode == 0 {
		return 1
	}
	return exitCode
}

// formatTiming formats a duration for boot-style output.
// Returns "(Xms)" or "(X.Xs)" format.
func formatTiming(d time.Duration) string {
	if d < time.Millisecond {
		return ""
	}
	if d < time.Second {
		return fmt.Sprintf("(%dms)", d.Milliseconds())
	}
	return fmt.Sprintf("(%.1fs)", d.Seconds())
}

// buildUoWData creates UoW data from the execution context for TUI visualization.
// Uses the same approach as buildExecutionTreeFromUnits to get proper unit IDs.
func buildUoWData(ctx *ExecutionContext) core.UoWData {
	if ctx.InitSummary == nil {
		return core.UoWData{}
	}

	// Try to get UoWs from UnitProvider for proper ID generation
	provider := GetUnitProvider(ctx.Config.Type)
	if provider != nil {
		units := provider(ctx)
		if len(units) > 0 {
			return buildUoWDataFromUnitSpecs(units)
		}
	}

	// Fallback: empty data
	return core.UoWData{}
}

// sendConfigReady sends early configuration metadata to the TUI.
// This enables progressive display of command context before module resolution.
func sendConfigReady(ctx *ExecutionContext) {
	repoConcurrency := ctx.RepoConfig.EffectiveParallelism(environments.IsCI())
	maxConcurrency := CalculateMaxConcurrency(ctx.Config.MaxConcurrency, repoConcurrency, ctx.Config.Turbo, ctx.Config.Sequential)

	parallelismMode := "devbox"
	if environments.IsCI() {
		parallelismMode = "ci"
	}

	ctx.Orchestrator.SendConfigReady(
		string(ctx.Config.Type),
		string(logging.GetExecutionContext()),
		parallelismMode,
		maxConcurrency,
		maxConcurrency, // Approximate until scheduler calculates real value
		ctx.Config.OutputDir,
	)
}

// buildUoWDataFromUnitSpecs builds UoW data using UnitSpec data for proper IDs.
func buildUoWDataFromUnitSpecs(units []workunit.UnitSpec) core.UoWData {
	// Build a map of module -> UoWUnits from unit specs
	moduleUoWs := make(map[string][]core.UoWUnit)
	moduleOrder := []string{}
	seenModules := make(map[string]bool)

	for _, spec := range units {
		module := spec.ID.Module
		entry := core.UoWUnit{
			ID:          spec.ID.Longname(),
			DisplayName: spec.ID.DisplayName(),
			Weight:      spec.Weight,
		}
		moduleUoWs[module] = append(moduleUoWs[module], entry)

		if !seenModules[module] {
			seenModules[module] = true
			moduleOrder = append(moduleOrder, module)
		}
	}

	// Build flat module list
	modules := make([]core.UoWModule, len(moduleOrder))
	for i, moduleName := range moduleOrder {
		modules[i] = core.UoWModule{
			Name:  moduleName,
			Units: moduleUoWs[moduleName],
		}
	}

	return core.UoWData{
		Modules: modules,
	}
}
