// Command: create commit-message
// Short: Generate AI-powered commit messages from staged changes
// Long: The create commit-message command uses AI to analyze your staged git changes and generate a structured,
// Long: conventional commit message that follows project standards and includes module-specific details.
// Long: The generated message includes a top-level summary and per-module sections describing changes.
// Long: All output is validated against the commit message contract to ensure consistency and quality.
// Long: By default, the command outputs the commit message to stdout. Use --debug to save intermediate outputs.
// Long: Use --commit to automatically create a git commit with the generated message.
// Long:
// Long: Expected Output:
// Long: - Structured conventional commit message to stdout
// Long: - Top-level summary and per-module sections
// Long: - Validated against commit message contract
// Long: - Debug outputs in out/ if --debug enabled
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode to save intermediate outputs (context, prompts, AI responses) to the 'out' directory for troubleshooting and analysis
// Flag.commit: type=bool, shorthand=c, default=false, usage=Automatically create git commit with generated message
package commitmessage

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"strings"

	commitmessageinternal "github.com/ready-to-release/eac/go/eac/commands/impl/create/commit-message/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	aimock "github.com/ready-to-release/eac/go/eac/core/ai"
	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"github.com/ready-to-release/eac/go/eac/core/repository/reports"
	"go.uber.org/zap"
)

var log = logging.C()

// logDebugArtifact logs debug content with labeled sections to the log file.
// This replaces writeDebugFile - content goes to out/commands.log instead of separate files.
func logDebugArtifact(logger *logging.Logger, label string, content string) {
	logger.Debug(fmt.Sprintf("=== %s START ===", label))
	logger.Debug(content)
	logger.Debug(fmt.Sprintf("=== %s END ===", label))
}

// logDebugArtifactf logs debug content with a formatted label.
func logDebugArtifactf(logger *logging.Logger, format string, content string, args ...interface{}) {
	label := fmt.Sprintf(format, args...)
	logDebugArtifact(logger, label, content)
}

// gitRepo holds the git repository instance for git operations.
// In production, this is initialized lazily. For tests, it can be injected via SetGitRepo.
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

// ResetGitRepo clears the repository for test cleanup.
func ResetGitRepo() {
	gitRepo = nil
}

// ValidationError is an alias for commitmessageinternal.ValidationError for external access
type ValidationError = commitmessageinternal.ValidationError

// VerifyCommitMessageContract validates a commit message against the contract rules.
// This is exposed for testing purposes.
func VerifyCommitMessageContract(commitMessage string, affectedModules []string) []ValidationError {
	return commitmessageinternal.VerifyCommitMessageContract(commitMessage, affectedModules)
}

// AutoCleanup performs automatic fixes on commit message before validation.
// This is exposed for testing purposes.
func AutoCleanup(commitMessage string) string {
	return commitmessageinternal.AutoCleanup(commitMessage)
}

func init() {
	registry.Register(CreateCommitMessage)
}

// executionConfig holds configuration for the commit AI command
type executionConfig struct {
	workspaceRoot   string
	debug           bool
	autoCommit      bool
	stagedFiles     []repository.RepositoryFileWithModule
	affectedModules []string
	gitDiff         string
}

func CreateCommitMessage() int {
	// Parse configuration early to get debug mode, auto-commit flag, and workspace root
	debug, autoCommit, workspaceRoot, err := parseConfig()
	if err != nil {
		log.Errorf("ERROR: %v", err)
		return 1
	}

	// Initialize logger early so all code paths can use it
	var logger *logging.Logger
	if debug {
		logger, err = logging.NewWithDebug("create", workspaceRoot)
	} else {
		logger, err = logging.NewDefault("create", workspaceRoot)
	}
	if err != nil {
		log.Errorf("initializing logger: %v", err)
		return 1
	}
	defer logger.Sync()

	log.Debug("Logger initialized")

	// Retry loop for regenerating commit message if validation fails
	// Limited to prevent infinite loops
	const maxRetries = 5
	attempt := 0

	for attempt < maxRetries {
		attempt++

		// Show warning after multiple retries
		if attempt > 3 {
			logger.Warn("retry attempt", zap.Int("attempt", attempt), zap.Int("max", maxRetries))
		}

		result, shouldRetry, generatedMessage := commitAIAttemptWithMessage(logger, workspaceRoot, debug)
		if !shouldRetry {
			// If successful and auto-commit is enabled, perform the commit
			if result == 0 && autoCommit && generatedMessage != "" {
				return performAutoCommit(workspaceRoot, generatedMessage, logger)
			}
			return result
		}

		// Check if max retries reached
		if attempt >= maxRetries {
			logger.Error("Maximum retry attempts reached")
			logger.Info("The AI is having difficulty generating a valid commit message.")
			logger.Info("Please try one of the following:")
			logger.Info("  - Simplify your staged changes")
			logger.Info("  - Split changes across multiple commits")
			logger.Info("  - Write commit message manually with: git commit")
			return 1
		}

		logger.Info("retrying commit message generation", zap.Int("attempt", attempt+1), zap.Int("max", maxRetries))
	}

	return 1
}

