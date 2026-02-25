package squashmessage

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/commands/repository/create/aiutil"
	squashmessageinternal "github.com/ready-to-release/eac/go/commands/repository/get/squash-message/internal"
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

// filterNoiseDiff strips diff sections for files that have no semantic value for AI
// generation: go.sum, package-lock.json, and any *.lock files. These files consist
// entirely of hash lines and content-address entries; including them wastes the diff
// budget without helping the AI understand what changed.
//
// The filter splits on "diff --git" headers, drops sections whose file path ends with a
// noise suffix, then rejoins the remaining sections.
func filterNoiseDiff(diff string) string {
	noiseSuffixes := []string{".sum", ".lock", "package-lock.json"}

	sections := strings.Split(diff, "diff --git ")
	var kept []string
	for i, section := range sections {
		if i == 0 {
			// Anything before the first "diff --git" header (usually empty)
			kept = append(kept, section)
			continue
		}

		// The first line of a section is the header, e.g. "a/go.sum b/go.sum"
		firstLine := section
		if idx := strings.Index(section, "\n"); idx != -1 {
			firstLine = section[:idx]
		}

		isNoise := false
		for _, suffix := range noiseSuffixes {
			if strings.HasSuffix(firstLine, suffix) {
				isNoise = true
				break
			}
		}

		if !isNoise {
			kept = append(kept, section)
		}
	}

	result := strings.Join(kept, "diff --git ")
	return strings.TrimSpace(result)
}

// buildSquashContext builds the context string for AI generation.
func buildSquashContext(currentBranch, baseBranch string, commits []git.CommitInfo, files []repository.RepositoryFileWithModule, diff, diffStats string, affectedModules []string) string {
	var buf strings.Builder

	// Branch information
	buf.WriteString("## Branch Information\n\n")
	buf.WriteString(fmt.Sprintf("Base branch: %s\n", baseBranch))
	buf.WriteString(fmt.Sprintf("Current branch: %s\n", currentBranch))
	buf.WriteString(fmt.Sprintf("Commits ahead: %d\n\n", len(commits)))

	// Commit history (capped at MaxPromptCommits to control prompt size)
	buf.WriteString("## Commit History\n\n")
	shownCommits := commits
	if len(commits) > squashmessageinternal.MaxPromptCommits {
		shownCommits = commits[:squashmessageinternal.MaxPromptCommits]
	}
	for i, commit := range shownCommits {
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
	if len(commits) > squashmessageinternal.MaxPromptCommits {
		buf.WriteString(fmt.Sprintf("... and %d more commits (subjects omitted for brevity)\n\n", len(commits)-squashmessageinternal.MaxPromptCommits))
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

	// Files table (capped at MaxPromptFiles)
	buf.WriteString("## Changed Files\n\n")
	buf.WriteString(buildFilesTable(files))
	buf.WriteString("\n\n")

	// Diff stats (truncated to MaxPromptDiffStatsSize)
	if diffStats != "" {
		buf.WriteString("## Diff Stats\n\n")
		if len(diffStats) > squashmessageinternal.MaxPromptDiffStatsSize {
			truncatedStats := diffStats[:squashmessageinternal.MaxPromptDiffStatsSize]
			lastNewline := strings.LastIndex(truncatedStats, "\n")
			if lastNewline > 0 {
				truncatedStats = truncatedStats[:lastNewline]
			}
			buf.WriteString(truncatedStats)
			buf.WriteString("\n... [STATS TRUNCATED]\n")
		} else {
			buf.WriteString(diffStats)
		}
		buf.WriteString("\n\n")
	}

	// Cumulative diff (noise-filtered and truncated)
	filteredDiff := filterNoiseDiff(diff)

	buf.WriteString("## Cumulative Diff\n\n")
	buf.WriteString("```diff\n")
	if len(filteredDiff) > squashmessageinternal.MaxPromptDiffSize {
		truncatedDiff := filteredDiff[:squashmessageinternal.MaxPromptDiffSize]
		// Find last complete line to avoid cutting mid-line
		lastNewline := strings.LastIndex(truncatedDiff, "\n")
		if lastNewline > 0 {
			truncatedDiff = truncatedDiff[:lastNewline]
		}
		buf.WriteString(truncatedDiff)
		buf.WriteString("\n\n... [DIFF TRUNCATED - ")
		buf.WriteString(fmt.Sprintf("%d KB of %d KB shown", squashmessageinternal.MaxPromptDiffSize/1024, len(filteredDiff)/1024))
		buf.WriteString("] ...\n")
		buf.WriteString("Note: Use diff stats and file list above to understand full scope of changes.\n")
	} else {
		buf.WriteString(filteredDiff)
	}
	buf.WriteString("\n```\n")

	return buf.String()
}

// buildFilesTable builds a markdown table of changed files, capped at MaxPromptFiles rows.
func buildFilesTable(files []repository.RepositoryFileWithModule) string {
	var buf strings.Builder
	buf.WriteString("| File | Modules |\n")
	buf.WriteString("|------|--------|\n")
	shown := files
	if len(files) > squashmessageinternal.MaxPromptFiles {
		shown = files[:squashmessageinternal.MaxPromptFiles]
	}
	for _, f := range shown {
		modulesStr := strings.Join(f.Modules, ", ")
		if modulesStr == "" {
			modulesStr = "NONE"
		}
		buf.WriteString(fmt.Sprintf("| %s | %s |\n", f.Name, modulesStr))
	}
	if len(files) > squashmessageinternal.MaxPromptFiles {
		buf.WriteString(fmt.Sprintf("| ... and %d more files | (omitted) |\n", len(files)-squashmessageinternal.MaxPromptFiles))
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
