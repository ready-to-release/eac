package cmdframework

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/adapters/tui"
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
//	    Type:       cmdframework.CommandTypeBuild,
//	    ActionVerb: "Building",
//	    OutputDir:  "out/build",
//	    // ... other config
//	}, myWorkerFunc, nil)
func Run(cfg *CommandConfig, worker WorkerFunc, hooks *Hooks) int {
	// ISOLATION CHECK: Fail-fast if running build/test/scan/lint within test scope
	// Tests should not actively run these commands - it's bad isolation.
	// Use --dry-run for planning/validation within tests.
	if os.Getenv(environments.EnvR2RTestScope) != "" && !cfg.DryRun {
		switch cfg.Type {
		case CommandTypeBuild, CommandTypeTest, CommandTypeScan, CommandTypeLint:
			log.Errorf("ISOLATION VIOLATION: %s command cannot run within test scope without --dry-run", cfg.Type)
			log.Errorf("Tests should not actively execute %s - use --dry-run for validation", cfg.Type)
			return 1
		}
	}

	if hooks == nil {
		hooks = &Hooks{}
	}

	ctx := &ExecutionContext{
		Config:                cfg,
		ComponentTypesDisplay: make(map[string]string),
	}

	// Phase 1a: Early init - TUI starts immediately (shows loading animation)
	if err := phaseInitEarly(ctx); err != nil {
		log.Errorf("Early initialization failed: %v", err)
		return 1
	}
	defer ctx.Cleanup()

	// Phase 1b: Deferred init - config, tools, etc. (TUI shows loading dots)
	if err := phaseInitDeferred(ctx); err != nil {
		log.Errorf("Initialization failed: %v", err)
		if cfg.UseTUI && ctx.Orchestrator != nil {
			// TUI is already running - show error and exit cleanly
			ctx.Orchestrator.SendInitLine("ERROR: " + err.Error())
			ctx.Orchestrator.SendSummary(&tui.SummaryData{
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
	// This fills in command name, parallelism mode, worker count before module resolution.
	if cfg.UseTUI && ctx.Orchestrator != nil {
		sendConfigReady(ctx)
	}

	// Hook: AfterInit
	if hooks.AfterInit != nil {
		if err := hooks.AfterInit(ctx); err != nil {
			log.Errorf("AfterInit hook failed: %v", err)
			return 1
		}
	}

	// Phase 2: Resolve Modules (skipped if SkipResolve is set)
	if !cfg.SkipResolve {
		if err := phaseResolve(ctx); err != nil {
			log.Errorf("Module resolution failed: %v", err)
			return 1
		}
		// Output boot status for resolve phase
		moduleCount := len(ctx.ModuleRegistry.All())
		ctx.WriteStatus(true, "Discovering modules... %d found %s", moduleCount, formatTiming(ctx.initTimings.ModuleDiscovery))
		if ctx.initTimings.ExecutionOrder > 0 {
			ctx.WriteStatus(true, "Calculating execution order... %s", formatTiming(ctx.initTimings.ExecutionOrder))
		}
	}

	// Hook: AfterResolve (always called - allows command to set up execution plan if SkipResolve)
	if hooks.AfterResolve != nil {
		hookStart := time.Now()
		if err := hooks.AfterResolve(ctx); err != nil {
			// Check for informational exit (user-facing output already written)
			var infoErr ErrInformationalExit
			if errors.As(err, &infoErr) {
				// Graceful exit - no additional error logging needed
				return 1
			}
			log.Errorf("AfterResolve hook failed: %v", err)
			return 1
		}
		// If change detection happened, output its timing
		if ctx.initTimings.ChangeDetection > 0 {
			ctx.WriteStatus(true, "Detecting changes... %s", formatTiming(ctx.initTimings.ChangeDetection))
		} else if time.Since(hookStart) > 10*time.Millisecond {
			// Generic hook timing for non-trivial hooks
			ctx.WriteStatus(true, "Preparing execution... %s", formatTiming(time.Since(hookStart)))
		}
	}

	// Phase 3: Verify Dependencies
	verifyStart := time.Now()
	if err := phaseVerify(ctx); err != nil {
		log.Errorf("Verification failed: %v", err)
		return 1
	}
	ctx.initTimings.DepsVerify = time.Since(verifyStart)
	if !cfg.SkipDeps && ctx.initTimings.DepsVerify > 5*time.Millisecond {
		ctx.WriteStatus(true, "Verifying dependencies... %s", formatTiming(ctx.initTimings.DepsVerify))
	}

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

	// Check if artifact validation failed (missing required build artifacts)
	// This only applies to test/scan commands - build command creates artifacts, it doesn't require them
	if cfg.Type != CommandTypeBuild && ctx.InitSummary != nil && ctx.InitSummary.ArtifactValidation != nil {
		av := ctx.InitSummary.ArtifactValidation
		if !av.AllPresent && len(av.MissingFrom) > 0 {
			ctx.WriteInit("")
			ctx.WriteInit("❌ Missing required artifacts from: %v", av.MissingFrom)
			ctx.WriteInit("   Run 'build' command first, or use --skip-depm to skip validation")
			// Stop TUI before returning
			if ctx.Orchestrator != nil && cfg.UseTUI {
				ctx.Orchestrator.WaitTUI()
				ctx.Orchestrator.StopTUI()
			}
			return 1
		}
	}

	// Hook: BeforeExecute
	if hooks.BeforeExecute != nil {
		if err := hooks.BeforeExecute(ctx); err != nil {
			log.Errorf("BeforeExecute hook failed: %v", err)
			return 1
		}
	}

	// Phase 4: Execute
	if err := phaseExecute(ctx, worker); err != nil {
		log.Errorf("Execution failed: %v", err)
		return 1
	}

	// Hook: AfterExecute - runs post-build tasks like manifest generation
	// In TUI mode: runs in background while TUI is displayed (TUI has its own exit hold)
	// In non-TUI mode: must complete before process exits
	type afterExecResult struct {
		err error
	}
	var afterExecDone chan afterExecResult
	if hooks.AfterExecute != nil {
		afterExecDone = make(chan afterExecResult, 1)
		go func() {
			afterExecStart := time.Now()
			log.Debugf("AfterExecute: starting")
			err := hooks.AfterExecute(ctx)
			if err != nil {
				log.Errorf("AfterExecute hook failed: %v", err)
			}
			log.Debugf("AfterExecute: completed in %v", time.Since(afterExecStart))
			afterExecDone <- afterExecResult{err: err}
			close(afterExecDone)
		}()
	}

	// Phase 5: Summary - handles TUI exit (user timer if active, else immediate)
	exitCode := phaseSummary(ctx, hooks.CustomSummary)

	// Wait for AfterExecute to complete and propagate errors
	// In TUI mode, we still wait for completion to ensure errors are captured
	// In non-TUI mode, we must wait here or the process exits before AfterExecute finishes
	if afterExecDone != nil {
		log.Debugf("Waiting for AfterExecute to complete...")
		result := <-afterExecDone
		log.Debugf("AfterExecute complete, exiting")
		// Propagate AfterExecute errors to exit code (only if execution succeeded)
		if result.err != nil && exitCode == 0 {
			exitCode = 1
		}
	}

	return exitCode
}

// RunSimple is a convenience wrapper for commands that don't need hooks.
func RunSimple(cfg *CommandConfig, worker WorkerFunc) int {
	return Run(cfg, worker, nil)
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
func buildUoWData(ctx *ExecutionContext) interfaces.UoWData {
	if ctx.InitSummary == nil {
		return interfaces.UoWData{}
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
	return interfaces.UoWData{}
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
func buildUoWDataFromUnitSpecs(units []workunit.UnitSpec) interfaces.UoWData {
	// Build a map of module -> UoWUnits from unit specs
	moduleUoWs := make(map[string][]interfaces.UoWUnit)
	moduleOrder := []string{}
	seenModules := make(map[string]bool)

	for _, spec := range units {
		module := spec.ID.Module
		entry := interfaces.UoWUnit{
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
	modules := make([]interfaces.UoWModule, len(moduleOrder))
	for i, moduleName := range moduleOrder {
		modules[i] = interfaces.UoWModule{
			Name:  moduleName,
			Units: moduleUoWs[moduleName],
		}
	}

	return interfaces.UoWData{
		Modules: modules,
	}
}
