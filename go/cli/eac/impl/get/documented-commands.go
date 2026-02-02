// Command: get documented-commands
// Description: Get EAC commands documented in markdown files
// Long: Scans all docs/ folder for EAC commands in bash, powershell, and pwsh code blocks.
// Long: Returns a mapping of commands to their documentation locations.
// Long:
// Long: Output Format:
// Long:   - command: The EAC command (e.g., "build", "get modules")
// Long:   - occurrences: List of file locations where the command appears
// Long:     - file: Relative path to the markdown file
// Long:     - line: Line number in the file
// Long:     - language: Code block language (bash, powershell, pwsh)
// Long:     - snippet: The actual command line from the code block
package get

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

func init() {
	registry.Register(GetDocumentedCommands)
}

// DocumentedCommand represents a command found in documentation.
type DocumentedCommand struct {
	Command     string              `yaml:"command" json:"command"`
	Occurrences []CommandOccurrence `yaml:"occurrences" json:"occurrences"`
}

// CommandOccurrence represents where a command appears in documentation.
type CommandOccurrence struct {
	File     string `yaml:"file" json:"file"`
	Line     int    `yaml:"line" json:"line"`
	Language string `yaml:"language" json:"language"`
	Snippet  string `yaml:"snippet" json:"snippet"`
}

// DocumentedCommandsReport is the output structure.
type DocumentedCommandsReport struct {
	Commands []DocumentedCommand `yaml:"commands" json:"commands"`
	Summary  CommandsSummary     `yaml:"summary" json:"summary"`
}

// CommandsSummary provides aggregate statistics.
type CommandsSummary struct {
	TotalCommands    int `yaml:"total_commands" json:"total_commands"`
	TotalOccurrences int `yaml:"total_occurrences" json:"total_occurrences"`
	TotalFiles       int `yaml:"total_files" json:"total_files"`
}

// GetDocumentedCommands scans docs for EAC commands in code blocks.
func GetDocumentedCommands() int {
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		workspaceRoot, err := registry.GetWorkspaceRoot()
		if err != nil {
			return nil, fmt.Errorf("failed to find repository root: %w", err)
		}

		docsPath := filepath.Join(workspaceRoot, "docs")

		// Scan all markdown files
		commandMap := make(map[string][]CommandOccurrence)
		filesScanned := make(map[string]bool)

		err = filepath.Walk(docsPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Only process markdown files
			if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
				return nil
			}

			occurrences, err := scanMarkdownFile(path, workspaceRoot)
			if err != nil {
				return nil // Continue scanning other files
			}

			for _, occ := range occurrences {
				commandMap[occ.command] = append(commandMap[occ.command], occ.occurrence)
				filesScanned[occ.occurrence.File] = true
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to scan docs: %w", err)
		}

		// Build sorted output
		var commands []DocumentedCommand
		totalOccurrences := 0

		// Get sorted command names
		commandNames := make([]string, 0, len(commandMap))
		for cmd := range commandMap {
			commandNames = append(commandNames, cmd)
		}
		sort.Strings(commandNames)

		for _, cmdName := range commandNames {
			occs := commandMap[cmdName]
			totalOccurrences += len(occs)
			commands = append(commands, DocumentedCommand{
				Command:     cmdName,
				Occurrences: occs,
			})
		}

		report := DocumentedCommandsReport{
			Commands: commands,
			Summary: CommandsSummary{
				TotalCommands:    len(commands),
				TotalOccurrences: totalOccurrences,
				TotalFiles:       len(filesScanned),
			},
		}

		return report, nil
	})
}

// commandMatch holds a parsed command and its location.
type commandMatch struct {
	command    string
	occurrence CommandOccurrence
}

