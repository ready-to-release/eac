package cmdframework

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/environments"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"github.com/ready-to-release/eac/go/eac/core/tool"
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
	orchConfig := orchestrator.Config{
		WorkspaceRoot:        workspaceRoot,
		OutputBaseDir:        ctx.Config.OutputDir,
		LogFileName:          ctx.Config.LogFileName,
		ActionVerb:           ctx.Config.ActionVerb,
		MaxConcurrency:       maxConcurrency,
		Turbo:                turboMultiplier,
		StatusUpdateInterval: 500, // 500ms for responsive feedback
		ModuleTypes:          ctx.ModuleTypes,
		ShowTimings:          ctx.Config.ShowTimings,
		DryRun:               ctx.Config.DryRun,
		TUI:                  ctx.Config.UseTUI,
		TUIHeight:            ctx.Config.TUIHeight,
		TUIASCIIMode:         ctx.Config.TUIASCIIMode,
	}

	// Create orchestrator
	orch := orchestrator.New(&orchConfig, nil) // Worker set later
	ctx.Orchestrator = orch
	ctx.AddCleanup(func() { orch.Close() })

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
