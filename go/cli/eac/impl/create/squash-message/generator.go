package squashmessage

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/create/aiutil"
	coreai "github.com/ready-to-release/eac/go/core/ai"
	"github.com/ready-to-release/eac/go/core/git"
	"github.com/ready-to-release/eac/go/core/repository"
)

// logDebugArtifact delegates to the shared AI utility for debug artifact logging.
func logDebugArtifact(label, content string) {
	aiutil.LogDebugArtifact(log, label, content)
}

// extractAffectedModules extracts unique module names from file list.
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

// buildSquashContext builds the context string for AI generation.
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

// buildFilesTable builds a markdown table of changed files.
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

// generateTopLevelMessage generates the top-level commit message using AI.
func generateTopLevelMessage(workspaceRoot, promptContext string) (string, error) {
	// Check for mock response from file-based mock system (subprocess testing)
	if mock, ok := coreai.GetMockResponse("squash-message"); ok {
		// Format mock JSON to squash message text
		formattedResult, err := FormatSquashMessage(mock)
		if err != nil {
			return "", fmt.Errorf("failed to format mock squash message: %w", err)
		}
		return strings.TrimSpace(formattedResult), nil
	}

	// Load squash prompt template with three-tier priority:
	// 1. Command flag (not applicable - internal function)
	// 2. Team override (.eac/templates/ai/commit-message/squash.md)
	// 3. System default (templates/ai/commit-message/squash.md)
	loader := coreai.NewContractLoader(workspaceRoot, coreai.TypeCommitMessage, "")
	promptTemplate, _, err := loader.LoadPrompt("squash", "")
	if err != nil {
		return "", fmt.Errorf("failed to load squash.md template: %w", err)
	}

	// Build full prompt
	prompt := string(promptTemplate) + "\n\n" + promptContext

	logDebugArtifact("SQUASH-PROMPT", prompt)

	// Execute AI with shared generation pipeline
	retryResult, err := aiutil.ExecuteGeneration(aiutil.GenerationParams{
		WorkspaceRoot:  workspaceRoot,
		Prompt:         prompt,
		TypeName:       coreai.TypeSquashMessage,
		SchemaFilename: "squash-message.schema.json",
	})
	if err != nil {
		return "", err
	}

	jsonResult := retryResult.Output
	logDebugArtifact("SQUASH-AI-JSON-RESPONSE", jsonResult)

	// Format JSON -> squash commit text (no AI)
	formattedResult, err := FormatSquashMessage(jsonResult)
	if err != nil {
		return "", fmt.Errorf("failed to format squash message: %w", err)
	}

	return strings.TrimSpace(formattedResult), nil
}

// generateModuleSections generates per-module sections (reuse commit-message logic).
func generateModuleSections(workspaceRoot string, affectedModules []string, files []repository.RepositoryFileWithModule, diff string) (map[string]string, error) {
	// For now, return empty map - module sections can be added later if needed
	// This keeps the implementation simple
	return make(map[string]string), nil
}

// assembleMessage assembles the final message from parts.
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