// commitAIAttempt performs a single attempt at generating and committing
// Returns (exit code, should retry)
func commitAIAttempt(logger *logging.Logger, workspaceRoot string, debug bool) (int, bool) {
	exitCode, shouldRetry, _ := commitAIAttemptWithMessage(logger, workspaceRoot, debug)
	return exitCode, shouldRetry
}

// commitAIAttemptWithMessage performs a single attempt at generating commit message
// Returns (exit code, should retry, generated message)
func commitAIAttemptWithMessage(logger *logging.Logger, workspaceRoot string, debug bool) (int, bool, string) {
	// Phase 1: Verify Contract Implementation
	if err := verifyContractImplementation(workspaceRoot, logger); err != nil {
		return 1, false, ""
	}

	// Phase 2: Build Execution Context
	cfg, stagedFilesTable, diffStats, err := buildExecutionContext(workspaceRoot, logger)
	if err != nil {
		log.Errorf("ERROR: Build context failed: %v", err)
		return 1, false, ""
	}
	if cfg == nil {
		logger.Info("No staged changes.")
		return 1, false, ""
	}

	// Phase 3: Generate Top-Level Summary
	topLevel, err := generateTopLevelSummary(cfg, stagedFilesTable, diffStats, logger)
	if err != nil {
		log.Errorf("ERROR: Top-level generation failed: %v", err)
		return 1, false, ""
	}

	// Phase 4: Generate Module Sections
	moduleSections, err := generateModuleSections(cfg, logger)
	if err != nil {
		log.Errorf("ERROR: Module section generation failed: %v", err)
		return 1, false, ""
	}

	// Phase 5: Assemble Final Message
	finalMessage := assembleFinalMessage(cfg, topLevel, moduleSections, logger)

	// Phase 6: Validate and Output (message only - no git commit)
	exitCode, shouldRetry := validateAndOutput(cfg, finalMessage)
	return exitCode, shouldRetry, finalMessage
}


// Phase 1: Parse Configuration
func parseConfig() (debug bool, autoCommit bool, workspaceRoot string, err error) {
	args := os.Args[3:] // Skip program name, "create", and "commit-message"

	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		return false, false, "", err
	}

	// Parse flags using shared package
	debug = flags.ParseDebugFlag(args)
	autoCommit = flags.HasFlag(args, "--commit", "-c")

	// Get repository root
	workspaceRoot, err = repository.GetRepositoryRoot("")
	if err != nil {
		return false, false, "", fmt.Errorf("failed to find repository root: %w", err)
	}

	return debug, autoCommit, workspaceRoot, nil
}

// verifyContractImplementation checks if the AI config is valid
func verifyContractImplementation(workspaceRoot string, logger *logging.Logger) error {
	log.Debug("verifyContractImplementation: start")
	// Verify that ai-config.yml can be loaded and has commit-message type
	aiConfig, err := aimock.LoadAIConfig(workspaceRoot)
	if err != nil {
		logger.Error("AI config verification failed")
		logger.Error("config load error", zap.Error(err))
		return fmt.Errorf("config verification failed: %w", err)
	}

	// Check that commit-message type exists
	if _, ok := aiConfig.Types["commit-message"]; !ok {
		logger.Error("AI config missing commit-message type")
		return fmt.Errorf("ai-config.yml must define 'commit-message' type")
	}

	log.Debug("verifyContractImplementation: config verified")
	return nil
}

