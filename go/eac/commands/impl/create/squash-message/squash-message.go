// Command: create squash-message
// Description: Generate squash commit message from branch commits
// Short: Generate squash commit message from branch commits
// Long: Analyzes all commits in the current branch compared to the base branch
// Long: and generates a comprehensive, cohesive commit message suitable for
// Long: squash merges in pull requests or local branch merging.
// Long:
// Long: This command examines the full commit history and cumulative diff
// Long: to create a message that accurately represents the entire feature
// Long: or change set, rather than individual commit details.
// Long:
// Long: The generated message is designed to be copied into GitHub's PR squash
// Long: merge UI for use when squashing and merging pull requests.
// Long:
// Long: Expected Output:
// Long: - Comprehensive commit message for squash merge
// Long: - Suitable for GitHub PR squash merge UI
// Long:
// Long: Example:
// Long:   create squash-message
// Long:   create squash-message --base=develop
// Long:   create squash-message --debug
// Flag.base: type=string, default=main, usage=Base branch for comparison
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode to save intermediate outputs
package squashmessage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/ai"
	"github.com/ready-to-release/eac/go/eac/commands/internal/ai/providers"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(CreateSquashMessage)
}

var log = logging.C()

// logDebugArtifact logs debug content with labeled sections to the log file.
// This replaces writeDebugFile - content goes to out/commands.log instead of separate files.
func logDebugArtifact(logger *logging.Logger, label string, content string) {
	logger.Debug(fmt.Sprintf("=== %s START ===", label))
	logger.Debug(content)
	logger.Debug(fmt.Sprintf("=== %s END ===", label))
}

// gitRepo holds the git repository instance for testing
var gitRepo git.GitRepository

// getGitRepo returns the git repository, initializing it if needed.
func getGitRepo(workspaceRoot string) (git.GitRepository, error) {
	if gitRepo != nil {
		return gitRepo, nil
	}
	repo, err := git.Open(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	return repo, nil
}

// SetGitRepo allows tests to inject a mock repository.
func SetGitRepo(repo git.GitRepository) {
	gitRepo = repo
}

// CreateSquashMessage is the main entry point for the create squash-message command.
func CreateSquashMessage() int {
	// Phase 1: Parse configuration
	config, err := parseConfig()
	if err != nil {
		log.Errorf("ERROR: %v", err)
		return 1
	}

	// Phase 2: Get workspace root (needed for logger)
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("ERROR: Failed to get workspace root: %v", err)
		return 1
	}

	// Phase 3: Initialize logger with workspace root
	var logger *logging.Logger
	if config.debug {
		logger, err = logging.NewWithDebug("create", workspaceRoot)
	} else {
		logger, err = logging.NewDefault("create", workspaceRoot)
	}
	if err != nil {
		log.Errorf("ERROR: Failed to initialize logger: %v", err)
		return 1
	}
	defer logger.Sync()

	log.Debug("Starting create squash-message command")

	// Phase 4: Open git repository
	repo, err := getGitRepo(workspaceRoot)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to open git repository: %v", err))
		return 1
	}

	// Phase 5: Get current branch
	currentBranch, err := repo.CurrentBranch()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get current branch: %v", err))
		return 1
	}
	logger.Debug(fmt.Sprintf("Current branch: %s", currentBranch))

	// Phase 6: Get branch commits
	logger.Info(fmt.Sprintf("Analyzing commits from %s...HEAD", config.baseBranch))
	commits, err := repo.GetBranchCommits(config.baseBranch)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get branch commits: %v", err))
		logger.Error(fmt.Sprintf("Ensure branch '%s' exists and you have commits ahead of it", config.baseBranch))
		return 1
	}
	logger.Debug(fmt.Sprintf("Found %d commits", len(commits)))

	// Phase 7: Get branch diff and files
	diff, err := repo.GetBranchDiff(config.baseBranch)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to get branch diff: %v", err))
		diff = ""
	}

	diffStats, err := repo.GetBranchDiffStats(config.baseBranch)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to get diff stats: %v", err))
		diffStats = ""
	}

	fileNames, err := repo.GetBranchFiles(config.baseBranch)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get branch files: %v", err))
		return 1
	}
	logger.Debug(fmt.Sprintf("Found %d changed files", len(fileNames)))

	// Phase 8: Get module mappings
	// Convert file names to FileInfo
	var fileInfos []repository.FileInfo
	for _, fileName := range fileNames {
		fileInfos = append(fileInfos, repository.FileInfo{
			Path:         fileName,
			AbsolutePath: filepath.Join(workspaceRoot, fileName),
			IsTracked:    true,
			IsIgnored:    false,
		})
	}

	// Enrich with module information
	filesWithModules, err := repository.EnrichFilesWithModules(fileInfos, workspaceRoot)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get module mappings: %v", err))
		return 1
	}

	affectedModules := extractAffectedModules(filesWithModules)
	logger.Debug(fmt.Sprintf("Affected modules: %v", affectedModules))

	// Phase 9: Build prompt context
	context := buildSquashContext(currentBranch, config.baseBranch, commits, filesWithModules, diff, diffStats, affectedModules)
	logDebugArtifact(logger, "SQUASH-CONTEXT", context)

	// Phase 10: Generate top-level message using AI
	logger.Info("Generating squash commit message using AI...")
	topLevelMessage, err := generateTopLevelMessage(workspaceRoot, logger, context)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to generate top-level message: %v", err))
		return 1
	}

	// Phase 11: Generate module sections (reuse commit-message logic)
	moduleSections, err := generateModuleSections(workspaceRoot, logger, affectedModules, filesWithModules, diff)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to generate module sections: %v", err))
		return 1
	}

	// Phase 12: Assemble final message
	finalMessage := assembleMessage(topLevelMessage, moduleSections)
	logDebugArtifact(logger, "SQUASH-FINAL", finalMessage)

	// Phase 13: Output message (validation skipped for now)
	fmt.Println(">>>>>>OUTPUT START<<<<<<")
	fmt.Println(finalMessage)

	logger.Debug("Create squash-message command completed successfully")
	return 0
}

