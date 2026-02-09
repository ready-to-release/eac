package cmdframework

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/display"
	"github.com/ready-to-release/eac/go/clibase/orchestrator"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

// DefaultTurboMultiplier is the turbo multiplier when --turbo is set without a value.
// 1.25 = 25% increase in pressure roof capacity.
const DefaultTurboMultiplier = 1.25

// CalculateMaxConcurrency determines the effective max concurrency based on configuration.
// Returns 0 for dynamic mode (orchestrator calculates from CPU×RAM).
// Only returns non-zero if user explicitly set a limit or sequential mode.
// Priority: sequential (1) > CLI flag (configConcurrency) > repo config (repoConcurrency) > dynamic (0)
func CalculateMaxConcurrency(configConcurrency, repoConcurrency int, turbo, sequential bool) int {
	// Sequential overrides everything
	if sequential {
		return 1
	}

	// If user explicitly set concurrency via CLI flag, use it
	if configConcurrency > 0 {
		return configConcurrency
	}

	// If repo config has a parallelism ceiling, use it
	if repoConcurrency > 0 {
		return repoConcurrency
	}

	// Return 0 for dynamic mode - orchestrator will calculate from CPU×RAM
	return 0
}

// CalculateTurboMultiplier returns the turbo multiplier for the pressure roof.
// Returns 1.0 for normal mode, DefaultTurboMultiplier for turbo mode.
func CalculateTurboMultiplier(turbo bool) float64 {
	if turbo {
		return DefaultTurboMultiplier
	}
	return 1.0
}

// phaseInitEarly performs minimal setup needed before the TUI can start.
// This runs synchronously and must complete quickly (<10ms) so the TUI
// appears immediately. Heavy work (config loading, tool init) is deferred
// to phaseInitDeferred.
func phaseInitEarly(ctx *ExecutionContext) error {
	ctx.StartTime = time.Now()

	// IMPORTANT: Create TUI console BEFORE starting output buffer.
	// The TUI console captures os.Stdout at creation time to write directly to terminal.
	// The output buffer then redirects os.Stdout to capture any leaked output from
	// subprocesses, but the TUI still writes to the real terminal.
	var tuiConsole display.Console
	useRegistryTUI := ctx.Config.CommandPath != "" && ctx.Config.UseTUI
	bootstrap := display.GetTUIBootstrap()
	if useRegistryTUI && bootstrap != nil && bootstrap.NewForCommand != nil {
		tuiConfig := display.Config{
			Height:       getTUIHeight(ctx.Config),
			BufferSize:   1000,
			RunPhaseName: ctx.Config.ActionVerb(),
			ASCIIMode:    ctx.Config.TUIASCIIMode,
			TUI3Demo:     ctx.Config.TUI3Demo,
			SkipTUIDelay: ctx.Config.SkipTUIDelay,
			CommandName:  ctx.Config.CommandPath,
			ActionType:   ctx.Config.Type,
		}
		tuiConsole = bootstrap.NewForCommand(ctx.Config.CommandPath, tuiConfig)
	}

	// Start output buffer AFTER TUI console is created - capture any stdout/stderr leaks
	// This must happen before any other initialization that might print
	ctx.initOutputBuffer()
	ctx.AddCleanup(func() { ctx.stopOutputBuffer() })

	// Create orchestrator with minimal config.
	// WorkspaceRoot, OutputBaseDir, MaxConcurrency, Turbo etc.
	// are populated later by phaseInitDeferred via UpdateConfig.
	orchConfig := orchestrator.Config{
		ActionType:           ctx.Config.Type,
		StatusUpdateInterval: 500, // 500ms for responsive feedback
		TUI:                  ctx.Config.UseTUI,
		TUIHeight:            ctx.Config.TUIHeight,
		TUIASCIIMode:         ctx.Config.TUIASCIIMode,
		TUI3Demo:             ctx.Config.TUI3Demo,
		SkipTUIDelay:         ctx.Config.SkipTUIDelay,
	}

	orch := orchestrator.New(&orchConfig, nil) // Worker set later
	ctx.Orchestrator = orch
	ctx.AddCleanup(func() { orch.Close() })

	// Wire up TUI console -> orchestrator -> observers
	if tuiConsole != nil && bootstrap != nil {
		// Unwrap the console if needed (e.g., parallel.Console -> ParallelConsole)
		innerConsole := tuiConsole
		if bootstrap.UnwrapConsole != nil {
			innerConsole = bootstrap.UnwrapConsole(tuiConsole)
		}

		orch.SetConsole(innerConsole)

		// Create TUI observer to receive execution events
		if bootstrap.NewObserver != nil {
			observer, writerFactory := bootstrap.NewObserver(innerConsole)
			if obs, ok := observer.(core.ExecutionObserver); ok {
				orch.AddObserver(obs)
			}
			if wf, ok := writerFactory.(core.WriterFactory); ok {
				orch.SetWriterFactory(wf)
			}
		}

		// Create TUI hooks for controlled interaction
		if bootstrap.NewHooks != nil {
			if hooks, ok := bootstrap.NewHooks(innerConsole).(core.TUIHooks); ok {
				ctx.TUIHooks = hooks
			}
		}
	} else if !ctx.Config.UseTUI {
		// No TUI - use console observer for plain text output
		consoleObserver := output.NewConsoleObserver()
		orch.AddObserver(consoleObserver)
	}

	// Initialize orchestrator output infrastructure
	if err := orch.Init(); err != nil {
		return fmt.Errorf("failed to initialize orchestrator: %w", err)
	}

	// START TUI IMMEDIATELY - shows loading state ("Initializing..." with animated dots)
	// The TUI already handles the "no data yet" case gracefully via renderInitPaneLoading().
	if ctx.Config.UseTUI {
		// TUI Hook: Pre-start configuration
		if ctx.TUIHooks != nil {
			if err := ctx.TUIHooks.BeforeStart(context.Background()); err != nil {
				return fmt.Errorf("TUI BeforeStart hook failed: %w", err)
			}
		}
		orch.StartTUI()
	}

	log.Debugf("TUI boot time: %v", time.Since(ctx.StartTime))

	return nil
}

