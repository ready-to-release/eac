package work

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	commitmessage "github.com/ready-to-release/eac/go/commands/repository/get/commit-message"
	"github.com/ready-to-release/eac/go/commands/repository/work/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/tool"
)

type workCommitCommand struct{}

var _ core.SimpleCommandPort = (*workCommitCommand)(nil)

func (c *workCommitCommand) Name() string { return "work commit" }

func (c *workCommitCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "work-commit",
		Short:         "Commit changes with AI-generated commit messages",
		Long: "Commits changes in the current workspace using AI to generate semantic commit messages.\n\nBy default, commits only staged changes. Use --all to stage all changes before committing.\nUses the commit command internally to generate high-quality commit messages that follow\nproject conventions and include module-specific details.",
		Notes: "Expected Output:\n  - Git commit created with AI-generated or custom message",
		Examples: []string{
			"eac work commit",
			"eac work commit --all",
			"eac work commit --message \"fix: resolve authentication bug\"",
			"eac work commit -m \"feat: add new feature\"",
		},
		Flags: []core.FlagSpec{
			{Name: "all", Type: "bool", Shorthand: "a", DefaultValue: "false", Usage: "Stage all changes before committing"},
			{Name: "message", Type: "string", Shorthand: "m", Usage: "Custom commit message (skips AI generation)"},
			{Name: "debug", Type: "bool", Shorthand: "d", DefaultValue: "false", Usage: "Enable debug mode (pass through to commit)"},
		},
	}
}

func (c *workCommitCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return Commit()
}

// Commit commits changes with AI-generated or custom message.
func Commit() int {
	startTime := time.Now()

	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Phase 1: Parse configuration
	config, err := parseCommitConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	defer func() { _ = config.base.Logger.Sync() }() //nolint:errcheck // best-effort sync

	logger := config.base.Logger
	logger.Debug("Phase 1: Starting parse configuration",
		zap.String("phase", "phase1"))
	logger.Debug("Phase 1: Completed",
		zap.String("phase", "phase1"),
		zap.Bool("stageAll", config.stageAll),
		zap.String("customMessage", config.customMessage),
		zap.Bool("debug", config.base.Debug))

	// Phase 2: Validate environment
	logger.Debug("Phase 2: Starting validate environment",
		zap.String("phase", "phase2"))
	if err := internal.EnsureInGitRepo(); err != nil {
		logger.Debug("Phase 2: Failed",
			zap.String("phase", "phase2"),
			zap.Error(err))
		logger.Error("Not in git repository", zap.Error(err))
		return 1
	}
	logger.Debug("Phase 2: Completed",
		zap.String("phase", "phase2"))

	// Phase 3: Stage changes if --all
	if config.stageAll {
		logger.Debug("Phase 3: Starting stage changes",
			zap.String("phase", "phase3"))
		if err := stageAllChanges(); err != nil {
			logger.Debug("Phase 3: Failed",
				zap.String("phase", "phase3"),
				zap.Error(err))
			logger.Error("Failed to stage changes", zap.Error(err))
			return 1
		}
		logger.Debug("Phase 3: Completed",
			zap.String("phase", "phase3"))
	} else {
		logger.Debug("Phase 3: Skipped (--all not specified)",
			zap.String("phase", "phase3"))
	}

	// Phase 4: Check for staged changes
	logger.Debug("Phase 4: Starting check for staged changes",
		zap.String("phase", "phase4"))
	hasStagedChanges, err := checkStagedChanges()
	if err != nil {
		logger.Debug("Phase 4: Failed",
			zap.String("phase", "phase4"),
			zap.Error(err))
		logger.Error("Failed to check staged changes", zap.Error(err))
		return 1
	}
	if !hasStagedChanges {
		logger.Debug("Phase 4: Failed",
			zap.String("phase", "phase4"),
			zap.Error(fmt.Errorf("no staged changes")))
		logger.Error("No staged changes")
		logger.Error("Use 'work commit --all' to stage and commit all changes")
		return 1
	}
	logger.Debug("Phase 4: Completed",
		zap.String("phase", "phase4"),
		zap.Bool("hasStagedChanges", hasStagedChanges))

	// Phase 5: Commit
	logger.Debug("Phase 5: Starting commit",
		zap.String("phase", "phase5"),
		zap.Bool("useCustomMessage", config.customMessage != ""))

	var exitCode int
	if config.customMessage != "" {
		// Use custom message
		exitCode = commitWithMessage(config.customMessage)
	} else {
		// Use AI to generate message
		logger.Debug("Delegating to commit message for AI generation")
		exitCode = commitWithAI(config.base.Debug)
	}

	if exitCode != 0 {
		logger.Debug("Phase 5: Failed",
			zap.String("phase", "phase5"),
			zap.Int("exitCode", exitCode))
		return exitCode
	}

	duration := time.Since(startTime)
	logger.Debug("Phase 5: Completed",
		zap.String("phase", "phase5"),
		zap.Duration("totalDuration", duration))

	return 0
}