// squashConfig holds the parsed command configuration
type squashConfig struct {
	baseBranch string
	debug      bool
}

// parseConfig parses command line arguments
func parseConfig() (*squashConfig, error) {
	args := os.Args[3:] // Skip "r2r", "create", "squash-message"

	cfg := &squashConfig{
		baseBranch: "main",
		debug:      false,
	}

	// Parse debug flag using shared package
	cfg.debug = flags.ParseDebugFlag(args)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--base=") {
			cfg.baseBranch = strings.TrimPrefix(arg, "--base=")
		} else if arg == "--base" {
			if i+1 < len(args) {
				cfg.baseBranch = args[i+1]
				i++
			} else {
				return nil, fmt.Errorf("--base requires a value")
			}
		} else if arg == "--debug" || arg == "-d" {
			// Already handled by shared flags package
		} else {
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	return cfg, nil
}

// extractAffectedModules extracts unique module names from file list
func extractAffectedModules(files []repository.RepositoryFileWithModule) []string {
	moduleSet := make(map[string]bool)
	for _, f := range files {
		for _, module := range f.Modules {
			moduleSet[module] = true
		}
	}

	var modules []string
	for module := range moduleSet {
		modules = append(modules, module)
	}
	return modules
}

// buildSquashContext builds the context string for AI generation
func buildSquashContext(currentBranch, baseBranch string, commits []git.CommitInfo, files []repository.RepositoryFileWithModule, diff, diffStats string, affectedModules []string) string {
	var buf strings.Builder

	// Branch information
	buf.WriteString("## Branch Information\n\n")
	buf.WriteString(fmt.Sprintf("Base branch: %s\n", baseBranch))
	buf.WriteString(fmt.Sprintf("Current branch: %s\n", currentBranch))
	buf.WriteString(fmt.Sprintf("Commits ahead: %d\n\n", len(commits)))

	// Commit history
	buf.WriteString("## Commit History\n\n")
	for i, commit := range commits {
		buf.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, commit.Subject, commit.ShortSHA))
		if commit.Message != commit.Subject {
			// Add body if it exists (excluding subject line)
			body := strings.TrimPrefix(commit.Message, commit.Subject)
			body = strings.TrimSpace(body)
			if body != "" {
				// Indent body lines
				bodyLines := strings.Split(body, "\n")
				for _, line := range bodyLines {
					if line != "" {
						buf.WriteString(fmt.Sprintf("   %s\n", line))
					}
				}
			}
		}
		buf.WriteString("\n")
	}

	// Module count
	buf.WriteString("## Module Count\n\n")
	if len(affectedModules) == 1 {
		buf.WriteString("1 (single-module)\n\n")
	} else {
		buf.WriteString(fmt.Sprintf("%d (multi-module)\n\n", len(affectedModules)))
	}

	// Affected modules
	buf.WriteString("## Affected Modules\n\n")
	for _, module := range affectedModules {
		buf.WriteString(fmt.Sprintf("- %s\n", module))
	}
	buf.WriteString("\n")

	// Files table
	buf.WriteString("## Changed Files\n\n")
	buf.WriteString(buildFilesTable(files))
	buf.WriteString("\n\n")

	// Diff stats
	if diffStats != "" {
		buf.WriteString("## Diff Stats\n\n")
		buf.WriteString(diffStats)
		buf.WriteString("\n\n")
	}

	// Cumulative diff (truncated)
	buf.WriteString("## Cumulative Diff\n\n")
	buf.WriteString("```diff\n")
	if len(diff) > 200000 {
		buf.WriteString(diff[:200000])
		buf.WriteString("\n... (diff truncated)\n")
	} else {
		buf.WriteString(diff)
	}
	buf.WriteString("\n```\n")

	return buf.String()
}

