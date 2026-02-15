package commitmessage

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	commitmessageinternal "github.com/ready-to-release/eac/go/commands/repository/create/commit-message/internal"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/repository/reports"
)

// executionConfig holds configuration for the commit AI command.
type executionConfig struct {
	workspaceRoot   string
	debug           bool
	autoCommit      bool
	stagedFiles     []repository.RepositoryFileWithModule
	affectedModules []string
	gitDiff         string
}

// buildExecutionContext gathers staged files, affected modules, and git diff.
func buildExecutionContext(deps *Deps, workspaceRoot string, debug bool) (*executionConfig, string, string, error) {
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
	affectedModules := extractAffectedModules(report)

	// Get git diff and stats
	log.Debug("buildExecutionContext: calling getGitDiffAndStats")
	gitDiff, diffStats, err := getGitDiffAndStats(deps, workspaceRoot)
	if err != nil {
		return nil, "", "", err
	}

	log.Debugf("Affected modules count: %d", len(affectedModules))
	for i, mod := range affectedModules {
		log.Debugf("  %d. %s", i+1, mod)
	}

	cfg := &executionConfig{
		workspaceRoot:   workspaceRoot,
		debug:           debug,
		stagedFiles:     report.AllFiles,
		affectedModules: affectedModules,
		gitDiff:         gitDiff,
	}

	return cfg, stagedFilesTable, diffStats, nil
}

// getStagedFilesReport retrieves staged files and builds a table representation.
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

// extractAffectedModules extracts and validates unique module names from file report.
func extractAffectedModules(report *reports.FilesModulesReport) []string {
	moduleSet := make(map[string]bool)
	for _, file := range report.AllFiles {
		for _, module := range file.Modules {
			// Validate module name: only lowercase letters, numbers, dashes, underscores
			if isValidModuleName(module) {
				moduleSet[module] = true
			} else {
				log.Warnf("skipping invalid module name: module=%s", module)
			}
		}
	}

	var affectedModules []string
	for module := range moduleSet {
		affectedModules = append(affectedModules, module)
	}

	return affectedModules
}

// getGitDiffAndStats retrieves git diff and diff stats for staged changes.
func getGitDiffAndStats(deps *Deps, workspaceRoot string) (string, string, error) {
	log.Debug("getGitDiffAndStats: start")
	log.Debug("getGitDiffAndStats: calling getGitRepo")
	repo, err := deps.GetGitRepo(workspaceRoot)
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
		log.Warnf("failed to get diff stats: %v", err)
		diffStats = ""
	}
	log.Debug("getGitDiffAndStats: StagedDiffStats complete")

	return diffOutput, strings.TrimSpace(diffStats), nil
}

// isValidModuleName checks if a module name is valid (lowercase letters, numbers, dashes, underscores only)
// Rejects paths with slashes or other special characters.
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

// buildTopLevelContext creates context for the top-level commit message agent.
func buildTopLevelContext(stagedFilesTable, gitDiff, diffStats string, affectedModules []string) string {
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

	// Git Diff - shows all code changes (truncated if too large for prompt)
	context.WriteString("## Git Diff\n\n")
	context.WriteString("```diff\n")
	if len(gitDiff) > commitmessageinternal.MaxPromptDiffSize {
		// Truncate diff to avoid exceeding Claude CLI prompt limits
		truncatedDiff := gitDiff[:commitmessageinternal.MaxPromptDiffSize]
		// Find last complete line to avoid cutting mid-line
		lastNewline := strings.LastIndex(truncatedDiff, "\n")
		if lastNewline > 0 {
			truncatedDiff = truncatedDiff[:lastNewline]
		}
		context.WriteString(truncatedDiff)
		context.WriteString("\n\n... [DIFF TRUNCATED - ")
		context.WriteString(fmt.Sprintf("%d KB of %d KB shown", commitmessageinternal.MaxPromptDiffSize/1024, len(gitDiff)/1024))
		context.WriteString("] ...\n")
		context.WriteString("Note: Use diff stats and file list above to understand full scope of changes.\n")
	} else {
		context.WriteString(gitDiff)
	}
	context.WriteString("\n```\n")

	return context.String()
}

// buildModuleContext creates context for a single module section agent.
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

	// Git diff filtered to this module's files (truncated if too large for prompt)
	filteredDiff := filterDiffForModule(fullDiff, moduleFiles)
	context.WriteString("## Git Diff\n\n")
	context.WriteString("```diff\n")
	if len(filteredDiff) > commitmessageinternal.MaxPromptDiffSize {
		// Truncate diff to avoid exceeding Claude CLI prompt limits
		truncatedDiff := filteredDiff[:commitmessageinternal.MaxPromptDiffSize]
		// Find last complete line to avoid cutting mid-line
		lastNewline := strings.LastIndex(truncatedDiff, "\n")
		if lastNewline > 0 {
			truncatedDiff = truncatedDiff[:lastNewline]
		}
		context.WriteString(truncatedDiff)
		context.WriteString("\n\n... [DIFF TRUNCATED - ")
		context.WriteString(fmt.Sprintf("%d KB of %d KB shown", commitmessageinternal.MaxPromptDiffSize/1024, len(filteredDiff)/1024))
		context.WriteString("] ...\n")
		context.WriteString("Note: Use file list above to understand full scope of changes.\n")
	} else {
		context.WriteString(filteredDiff)
	}
	context.WriteString("\n```\n")

	return context.String()
}

// filterDiffForModule extracts only the diff chunks for files belonging to a specific module.
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