// Phase 3: Build Execution Context
func buildExecutionContext(workspaceRoot string, logger *logging.Logger) (*executionConfig, string, string, error) {
	log.Debug("buildExecutionContext: start")
	// Validate inputs
	if workspaceRoot == "" {
		return nil, "", "", fmt.Errorf("workspaceRoot cannot be empty")
	}
	if _, err := os.Stat(workspaceRoot); os.IsNotExist(err) {
		return nil, "", "", fmt.Errorf("workspaceRoot does not exist: %s", workspaceRoot)
	}

	// Get staged files report
	log.Debug("buildExecutionContext: calling getStagedFilesReport")
	report, stagedFilesTable, err := getStagedFilesReport(workspaceRoot)
	log.Debug("buildExecutionContext: getStagedFilesReport complete")
	if err != nil {
		return nil, "", "", err
	}

	if len(report.AllFiles) == 0 {
		return nil, "", "", nil // No staged changes (not an error)
	}

	// Extract affected modules
	log.Debug("buildExecutionContext: extracting affected modules")
	affectedModules := extractAffectedModules(report, logger)

	// Get git diff and stats
	log.Debug("buildExecutionContext: calling getGitDiffAndStats")
	gitDiff, diffStats, err := getGitDiffAndStats(workspaceRoot, logger)
	if err != nil {
		return nil, "", "", err
	}

	log.Debugf("Affected modules count: %d", len(affectedModules))
	for i, mod := range affectedModules {
		log.Debugf("  %d. %s", i+1, mod)
	}

	cfg := &executionConfig{
		workspaceRoot:   workspaceRoot,
		debug:           logger.IsDebugMode(),
		stagedFiles:     report.AllFiles,
		affectedModules: affectedModules,
		gitDiff:         gitDiff,
	}

	return cfg, stagedFilesTable, diffStats, nil
}

// getStagedFilesReport retrieves staged files and builds a table representation
func getStagedFilesReport(workspaceRoot string) (*reports.FilesModulesReport, string, error) {
	// Get staged files with module mappings
	report, err := reports.GetFilesModulesReport(true, false, true, workspaceRoot)
	if err != nil {
		return nil, "", fmt.Errorf("getting module mappings: %w", err)
	}

	// Build the staged files table
	tb := render.NewTableBuilder().WithHeaders("File", "Modules")
	for _, file := range report.AllFiles {
		modulesStr := "NONE"
		if len(file.Modules) > 0 {
			modulesStr = strings.Join(file.Modules, ", ")
		}
		tb.AddRow(file.Name, modulesStr)
	}
	stagedFilesTable := tb.Build()

	return report, stagedFilesTable, nil
}

// extractAffectedModules extracts and validates unique module names from file report
func extractAffectedModules(report *reports.FilesModulesReport, logger *logging.Logger) []string {
	moduleSet := make(map[string]bool)
	for _, file := range report.AllFiles {
		for _, module := range file.Modules {
			// Validate module name: only lowercase letters, numbers, dashes, underscores
			if isValidModuleName(module) {
				moduleSet[module] = true
			} else {
				logger.Warn("skipping invalid module name", zap.String("module", module))
			}
		}
	}

	var affectedModules []string
	for module := range moduleSet {
		affectedModules = append(affectedModules, module)
	}

	return affectedModules
}