// phaseInitDeferred performs the heavy initialization after the TUI is already running.
// Progress is reported to the TUI via WriteStatus, which sends phase line messages
// that feed into the init pane. The loading animation continues until InitSummary arrives.
func phaseInitDeferred(ctx *ExecutionContext) error {
	deferredStart := time.Now()

	// Find workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}
	ctx.WorkspaceRoot = workspaceRoot

	// Load EAC configuration
	configStart := time.Now()
	eacCfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return fmt.Errorf("failed to load EAC config: %w", err)
	}
	ctx.EACConfig = eacCfg
	ctx.RepoConfig = eacCfg.Repository
	ctx.initTimings.ConfigLoad = time.Since(configStart)

	// Initialize tool system bridges
	toolStart := time.Now()
	configRoot := filepath.Join(workspaceRoot, ".eac")
	if err := tool.InitializeGlobalBridges(workspaceRoot, configRoot); err != nil {
		log.Debugf("Tool bridge initialization skipped: %v", err)
		// Continue - tool config is optional, native handlers will still work
	}
	ctx.ToolSystem = tool.GlobalToolSystem()
	ctx.initTimings.ToolInit = time.Since(toolStart)

	// Create output directory
	outputDir := filepath.Join(workspaceRoot, ctx.Config.OutputDir)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine max concurrency using shared calculation
	repoConcurrency := ctx.RepoConfig.EffectiveParallelism(environments.IsCI())
	maxConcurrency := CalculateMaxConcurrency(ctx.Config.MaxConcurrency, repoConcurrency, ctx.Config.Turbo, ctx.Config.Sequential)
	turboMultiplier := CalculateTurboMultiplier(ctx.Config.Turbo)

	if ctx.Config.Turbo && !ctx.Config.Sequential {
		log.Debugf("Turbo mode enabled: %.2fx pressure multiplier", turboMultiplier)
	}

	// Update orchestrator with loaded config values
	ctx.Orchestrator.UpdateConfig(orchestrator.ConfigUpdate{
		WorkspaceRoot:         workspaceRoot,
		OutputBaseDir:         ctx.Config.OutputDir,
		MaxConcurrency:        maxConcurrency,
		Turbo:                 turboMultiplier,
		ComponentTypesDisplay: ctx.ComponentTypesDisplay,
		ShowTimings:           ctx.Config.ShowTimings,
		DryRun:                ctx.Config.DryRun,
	})

	// Configure logging
	logName := string(ctx.Config.Type)
	if err := logging.ConfigureLoggingSimple(workspaceRoot, logName, nil, ctx.Config.DebugMode); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	ctx.AddCleanup(func() { logging.CloseLogging() })

	// Start async dependency verification (e.g., `docker info`) in background.
	// This overlaps with module resolution and change detection (~1-3s),
	// hiding the Docker check latency (~100-500ms) from the critical path.
	ctx.asyncDepsResult = startAsyncDepsCheck(ctx)

	log.Debugf("%s logging configured: debugMode=%v, useTUI=%v",
		ctx.Config.ActionVerb(), ctx.Config.DebugMode, ctx.Config.UseTUI)
	log.Debugf("Deferred init time: %v", time.Since(deferredStart))

	return nil
}

// getTUIHeight returns the TUI height, using default if not specified.
func getTUIHeight(cfg *CommandConfig) int {
	if cfg.TUIHeight > 0 {
		return cfg.TUIHeight
	}
	return display.DefaultHeight
}
