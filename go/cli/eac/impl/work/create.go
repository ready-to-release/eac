// Command: work create
// Short: Create a new workspace for parallel development
// Long: Creates a new git worktree in a sibling directory for parallel development with Claude.
// Long:
// Long: The workspace is created in a sibling directory with the naming pattern:
// Long:   <repo-name>-<branch-name>
// Long:
// Long: This allows you to work on multiple features simultaneously with separate Claude Code sessions.
// Long: Use --debug to enable detailed logging to out/logs/work/.
// Long:
// Long: Expected Output:
// Long:   - New git worktree in sibling directory
// Long:   - Ready for parallel Claude Code session
// Long:
// Long: Example:
// Long:   work create feature/authentication
// Long:   work create bugfix/issue-123 --from=develop
// Long:   work create feature/api --path=../custom-path
// Long:   work create feature/test --debug
// Flag.from: type=string, default=main, usage=Base branch to create from
// Flag.path: type=string, usage=Custom path for workspace (default: ../<repo>-<branch>)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
package work

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/ready-to-release/eac/go/cli/eac/impl/work/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
)

// Create creates a new workspace (git worktree) for parallel development.
func Create() int {
	commandStart := time.Now()

	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Phase 1: Parse configuration
	phase1Start := time.Now()
	config, err := parseCreateConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	defer func() { _ = config.base.Logger.Sync() }() //nolint:errcheck // best-effort logger sync

	config.base.Logger.Debug("Phase 1: Starting configuration parsing",
		zap.String("phase", "phase1"),
		zap.String("description", "parse configuration"))

	config.base.Logger.Debug("Phase 1: Completed",
		zap.String("phase", "phase1"),
		zap.Duration("duration", time.Since(phase1Start)),
		zap.String("branchName", config.branchName),
		zap.String("baseBranch", config.baseBranch),
		zap.String("worktreePath", config.worktreePath),
		zap.Bool("debug", config.base.Debug))

	// Phase 2: Validate environment
	phase2Start := time.Now()
	config.base.Logger.Debug("Phase 2: Starting environment validation",
		zap.String("phase", "phase2"),
		zap.String("description", "validate environment"),
		zap.String("branchName", config.branchName),
		zap.String("baseBranch", config.baseBranch))

	if err := validateCreateEnvironment(config); err != nil {
		config.base.Logger.Debug("Phase 2: Failed",
			zap.String("phase", "phase2"),
			zap.Duration("duration", time.Since(phase2Start)),
			zap.Error(err))
		config.base.Logger.Error(fmt.Sprintf("Validation failed: %v", err))
		return 1
	}

	config.base.Logger.Debug("Phase 2: Completed",
		zap.String("phase", "phase2"),
		zap.Duration("duration", time.Since(phase2Start)))

	// Phase 3: Create worktree
	phase3Start := time.Now()
	config.base.Logger.Debug("Phase 3: Starting worktree creation",
		zap.String("phase", "phase3"),
		zap.String("description", "create worktree"),
		zap.String("worktreePath", config.worktreePath),
		zap.String("branchName", config.branchName),
		zap.String("baseBranch", config.baseBranch))

	worktreePath, err := createWorktree(config)
	if err != nil {
		config.base.Logger.Debug("Phase 3: Failed",
			zap.String("phase", "phase3"),
			zap.Duration("duration", time.Since(phase3Start)),
			zap.Error(err))
		config.base.Logger.Error(fmt.Sprintf("Failed to create worktree: %v", err))
		return 1
	}

	config.base.Logger.Debug("Phase 3: Completed",
		zap.String("phase", "phase3"),
		zap.Duration("duration", time.Since(phase3Start)),
		zap.String("worktreePath", worktreePath))

	// Phase 4: Output success
	phase4Start := time.Now()
	config.base.Logger.Debug("Phase 4: Starting success output",
		zap.String("phase", "phase4"),
		zap.String("description", "output success"))

	outputCreateSuccess(worktreePath, config.branchName)

	config.base.Logger.Debug("Phase 4: Completed",
		zap.String("phase", "phase4"),
		zap.Duration("duration", time.Since(phase4Start)))

	config.base.Logger.Debug("Work create command completed successfully",
		zap.Duration("totalDuration", time.Since(commandStart)))
	return 0
}

