// Command: commit-ai
// Description: Generate commit message using AI with staged changes and module mappings
// Flags: --debug (save intermediate outputs and show debug info)
// HasSideEffects: false
package commit

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	commitmessage "github.com/ready-to-release/eac/src/commands/impl/commit/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/commands/internal/render"
	"github.com/ready-to-release/eac/src/core/ai"
	"github.com/ready-to-release/eac/src/core/ai/providers"
	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/ready-to-release/eac/src/core/repository/reports"
)

// Embedded built-in prompts (compiled into binary)
//go:embed prompts/top-level.md
var builtinTopLevelPrompt string

//go:embed prompts/module.md
var builtinModulePrompt string

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

// debugWriter handles debug file output with consistent error handling.
//
// Thread-safe: All methods use mutex to protect concurrent file operations.
// This is required for parallel module section generation where multiple
// goroutines may write debug files simultaneously.
type debugWriter struct {
	enabled       bool
	workspaceRoot string
	mu            sync.Mutex // Protects file write operations
}

func newDebugWriter(enabled bool, workspaceRoot string) *debugWriter {
	return &debugWriter{
		enabled:       enabled,
		workspaceRoot: workspaceRoot,
	}
}

func (d *debugWriter) write(filename string, content string) {
	if !d.enabled {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	debugFile := filepath.Join(d.workspaceRoot, "out", filename)
	if err := os.WriteFile(debugFile, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Failed to write debug file %s: %v\n", debugFile, err)
	} else {
		fmt.Fprintf(os.Stderr, "🔍 DEBUG: Saved to %s\n", debugFile)
	}
}

func (d *debugWriter) writef(format string, content string, args ...interface{}) {
	if !d.enabled {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	filename := fmt.Sprintf(format, args...)
	debugFile := filepath.Join(d.workspaceRoot, "out", filename)
	if err := os.WriteFile(debugFile, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Failed to write debug file %s: %v\n", debugFile, err)
	} else {
		fmt.Fprintf(os.Stderr, "🔍 DEBUG: Saved to %s\n", debugFile)
	}
}

func (d *debugWriter) log(format string, args ...interface{}) {
	if !d.enabled {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	fmt.Fprintf(os.Stderr, "🔍 DEBUG: "+format+"\n", args...)
}

func CommitAI() int {
	// Phase 1: Parse Configuration
	debug, workspaceRoot, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 2: Verify Contract Implementation
	if err := verifyContractImplementation(workspaceRoot); err != nil {
		return 1
	}

	// Create debug writer
	debugWriter := newDebugWriter(debug, workspaceRoot)

	// Phase 3: Build Execution Context
	cfg, stagedFilesTable, diffStats, err := buildExecutionContext(workspaceRoot, debugWriter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if cfg == nil {
		fmt.Println("No staged changes.")
		return 0
	}

	// Phase 4: Generate Top-Level Summary
	topLevel, err := generateTopLevelSummary(cfg, stagedFilesTable, diffStats, debugWriter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		return 1
	}

	// Phase 5: Generate Module Sections
	moduleSections, err := generateModuleSections(cfg, debugWriter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		return 1
	}

	// Phase 6: Assemble Final Message
	finalMessage := assembleFinalMessage(cfg, topLevel, moduleSections, debugWriter)

	// Phase 7: Validate and Output
	return validateAndOutput(cfg, finalMessage)
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
	contractPath := filepath.Join(workspaceRoot, "contracts/commit-message/0.1.0/structure.yml")
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

	// Extract unique modules
	moduleSet := make(map[string]bool)
	for _, file := range report.AllFiles {
		for _, module := range file.Modules {
			moduleSet[module] = true
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
	var rawOutput string
	err := commitmessage.WithProgress("🤖 Generating top-level commit summary...", func() error {
		// Load prompt
		agentContent, err := loadPromptWithFallback("top-level", cfg.workspaceRoot)
		if err != nil {
			return fmt.Errorf("failed to load prompt: %w", err)
		}

		model := extractModelFromAgent(agentContent)

		// Create executor
		executor := ai.NewExecutor(cfg.workspaceRoot)
		providers.RegisterBuiltIn(executor)

		// Build full prompt
		fullPrompt := agentContent + "\n\n>>>>>>>>>>INPUT STARTS NOW<<<<<<<<<<<\n\n" + topLevelContext

		// Execute
		ctx := context.Background()
		var opts []ai.Option
		if model != "" {
			opts = append(opts, ai.WithModel(model))
		}
		output, err := executor.Execute(ctx, fullPrompt, opts...)
		if err != nil {
			return fmt.Errorf("AI execution failed: %w", err)
		}

		// Save raw output before any processing
		rawOutput = strings.TrimSpace(output)

		// Strip noise
		topLevelOutput = stripAgentNoise(rawOutput, "top-level")

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("running commit-message-top-level agent: %w", err)
	}

	debugWriter.write("debug-top-level-raw.md", rawOutput)
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

// Phase 7: Validate and Output
func validateAndOutput(cfg *executionConfig, message string) int {
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

	// Output for VSCode extension to detect
	fmt.Println(">>>>>>OUTPUT START<<<<<<")
	fmt.Println(message)
	fmt.Println("\n---")

	// Print verification results
	if len(validationErrors) == 0 {
		fmt.Println() // Just a blank line
		return 0
	}

	// Show validation errors/warnings
	if errorCount > 0 {
		fmt.Printf("❌ Found %d contract violation(s):\n\n", errorCount)
	}
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

	if errorCount > 0 {
		return 1
	}

	return 0
}

// generateWithPrompt executes an AI prompt using the configured provider.
//
// This function orchestrates prompt-based text generation by:
//   1. Loading the prompt file with fallback (user override → embedded)
//   2. Extracting model preferences from frontmatter
//   3. Combining prompt instructions with user-provided context
//   4. Invoking the AI provider via the executor abstraction layer
//   5. Stripping agent noise (initialization messages, conversational wrappers, markdown)
//   6. Returning clean output ready for contract validation
//
// The prompt type is automatically inferred from the name to apply
// type-specific noise filtering:
//   - "top-level" → expects "<type>(<scope>): <summary>" with Auditor-Summary
//   - "module" → expects plain text module name with dashes separator
//
// Parameters:
//   - promptName:    Name of the prompt file without extension
//                    (e.g., "top-level", "module")
//   - userPrompt:    Context and input for the prompt
//                    (e.g., staged files table, git diff, module list, diff stats)
//   - workspaceRoot: Repository root directory for executor context
//
// Returns:
//   - Generated output with agent initialization noise removed
//   - Error if prompt file cannot be read or AI execution fails
//
// Example:
//
//	topLevelContext := buildTopLevelContext(files, diff, diffStats, modules)
//	output, err := generateWithPrompt(
//	    "top-level",
//	    topLevelContext,
//	    "/workspace/root",
//	)
// loadPromptWithFallback implements two-tier prompt loading:
// 1. User override: .r2r/prompts/commit/<name>.md
// 2. Built-in: embedded prompts/<name>.md
func loadPromptWithFallback(promptName string, workspaceRoot string) (string, error) {
	// Tier 1: Check for user override in .r2r/prompts/commit/
	userOverridePath := filepath.Join(workspaceRoot, ".r2r", "prompts", "commit", promptName+".md")
	if content, err := os.ReadFile(userOverridePath); err == nil {
		return string(content), nil
	}

	// Tier 2: Use built-in embedded prompt
	switch promptName {
	case "top-level":
		return builtinTopLevelPrompt, nil
	case "module":
		return builtinModulePrompt, nil
	default:
		return "", fmt.Errorf("unknown prompt name: %s", promptName)
	}
}

// generateWithPrompt generates output using the two-tier prompt loading system
func generateWithPrompt(promptName string, userPrompt string, workspaceRoot string) (string, error) {
	// Load prompt using three-tier system
	agentContent, err := loadPromptWithFallback(promptName, workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	model := extractModelFromAgent(agentContent)

	// Create executor and register providers
	executor := ai.NewExecutor(workspaceRoot)
	providers.RegisterBuiltIn(executor)

	// Build full prompt: agent instructions + user input
	fullPrompt := agentContent + "\n\n>>>>>>>>>>INPUT STARTS NOW<<<<<<<<<<<\n\n" + userPrompt

	// Prepare options
	var opts []ai.Option
	if model != "" {
		opts = append(opts, ai.WithModel(model))
	}

	// Execute with context
	ctx := context.Background()
	output, err := executor.Execute(ctx, fullPrompt, opts...)
	if err != nil {
		return "", fmt.Errorf("AI execution failed: %w", err)
	}

	// Trim output
	output = strings.TrimSpace(output)

	// Remove common agent initialization noise
	output = stripAgentNoise(output, promptName)

	return output, nil
}


// Note: standardCommitTypes is now defined in internal/constants.go as StandardCommitTypes

// commitPatternMatcher handles pattern matching for different agent types
type commitPatternMatcher struct {
	agentType string
}

func (m *commitPatternMatcher) findFirstValidLine(lines []string) int {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if m.agentType == "top-level" {
			if m.isValidTopLevelLine(trimmed) {
				return i
			}
		} else if m.agentType == "module" {
			if m.isValidModuleLine(trimmed) {
				return i
			}
		} else {
			// Unknown agent type - use generic heuristic
			if m.isValidTopLevelLine(trimmed) {
				return i
			}
		}
	}
	return -1
}

func (m *commitPatternMatcher) isValidTopLevelLine(trimmed string) bool {
	// Conventional commit with scope
	// Format: <type>(<scope>): <summary>
	// Examples: feat(multi-module): add feature, fix(cli): resolve bug
	for _, commitType := range commitmessage.StandardCommitTypes {
		// Look for <type>(<scope>): pattern
		pattern := commitType + "("
		if strings.HasPrefix(trimmed, pattern) {
			// Verify it has closing paren and colon
			if strings.Contains(trimmed, "):") {
				return true
			}
		}
	}

	return false
}

func (m *commitPatternMatcher) isValidModuleLine(trimmed string) bool {
	// Plain module name (followed by dashes on next line)
	// Must be:
	// - Not empty
	// - Not markdown headers (no #)
	// - Not dashes
	// - Not conventional commit format (no : in first part)
	// - Lowercase alphanumeric with dashes/underscores only
	if trimmed != "" &&
		!strings.HasPrefix(trimmed, "#") &&
		trimmed != "---" &&
		!strings.Contains(trimmed, ":") {
		// Check if it looks like a module name (lowercase alphanumeric with dashes)
		isModuleName := true
		for _, ch := range trimmed {
			if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
				isModuleName = false
				break
			}
		}
		if isModuleName {
			return true
		}
	}

	return false
}

// filterWithAntiCorruptionRules applies anti-corruption rules to filter noise
func filterWithAntiCorruptionRules(lines []string, rules *commitmessage.AntiCorruptionRules) string {
	var cleaned []string
	foundContent := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip bullet points (meta-descriptions) before any real content found
		if !foundContent && strings.HasPrefix(trimmed, "- ") {
			continue
		}

		// Skip markdown bold instructions (e.g., **CRITICAL**:)
		if !foundContent && strings.HasPrefix(trimmed, "**") {
			continue
		}

		// Skip lines matching forbidden prefixes
		shouldSkip := false
		for _, prefix := range rules.ForbiddenPrefixes {
			if strings.HasPrefix(trimmed, prefix) {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			continue
		}

		// Skip lines containing forbidden strings
		for _, contains := range rules.ForbiddenContains {
			if strings.Contains(trimmed, contains) {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			continue
		}

		// Skip lines containing forbidden emojis
		for _, emoji := range rules.ForbiddenEmojis {
			if strings.Contains(trimmed, emoji) {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			continue
		}

		// Skip agent signatures (end of content markers)
		for _, signature := range rules.AgentSignatures {
			if strings.HasPrefix(trimmed, signature) {
				return strings.TrimSpace(strings.Join(cleaned, "\n"))
			}
		}

		// Skip horizontal rules before content starts
		if !foundContent && (trimmed == "---" || trimmed == "___" || trimmed == "***") {
			continue
		}

		// Skip empty lines before content starts
		if !foundContent && trimmed == "" {
			continue
		}

		// Skip bold announcement lines
		if strings.HasPrefix(trimmed, "**") && strings.HasSuffix(trimmed, "**") {
			continue
		}

		// Content has started
		if trimmed != "" {
			foundContent = true
		}

		cleaned = append(cleaned, line)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

// filterWithHardcodedRules applies hardcoded noise patterns when YAML rules unavailable
func filterWithHardcodedRules(lines []string) string {
	var cleaned []string
	foundContent := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip common noise patterns (hardcoded)
		if strings.Contains(trimmed, "**Initialized and ready") ||
			strings.Contains(trimmed, "**INITIALIZED") ||
			strings.Contains(trimmed, "Loading project context") ||
			strings.Contains(trimmed, "ready to assist") ||
			(len(trimmed) > 0 && trimmed[0] > 127 && strings.Contains(trimmed, "**")) {
			continue
		}

		// Skip horizontal rules before content starts
		if !foundContent && (trimmed == "---" || trimmed == "___" || trimmed == "***") {
			continue
		}

		// Skip empty lines before content starts
		if !foundContent && trimmed == "" {
			continue
		}

		// Content has started
		if trimmed != "" {
			foundContent = true
		}

		cleaned = append(cleaned, line)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

// stripAgentNoise removes common initialization/greeting patterns from agent output
// by scanning for the first valid commit message line based on agent type.
//
// This function loads anti-corruption rules from contracts/commit-message/0.1.0/anti-corruption.yml
// and applies them to filter out conversational wrappers, initialization messages, and other noise.
func stripAgentNoise(output string, agentType string) string {
	// Remove wrapping markdown code fences if entire output is wrapped
	output = strings.TrimSpace(output)
	if strings.HasPrefix(output, "```") {
		// Remove opening fence (may have language identifier)
		lines := strings.Split(output, "\n")
		if len(lines) > 0 {
			lines = lines[1:] // Skip first line with ```
		}
		// Remove closing fence
		if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
			lines = lines[:len(lines)-1]
		}
		output = strings.Join(lines, "\n")
	}

	lines := strings.Split(output, "\n")
	matcher := &commitPatternMatcher{agentType: agentType}

	// Try to find first valid line using pattern matching
	firstValidLineIdx := matcher.findFirstValidLine(lines)
	if firstValidLineIdx != -1 {
		// Found valid line - cut everything before it
		return strings.TrimSpace(strings.Join(lines[firstValidLineIdx:], "\n"))
	}

	// No valid pattern found - use anti-corruption rules
	workspaceRoot, _ := repository.GetRepositoryRoot("")
	if workspaceRoot == "" {
		return filterWithHardcodedRules(lines)
	}

	rulesPath := filepath.Join(workspaceRoot, "contracts/commit-message/0.1.0/anti-corruption.yml")
	rules, err := commitmessage.LoadAntiCorruptionRules(rulesPath)
	if err != nil {
		return filterWithHardcodedRules(lines)
	}

	return filterWithAntiCorruptionRules(lines, rules)
}

// stripAgentNoiseHardcoded is a fallback function when anti-corruption rules can't be loaded
// This maintains backward compatibility if the YAML file is missing
func stripAgentNoiseHardcoded(output string, agentType string) string {
	lines := strings.Split(output, "\n")
	matcher := &commitPatternMatcher{agentType: agentType}

	// Try to find first valid line using pattern matching
	firstValidLineIdx := matcher.findFirstValidLine(lines)
	if firstValidLineIdx != -1 {
		return strings.TrimSpace(strings.Join(lines[firstValidLineIdx:], "\n"))
	}

	// No valid pattern - use hardcoded rules
	return filterWithHardcodedRules(lines)
}

// extractModelFromAgent parses agent frontmatter and extracts the model field
func extractModelFromAgent(agentContent string) string {
	lines := strings.Split(agentContent, "\n")
	inFrontmatter := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}

		if inFrontmatter && strings.HasPrefix(trimmed, "model:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}

// buildTopLevelContext creates context for the top-level commit message agent
func buildTopLevelContext(stagedFilesTable string, gitDiff string, diffStats string, affectedModules []string) string {
	var context bytes.Buffer

	// Module Count and List
	context.WriteString("## Module Count\n\n")
	if len(affectedModules) == 1 {
		context.WriteString("1 (single-module)\n\n")
	} else {
		context.WriteString(fmt.Sprintf("%d (multi-module)\n\n", len(affectedModules)))
	}

	// Affected Modules list
	context.WriteString("## Affected Modules\n\n")
	for _, module := range affectedModules {
		context.WriteString(fmt.Sprintf("- %s\n", module))
	}
	context.WriteString("\n")

	// Staged Files - shows all file-to-module mappings
	context.WriteString("## Staged Files\n\n")
	context.WriteString(stagedFilesTable)
	context.WriteString("\n\n")

	// Diff Stats - shows summary of changes
	if diffStats != "" {
		context.WriteString("## Diff Stats\n\n")
		context.WriteString(diffStats)
		context.WriteString("\n\n")
	}

	// Git Diff - shows all code changes
	context.WriteString("## Git Diff\n\n")
	context.WriteString("```diff\n")
	context.WriteString(gitDiff)
	context.WriteString("\n```\n")

	return context.String()
}

// buildModuleContext creates context for a single module section agent
func buildModuleContext(moduleName string, moduleFiles []repository.RepositoryFileWithModule, fullDiff string) string {
	var context bytes.Buffer

	// Module Name
	context.WriteString("## Module Name\n\n")
	context.WriteString(moduleName)
	context.WriteString("\n\n")

	// Files for this module
	context.WriteString("## Files\n\n")
	tb := render.NewTableBuilder().
		WithHeaders("File")

	for _, file := range moduleFiles {
		tb.AddRow(file.Name)
	}
	context.WriteString(tb.Build())
	context.WriteString("\n\n")

	// Git diff filtered to this module's files
	filteredDiff := filterDiffForModule(fullDiff, moduleFiles)
	context.WriteString("## Git Diff\n\n")
	context.WriteString("```diff\n")
	context.WriteString(filteredDiff)
	context.WriteString("\n```\n")

	return context.String()
}

// filterDiffForModule extracts only the diff chunks for files belonging to a specific module
func filterDiffForModule(fullDiff string, moduleFiles []repository.RepositoryFileWithModule) string {
	// Create a set of file names for quick lookup
	fileSet := make(map[string]bool)
	for _, file := range moduleFiles {
		fileSet[file.Name] = true
	}

	var result bytes.Buffer
	lines := strings.Split(fullDiff, "\n")

	inRelevantFile := false
	var currentChunk bytes.Buffer

	for _, line := range lines {
		// Detect diff file header
		if strings.HasPrefix(line, "diff --git") {
			// Save previous chunk if it was relevant
			if inRelevantFile && currentChunk.Len() > 0 {
				result.WriteString(currentChunk.String())
			}

			// Reset for new file
			currentChunk.Reset()
			inRelevantFile = false

			// Check if this file belongs to the module
			// Extract filename from "diff --git a/path/to/file b/path/to/file"
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				// Remove "a/" prefix from path
				filePath := strings.TrimPrefix(parts[2], "a/")
				if fileSet[filePath] {
					inRelevantFile = true
					currentChunk.WriteString(line + "\n")
				}
			}
		} else if inRelevantFile {
			currentChunk.WriteString(line + "\n")
		}
	}

	// Don't forget the last chunk
	if inRelevantFile && currentChunk.Len() > 0 {
		result.WriteString(currentChunk.String())
	}

	return strings.TrimSpace(result.String())
}

// addMissingModules adds stub sections for any modules that are missing from the commit message
func addMissingModules(commitMessage string, affectedModules []string, allFiles []repository.RepositoryFileWithModule, gitDiff string) string {
	// Parse existing commit message to find which modules already have sections
	foundModules := make(map[string]bool)
	lines := strings.Split(commitMessage, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect module name followed by dashes
		if i < len(lines)-1 {
			nextLine := strings.TrimSpace(lines[i+1])
			if isModuleNameLine(trimmed) && isDashesLine(nextLine) {
				foundModules[trimmed] = true
			}
		}
	}

	// Find missing modules
	var missingModules []string
	for _, module := range affectedModules {
		if !foundModules[module] {
			missingModules = append(missingModules, module)
		}
	}

	// If no missing modules, return original
	if len(missingModules) == 0 {
		return commitMessage
	}

	// Build sections for missing modules
	var result bytes.Buffer
	result.WriteString(commitMessage)

	// Check if commit message already ends with separator
	endsWithSeparator := strings.HasSuffix(strings.TrimRight(commitMessage, " \t\n"), "---")

	for i, module := range missingModules {
		// Get files for this module
		var moduleFiles []repository.RepositoryFileWithModule
		for _, file := range allFiles {
			for _, fileModule := range file.Modules {
				if fileModule == module {
					moduleFiles = append(moduleFiles, file)
					break
				}
			}
		}

		// Build file list for subject line
		var fileNames []string
		for _, file := range moduleFiles {
			fileNames = append(fileNames, file.Name)
		}
		filesStr := "CHANGED FILES"
		if len(fileNames) > 0 {
			filesStr = strings.Join(fileNames, ", ")
			// Truncate if too long
			if len(filesStr) > 40 {
				filesStr = filesStr[:37] + "..."
			}
		}

		// Add separator before stub sections
		if i == 0 {
			// First stub: only add separator if commit message doesn't already end with one
			if !endsWithSeparator {
				result.WriteString("\n\n---\n\n")
			} else {
				result.WriteString("\n\n")
			}
		} else {
			// Subsequent stubs: always add separator to separate from previous stub
			result.WriteString("\n\n---\n\n")
		}

		// Module header: plain text with dashes
		result.WriteString(fmt.Sprintf("%s\n", module))
		// Add dashes (at least as long as module name)
		dashes := strings.Repeat("-", len(module))
		if len(dashes) < 9 {
			dashes = "---------"
		}
		result.WriteString(fmt.Sprintf("%s\n", dashes))
		result.WriteString(fmt.Sprintf("%s: chore: %s\n\n", module, filesStr))
		result.WriteString("Module changes not described by AI agent.\n")
	}

	return result.String()
}

// combineCommitSections combines top-level section and module sections into final commit message
func combineCommitSections(topLevel string, moduleSections []string) string {
	var result bytes.Buffer

	// Top-level section (trim trailing whitespace)
	result.WriteString(strings.TrimRight(topLevel, " \t\n"))

	// Only add module sections if there are any (multi-module commits only)
	if len(moduleSections) > 0 {
		// Filter out empty sections first
		var nonEmptySections []string
		for _, section := range moduleSections {
			trimmed := strings.TrimSpace(section)
			if trimmed != "" {
				nonEmptySections = append(nonEmptySections, section)
			}
		}

		// Only proceed if we have non-empty sections
		if len(nonEmptySections) > 0 {
			result.WriteString("\n\n")

			// Module sections with --- separators
			for i, section := range nonEmptySections {
				// Trim trailing whitespace from each section
				trimmedSection := strings.TrimRight(section, " \t\n")
				result.WriteString(trimmedSection)

				// Add separator between modules (but not after the last one)
				if i < len(nonEmptySections)-1 {
					result.WriteString("\n\n---\n\n")
				}
			}
		}
	}

	return result.String()
}

// isModuleNameLine checks if a line looks like a module name
func isModuleNameLine(s string) bool {
	if s == "" || len(s) > commitmessage.MaxModuleNameLength {
		return false
	}

	for _, ch := range s {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			return false
		}
	}

	return true
}

// isDashesLine checks if a line consists only of dashes
func isDashesLine(s string) bool {
	if len(s) < commitmessage.MinDashesLength {
		return false
	}

	for _, ch := range s {
		if ch != '-' {
			return false
		}
	}

	return true
}

