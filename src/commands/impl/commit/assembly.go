package commit

import (
	"bytes"
	"fmt"
	"strings"

	commitmessage "github.com/ready-to-release/eac/src/commands/impl/commit/internal"
	"github.com/ready-to-release/eac/src/core/repository"
)

// combineCommitSections combines top-level section and module sections into final commit message
func combineCommitSections(topLevel string, moduleSections []string) string {
	var result bytes.Buffer

	// Top-level section (trim trailing whitespace)
	result.WriteString(strings.TrimRight(topLevel, " \t\n"))

	// Only add module sections if there are any (multi-module commits only)
	if len(moduleSections) > 0 {
		// Filter out empty sections and deduplicate by module name
		seenModules := make(map[string]bool)
		var uniqueSections []string

		for _, section := range moduleSections {
			trimmed := strings.TrimSpace(section)
			if trimmed == "" {
				continue
			}

			// Extract module name from section (first line)
			moduleName := extractModuleName(trimmed)
			if moduleName == "" {
				// Can't extract module name, include it anyway
				uniqueSections = append(uniqueSections, section)
				continue
			}

			// Skip if we've already seen this module
			if seenModules[moduleName] {
				continue
			}

			seenModules[moduleName] = true
			uniqueSections = append(uniqueSections, section)
		}

		// Only proceed if we have unique non-empty sections
		if len(uniqueSections) > 0 {
			result.WriteString("\n\n")

			// Module sections separated by horizontal rules
			for i, section := range uniqueSections {
				// Trim trailing whitespace from each section
				trimmedSection := strings.TrimRight(section, " \t\n")
				result.WriteString(trimmedSection)

				// Add separator between modules (but not after the last one)
				if i < len(uniqueSections)-1 {
					result.WriteString("\n\n---\n\n")
				}
			}
		}
	}

	return result.String()
}

// extractModuleName extracts the module name from a module section (first line)
func extractModuleName(section string) string {
	lines := strings.Split(section, "\n")
	if len(lines) == 0 {
		return ""
	}

	firstLine := strings.TrimSpace(lines[0])

	// Check if this looks like a module name (lowercase alphanumeric with dashes/underscores)
	if isModuleNameLine(firstLine) {
		return firstLine
	}

	return ""
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

		// Add blank lines before stub sections
		if i == 0 {
			// First stub: add blank lines
			result.WriteString("\n\n")
		} else {
			// Subsequent stubs: add blank lines to separate from previous stub
			result.WriteString("\n\n")
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