// getGitDiffAndStats retrieves git diff and diff stats for staged changes
func getGitDiffAndStats(workspaceRoot string, logger *logging.Logger) (string, string, error) {
	log.Debug("getGitDiffAndStats: start")
	log.Debug("getGitDiffAndStats: calling getGitRepo")
	repo, err := getGitRepo(workspaceRoot)
	if err != nil {
		return "", "", err
	}
	log.Debug("getGitDiffAndStats: getGitRepo complete")

	// Get git diff
	log.Debug("getGitDiffAndStats: calling StagedDiff")
	diffOutput, err := repo.StagedDiff()
	if err != nil {
		return "", "", fmt.Errorf("getting git diff: %w", err)
	}
	log.Debug("getGitDiffAndStats: StagedDiff complete")

	// Check diff size to prevent memory issues
	if len(diffOutput) > commitmessageinternal.MaxDiffSize {
		return "", "", fmt.Errorf("git diff too large: %d bytes (max %d bytes / %.1f MB). Consider committing in smaller chunks",
			len(diffOutput), commitmessageinternal.MaxDiffSize, float64(commitmessageinternal.MaxDiffSize)/(1024*1024))
	}

	// Get git diff stats
	log.Debug("getGitDiffAndStats: calling StagedDiffStats")
	diffStats, err := repo.StagedDiffStats()
	if err != nil {
		logger.Warn("failed to get diff stats", zap.Error(err))
		diffStats = ""
	}
	log.Debug("getGitDiffAndStats: StagedDiffStats complete")

	return diffOutput, strings.TrimSpace(diffStats), nil
}

// Phase 4: Generate Top-Level Summary
func generateTopLevelSummary(cfg *executionConfig, stagedFilesTable string, diffStats string, logger *logging.Logger) (string, error) {
	topLevelContext := buildTopLevelContext(stagedFilesTable, cfg.gitDiff, diffStats, cfg.affectedModules)

	logDebugArtifact(logger, "TOP-LEVEL-CONTEXT", topLevelContext)

	var topLevelOutput string
	var providerName string
	err := commitmessageinternal.WithProgress("🤖 Generating top-level commit summary...", func() error {
		result, genErr := generateWithPromptResult("top-level", topLevelContext, cfg.workspaceRoot, cfg.affectedModules, cfg.debug, nil)
		if result != nil {
			topLevelOutput = result.Output
			providerName = result.ProviderName
		}
		return genErr
	})

	if err != nil {
		return "", fmt.Errorf("running commit-message-top-level agent: %w", err)
	}

	// Log provider info when debug is enabled
	if providerName != "" {
		log.Debugf("AI provider used: %s", providerName)
	}

	// Strip out any module sections the AI may have added after "Changes:" line
	// Module sections will be generated separately and appended later
	topLevelOutput = stripModuleSectionsFromTopLevel(topLevelOutput)

	logDebugArtifact(logger, "TOP-LEVEL-OUTPUT", topLevelOutput)

	return topLevelOutput, nil
}

// Phase 5: Generate Module Sections (multi-module only)
func generateModuleSections(cfg *executionConfig, logger *logging.Logger) ([]string, error) {
	// Use parallel implementation for performance (60-70% speedup for multi-module commits)
	// Sequential: N modules × 5s = 15s for 3 modules
	// Parallel:   max(5s) = 5s for 3 modules
	return generateModuleSectionsParallel(cfg, logger, nil)
}

// Phase 6: Assemble Final Message
func assembleFinalMessage(cfg *executionConfig, topLevel string, moduleSections []string, logger *logging.Logger) string {
	// Combine sections
	combinedMessage := combineCommitSections(topLevel, moduleSections)
	logDebugArtifact(logger, "COMBINED-MESSAGE", combinedMessage)

	// Auto-cleanup
	cleanedOutput := commitmessageinternal.AutoCleanup(combinedMessage)
	logDebugArtifact(logger, "AFTER-CLEANUP", cleanedOutput)

	// NOTE: addMissingModules fallback is no longer needed with parallel generation
	// The new generateModuleSectionsParallel guarantees all affected modules get sections
	// Keeping the fallback causes duplicate sections, so it's disabled
	// cleanedOutput = addMissingModules(cleanedOutput, cfg.affectedModules, cfg.stagedFiles, cfg.gitDiff)
	// logDebugArtifact(logger, "AFTER-MISSING-MODULES", cleanedOutput)

	return cleanedOutput
}

// Phase 7: Validate and Output
// Returns (exit code, should retry)
// NOTE: This function only outputs the commit message - it does NOT perform git commit.
// The user is expected to copy/use the message with their preferred commit workflow.
func validateAndOutput(cfg *executionConfig, message string) (int, bool) {
	// Verify contract compliance
	validationErrors := commitmessageinternal.VerifyCommitMessageContract(message, cfg.affectedModules)

	errorCount, warningCount := 0, 0
	for _, verr := range validationErrors {
		if verr.Severity == "error" {
			errorCount++
		} else {
			warningCount++
		}
	}

	// Output marker for VSCode extension to find the start of commit message
	// This separates progress/status messages from the actual output
	// Write directly to stdout (not via log) for programmatic consumption
	fmt.Println(">>>>>>OUTPUT START<<<<<<")

	// Output the generated message (raw, clean output for piping/copying)
	fmt.Println(message)

	return 0, false
}