// scanMarkdownFile scans a markdown file for EAC commands in code blocks.
func scanMarkdownFile(filePath, workspaceRoot string) ([]commandMatch, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []commandMatch
	scanner := bufio.NewScanner(file)

	// Regex patterns
	codeBlockStart := regexp.MustCompile("^```(bash|powershell|pwsh)\\s*$")
	codeBlockEnd := regexp.MustCompile("^```\\s*$")

	// Command patterns - match r2r eac or just eac commands
	// Handles: r2r eac <cmd>, eac <cmd>, r2r <cmd>
	commandPattern := regexp.MustCompile(`^\s*(?:r2r\s+)?eac\s+(.+)$`)
	r2rPattern := regexp.MustCompile(`^\s*r2r\s+(?:eac\s+)?(.+)$`)

	lineNum := 0
	inCodeBlock := false
	currentLanguage := ""

	// Get relative path for output
	relPath, err := filepath.Rel(workspaceRoot, filePath)
	if err != nil {
		relPath = filePath
	}
	// Normalize path separators
	relPath = filepath.ToSlash(relPath)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for code block start
		if match := codeBlockStart.FindStringSubmatch(line); match != nil {
			inCodeBlock = true
			currentLanguage = match[1]
			continue
		}

		// Check for code block end
		if codeBlockEnd.MatchString(line) && inCodeBlock {
			inCodeBlock = false
			currentLanguage = ""
			continue
		}

		// If in a code block, look for commands
		if inCodeBlock {
			// Skip comments and empty lines
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}

			// Try to match eac command
			var cmdMatch []string
			if cmdMatch = commandPattern.FindStringSubmatch(line); cmdMatch != nil {
				// Matched "eac <cmd>" or "r2r eac <cmd>"
			} else if cmdMatch = r2rPattern.FindStringSubmatch(line); cmdMatch != nil {
				// Matched "r2r <cmd>"
			}

			if cmdMatch != nil {
				// Extract the command (without arguments that look like values)
				cmdParts := extractCommand(cmdMatch[1])
				if cmdParts != "" {
					matches = append(matches, commandMatch{
						command: cmdParts,
						occurrence: CommandOccurrence{
							File:     relPath,
							Line:     lineNum,
							Language: currentLanguage,
							Snippet:  strings.TrimSpace(line),
						},
					})
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}

// extractCommand extracts the command name from a command line
// e.g., "build src-auth --version v1.2.0" -> "build"
// e.g., "get modules --as-yaml" -> "get modules"
// e.g., "show artifacts src-auth" -> "show artifacts".
func extractCommand(cmdLine string) string {
	// Split by whitespace
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return ""
	}

	// Known command prefixes that take subcommands
	multiWordPrefixes := map[string]bool{
		"get":       true,
		"show":      true,
		"validate":  true,
		"create":    true,
		"update":    true,
		"release":   true,
		"pipeline":  true,
		"work":      true,
		"templates": true,
		"serve":     true,
		"scan":      true,
	}

	// Build command, stopping at flags or arguments
	var cmdParts []string
	for i, part := range parts {
		// Stop at flags
		if strings.HasPrefix(part, "-") {
			break
		}

		// First part is always the command
		if i == 0 {
			cmdParts = append(cmdParts, part)
			continue
		}

		// If first part is a multi-word prefix and this looks like a subcommand
		if i == 1 && multiWordPrefixes[parts[0]] {
			// Check if it looks like a subcommand (not a module name or path)
			if !looksLikeArgument(part) {
				cmdParts = append(cmdParts, part)
				continue
			}
		}

		// Stop at what looks like an argument
		break
	}

	return strings.Join(cmdParts, " ")
}

// looksLikeArgument checks if a string looks like a command argument rather than a subcommand.
func looksLikeArgument(s string) bool {
	// Module names often contain hyphens but so do subcommands
	// File paths contain slashes or dots
	if strings.Contains(s, "/") || strings.Contains(s, "\\") {
		return true
	}
	// Version numbers
	if strings.HasPrefix(s, "v") && len(s) > 1 && (s[1] >= '0' && s[1] <= '9') {
		return true
	}
	// Numbers
	if len(s) > 0 && s[0] >= '0' && s[0] <= '9' {
		return true
	}
	// Known module name patterns (src-, ext-, eac-, r2r-, etc.)
	modulePatterns := []string{"src-", "ext-", "eac-", "r2r-", "docs-"}
	for _, prefix := range modulePatterns {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
