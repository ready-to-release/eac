// Command: commit message
// Description: Generate commit message using AI with staged changes and module mappings
// Short: Generate AI-powered commit messages from staged changes
// Long: The commit message command uses AI to analyze your staged git changes and generate a structured,
// Long: conventional commit message that follows project standards and includes module-specific details.
// Long: The generated message includes a top-level summary and per-module sections describing changes.
// Long: All output is validated against the commit message contract to ensure consistency and quality.
// Long: By default, the command outputs the commit message to stdout. Use --debug to save intermediate outputs.
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode to save intermediate outputs (context, prompts, AI responses) to the 'out' directory for troubleshooting and analysis
// Flags: --debug (save intermediate outputs and show debug info)
// HasSideEffects: false
package commit

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	commitmessage "github.com/ready-to-release/eac/src/commands/impl/commit/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/commands/internal/render"
	"github.com/ready-to-release/eac/src/core/git"
	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/ready-to-release/eac/src/core/repository/reports"
)

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

// ValidationError is an alias for commitmessage.ValidationError for external access
type ValidationError = commitmessage.ValidationError

// VerifyCommitMessageContract validates a commit message against the contract rules.
// This is exposed for testing purposes.
func VerifyCommitMessageContract(commitMessage string, affectedModules []string) []ValidationError {
	return commitmessage.VerifyCommitMessageContract(commitMessage, affectedModules)
}

// AutoCleanup performs automatic fixes on commit message before validation.
// This is exposed for testing purposes.
func AutoCleanup(commitMessage string) string {
	return commitmessage.AutoCleanup(commitMessage)
}

func init() {
	registry.Register(CommitMessage)
}

// executionConfig holds configuration for the commit AI command
type executionConfig struct {
	workspaceRoot   string
	debug           bool
	stagedFiles     []repository.RepositoryFileWithModule
	affectedModules []string
	gitDiff         string
}

func CommitMessage() int {
	// Retry loop for regenerating commit message if validation fails
	// Limited to prevent infinite loops
	const maxRetries = 5
	attempt := 0

	for attempt < maxRetries {
		attempt++

		// Show warning after multiple retries
		if attempt > 3 {
			fmt.Printf("\n⚠️  Retry attempt %d/%d\n", attempt, maxRetries)
		}

		result, shouldRetry := commitAIAttempt()
		if !shouldRetry {
			return result
		}

		// Check if max retries reached
		if attempt >= maxRetries {
			fmt.Println("\n❌ Maximum retry attempts reached.")
			fmt.Println("   The AI is having difficulty generating a valid commit message.")
			fmt.Println("   Please try one of the following:")
			fmt.Println("   - Simplify your staged changes")
			fmt.Println("   - Split changes across multiple commits")
			fmt.Println("   - Write commit message manually with: git commit")
			return 1
		}

		fmt.Printf("\n🔄 Retrying commit message generation (%d/%d)...\n", attempt+1, maxRetries)
	}

	return 1
}

// commitAIAttempt performs a single attempt at generating and committing
// Returns (exit code, should retry)
func commitAIAttempt() (int, bool) {
	// Phase 1: Parse Configuration
	debug, workspaceRoot, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1, false
	}

	// Phase 2: Verify Contract Implementation
	if err := verifyContractImplementation(workspaceRoot); err != nil {
		return 1, false
	}

	// Create debug writer
	debugWriter := newDebugWriter(debug, workspaceRoot)

	// Phase 3: Build Execution Context
	cfg, stagedFilesTable, diffStats, err := buildExecutionContext(workspaceRoot, debugWriter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1, false
	}
	if cfg == nil {
		fmt.Println("No staged changes.")
		return 0, false
	}

	// Phase 4: Generate Top-Level Summary
	topLevel, err := generateTopLevelSummary(cfg, stagedFilesTable, diffStats, debugWriter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		return 1, false
	}

	// Phase 5: Generate Module Sections
	moduleSections, err := generateModuleSections(cfg, debugWriter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		return 1, false
	}

	// Phase 6: Assemble Final Message
	finalMessage := assembleFinalMessage(cfg, topLevel, moduleSections, debugWriter)

	// Phase 7: Validate and Output (message only - no git commit)
	return validateAndOutput(cfg, finalMessage)
}