// buildFilesTable builds a markdown table of changed files
func buildFilesTable(files []repository.RepositoryFileWithModule) string {
	var buf strings.Builder
	buf.WriteString("| File | Modules |\n")
	buf.WriteString("|------|--------|\n")
	for _, f := range files {
		modulesStr := strings.Join(f.Modules, ", ")
		if modulesStr == "" {
			modulesStr = "NONE"
		}
		buf.WriteString(fmt.Sprintf("| %s | %s |\n", f.Name, modulesStr))
	}
	return buf.String()
}

// generateTopLevelMessage generates the top-level commit message using AI
func generateTopLevelMessage(workspaceRoot string, logger *logging.Logger, promptContext string) (string, error) {
	// Load squash prompt template with three-tier priority:
	// 1. Command flag (not applicable - internal function)
	// 2. Team override (.r2r/eac/templates/ai/commit-message/squash.md)
	// 3. System default (templates/ai/commit-message/squash.md)
	// Note: "squash" is a variant prompt (convention adds .md automatically)
	loader := contracts.NewContractLoader(workspaceRoot, "ai/commit-message", "")
	promptTemplate, _, err := loader.LoadPrompt("squash", "")
	if err != nil {
		return "", fmt.Errorf("failed to load squash.md template: %w", err)
	}

	// Build full prompt
	prompt := string(promptTemplate) + "\n\n" + promptContext

	logDebugArtifact(logger, "SQUASH-PROMPT", prompt)

	// Execute AI (includes test provider support for mocking)
	executor := ai.NewExecutor(workspaceRoot)
	providers.RegisterBuiltIn(executor)
	ctx := context.Background()
	result, err := executor.Execute(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("AI execution failed: %w", err)
	}

	logDebugArtifact(logger, "SQUASH-AI-RESPONSE", result)

	return strings.TrimSpace(result), nil
}

// generateModuleSections generates per-module sections (reuse commit-message logic)
func generateModuleSections(workspaceRoot string, logger *logging.Logger, affectedModules []string, files []repository.RepositoryFileWithModule, diff string) (map[string]string, error) {
	// For now, return empty map - module sections can be added later if needed
	// This keeps the implementation simple
	return make(map[string]string), nil
}

// assembleMessage assembles the final message from parts
func assembleMessage(topLevel string, moduleSections map[string]string) string {
	var buf strings.Builder
	buf.WriteString(topLevel)

	// Add module sections if any
	for _, section := range moduleSections {
		buf.WriteString("\n\n")
		buf.WriteString(section)
	}

	return buf.String()
}
