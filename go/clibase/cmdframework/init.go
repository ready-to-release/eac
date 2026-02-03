package cmdframework

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/go/clibase/orchestrator"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/adapters/tui/parallel"
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

// phaseInit handles the initialization phase:
// - Find workspace root
// - Load EAC configuration
// - Configure logging
// - Set up orchestrator
// - Initialize TUI if enabled.
func phaseInit(ctx *ExecutionContext) error {
	ctx.StartTime = time.Now()

	// IMPORTANT: Create TUI console BEFORE starting output buffer.
	// The TUI console captures os.Stdout at creation time to write directly to terminal.
	// The output buffer then redirects os.Stdout to capture any leaked output from
	// subprocesses, but the TUI still writes to the real terminal.
	var tuiConsole tui.Console
	useRegistryTUI := ctx.Config.CommandPath != "" && ctx.Config.UseTUI
	if useRegistryTUI {
		tuiConfig := tui.Config{
			Height:       getTUIHeight(ctx.Config),
			BufferSize:   1000,
			RunPhaseName: ctx.Config.ActionVerb,
			ASCIIMode:    ctx.Config.TUIASCIIMode,
			TUI3Demo:     ctx.Config.TUI3Demo,
			SkipTUIDelay: ctx.Config.SkipTUIDelay,
			CommandName:  ctx.Config.CommandPath,
			CommandType:  string(ctx.Config.Type),
		}
		tuiConsole = tui.NewForCommand(ctx.Config.CommandPath, tuiConfig)
	}

	// Start output buffer AFTER TUI console is created - capture any stdout/stderr leaks
	// This must happen before any other initialization that might print
	ctx.initOutputBuffer()
	ctx.AddCleanup(func() { ctx.stopOutputBuffer() })

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
	// This loads tool-config.yml and integrates YAML-defined tools with native handlers
	toolStart := time.Now()
	configRoot := filepath.Join(workspaceRoot, ".eac")
	if err := tool.InitializeGlobalBridges(workspaceRoot, configRoot); err != nil {
		log.Debugf("Tool bridge initialization skipped: %v", err)
		// Continue - tool config is optional, native handlers will still work
	}
	ctx.initTimings.ToolInit = time.Since(toolStart)

	// Create output directory
	outputDir := filepath.Join(workspaceRoot, ctx.Config.OutputDir)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine max concurrency using shared calculation
	// 0 = dynamic (orchestrator calculates from CPU×RAM×turbo)
	repoConcurrency := ctx.RepoConfig.EffectiveParallelism(environments.IsCI())
	maxConcurrency := CalculateMaxConcurrency(ctx.Config.MaxConcurrency, repoConcurrency, ctx.Config.Turbo, ctx.Config.Sequential)
	turboMultiplier := CalculateTurboMultiplier(ctx.Config.Turbo)

	// Log turbo mode if enabled
	if ctx.Config.Turbo && !ctx.Config.Sequential {
		log.Debugf("Turbo mode enabled: %.2fx pressure multiplier", turboMultiplier)
	}

	// Configure orchestrator
	// If using registry, set TUI: false so orchestrator doesn't create its own console
	orchConfig := orchestrator.Config{
		WorkspaceRoot:        workspaceRoot,
		OutputBaseDir:        ctx.Config.OutputDir,
		LogFileName:          ctx.Config.LogFileName,
		ActionVerb:           ctx.Config.ActionVerb,
		MaxConcurrency:       maxConcurrency,
		Turbo:                turboMultiplier,
		StatusUpdateInterval: 500, // 500ms for responsive feedback
		ComponentTypesDisplay: ctx.ComponentTypesDisplay,
		ShowTimings:          ctx.Config.ShowTimings,
		DryRun:               ctx.Config.DryRun,
		TUI:                  ctx.Config.UseTUI, // Always set TUI flag for output suppression (console may be created by registry)
		TUIHeight:            ctx.Config.TUIHeight,
		TUIASCIIMode:         ctx.Config.TUIASCIIMode,
		TUI3Demo:             ctx.Config.TUI3Demo,
		SkipTUIDelay:         ctx.Config.SkipTUIDelay,
		Layered:              ctx.Config.Layered, // Controls LayerMode: true=Strict, false=None
	}

	// Create orchestrator
	orch := orchestrator.New(&orchConfig, nil) // Worker set later
	ctx.Orchestrator = orch
	ctx.AddCleanup(func() { orch.Close() })

	// If TUI console was created earlier, set it on orchestrator
	if tuiConsole != nil {
		// If it's a parallel.Console, extract the inner tui.Console for the orchestrator
		if pc, ok := tuiConsole.(*parallel.Console); ok {
			orch.SetConsole(pc.Inner())

			// Create TUI observer to receive execution events
			tuiObserver := tui.NewTUIObserver(pc.Inner())
			orch.AddObserver(tuiObserver)
			orch.SetWriterFactory(tuiObserver)

			// Create TUI hooks for controlled interaction
			ctx.TUIHooks = tui.NewTUIHooks(pc.Inner())
		}
		// Note: If it's a different TUI type (e.g., default TUI), the orchestrator
		// won't have a console set and will run in non-TUI mode. This is expected
		// for interactive selection TUIs that don't need parallel task display.
	} else if !ctx.Config.UseTUI {
		// No TUI - use console observer for plain text output
		consoleObserver := output.NewConsoleObserver()
		orch.AddObserver(consoleObserver)
	}

	// Initialize TUI (but don't show yet - that happens after init phases complete)
	if ctx.Config.UseTUI {
		if err := orch.Init(); err != nil {
			return fmt.Errorf("failed to initialize orchestrator: %w", err)
		}
	}

	// Configure logging
	// Debug always goes to file, also to console if --debug flag set
	logName := string(ctx.Config.Type)
	if err := logging.ConfigureLoggingSimple(workspaceRoot, logName, nil, ctx.Config.DebugMode); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	ctx.AddCleanup(func() { logging.CloseLogging() })

	log.Debugf("%s logging configured: debugMode=%v, useTUI=%v",
		ctx.Config.ActionVerb, ctx.Config.DebugMode, ctx.Config.UseTUI)

	return nil
}

// getTUIHeight returns the TUI height, using default if not specified.
func getTUIHeight(cfg *CommandConfig) int {
	if cfg.TUIHeight > 0 {
		return cfg.TUIHeight
	}
	return tui.DefaultHeight
}