// Phase 1: Parse Configuration
func parseConfig() (bool, string, error) {
	// Parse flags
	debug := false
	for _, arg := range os.Args[3:] { // Skip program name, "commit", and "message"
		if arg == "--debug" {
			debug = true
		}
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return false, "", fmt.Errorf("failed to find repository root: %w", err)
	}

	return debug, workspaceRoot, nil
}

// Phase 2: Verify Contract Implementation
func verifyContractImplementation(workspaceRoot string) error {
	contractPath := filepath.Join(workspaceRoot, "contracts/ai/commit-message/0.1.0/contract.yml")
	contractErrors := commitmessage.VerifyContractImplementation(contractPath)
	if len(contractErrors) > 0 {
		fmt.Fprintf(os.Stderr, "❌ Contract implementation verification failed:\n")
		for _, err := range contractErrors {
			fmt.Fprintf(os.Stderr, "  - [%s] %s\n", err.Code, err.Message)
		}
		return fmt.Errorf("contract verification failed")
	}
	return nil
}

// Phase 3: Build Execution Context
func buildExecutionContext(workspaceRoot string, debugWriter *debugWriter) (*executionConfig, string, string, error) {
	// Validate inputs
	if workspaceRoot == "" {
		return nil, "", "", fmt.Errorf("workspaceRoot cannot be empty")
	}
	if _, err := os.Stat(workspaceRoot); os.IsNotExist(err) {
		return nil, "", "", fmt.Errorf("workspaceRoot does not exist: %s", workspaceRoot)
	}

	// Get staged files report
	report, stagedFilesTable, err := getStagedFilesReport(workspaceRoot)
	if err != nil {
		return nil, "", "", err
	}

	if len(report.AllFiles) == 0 {
		return nil, "", "", nil // No staged changes (not an error)
	}

	// Extract affected modules
	affectedModules := extractAffectedModules(report, debugWriter)

	// Get git diff and stats
	gitDiff, diffStats, err := getGitDiffAndStats(workspaceRoot, debugWriter)
	if err != nil {
		return nil, "", "", err
	}

	debugWriter.log("Affected modules count: %d", len(affectedModules))
	for i, mod := range affectedModules {
		debugWriter.log("  %d. %s", i+1, mod)
	}

	cfg := &executionConfig{
		workspaceRoot:   workspaceRoot,
		debug:           debugWriter.enabled,
		stagedFiles:     report.AllFiles,
		affectedModules: affectedModules,
		gitDiff:         gitDiff,
	}

	return cfg, stagedFilesTable, diffStats, nil
}