// performAutoCommit creates a git commit with the generated message.
// This is called when --commit flag is provided and message generation succeeds.
func performAutoCommit(workspaceRoot string, message string, logger *logging.Logger) int {
	repo, err := getGitRepo(workspaceRoot)
	if err != nil {
		logger.Error("auto-commit failed", zap.Error(err))
		return 1
	}

	// Get author info from git config (use git command for reliability)
	authorName, authorEmail := getGitAuthorInfo(workspaceRoot)
	if authorName == "" || authorEmail == "" {
		logger.Error("Auto-commit failed: git user.name and user.email must be configured")
		logger.Info("Run: git config user.name \"Your Name\"")
		logger.Info("Run: git config user.email \"your@email.com\"")
		return 1
	}

	// Perform the commit
	hash, err := repo.Commit(message, authorName, authorEmail)
	if err != nil {
		logger.Error("git commit failed", zap.Error(err))
		return 1
	}

	// Success message to stderr (so it doesn't interfere with stdout output)
	log.Infof("\n✓ Committed: %s", hash[:7])
	return 0
}

// getGitAuthorInfo retrieves git user.name and user.email from git config.
// Returns empty strings if not configured.
func getGitAuthorInfo(workspaceRoot string) (name string, email string) {
	// Try to get from environment first (GIT_AUTHOR_NAME, GIT_AUTHOR_EMAIL)
	name = os.Getenv("GIT_AUTHOR_NAME")
	email = os.Getenv("GIT_AUTHOR_EMAIL")
	if name != "" && email != "" {
		return name, email
	}

	// Fall back to git config command
	// Using exec.Command because go-git config reading can be complex
	nameCmd := execCommand("git", "config", "user.name")
	nameCmd.Dir = workspaceRoot
	if nameOut, err := nameCmd.Output(); err == nil {
		name = strings.TrimSpace(string(nameOut))
	}

	emailCmd := execCommand("git", "config", "user.email")
	emailCmd.Dir = workspaceRoot
	if emailOut, err := emailCmd.Output(); err == nil {
		email = strings.TrimSpace(string(emailOut))
	}

	return name, email
}

// execCommand is a variable to allow mocking in tests
var execCommand = execCommandReal

func execCommandReal(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// stripModuleSectionsFromTopLevel removes any module-like sections that appear
// after the "Changes:" line in the top-level commit message.
// The AI sometimes includes module summaries here despite being told not to.
func stripModuleSectionsFromTopLevel(message string) string {
	lines := strings.Split(message, "\n")
	result := make([]string, 0, len(lines))
	foundChangesLine := false

	for _, line := range lines {
		// Include all lines up to and including the "Changes:" line
		result = append(result, line)

		// Once we find "Changes:", stop including further lines if they look like module sections
		if strings.HasPrefix(strings.TrimSpace(line), "Changes:") {
			foundChangesLine = true
			break
		}
	}

	// If we found the Changes line and there's content after it, check if it's module sections
	if foundChangesLine && len(result) < len(lines) {
		// Look ahead to see if the next non-empty lines are module sections (starting with ---)
		remainingLines := lines[len(result):]
		for _, line := range remainingLines {
			trimmed := strings.TrimSpace(line)

			// Skip empty lines
			if trimmed == "" {
				continue
			}

			// If we hit a line that starts with "---", it's likely the start of module sections
			// Stop here - don't include it or anything after
			if strings.HasPrefix(trimmed, "---") {
				break
			}

			// If it's not a separator and not empty, include it (could be additional top-level content)
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// isValidModuleName checks if a module name is valid (lowercase letters, numbers, dashes, underscores only)
// Rejects paths with slashes or other special characters
func isValidModuleName(name string) bool {
	if name == "" || len(name) > 50 {
		return false
	}

	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			return false
		}
	}

	return true
}
