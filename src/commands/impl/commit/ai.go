// Command: commit-ai
// Description: Generate commit message using AI with staged changes and module mappings
// Short: Generate AI-powered commit messages from staged changes
// Long: The commit-ai command uses AI to analyze your staged git changes and generate a structured,
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
	"os/exec"
	"path/filepath"
	"strings"

	commitmessage "github.com/ready-to-release/eac/src/commands/impl/commit/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/commands/internal/render"
	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/ready-to-release/eac/src/core/repository/reports"
)

func init() {
	registry.Register(CommitAI)
}

// executionConfig holds configuration for the commit AI command
type executionConfig struct {
	workspaceRoot   string
	debug           bool
	stagedFiles     []repository.RepositoryFileWithModule
	affectedModules []string
	gitDiff         string
}

func CommitAI() int {
	// Retry loop for regenerating commit message if validation fails
	for {
		result, shouldRetry := commitAIAttempt()
		if !shouldRetry {
			return result
		}
		fmt.Println("\n🔄 Retrying commit message generation...")
	}
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

	// Phase 7: Validate and Output (with interactive prompt on failure)
	return validateAndCommit(cfg, finalMessage)
}

// Phase 1: Parse Configuration
func parseConfig() (bool, string, error) {
	// Parse flags
	debug := false
	for _, arg := range os.Args[2:] { // Skip program name and "commit-ai"
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
	// Get staged files with module mappings
	report, err := reports.GetFilesModulesReport(true, false, true, workspaceRoot, "0.1.0")
	if err != nil {
		return nil, "", "", fmt.Errorf("getting module mappings: %w", err)
	}

	if len(report.AllFiles) == 0 {
		return nil, "", "", nil // No staged changes (not an error)
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

	// Extract unique modules and validate them
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

	// Get git diff
	diffCmd := exec.Command("git", "diff", "--staged")
	diffCmd.Dir = workspaceRoot
	diffOutput, err := diffCmd.Output()
	if err != nil {
		return nil, "", "", fmt.Errorf("getting git diff: %w", err)
	}

	// Check diff size to prevent memory issues
	if len(diffOutput) > commitmessage.MaxDiffSize {
		return nil, "", "", fmt.Errorf("git diff too large: %d bytes (max %d bytes / %.1f MB). Consider committing in smaller chunks",
			len(diffOutput), commitmessage.MaxDiffSize, float64(commitmessage.MaxDiffSize)/(1024*1024))
	}

	// Get git diff stats
	statsCmd := exec.Command("git", "diff", "--staged", "--stat")
	statsCmd.Dir = workspaceRoot
	statsOutput, err := statsCmd.Output()
	if err != nil {
		debugWriter.log("Warning: Failed to get diff stats: %v", err)
		statsOutput = []byte("")
	}
	diffStats := strings.TrimSpace(string(statsOutput))

	debugWriter.log("Affected modules count: %d", len(affectedModules))
	for i, mod := range affectedModules {
		debugWriter.log("  %d. %s", i+1, mod)
	}

	cfg := &executionConfig{
		workspaceRoot:   workspaceRoot,
		debug:           debugWriter.enabled,
		stagedFiles:     report.AllFiles,
		affectedModules: affectedModules,
		gitDiff:         string(diffOutput),
	}

	return cfg, stagedFilesTable, diffStats, nil
}

// Phase 4: Generate Top-Level Summary
func generateTopLevelSummary(cfg *executionConfig, stagedFilesTable string, diffStats string, debugWriter *debugWriter) (string, error) {
	topLevelContext := buildTopLevelContext(stagedFilesTable, cfg.gitDiff, diffStats, cfg.affectedModules)

	debugWriter.write("debug-top-level-context.md", topLevelContext)

	var topLevelOutput string
	err := commitmessage.WithProgress("🤖 Generating top-level commit summary...", func() error {
		result, genErr := generateWithPrompt("top-level", topLevelContext, cfg.workspaceRoot, cfg.affectedModules, cfg.debug)
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
	return generateModuleSectionsParallel(cfg, debugWriter)
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

// Phase 7: Validate and Commit (with interactive prompt)
// Returns (exit code, should retry)
func validateAndCommit(cfg *executionConfig, message string) (int, bool) {
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

	// Show the generated message
	fmt.Println("\n📝 Generated commit message:")
	fmt.Println("---")
	fmt.Println(message)
	fmt.Println("---")

	// If valid (no errors, only warnings or clean), proceed with commit
	if errorCount == 0 {
		if warningCount > 0 {
			fmt.Printf("\n⚠️  Found %d warning(s):\n\n", warningCount)
			for _, verr := range validationErrors {
				fmt.Printf("⚠️  %s\n", verr.Error())
			}
			fmt.Println()
		}
		return performCommit(message), false
	}

	// Show validation errors
	fmt.Printf("\n❌ Found %d contract violation(s):\n\n", errorCount)
	if warningCount > 0 {
		fmt.Printf("⚠️  Found %d warning(s):\n\n", warningCount)
	}

	for _, verr := range validationErrors {
		icon := "❌"
		if verr.Severity == "warning" {
			icon = "⚠️ "
		}
		fmt.Printf("%s %s\n", icon, verr.Error())
	}

	// Prompt user for action
	fmt.Println()
	response := promptYNR("Use this message anyway?")

	switch response {
	case "y":
		// User wants to use the message despite validation errors
		fmt.Println("✓ Proceeding with commit (ignoring validation errors)...")
		return performCommit(message), false
	case "n":
		// User wants to abort and write their own message
		fmt.Println("❌ Commit aborted.")
		fmt.Println("   Run 'git commit' or 'work commit --message \"your message\"' to write your own.")
		return 1, false
	case "r":
		// User wants to retry AI generation
		return 0, true
	default:
		// Should never happen, but treat as abort
		return 1, false
	}
}

// performCommit executes git commit with the given message
func performCommit(message string) int {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: failed to create commit: %v\n", err)
		return 1
	}
	fmt.Println("\n✓ Commit created successfully")
	return 0
}

// promptYNR prompts the user with a yes/no/retry question
// Returns "y", "n", or "r"
func promptYNR(question string) string {
	fmt.Printf("%s (y/n/r): ", question)

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))

	switch response {
	case "y", "yes":
		return "y"
	case "n", "no":
		return "n"
	case "r", "retry":
		return "r"
	default:
		fmt.Printf("Invalid input '%s'. Please enter y (yes), n (no), or r (retry).\n", response)
		return promptYNR(question) // Ask again
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