// getStagedFilesReport retrieves staged files and builds a table representation
func getStagedFilesReport(workspaceRoot string) (*reports.FilesModulesReport, string, error) {
	// Get staged files with module mappings
	report, err := reports.GetFilesModulesReport(true, false, true, workspaceRoot, "0.1.0")
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
func extractAffectedModules(report *reports.FilesModulesReport, debugWriter *debugWriter) []string {
	moduleSet := make(map[string]bool)
	for _, file := range report.AllFiles {
		for _, module := range file.Modules {
			// Validate module name: only lowercase letters, numbers, dashes, underscores
			if isValidModuleName(module) {
				moduleSet[module] = true
			} else {
				debugWriter.log("WARNING: Skipping invalid module name: %s", module)
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
func getGitDiffAndStats(workspaceRoot string, debugWriter *debugWriter) (string, string, error) {
	repo, err := getGitRepo(workspaceRoot)
	if err != nil {
		return "", "", err
	}

	// Get git diff
	diffOutput, err := repo.StagedDiff()
	if err != nil {
		return "", "", fmt.Errorf("getting git diff: %w", err)
	}

	// Check diff size to prevent memory issues
	if len(diffOutput) > commitmessage.MaxDiffSize {
		return "", "", fmt.Errorf("git diff too large: %d bytes (max %d bytes / %.1f MB). Consider committing in smaller chunks",
			len(diffOutput), commitmessage.MaxDiffSize, float64(commitmessage.MaxDiffSize)/(1024*1024))
	}

	// Get git diff stats
	diffStats, err := repo.StagedDiffStats()
	if err != nil {
		debugWriter.log("Warning: Failed to get diff stats: %v", err)
		diffStats = ""
	}

	return diffOutput, strings.TrimSpace(diffStats), nil
}

// Phase 4: Generate Top-Level Summary
func generateTopLevelSummary(cfg *executionConfig, stagedFilesTable string, diffStats string, debugWriter *debugWriter) (string, error) {
	topLevelContext := buildTopLevelContext(stagedFilesTable, cfg.gitDiff, diffStats, cfg.affectedModules)

	debugWriter.write("debug-top-level-context.md", topLevelContext)

	var topLevelOutput string
	err := commitmessage.WithProgress("🤖 Generating top-level commit summary...", func() error {
		result, genErr := generateWithPrompt("top-level", topLevelContext, cfg.workspaceRoot, cfg.affectedModules, cfg.debug, nil)
		topLevelOutput = result
		return genErr
	})

	if err != nil {
		return "", fmt.Errorf("running commit-message-top-level agent: %w", err)
	}

	// Strip out any module sections the AI may have added after "Changes:" line
	// Module sections will be generated separately and appended later
	topLevelOutput = stripModuleSectionsFromTopLevel(topLevelOutput)

	debugWriter.write("debug-top-level-output.md", topLevelOutput)

	return topLevelOutput, nil
}

// Phase 5: Generate Module Sections (multi-module only)
func generateModuleSections(cfg *executionConfig, debugWriter *debugWriter) ([]string, error) {
	// Use parallel implementation for performance (60-70% speedup for multi-module commits)
	// Sequential: N modules × 5s = 15s for 3 modules
	// Parallel:   max(5s) = 5s for 3 modules
	return generateModuleSectionsParallel(cfg, debugWriter, nil)
}

// Phase 6: Assemble Final Message
func assembleFinalMessage(cfg *executionConfig, topLevel string, moduleSections []string, debugWriter *debugWriter) string {
	// Combine sections
	combinedMessage := combineCommitSections(topLevel, moduleSections)
	debugWriter.write("debug-combined-message.md", combinedMessage)

	// Auto-cleanup
	cleanedOutput := commitmessage.AutoCleanup(combinedMessage)
	debugWriter.write("debug-after-cleanup.md", cleanedOutput)

	// Add missing modules
	cleanedOutput = addMissingModules(cleanedOutput, cfg.affectedModules, cfg.stagedFiles, cfg.gitDiff)
	debugWriter.write("debug-after-missing-modules.md", cleanedOutput)

	return cleanedOutput
}

// Phase 7: Validate and Output
// Returns (exit code, should retry)
// NOTE: This function only outputs the commit message - it does NOT perform git commit.
// The user is expected to copy/use the message with their preferred commit workflow.
func validateAndOutput(cfg *executionConfig, message string) (int, bool) {
	// Verify contract compliance
	validationErrors := commitmessage.VerifyCommitMessageContract(message, cfg.affectedModules)

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
	fmt.Println(">>>>>>OUTPUT START<<<<<<")

	// Output the generated message (raw, clean output for piping/copying)
	fmt.Print(message)

	return 0, false
}

// promptYN prompts the user with a yes/no question
// Returns "y" or "n"
func promptYN(question string) string {
	return promptYNWithRetries(question, 0)
}

// promptYNWithRetries prompts with retry limit to prevent infinite recursion
func promptYNWithRetries(question string, attempt int) string {
	const maxAttempts = 3

	if attempt >= maxAttempts {
		fmt.Printf("\nToo many invalid inputs. Defaulting to 'no'.\n")
		return "n"
	}

	fmt.Printf("%s (y/n): ", question)

	var response string
	_, err := fmt.Scanln(&response)

	// If we can't read from stdin (non-interactive), default to "no"
	if err != nil {
		fmt.Printf("\nNo input available (non-interactive mode). Defaulting to 'no'.\n")
		return "n"
	}

	response = strings.ToLower(strings.TrimSpace(response))

	// If response is empty (stdin exhausted), default to "no"
	if response == "" {
		fmt.Printf("\nEmpty input received. Defaulting to 'no'.\n")
		return "n"
	}

	switch response {
	case "y", "yes":
		return "y"
	case "n", "no":
		return "n"
	default:
		fmt.Printf("Invalid input '%s'. Please enter y (yes) or n (no).\n", response)
		return promptYNWithRetries(question, attempt+1)
	}
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
