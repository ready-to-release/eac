// Command: create pr
// Short: Create pull request with AI-generated description
// Long: Creates a pull request for the current workspace branch with an AI-generated
// Long: title and description based on all commits in the branch.
// Long:
// Long: This command:
// Long:   1. Validates the workspace is ready for PR
// Long:   2. Pushes the branch to origin if needed
// Long:   3. Analyzes all commits to generate PR title and description
// Long:   4. Creates the pull request using GitHub CLI
// Long:
// Long: Requires GitHub CLI (gh) to be installed and authenticated.
// Long:
// Long: Expected Output:
// Long: - PR created on GitHub
// Long: - PR URL printed to stdout
// Long: - AI-generated title and description based on commits
// Long:
// Long: Example:
// Long:   create pr
// Long:   create pr --target=develop
// Long:   create pr --title "Add authentication feature"
// Long:   create pr --debug
// Flag.target: type=string, default=main, usage=Target branch for the pull request
// Flag.title: type=string, usage=Custom PR title (description still AI-generated)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode for AI generation
package work

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/ready-to-release/eac/go/eac/commands/impl/work/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

func init() {
	registry.Register(CreatePR)
}

// prFlags defines valid flags for the pr command

// Intent: Create pull request with AI-generated description
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear flow: validate → push → generate description → create PR
//   - Uses gh CLI for PR creation
//   - AI generates comprehensive description from all commits
//
// Easy to change:
//   - PR generation logic isolated
//   - Target branch configurable
//   - Custom title supported
//
// Hard to break:
//   - Validates uncommitted changes
//   - Checks gh CLI availability
//   - Ensures commits exist before creating PR
//   - Pushes branch automatically if needed

