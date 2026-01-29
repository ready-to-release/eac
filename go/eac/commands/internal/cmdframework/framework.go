package cmdframework

import (
	"fmt"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/logging"
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
	if hooks == nil {
		hooks = &Hooks{}
	}

	ctx := &ExecutionContext{
		Config:      cfg,
		ModuleTypes: make(map[string]string),
	}

	// Phase 1: Initialize
	if err := phaseInit(ctx); err != nil {
		log.Errorf("Initialization failed: %v", err)
		return 1
	}
	defer ctx.Cleanup()

	// Output boot status for init phase
	ctx.WriteStatus(true, "Loading configuration... %s", formatTiming(ctx.initTimings.ConfigLoad))
	if ctx.initTimings.ToolInit > 0 {
		ctx.WriteStatus(true, "Initializing tool system... %s", formatTiming(ctx.initTimings.ToolInit))
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

	// Start TUI display now that init is complete
	if cfg.UseTUI {
		ctx.WriteStatus(true, "Booting TUI...")
		ctx.Orchestrator.StartTUI()

		// Send init summary to TUI
		if ctx.InitSummary != nil {
			tuiSummary := convertToTUIInitSummary(ctx.InitSummary)
			ctx.Orchestrator.SetInitSummary(tuiSummary)
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

	// Hook: AfterExecute
	if hooks.AfterExecute != nil {
		if err := hooks.AfterExecute(ctx); err != nil {
			log.Errorf("AfterExecute hook failed: %v", err)
			// Continue to summary even if hook fails
		}
	}

	// Phase 5: Summary
	return phaseSummary(ctx, hooks.CustomSummary)
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
