// Command: commit-ai
// Description: Show staged changes with their module mappings for AI commit message generation
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
	"github.com/ready-to-release/eac/src/repository/reports"
)

func init() {
	Register("commit-ai", CommitAI)
}

func CommitAI() int {
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

	// Get git diff for staged changes (do not print anything yet)
	diffCmd := exec.Command("git", "diff", "--staged")
	diffCmd.Dir = "../.."
	diffOutput, err := diffCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting git diff: %v\n", err)
		return 1
	}
	gitDiff := string(diffOutput)

	// Build context for agent
	contextPrompt := buildAgentContext(stagedFilesTable, gitDiff)

	// LEVER 2: Invoke commit-message-generator agent
	var msgOutput string
	err = commitmessage.WithProgress("🤖 Analyzing changes and generating commit message...", func() error {
		output, err := callClaudeAgentAPI("../../.claude/agents/commit-message-generator.md", contextPrompt)
		msgOutput = output
		return err
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error running commit-message-generator: %v\n", err)
		return 1
	}

	// LEVER 3: Inject files table placeholder (before cleanup)
	rawAgentOutput := msgOutput // Save raw output for debugging
	msgOutput = strings.Replace(msgOutput, "\\filetable-placeholder", stagedFilesTable, 1)

	// LEVER 4: Auto-cleanup before verification (silent)
	cleanedOutput := commitmessage.AutoCleanup(msgOutput)

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

	_ = rawAgentOutput // Use the variable to avoid unused error

	// Output for VSCode extension to detect
	fmt.Println(">>>>>>OUTPUT START<<<<<<")
	fmt.Println(cleanedOutput)
	fmt.Println("\n---\n")

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
				strings.HasPrefix(trimmed, "Now I") ||
				strings.HasPrefix(trimmed, "You are now") ||
				strings.HasPrefix(trimmed, "The title ") ||
				strings.HasPrefix(trimmed, "The corrected ") ||
				strings.HasPrefix(trimmed, "The generated ") ||
				strings.HasPrefix(trimmed, "The changes ") ||
				strings.HasPrefix(trimmed, "After reviewing") ||
				strings.Contains(trimmed, "ready to assist") || // Skip assistant messages
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
			if trimmed == "```" || trimmed == "```markdown" {
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

// buildAgentContext creates the full context for agents including staged files table and git diff
func buildAgentContext(stagedFilesTable string, gitDiff string) string {
	var context bytes.Buffer

	context.WriteString("## Staged Files with Module Mappings\n\n")
	context.WriteString(stagedFilesTable)
	context.WriteString("\n\n---\n\n")

	context.WriteString("## Git Diff (Staged Changes)\n\n")
	context.WriteString("```diff\n")
	context.WriteString(gitDiff)
	context.WriteString("```\n\n")

	context.WriteString("---\n\n")
	context.WriteString("**INSTRUCTIONS:**\n")
	context.WriteString("- The staged files table shows which files are changed and their module mappings\n")
	context.WriteString("- The git diff shows the actual code changes\n")
	context.WriteString("- Extract code snippets from the diff (lines starting with +) for the code extract section\n")
	context.WriteString("- Focus on the most significant changes (5-15 lines per module)\n\n")

	return context.String()
}