// CreatePR creates a pull request for the current workspace
func CreatePR() int {
	cmdStart := time.Now()

	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Phase 1: Parse configuration
	phase1Start := time.Now()
	config, err := parsePRConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	defer config.base.Logger.Sync()

	config.base.Logger.Debug("Phase 1: Starting configuration parsing",
		zap.String("phase", "phase1"))

	config.base.Logger.Debug("Starting work pr command",
		zap.String("currentBranch", config.currentBranch),
		zap.String("targetBranch", config.targetBranch),
		zap.String("customTitle", config.customTitle),
		zap.Bool("debug", config.base.Debug))

	config.base.Logger.Debug("Phase 1: Completed",
		zap.String("phase", "phase1"),
		zap.Duration("duration", time.Since(phase1Start)))

	// Phase 2: Validate environment
	phase2Start := time.Now()
	config.base.Logger.Debug("Phase 2: Starting environment validation",
		zap.String("phase", "phase2"),
		zap.String("currentBranch", config.currentBranch),
		zap.String("targetBranch", config.targetBranch))

	if err := validatePREnvironment(config); err != nil {
		config.base.Logger.Debug("Phase 2: Failed",
			zap.String("phase", "phase2"),
			zap.Error(err),
			zap.Duration("duration", time.Since(phase2Start)))
		config.base.Logger.Error(fmt.Sprintf("Validation failed: %v", err))
		return 1
	}

	config.base.Logger.Debug("Phase 2: Completed",
		zap.String("phase", "phase2"),
		zap.Duration("duration", time.Since(phase2Start)))

	// Phase 3: Check for commits
	phase3Start := time.Now()
	config.base.Logger.Debug("Phase 3: Starting commit count check",
		zap.String("phase", "phase3"),
		zap.String("targetBranch", config.targetBranch))

	config.base.Logger.Info("Checking branch status...")
	commitCount, err := getCommitCount(config)
	if err != nil {
		config.base.Logger.Debug("Phase 3: Failed",
			zap.String("phase", "phase3"),
			zap.Error(err),
			zap.Duration("duration", time.Since(phase3Start)))
		config.base.Logger.Error(fmt.Sprintf("Failed to count commits: %v", err))
		return 1
	}
	if commitCount == 0 {
		config.base.Logger.Debug("Phase 3: Failed - no commits",
			zap.String("phase", "phase3"),
			zap.Int("commitCount", commitCount),
			zap.Duration("duration", time.Since(phase3Start)))
		config.base.Logger.Error(fmt.Sprintf("No commits ahead of %s", config.targetBranch))
		config.base.Logger.Error("Your branch has no new commits to create a PR from")
		return 1
	}

	config.base.Logger.Debug("Phase 3: Completed",
		zap.String("phase", "phase3"),
		zap.Int("commitCount", commitCount),
		zap.Duration("duration", time.Since(phase3Start)))

	// Phase 4: Push branch to origin
	phase4Start := time.Now()
	config.base.Logger.Debug("Phase 4: Starting branch push",
		zap.String("phase", "phase4"),
		zap.String("branch", config.currentBranch))

	config.base.Logger.Info("Pushing to origin...")
	if err := pushBranch(config); err != nil {
		config.base.Logger.Debug("Phase 4: Failed",
			zap.String("phase", "phase4"),
			zap.Error(err),
			zap.Duration("duration", time.Since(phase4Start)))
		config.base.Logger.Error(fmt.Sprintf("Failed to push: %v", err))
		return 1
	}

	config.base.Logger.Debug("Phase 4: Completed",
		zap.String("phase", "phase4"),
		zap.Duration("duration", time.Since(phase4Start)))

	// Phase 5: Generate PR title and description
	phase5Start := time.Now()
	config.base.Logger.Debug("Phase 5: Starting PR content generation",
		zap.String("phase", "phase5"),
		zap.String("targetBranch", config.targetBranch),
		zap.Int("commitCount", commitCount))

	config.base.Logger.Info("\nGenerating PR description...")
	title, description, err := generatePRContent(config)
	if err != nil {
		config.base.Logger.Debug("Phase 5: Failed",
			zap.String("phase", "phase5"),
			zap.Error(err),
			zap.Duration("duration", time.Since(phase5Start)))
		config.base.Logger.Error(fmt.Sprintf("Failed to generate PR content: %v", err))
		return 1
	}

	// Use custom title if provided
	if config.customTitle != "" {
		config.base.Logger.Debug("Using custom title",
			zap.String("customTitle", config.customTitle),
			zap.String("generatedTitle", title))
		title = config.customTitle
	}

	config.base.Logger.Debug("Phase 5: Completed",
		zap.String("phase", "phase5"),
		zap.String("title", title),
		zap.Int("descriptionLength", len(description)),
		zap.Duration("duration", time.Since(phase5Start)))

	// Phase 6: Create pull request
	phase6Start := time.Now()
	config.base.Logger.Debug("Phase 6: Starting pull request creation",
		zap.String("phase", "phase6"),
		zap.String("head", config.currentBranch),
		zap.String("base", config.targetBranch),
		zap.String("title", title))

	config.base.Logger.Info("Creating pull request...")
	prURL, err := createPullRequest(config.base.Logger, title, description, config.currentBranch, config.targetBranch)
	if err != nil {
		config.base.Logger.Debug("Phase 6: Failed",
			zap.String("phase", "phase6"),
			zap.Error(err),
			zap.Duration("duration", time.Since(phase6Start)))
		config.base.Logger.Error(fmt.Sprintf("Failed to create PR: %v", err))
		return 1
	}

	config.base.Logger.Debug("Phase 6: Completed",
		zap.String("phase", "phase6"),
		zap.String("prURL", prURL),
		zap.Duration("duration", time.Since(phase6Start)))

	// Phase 7: Success
	config.base.Logger.Debug("Phase 7: Finalizing",
		zap.String("phase", "phase7"),
		zap.String("prURL", prURL))

	config.base.Logger.Info("")
	config.base.Logger.Info(fmt.Sprintf("✓ Pull request created: %s", prURL))

	config.base.Logger.Debug("Phase 7: Completed",
		zap.String("phase", "phase7"))

	config.base.Logger.Debug("Work pr command completed successfully",
		zap.Duration("totalDuration", time.Since(cmdStart)))

	return 0
}

// prConfig holds configuration for the pr command
type prConfig struct {
	base          *internal.BaseConfig
	targetBranch  string
	customTitle   string
	currentBranch string
}

