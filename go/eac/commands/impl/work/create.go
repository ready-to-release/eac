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

	"go.uber.org/zap"

	"github.com/ready-to-release/eac/go/eac/commands/impl/work/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

func init() {
	registry.Register(Create)
}

// Intent: Create a new workspace (git worktree) for parallel development
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear sequential flow: validate → check branch → create worktree → output
//   - Descriptive variable names and error messages
//   - Single responsibility functions
//
// Easy to change:
//   - Path generation isolated in internal package
//   - Branch validation separated from worktree creation
//   - Configuration parsing in dedicated function
//
// Hard to break:
//   - Validates git repo before proceeding
//   - Checks for existing branches and worktrees
//   - Clear error messages with actionable guidance
//   - Tests cover success and error cases

// Create creates a new workspace (git worktree) for parallel development
func Create() int {
	// Phase 1: Parse configuration
	config, err := parseCreateConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	defer config.base.Logger.Sync()

	config.base.Logger.Debug("Starting work create command",
		zap.String("branchName", config.branchName),
		zap.String("baseBranch", config.baseBranch),
		zap.String("worktreePath", config.worktreePath),
		zap.Bool("debug", config.base.Debug))

	// Phase 2: Validate environment
	if err := validateCreateEnvironment(config); err != nil {
		config.base.Logger.Error(fmt.Sprintf("Validation failed: %v", err))
		return 1
	}

	// Phase 3: Create worktree
	worktreePath, err := createWorktree(config)
	if err != nil {
		config.base.Logger.Error(fmt.Sprintf("Failed to create worktree: %v", err))
		return 1
	}

	// Phase 4: Output success
	outputCreateSuccess(config.base.Logger, worktreePath, config.branchName)
	config.base.Logger.Debug("Work create command completed successfully")
	return 0
}

// createConfig holds configuration for the create command
type createConfig struct {
	base         *internal.BaseConfig
	branchName   string
	baseBranch   string
	customPath   string
	repoName     string
	worktreePath string
}

// parseCreateConfig parses command line arguments
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

// validateCreateEnvironment validates the environment before creating worktree
func validateCreateEnvironment(config *createConfig) error {
	// Check we're in a git repository
	if err := internal.EnsureInGitRepo(); err != nil {
		return err
	}

	// Check if branch already exists
	exists, err := config.base.GitOps.BranchExists(config.branchName)
	if err != nil {
		return fmt.Errorf("failed to check branch: %w", err)
	}
	if exists {
		return fmt.Errorf("branch '%s' already exists", config.branchName)
	}

	// Check if worktree already exists for this branch
	wtExists, err := config.base.GitOps.WorktreeExists(config.branchName)
	if err != nil {
		return fmt.Errorf("failed to check worktree: %w", err)
	}
	if wtExists {
		return fmt.Errorf("worktree already exists for branch '%s'", config.branchName)
	}

	// Check if base branch exists
	baseExists, err := config.base.GitOps.BranchExists(config.baseBranch)
	if err != nil {
		return fmt.Errorf("failed to check base branch: %w", err)
	}
	if !baseExists {
		return fmt.Errorf("base branch '%s' does not exist", config.baseBranch)
	}

	return nil
}

// createWorktree creates the actual git worktree
func createWorktree(config *createConfig) (string, error) {
	// Get absolute path for worktree
	absWorktreePath, err := filepath.Abs(config.worktreePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve worktree path: %w", err)
	}

	// Create worktree using GitOps
	if err := config.base.GitOps.CreateWorktree(config.worktreePath, config.branchName, config.baseBranch); err != nil {
		return "", err
	}

	config.base.Logger.Debug(fmt.Sprintf("Created worktree at %s for branch %s", absWorktreePath, config.branchName))
	return absWorktreePath, nil
}

// outputCreateSuccess outputs success message with next steps
func outputCreateSuccess(logger *logging.Logger, worktreePath, branchName string) {
	logger.Info(fmt.Sprintf("✓ Created worktree at: %s", worktreePath))
	logger.Info(fmt.Sprintf("  Branch: %s", branchName))
	logger.Info("")
	logger.Info("Start Claude:")

	// Convert to forward slashes for cross-platform compatibility
	displayPath := filepath.ToSlash(worktreePath)
	logger.Info(fmt.Sprintf("  cd %s && claude-code", displayPath))
}
