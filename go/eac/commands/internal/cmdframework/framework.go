package cmdframework

import (
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
	}

	// Hook: AfterResolve (always called - allows command to set up execution plan if SkipResolve)
	if hooks.AfterResolve != nil {
		if err := hooks.AfterResolve(ctx); err != nil {
			log.Errorf("AfterResolve hook failed: %v", err)
			return 1
		}
	}

	// Phase 3: Verify Dependencies
	if err := phaseVerify(ctx); err != nil {
		log.Errorf("Verification failed: %v", err)
		return 1
	}

	// Display init summary
	displayInitSummary(ctx)

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
