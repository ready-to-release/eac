// Command: work commit
// Short: Commit changes with AI-generated commit messages
// Long: Commits changes in the current workspace using AI to generate semantic commit messages.
// Long:
// Long: By default, commits only staged changes. Use --all to stage all changes before committing.
// Long: Uses the commit command internally to generate high-quality commit messages that follow
// Long: project conventions and include module-specific details.
// Long:
// Long: Example:
// Long:   work commit
// Long:   work commit --all
// Long:   work commit --message "fix: resolve authentication bug"
// Long:   work commit -m "feat: add new feature"
// Flag.all: type=bool, shorthand=a, default=false, usage=Stage all changes before committing
// Flag.message: type=string, shorthand=m, usage=Custom commit message (skips AI generation)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode (pass through to commit)
// HasSideEffects: true
package work

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ready-to-release/eac/src/commands/impl/commit"
	"github.com/ready-to-release/eac/src/commands/impl/work/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/repository"
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 2: Validate environment
	if err := internal.EnsureInGitRepo(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 3: Stage changes if --all
	if config.stageAll {
		if err := stageAllChanges(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	}

	// Phase 4: Check for staged changes
	hasStagedChanges, err := checkStagedChanges()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if !hasStagedChanges {
		fmt.Fprintf(os.Stderr, "Error: No staged changes\n")
		fmt.Fprintf(os.Stderr, "Use 'work commit --all' to stage and commit all changes\n")
		return 1
	}

	// Phase 5: Commit
	if config.customMessage != "" {
		// Use custom message
		return commitWithMessage(config.customMessage)
	}

	// Use AI to generate message
	return commitWithAI(config.debug)
}

// commitConfig holds configuration for the commit command
type commitConfig struct {
	stageAll      bool
	customMessage string
	debug         bool
	repoRoot      string
}

// parseCommitConfig parses command line arguments
func parseCommitConfig() (*commitConfig, error) {
	args := os.Args[3:] // Skip program name, "work", "commit"

	config := &commitConfig{}

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--all", "-a":
			config.stageAll = true
		case "--debug", "-d":
			config.debug = true
		case "--message", "-m":
			if i+1 < len(args) {
				config.customMessage = args[i+1]
				i++ // Skip next arg
			} else {
				return nil, fmt.Errorf("--message requires a value")
			}
		}
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}
	config.repoRoot = repoRoot

	return config, nil
}

// stageAllChanges stages all changes in the working directory
func stageAllChanges() error {
	cmd := exec.Command("git", "add", ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stage changes: %w\nOutput: %s", err, string(output))
	}
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
func commitWithMessage(message string) int {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create commit: %v\n", err)
		return 1
	}
	return 0
}

// commitWithAI calls commit to generate and create commit
func commitWithAI(debug bool) int {
	// Prepare args for commit
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args to simulate: r2r commit [--debug]
	if debug {
		os.Args = []string{"r2r", "commit", "--debug"}
	} else {
		os.Args = []string{"r2r", "commit"}
	}

	// Call commit
	return commit.CommitAI()
}