// parsePRConfig parses command line arguments
func parsePRConfig() (*prConfig, error) {
	args := os.Args[3:] // Skip program name, "work", "pr"

	// Parse base config (debug flag, repo root, logger, git ops)
	baseConfig, err := internal.ParseBaseConfig(args)
	if err != nil {
		return nil, err
	}

	config := &prConfig{
		base:         baseConfig,
		targetBranch: "main",
	}

	// Parse flags
	if targetValue := flags.GetFlagValue(args, "--target"); targetValue != "" {
		config.targetBranch = targetValue
	}
	if titleValue := flags.GetFlagValue(args, "--title"); titleValue != "" {
		config.customTitle = titleValue
	}

	// Also check for --title with space
	for i := 0; i < len(args); i++ {
		if args[i] == "--title" {
			if i+1 < len(args) {
				config.customTitle = args[i+1]
				break
			} else {
				return nil, fmt.Errorf("--title requires a value")
			}
		}
	}

	// Get current branch from current working directory (not repoRoot)
	// This ensures we get the correct branch in worktree environments
	// Check R2R_PWD first (for test isolation)
	cwd := os.Getenv("R2R_PWD")
	if cwd == "" {
		// Fall back to actual working directory
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}
	currentBranch, err := config.base.GitOps.GetCurrentBranch(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	config.currentBranch = currentBranch

	return config, nil
}

// validatePREnvironment validates the environment before creating PR
func validatePREnvironment(config *prConfig) error {
	// Check we're in a git repository
	config.base.Logger.Debug("Checking git repository status")
	if err := internal.EnsureInGitRepo(); err != nil {
		config.base.Logger.Debug("Not in git repository", zap.Error(err))
		return err
	}

	// Prevent creating PR from main
	config.base.Logger.Debug("Checking branch is not main/master",
		zap.String("currentBranch", config.currentBranch))
	if config.currentBranch == "main" || config.currentBranch == "master" {
		return fmt.Errorf("cannot create PR from main\nYou are on the main branch. Switch to a workspace first.")
	}

	// Check for uncommitted changes
	// Check R2R_PWD first (for test isolation)
	config.base.Logger.Debug("Checking for uncommitted changes")
	cwd := os.Getenv("R2R_PWD")
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	clean, err := config.base.GitOps.IsWorktreeClean(cwd)
	if err != nil {
		config.base.Logger.Debug("Failed to check worktree status", zap.Error(err))
		return fmt.Errorf("failed to check working tree status: %w", err)
	}
	if !clean {
		config.base.Logger.Debug("Uncommitted changes detected")
		return fmt.Errorf("uncommitted changes detected\nCommit your changes first before creating a PR")
	}
	config.base.Logger.Debug("Worktree is clean")

	// Check if gh CLI is available
	config.base.Logger.Debug("Checking GitHub CLI availability")
	if err := checkGHCLI(); err != nil {
		config.base.Logger.Debug("GitHub CLI check failed", zap.Error(err))
		return err
	}
	config.base.Logger.Debug("GitHub CLI is available")

	return nil
}

// checkGHCLI checks if GitHub CLI is installed and available
func checkGHCLI() error {
	cmd := exec.Command("gh", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh CLI not found\nInstall GitHub CLI: https://cli.github.com/")
	}
	return nil
}

// getCommitCount returns the number of commits ahead of target branch
func getCommitCount(config *prConfig) (int, error) {
	config.base.Logger.Debug("Getting commit count",
		zap.String("targetBranch", config.targetBranch))

	count, err := config.base.GitOps.GetCommitCount(fmt.Sprintf("origin/%s", config.targetBranch), "HEAD")
	if err != nil {
		config.base.Logger.Debug("Failed to get commit count", zap.Error(err))
		return 0, fmt.Errorf("failed to count commits: %w", err)
	}

	config.base.Logger.Debug("Commit count retrieved",
		zap.Int("count", count))
	return count, nil
}

// pushBranch pushes the current branch to origin
func pushBranch(config *prConfig) error {
	config.base.Logger.Debug("Pushing branch to origin",
		zap.String("branch", config.currentBranch))

	if err := config.base.GitOps.PushBranch(config.currentBranch, false); err != nil {
		config.base.Logger.Debug("Failed to push branch", zap.Error(err))
		return err
	}

	config.base.Logger.Debug("Branch pushed successfully",
		zap.String("branch", config.currentBranch))
	return nil
}

// generatePRContent generates the PR title and description using AI
func generatePRContent(config *prConfig) (string, string, error) {
	// Get all commits from the branch
	config.base.Logger.Debug("Getting commit messages",
		zap.String("range", fmt.Sprintf("origin/%s..HEAD", config.targetBranch)))

	cmd := exec.Command("git", "log", fmt.Sprintf("origin/%s..HEAD", config.targetBranch), "--pretty=format:%s")
	output, err := cmd.Output()
	if err != nil {
		config.base.Logger.Debug("Failed to get commit messages", zap.Error(err))
		return "", "", fmt.Errorf("failed to get commit messages: %w", err)
	}

	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	config.base.Logger.Debug("Retrieved commit messages",
		zap.Int("commitCount", len(commits)))

	// Get the diff for the entire branch
	config.base.Logger.Debug("Getting diff statistics",
		zap.String("range", fmt.Sprintf("origin/%s...HEAD", config.targetBranch)))

	cmd = exec.Command("git", "diff", fmt.Sprintf("origin/%s...HEAD", config.targetBranch), "--stat")
	diffOutput, err := cmd.Output()
	if err != nil {
		config.base.Logger.Debug("Failed to get diff", zap.Error(err))
		return "", "", fmt.Errorf("failed to get diff: %w", err)
	}

	config.base.Logger.Debug("Retrieved diff statistics",
		zap.Int("diffLength", len(diffOutput)))

	// Generate title from first commit or branch name
	config.base.Logger.Debug("Generating PR title",
		zap.Int("commitCount", len(commits)),
		zap.String("branchName", config.currentBranch))

	title := generatePRTitle(commits, config.currentBranch)
	config.base.Logger.Debug("Generated PR title",
		zap.String("title", title))

	// Generate description
	config.base.Logger.Debug("Generating PR description")
	description := generatePRDescription(commits, string(diffOutput))
	config.base.Logger.Debug("Generated PR description",
		zap.Int("descriptionLength", len(description)))

	return title, description, nil
}

// generatePRTitle generates a PR title from commits or branch name
func generatePRTitle(commits []string, branchName string) string {
	if len(commits) > 0 && commits[0] != "" {
		// Use first commit message as title
		return commits[0]
	}

	// Fall back to formatted branch name
	// Convert feature/authentication -> Add authentication
	parts := strings.Split(branchName, "/")
	if len(parts) > 1 {
		feature := strings.ReplaceAll(parts[1], "-", " ")
		feature = strings.Title(feature)
		return fmt.Sprintf("Add %s", feature)
	}

	return fmt.Sprintf("Changes from %s", branchName)
}

// generatePRDescription generates a PR description from commits and diff
func generatePRDescription(commits []string, diffStat string) string {
	var sb strings.Builder

	sb.WriteString("## Summary\n\n")

	// List commits
	if len(commits) > 1 {
		sb.WriteString(fmt.Sprintf("This PR includes %d commits:\n\n", len(commits)))
		for _, commit := range commits {
			if commit != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", commit))
			}
		}
	} else if len(commits) == 1 {
		sb.WriteString(fmt.Sprintf("%s\n", commits[0]))
	}

	sb.WriteString("\n## Changes\n\n")
	sb.WriteString("```\n")
	sb.WriteString(diffStat)
	sb.WriteString("\n```\n")

	sb.WriteString("\n## Test Plan\n\n")
	sb.WriteString("- [ ] Manual testing completed\n")
	sb.WriteString("- [ ] Unit tests pass\n")
	sb.WriteString("- [ ] Integration tests pass\n")

	return sb.String()
}

// createPullRequest creates a pull request using gh CLI
func createPullRequest(logger *logging.Logger, title, description, head, base string) (string, error) {
	logger.Debug("Executing gh pr create command",
		zap.String("title", title),
		zap.String("head", head),
		zap.String("base", base),
		zap.Int("descriptionLength", len(description)))

	cmd := exec.Command("gh", "pr", "create",
		"--title", title,
		"--body", description,
		"--base", base,
		"--head", head,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Failed to execute gh pr create",
			zap.Error(err),
			zap.String("output", string(output)))
		return "", fmt.Errorf("failed to create pull request: %w\nOutput: %s", err, string(output))
	}

	// Extract PR URL from output
	prURL := strings.TrimSpace(string(output))
	logger.Debug("Pull request created successfully",
		zap.String("prURL", prURL))
	return prURL, nil
}
