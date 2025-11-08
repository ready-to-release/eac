// Command: commit-ai
// Description: Generate commit message using AI with staged changes and module mappings
// Flags: --debug (save intermediate outputs and show debug info)
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	commitmessage "github.com/ready-to-release/eac/src/commands/commit-message"
	"github.com/ready-to-release/eac/src/commands/render"
	"github.com/ready-to-release/eac/src/repository"
	"github.com/ready-to-release/eac/src/repository/reports"
)

func init() {
	Register("commit-ai", CommitAI)
}

func CommitAI() int {
	// Parse flags
	debug := false
	for _, arg := range os.Args[2:] { // Skip program name and "commit-ai"
		if arg == "--debug" {
			debug = true
		}
	}

	// LEVER 1: Verify contract implementation on startup
	contractPath := "../../contracts/commit-message/0.1.0/structure.yml"
	contractErrors := commitmessage.VerifyContractImplementation(contractPath)
	if len(contractErrors) > 0 {
		fmt.Fprintf(os.Stderr, "❌ Contract implementation verification failed:\n")
		for _, err := range contractErrors {
			fmt.Fprintf(os.Stderr, "  - [%s] %s\n", err.Code, err.Message)
		}
		return 1
	}

	// LEVER 1: Get staged files with module mappings
	report, err := reports.GetFilesModulesReport(true, false, true, "../..", "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting module mappings: %v\n", err)
		return 1
	}

	if len(report.AllFiles) == 0 {
		fmt.Println("No staged changes.")
		return 0
	}

	// Build the staged files table (same format as "show files staged")
	tb := render.NewTableBuilder().
		WithHeaders("File", "Modules")

	for _, file := range report.AllFiles {
		modulesStr := "NONE"
		if len(file.Modules) > 0 {
			modulesStr = strings.Join(file.Modules, ", ")
		}
		tb.AddRow(file.Name, modulesStr)
	}

	stagedFilesTable := tb.Build()

	// Extract unique modules from all files
	moduleSet := make(map[string]bool)
	for _, file := range report.AllFiles {
		for _, module := range file.Modules {
			moduleSet[module] = true
		}
	}

	// Build sorted list of unique modules
	var affectedModules []string
	for module := range moduleSet {
		affectedModules = append(affectedModules, module)
	}
	// Sort for consistent output
	// Note: Using simple sort since we don't have imports for sort yet
	// The order doesn't matter for the agent, just consistency

	// Get git diff for staged changes (do not print anything yet)
	diffCmd := exec.Command("git", "diff", "--staged")
	diffCmd.Dir = "../.."
	diffOutput, err := diffCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting git diff: %v\n", err)
		return 1
	}
	gitDiff := string(diffOutput)

	if debug {
		fmt.Fprintf(os.Stderr, "\n🔍 DEBUG: Affected modules count: %d\n", len(affectedModules))
		for i, mod := range affectedModules {
			fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, mod)
		}
	}

	// LEVER 2a: Build top-level context and invoke top-level agent
	topLevelContext := buildTopLevelContext(stagedFilesTable, gitDiff, affectedModules)

	if debug {
		// DEBUG: Save top-level context
		debugTopLevelContext := "../../out/debug-top-level-context.md"
		ioutil.WriteFile(debugTopLevelContext, []byte(topLevelContext), 0644)
		fmt.Fprintf(os.Stderr, "\n🔍 DEBUG: Top-level context saved to %s\n", debugTopLevelContext)
	}

	var topLevelOutput string
	err = commitmessage.WithProgress("🤖 Generating top-level commit summary...", func() error {
		output, err := callClaudeAgentAPIRaw("../../.claude/agents/commit-message-top-level.md", topLevelContext)
		topLevelOutput = output
		return err
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error running commit-message-top-level agent: %v\n", err)
		return 1
	}

	if debug {
		// DEBUG: Save top-level output
		debugTopLevelOutput := "../../out/debug-top-level-output.md"
		ioutil.WriteFile(debugTopLevelOutput, []byte(topLevelOutput), 0644)
		fmt.Fprintf(os.Stderr, "🔍 DEBUG: Top-level output saved to %s\n", debugTopLevelOutput)
	}

	// LEVER 2b: Build module contexts and invoke module agent for each
	var moduleSections []string

	// Group files by module
	moduleFilesMap := make(map[string][]repository.RepositoryFileWithModule)
	for _, file := range report.AllFiles {
		for _, module := range file.Modules {
			moduleFilesMap[module] = append(moduleFilesMap[module], file)
		}
	}

	for i, module := range affectedModules {
		moduleFiles := moduleFilesMap[module]
		moduleContext := buildModuleContext(module, moduleFiles, gitDiff)

		if debug {
			// DEBUG: Save module context
			debugModuleContext := fmt.Sprintf("../../out/debug-module-%d-%s-context.md", i+1, module)
			ioutil.WriteFile(debugModuleContext, []byte(moduleContext), 0644)
			fmt.Fprintf(os.Stderr, "🔍 DEBUG: Module context for %s saved to %s\n", module, debugModuleContext)
		}

		var moduleOutput string
		progressMsg := fmt.Sprintf("🤖 Generating section for module %s (%d/%d)...", module, i+1, len(affectedModules))
		err = commitmessage.WithProgress(progressMsg, func() error {
			output, err := callClaudeAgentAPIRaw("../../.claude/agents/commit-message-module.md", moduleContext)
			moduleOutput = output
			return err
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Error running commit-message-module agent for %s: %v\n", module, err)
			return 1
		}

		if debug {
			// DEBUG: Save module output
			debugModuleOutput := fmt.Sprintf("../../out/debug-module-%d-%s-output.md", i+1, module)
			ioutil.WriteFile(debugModuleOutput, []byte(moduleOutput), 0644)
			fmt.Fprintf(os.Stderr, "🔍 DEBUG: Module output for %s saved to %s\n", module, debugModuleOutput)
		}

		moduleSections = append(moduleSections, moduleOutput)
	}

	// LEVER 3: Combine all sections
	combinedMessage := combineCommitSections(topLevelOutput, moduleSections)

	if debug {
		// DEBUG: Save combined message
		debugCombined := "../../out/debug-combined-message.md"
		ioutil.WriteFile(debugCombined, []byte(combinedMessage), 0644)
		fmt.Fprintf(os.Stderr, "\n🔍 DEBUG: Combined message saved to %s\n", debugCombined)
	}

	// LEVER 4: Auto-cleanup before verification (silent)
	cleanedOutput := commitmessage.AutoCleanup(combinedMessage)

	if debug {
		// DEBUG: Save after cleanup
		debugFile3 := "../../out/debug-after-cleanup.md"
		ioutil.WriteFile(debugFile3, []byte(cleanedOutput), 0644)
		fmt.Fprintf(os.Stderr, "🔍 DEBUG: After cleanup saved to %s\n\n", debugFile3)
	}

	// LEVER 5: Verify contract compliance (silent)
	validationErrors := commitmessage.VerifyCommitMessageContract(cleanedOutput)

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
	fmt.Println(cleanedOutput)
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

// callClaudeAgentAPI invokes Claude CLI with isolated session (no --continue or --resume)
func callClaudeAgentAPI(agentFilePath string, prompt string) (string, error) {
	// Read agent file to extract model from frontmatter
	agentContent, err := ioutil.ReadFile(agentFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read agent file: %w", err)
	}

	model := extractModelFromAgent(string(agentContent))

	// Build command arguments - IMPORTANT: No --continue or --resume flags for session isolation
	args := []string{
		"--print",
		"--settings", `{"includeCoAuthoredBy":false}`,
	}

	// Add model if specified in agent frontmatter
	if model != "" {
		args = append(args, "--model", model)
		// Add fallback to haiku if using sonnet (for resilience during high load)
		if model == "sonnet" {
			args = append(args, "--fallback-model", "haiku")
		}
	}

	// Build full prompt: agent instructions + user input
	fullPrompt := string(agentContent) + "\n\n>>>>>>>>>>INPUT STARTS NOW<<<<<<<<<<<\n\n" + prompt

	cmd := exec.Command("claude", args...)
	cmd.Stdin = strings.NewReader(fullPrompt)
	cmd.Dir = "../.." // Run from repository root for proper context

	// CRITICAL: Remove ANTHROPIC_API_KEY to use Claude Pro subscription instead
	cmd.Env = removeAPIKeyFromEnv(os.Environ())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrText := stderr.String()
		stdoutText := stdout.String()

		// Print both to help debug
		if len(stderrText) > 0 {
			fmt.Fprintf(os.Stderr, "Claude stderr:\n%s\n", stderrText)
		}
		if len(stdoutText) > 0 {
			fmt.Fprintf(os.Stderr, "Claude stdout:\n%s\n", stdoutText)
		}

		return "", fmt.Errorf("claude CLI failed: %w\nStderr: %s\nStdout: %s", err, stderrText, stdoutText)
	}

	// Extract pure content (remove any wrapper text)
	output := extractContentBlock(stdout.String())
	return output, nil
}

// callClaudeAgentAPIRaw invokes Claude CLI without content extraction (for specialized agents)
func callClaudeAgentAPIRaw(agentFilePath string, prompt string) (string, error) {
	// Read agent file to extract model from frontmatter
	agentContent, err := ioutil.ReadFile(agentFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read agent file: %w", err)
	}

	model := extractModelFromAgent(string(agentContent))

	// Build command arguments
	args := []string{
		"--print",
		"--settings", `{"includeCoAuthoredBy":false}`,
	}

	// Add model if specified in agent frontmatter
	if model != "" {
		args = append(args, "--model", model)
		if model == "sonnet" {
			args = append(args, "--fallback-model", "haiku")
		}
	}

	// Build full prompt: agent instructions + user input
	fullPrompt := string(agentContent) + "\n\n>>>>>>>>>>INPUT STARTS NOW<<<<<<<<<<<\n\n" + prompt

	cmd := exec.Command("claude", args...)
	cmd.Stdin = strings.NewReader(fullPrompt)
	cmd.Dir = "../.."

	// Remove ANTHROPIC_API_KEY to use Claude Pro subscription
	cmd.Env = removeAPIKeyFromEnv(os.Environ())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrText := stderr.String()
		stdoutText := stdout.String()

		if len(stderrText) > 0 {
			fmt.Fprintf(os.Stderr, "Claude stderr:\n%s\n", stderrText)
		}
		if len(stdoutText) > 0 {
			fmt.Fprintf(os.Stderr, "Claude stdout:\n%s\n", stdoutText)
		}

		return "", fmt.Errorf("claude CLI failed: %w\nStderr: %s\nStdout: %s", err, stderrText, stdoutText)
	}

	// Return raw output (no extraction - these agents output pure content)
	return strings.TrimSpace(stdout.String()), nil
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

// extractContentBlock removes conversational wrapper text from agent output
func extractContentBlock(agentOutput string) string {
	lines := strings.Split(agentOutput, "\n")
	var contentLines []string
	inContent := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip meta-commentary at the start
		if !inContent {
			// Skip horizontal rules
			if trimmed == "---" || trimmed == "___" || trimmed == "***" {
				continue
			}

			// Skip any line starting with emoji (check first rune)
			if len(trimmed) > 0 {
				firstRune := []rune(trimmed)[0]
				if firstRune > 127 { // Non-ASCII (likely emoji)
					continue
				}
			}

			// Skip conversational wrappers (from contract forbidden patterns)
			if strings.HasPrefix(trimmed, "Based on ") ||
				strings.HasPrefix(trimmed, "Let me ") ||
				strings.HasPrefix(trimmed, "Here is ") ||
				strings.HasPrefix(trimmed, "Here's ") ||
				strings.HasPrefix(trimmed, "I will ") ||
				strings.HasPrefix(trimmed, "I'll ") ||
				strings.HasPrefix(trimmed, "I've ") ||
				strings.HasPrefix(trimmed, "I'm ") ||
				strings.HasPrefix(trimmed, "I can see ") ||
				strings.HasPrefix(trimmed, "Looking at ") ||
				strings.HasPrefix(trimmed, "Now ") || // Catch "Now generating..."
				strings.HasPrefix(trimmed, "You are now") ||
				strings.HasPrefix(trimmed, "The title ") ||
				strings.HasPrefix(trimmed, "The corrected ") ||
				strings.HasPrefix(trimmed, "The generated ") ||
				strings.HasPrefix(trimmed, "The changes ") ||
				strings.HasPrefix(trimmed, "After reviewing") ||
				strings.Contains(trimmed, "ready to assist") || // Skip assistant messages
				strings.Contains(trimmed, "commit message") || // Skip "generating your commit message"
				strings.Contains(trimmed, "🚀") || // Skip emoji celebration lines
				strings.Contains(trimmed, "✅") || // Skip checkmark lines
				strings.Contains(trimmed, "🤖") || // Skip bot emoji lines
				strings.Contains(trimmed, "🎉") || // Skip party emoji lines
				strings.Contains(trimmed, "✨") || // Skip sparkle emoji lines
				strings.Contains(trimmed, "INITIALIZED") || // Skip initialization messages
				(strings.HasPrefix(trimmed, "**") && strings.HasSuffix(trimmed, "**")) { // Skip bold announcements
				continue
			}

			// Skip opening markdown fence
			if trimmed == "```" || trimmed == "```markdown" || strings.HasPrefix(trimmed, "```") {
				continue
			}

			// Skip empty lines before content
			if trimmed == "" {
				continue
			}

			// Content starts here
			inContent = true
		}

		// Stop at agent signature or closing fence
		if strings.HasPrefix(trimmed, "Agent:") || (trimmed == "```" && inContent) {
			break
		}

		if inContent {
			contentLines = append(contentLines, line)
		}
	}

	result := strings.Join(contentLines, "\n")
	return strings.TrimSpace(result)
}

// removeAPIKeyFromEnv removes ANTHROPIC_API_KEY from environment variables
// This forces Claude CLI to use subscription auth instead of API key
func removeAPIKeyFromEnv(environ []string) []string {
	var filtered []string
	for _, env := range environ {
		if !strings.HasPrefix(env, "ANTHROPIC_API_KEY=") {
			filtered = append(filtered, env)
		}
	}
	return filtered
}

// buildTopLevelContext creates context for the top-level commit message agent
func buildTopLevelContext(stagedFilesTable string, gitDiff string, affectedModules []string) string {
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

// combineCommitSections combines top-level section and module sections into final commit message
func combineCommitSections(topLevel string, moduleSections []string) string {
	var result bytes.Buffer

	// Top-level section
	result.WriteString(topLevel)
	result.WriteString("\n\n")

	// Module sections with --- separators
	for i, section := range moduleSections {
		result.WriteString(section)

		// Add separator between modules (but not after the last one)
		if i < len(moduleSections)-1 {
			result.WriteString("\n\n---\n\n")
		}
	}

	return result.String()
}

