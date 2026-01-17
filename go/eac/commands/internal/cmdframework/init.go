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
)

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
	eacCfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return fmt.Errorf("failed to load EAC config: %w", err)
	}
	ctx.EACConfig = eacCfg
	ctx.RepoConfig = eacCfg.Repository

	// Create output directory
	outputDir := filepath.Join(workspaceRoot, ctx.Config.OutputDir)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine max concurrency
	maxConcurrency := ctx.Config.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = ctx.RepoConfig.EffectiveParallelism(environments.IsCI())
	}
	if ctx.Config.Sequential {
		maxConcurrency = 1
	}

	// Configure orchestrator
	orchConfig := orchestrator.Config{
		WorkspaceRoot:        workspaceRoot,
		OutputBaseDir:        ctx.Config.OutputDir,
		LogFileName:          ctx.Config.LogFileName,
		ActionVerb:           ctx.Config.ActionVerb,
		MaxConcurrency:       maxConcurrency,
		StatusUpdateInterval: 500, // 500ms for responsive feedback
		ModuleTypes:          ctx.ModuleTypes,
		ShowTimings:          ctx.Config.ShowTimings,
		DryRun:               ctx.Config.DryRun,
		TUI:                  ctx.Config.UseTUI,
		TUIHeight:            ctx.Config.TUIHeight,
	}

	// Create orchestrator
	orch := orchestrator.New(orchConfig, nil) // Worker set later
	ctx.Orchestrator = orch
	ctx.AddCleanup(func() { orch.Close() })

	// Initialize and start TUI if enabled
	if ctx.Config.UseTUI {
		if err := orch.Init(); err != nil {
			return fmt.Errorf("failed to initialize orchestrator: %w", err)
		}
		orch.StartTUI()

		// Get TUI writer for init phase output
		ctx.tuiWriter = orch.GetTUIWriter(tui.PhaseInit)

		// Brief pause to ensure TUI is ready
		time.Sleep(10 * time.Millisecond)
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