// createConfig holds configuration for the create command.
type createConfig struct {
	base         *internal.BaseConfig
	branchName   string
	baseBranch   string
	customPath   string
	repoName     string
	worktreePath string
}

// parseCreateConfig parses command line arguments.
func parseCreateConfig() (*createConfig, error) {
	args := os.Args[3:] // Skip program name, "work", "create"

	// Parse base config (debug flag, repo root, logger, git ops)
	baseConfig, err := internal.ParseBaseConfig(args)
	if err != nil {
		return nil, err
	}

	// Get positional arguments (non-flags)
	positionalArgs := flags.GetPositionalArgs(args)
	if len(positionalArgs) == 0 {
		return nil, fmt.Errorf("branch name is required\nUsage: work create <branch-name> [--from=main] [--path=<path>] [--debug]")
	}

	config := &createConfig{
		base:       baseConfig,
		branchName: positionalArgs[0],
		baseBranch: "main",
	}

	// Parse custom flags
	if fromValue := flags.GetFlagValue(args, "--from"); fromValue != "" {
		config.baseBranch = fromValue
	}
	if pathValue := flags.GetFlagValue(args, "--path"); pathValue != "" {
		config.customPath = pathValue
	}

	config.repoName = internal.GetRepoName(config.base.RepoRoot)

	// Determine worktree path
	if config.customPath != "" {
		absPath, err := internal.GetAbsolutePath(config.customPath)
		if err != nil {
			return nil, fmt.Errorf("invalid path: %w", err)
		}
		config.worktreePath = absPath
	} else {
		config.worktreePath = internal.GenerateWorktreePath(config.repoName, config.branchName)
	}

	return config, nil
}

// validateCreateEnvironment validates the environment before creating worktree.
func validateCreateEnvironment(config *createConfig) error {
	logger := config.base.Logger

	// Sub-phase 2.1: Check git repository
	subPhase1Start := time.Now()
	logger.Debug("Phase 2.1: Starting git repository check",
		zap.String("phase", "phase2.1"),
		zap.String("description", "check git repository"))

	if err := internal.EnsureInGitRepo(); err != nil {
		logger.Debug("Phase 2.1: Failed",
			zap.String("phase", "phase2.1"),
			zap.Duration("duration", time.Since(subPhase1Start)),
			zap.Error(err))
		return err
	}

	logger.Debug("Phase 2.1: Completed",
		zap.String("phase", "phase2.1"),
		zap.Duration("duration", time.Since(subPhase1Start)))

	// Sub-phase 2.2: Check if branch already exists
	subPhase2Start := time.Now()
	logger.Debug("Phase 2.2: Starting branch existence check",
		zap.String("phase", "phase2.2"),
		zap.String("description", "check branch existence"),
		zap.String("branchName", config.branchName))

	exists, err := config.base.GitOps.BranchExists(config.branchName)
	if err != nil {
		logger.Debug("Phase 2.2: Failed",
			zap.String("phase", "phase2.2"),
			zap.Duration("duration", time.Since(subPhase2Start)),
			zap.Error(err))
		return fmt.Errorf("failed to check branch: %w", err)
	}
	if exists {
		logger.Debug("Phase 2.2: Failed",
			zap.String("phase", "phase2.2"),
			zap.Duration("duration", time.Since(subPhase2Start)),
			zap.String("reason", "branch already exists"))
		return fmt.Errorf("branch '%s' already exists", config.branchName)
	}

	logger.Debug("Phase 2.2: Completed",
		zap.String("phase", "phase2.2"),
		zap.Duration("duration", time.Since(subPhase2Start)),
		zap.Bool("exists", false))

	// Sub-phase 2.3: Check if worktree already exists for this branch
	subPhase3Start := time.Now()
	logger.Debug("Phase 2.3: Starting worktree existence check",
		zap.String("phase", "phase2.3"),
		zap.String("description", "check worktree existence"),
		zap.String("branchName", config.branchName))

	wtExists, err := config.base.GitOps.WorktreeExists(config.branchName)
	if err != nil {
		logger.Debug("Phase 2.3: Failed",
			zap.String("phase", "phase2.3"),
			zap.Duration("duration", time.Since(subPhase3Start)),
			zap.Error(err))
		return fmt.Errorf("failed to check worktree: %w", err)
	}
	if wtExists {
		logger.Debug("Phase 2.3: Failed",
			zap.String("phase", "phase2.3"),
			zap.Duration("duration", time.Since(subPhase3Start)),
			zap.String("reason", "worktree already exists"))
		return fmt.Errorf("worktree already exists for branch '%s'", config.branchName)
	}

	logger.Debug("Phase 2.3: Completed",
		zap.String("phase", "phase2.3"),
		zap.Duration("duration", time.Since(subPhase3Start)),
		zap.Bool("exists", false))

	// Sub-phase 2.4: Check if base branch exists
	subPhase4Start := time.Now()
	logger.Debug("Phase 2.4: Starting base branch existence check",
		zap.String("phase", "phase2.4"),
		zap.String("description", "check base branch existence"),
		zap.String("baseBranch", config.baseBranch))

	baseExists, err := config.base.GitOps.BranchExists(config.baseBranch)
	if err != nil {
		logger.Debug("Phase 2.4: Failed",
			zap.String("phase", "phase2.4"),
			zap.Duration("duration", time.Since(subPhase4Start)),
			zap.Error(err))
		return fmt.Errorf("failed to check base branch: %w", err)
	}
	if !baseExists {
		logger.Debug("Phase 2.4: Failed",
			zap.String("phase", "phase2.4"),
			zap.Duration("duration", time.Since(subPhase4Start)),
			zap.String("reason", "base branch does not exist"))
		return fmt.Errorf("base branch '%s' does not exist", config.baseBranch)
	}

	logger.Debug("Phase 2.4: Completed",
		zap.String("phase", "phase2.4"),
		zap.Duration("duration", time.Since(subPhase4Start)),
		zap.Bool("exists", true))

	return nil
}

