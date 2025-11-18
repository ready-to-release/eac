package commit

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/render"
	"github.com/ready-to-release/eac/src/core/repository"
)

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