// commitConfig holds configuration for the commit command.
type commitConfig struct {
	base          *internal.BaseConfig
	stageAll      bool
	customMessage string
}

// parseCommitConfig parses command line arguments.
func parseCommitConfig() (*commitConfig, error) {
	args := os.Args[3:] // Skip program name, "work", "commit"

	// Parse base config (debug flag, repo root, logger, git ops)
	baseConfig, err := internal.ParseBaseConfig(args)
	if err != nil {
		return nil, err
	}

	config := &commitConfig{
		base: baseConfig,
	}

	// Parse flags
	config.stageAll = flags.HasFlag(args, "--all", "-a")

	// Parse --message/-m flag
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--message" || arg == "-m" {
			if i+1 < len(args) {
				config.customMessage = args[i+1]
				break
			} else {
				return nil, fmt.Errorf("--message requires a value")
			}
		}
	}

	return config, nil
}

// stageAllChanges stages all changes in the working directory.
func stageAllChanges() error {
	ts := tool.GlobalToolSystem()
	if ts == nil {
		return fmt.Errorf("tool system not initialized")
	}
	_, err := ts.RunTool(context.Background(), "git", ".", "add", ".")
	if err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}
	log.Debugf("Staged all changes")
	return nil
}

// checkStagedChanges checks if there are any staged changes.
func checkStagedChanges() (bool, error) {
	ts := tool.GlobalToolSystem()
	if ts == nil {
		return false, fmt.Errorf("tool system not initialized")
	}
	_, exitCode, err := ts.RunToolCombined(context.Background(), "git", ".", "diff", "--cached", "--quiet")
	if err != nil {
		return false, fmt.Errorf("failed to check staged changes: %w", err)
	}
	// Exit code 1 means there are differences (staged changes exist)
	// Exit code 0 means no differences (no staged changes)
	return exitCode == 1, nil
}

// commitWithMessage creates a commit with a custom message.
func commitWithMessage(message string) int {
	ts := tool.GlobalToolSystem()
	if ts == nil {
		log.Errorf("Failed to create commit: tool system not initialized")
		return 1
	}
	_, err := ts.RunTool(context.Background(), "git", ".", "commit", "-m", message)
	if err != nil {
		log.Errorf("Failed to create commit: error=%v", err)
		return 1
	}
	log.Debugf("Created commit with custom message")
	return 0
}

// commitWithAI calls commit message to generate commit message.
func commitWithAI(debug bool) int {
	// Prepare args for commit message
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args to simulate: clie commit message [--debug]
	if debug {
		os.Args = []string{"clie", "commit", "message", "--debug"}
	} else {
		os.Args = []string{"clie", "commit", "message"}
	}

	// Call commit message
	return commitmessage.GetCommitMessage()
}