// createWorktree creates the actual git worktree.
func createWorktree(config *createConfig) (string, error) {
	logger := config.base.Logger

	// Sub-phase 3.1: Resolve absolute path
	subPhase1Start := time.Now()
	logger.Debug("Phase 3.1: Starting path resolution",
		zap.String("phase", "phase3.1"),
		zap.String("description", "resolve absolute path"),
		zap.String("worktreePath", config.worktreePath))

	absWorktreePath, err := filepath.Abs(config.worktreePath)
	if err != nil {
		logger.Debug("Phase 3.1: Failed",
			zap.String("phase", "phase3.1"),
			zap.Duration("duration", time.Since(subPhase1Start)),
			zap.Error(err))
		return "", fmt.Errorf("failed to resolve worktree path: %w", err)
	}

	logger.Debug("Phase 3.1: Completed",
		zap.String("phase", "phase3.1"),
		zap.Duration("duration", time.Since(subPhase1Start)),
		zap.String("absWorktreePath", absWorktreePath))

	// Sub-phase 3.2: Create worktree using GitOps
	subPhase2Start := time.Now()
	logger.Debug("Phase 3.2: Starting git worktree creation",
		zap.String("phase", "phase3.2"),
		zap.String("description", "create git worktree"),
		zap.String("worktreePath", config.worktreePath),
		zap.String("branchName", config.branchName),
		zap.String("baseBranch", config.baseBranch))

	if err := config.base.GitOps.CreateWorktree(config.worktreePath, config.branchName, config.baseBranch); err != nil {
		logger.Debug("Phase 3.2: Failed",
			zap.String("phase", "phase3.2"),
			zap.Duration("duration", time.Since(subPhase2Start)),
			zap.Error(err))
		return "", err
	}

	logger.Debug("Phase 3.2: Completed",
		zap.String("phase", "phase3.2"),
		zap.Duration("duration", time.Since(subPhase2Start)),
		zap.String("absWorktreePath", absWorktreePath),
		zap.String("branchName", config.branchName))

	return absWorktreePath, nil
}

// outputCreateSuccess outputs success message with next steps.
func outputCreateSuccess(worktreePath, branchName string) {
	log.Infof("✓ Created worktree at: %s", worktreePath)
	log.Infof("  Branch: %s", branchName)
	log.Info("")
	log.Info("Start Claude:")

	// Convert to forward slashes for cross-platform compatibility
	displayPath := filepath.ToSlash(worktreePath)
	log.Infof("  cd %s && claude-code", displayPath)
}
