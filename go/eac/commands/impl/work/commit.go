// Command: work commit
// Short: Commit changes with AI-generated commit messages
// Long: Commits changes in the current workspace using AI to generate semantic commit messages.
// Long:
// Long: By default, commits only staged changes. Use --all to stage all changes before committing.
// Long: Uses the commit command internally to generate high-quality commit messages that follow
// Long: project conventions and include module-specific details.
// Long:
// Long: Expected Output:
// Long:   - Git commit created with AI-generated or custom message
// Long:
// Long: Example:
// Long:   work commit
// Long:   work commit --all
// Long:   work commit --message "fix: resolve authentication bug"
// Long:   work commit -m "feat: add new feature"
// Flag.all: type=bool, shorthand=a, default=false, usage=Stage all changes before committing
// Flag.message: type=string, shorthand=m, usage=Custom commit message (skips AI generation)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode (pass through to commit)
package work

import (
	"fmt"
	"os"
	"os/exec"

	"go.uber.org/zap"

	commitmessage "github.com/ready-to-release/eac/go/eac/commands/impl/create/commit-message"
	"github.com/ready-to-release/eac/go/eac/commands/impl/work/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

func init() {
	registry.Register(Commit)
}

// Intent: Commit changes with AI-generated messages in a workspace
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear flow: validate → stage (if --all) → generate message → commit
//   - Reuses existing commit for message generation
//   - Flags mirror git commit conventions (--all, --message)
//
// Easy to change:
//   - Staging logic isolated
//   - Message generation delegated to commit
//   - Custom message path separate from AI path
//
// Hard to break:
//   - Validates git repo and changes before proceeding
//   - Clear error messages for each failure case
//   - Tests verify all code paths

// Commit commits changes with AI-generated or custom message
func Commit() int {
	// Phase 1: Parse configuration
	config, err := parseCommitConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	defer config.base.Logger.Sync()

	config.base.Logger.Debug("Starting work commit command",
		zap.Bool("stageAll", config.stageAll),
		zap.String("customMessage", config.customMessage),
		zap.Bool("debug", config.base.Debug))

	// Phase 2: Validate environment
	if err := internal.EnsureInGitRepo(); err != nil {
		config.base.Logger.Error(fmt.Sprintf("Not in git repository: %v", err))
		return 1
	}

	// Phase 3: Stage changes if --all
	if config.stageAll {
		if err := stageAllChanges(config.base.Logger); err != nil {
			config.base.Logger.Error(fmt.Sprintf("Failed to stage changes: %v", err))
			return 1
		}
	}

	// Phase 4: Check for staged changes
	hasStagedChanges, err := checkStagedChanges()
	if err != nil {
		config.base.Logger.Error(fmt.Sprintf("Failed to check staged changes: %v", err))
		return 1
	}
	if !hasStagedChanges {
		config.base.Logger.Error("No staged changes")
		config.base.Logger.Error("Use 'work commit --all' to stage and commit all changes")
		return 1
	}

	// Phase 5: Commit
	if config.customMessage != "" {
		// Use custom message
		return commitWithMessage(config.base.Logger, config.customMessage)
	}

	// Use AI to generate message
	config.base.Logger.Debug("Delegating to commit message for AI generation")
	return commitWithAI(config.base.Debug)
}

// commitConfig holds configuration for the commit command
type commitConfig struct {
	base          *internal.BaseConfig
	stageAll      bool
	customMessage string
}

// parseCommitConfig parses command line arguments
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

// stageAllChanges stages all changes in the working directory
func stageAllChanges(logger *logging.Logger) error {
	cmd := exec.Command("git", "add", ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stage changes: %w\nOutput: %s", err, string(output))
	}
	logger.Debug("Staged all changes")
	return nil
}

// checkStagedChanges checks if there are any staged changes
func checkStagedChanges() (bool, error) {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 1 means there are differences (staged changes exist)
			if exitErr.ExitCode() == 1 {
				return true, nil
			}
		}
		return false, fmt.Errorf("failed to check staged changes: %w", err)
	}
	// Exit code 0 means no differences (no staged changes)
	return false, nil
}

// commitWithMessage creates a commit with a custom message
func commitWithMessage(logger *logging.Logger, message string) int {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create commit: %v", err))
		return 1
	}
	logger.Debug("Created commit with custom message")
	return 0
}

// commitWithAI calls commit message to generate commit message
func commitWithAI(debug bool) int {
	// Prepare args for commit message
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args to simulate: r2r commit message [--debug]
	if debug {
		os.Args = []string{"r2r", "commit", "message", "--debug"}
	} else {
		os.Args = []string{"r2r", "commit", "message"}
	}

	// Call commit message
	return commitmessage.CreateCommitMessage()
}
