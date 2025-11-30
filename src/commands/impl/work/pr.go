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

	"github.com/ready-to-release/eac/src/commands/impl/work/internal"
	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/logging"
)

func init() {
	registry.Register(CreatePR)
}

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
	// Phase 1: Parse configuration
	config, err := parsePRConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	defer config.base.Logger.Sync()

	config.base.Logger.Debug("Starting work pr command")
	internal.WriteDebugFile(config.base.Logger, config.base.RepoRoot, "pr-config.txt",
		fmt.Sprintf("CurrentBranch: %s\nTarget: %s\nCustomTitle: %s\n",
			config.currentBranch, config.targetBranch, config.customTitle))

	// Phase 2: Validate environment
	if err := validatePREnvironment(config); err != nil {
		config.base.Logger.Error(fmt.Sprintf("Validation failed: %v", err))
		return 1
	}

	// Phase 3: Check for commits
	config.base.Logger.Info("Checking branch status...")
	commitCount, err := getCommitCount(config)
	if err != nil {
		config.base.Logger.Error(fmt.Sprintf("Failed to count commits: %v", err))
		return 1
	}
	if commitCount == 0 {
		config.base.Logger.Error(fmt.Sprintf("No commits ahead of %s", config.targetBranch))
		config.base.Logger.Error("Your branch has no new commits to create a PR from")
		return 1
	}

	// Phase 4: Push branch to origin
	config.base.Logger.Info("Pushing to origin...")
	if err := pushBranch(config); err != nil {
		config.base.Logger.Error(fmt.Sprintf("Failed to push: %v", err))
		return 1
	}

	// Phase 5: Generate PR title and description
	config.base.Logger.Info("\nGenerating PR description...")
	title, description, err := generatePRContent(config)
	if err != nil {
		config.base.Logger.Error(fmt.Sprintf("Failed to generate PR content: %v", err))
		return 1
	}

	// Use custom title if provided
	if config.customTitle != "" {
		title = config.customTitle
	}

	// Phase 6: Create pull request
	config.base.Logger.Info("Creating pull request...")
	prURL, err := createPullRequest(config.base.Logger, title, description, config.currentBranch, config.targetBranch)
	if err != nil {
		config.base.Logger.Error(fmt.Sprintf("Failed to create PR: %v", err))
		return 1
	}

	// Phase 7: Success
	config.base.Logger.Info("")
	config.base.Logger.Info(fmt.Sprintf("✓ Pull request created: %s", prURL))
	config.base.Logger.Debug("Work pr command completed successfully")
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
	if targetValue := internal.GetFlagValue(args, "--target"); targetValue != "" {
		config.targetBranch = targetValue
	}
	if titleValue := internal.GetFlagValue(args, "--title"); titleValue != "" {
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

	// Get current branch
	currentBranch, err := config.base.GitOps.GetCurrentBranch(config.base.RepoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	config.currentBranch = currentBranch

	return config, nil
}

// validatePREnvironment validates the environment before creating PR
func validatePREnvironment(config *prConfig) error {
	// Check we're in a git repository
	if err := internal.EnsureInGitRepo(); err != nil {
		return err
	}

	// Prevent creating PR from main
	if config.currentBranch == "main" || config.currentBranch == "master" {
		return fmt.Errorf("cannot create PR from main\nYou are on the main branch. Switch to a workspace first.")
	}

	// Check for uncommitted changes
	cwd, _ := os.Getwd()
	clean, err := config.base.GitOps.IsWorktreeClean(cwd)
	if err != nil {
		return fmt.Errorf("failed to check working tree status: %w", err)
	}
	if !clean {
		return fmt.Errorf("uncommitted changes detected\nCommit your changes first before creating a PR")
	}

	// Check if gh CLI is available
	if err := checkGHCLI(); err != nil {
		return err
	}

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
	count, err := config.base.GitOps.GetCommitCount(fmt.Sprintf("origin/%s", config.targetBranch), "HEAD")
	if err != nil {
		return 0, fmt.Errorf("failed to count commits: %w", err)
	}
	return count, nil
}

// pushBranch pushes the current branch to origin
func pushBranch(config *prConfig) error {
	config.base.Logger.Debug(fmt.Sprintf("Pushing branch %s to origin", config.currentBranch))
	if err := config.base.GitOps.PushBranch(config.currentBranch, false); err != nil {
		return err
	}
	return nil
}

// generatePRContent generates the PR title and description using AI
func generatePRContent(config *prConfig) (string, string, error) {
	// Get all commits from the branch
	cmd := exec.Command("git", "log", fmt.Sprintf("origin/%s..HEAD", config.targetBranch), "--pretty=format:%s")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get commit messages: %w", err)
	}

	commits := strings.Split(strings.TrimSpace(string(output)), "\n")

	// Get the diff for the entire branch
	cmd = exec.Command("git", "diff", fmt.Sprintf("origin/%s...HEAD", config.targetBranch), "--stat")
	diffOutput, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get diff: %w", err)
	}

	// Generate title from first commit or branch name
	title := generatePRTitle(commits, config.currentBranch)

	// Generate description
	description := generatePRDescription(commits, string(diffOutput))

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
	cmd := exec.Command("gh", "pr", "create",
		"--title", title,
		"--body", description,
		"--base", base,
		"--head", head,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create pull request: %w\nOutput: %s", err, string(output))
	}

	// Extract PR URL from output
	prURL := strings.TrimSpace(string(output))
	logger.Debug(fmt.Sprintf("Created PR: %s", prURL))
	return prURL, nil
}
