package books

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/paths"
	"gopkg.in/yaml.v3"
)

// ErrCommandNotFound indicates a command marker references a non-existent command.
var ErrCommandNotFound = errors.New("command not found")

// CommandHelp represents parsed help output for a command.
type CommandHelp struct {
	Name        string
	Description string
	Usage       string
	Arguments   []FlagArg
	Flags       []FlagArg
	Notes       string
	Examples    string
}

// FlagArg represents a flag or argument with description.
type FlagArg struct {
	Name        string
	Description string
}

// CommandInfo represents a command from get valid-commands.
type CommandInfo struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
}

// cmdMarkerPattern matches command markers in markdown
// Variants:
//   - <!-- book:cmd build --> - Single command help
//   - <!-- book:cmd-group validate --> - All subcommands under validate
//   - <!-- book:cmd-all --> - Full command reference
//   - <!-- book:cmd-toc --> - Table of contents with links
//   - <!-- book:categories-table --> - Category quick reference table
//   - <!-- book:categories-list --> - Category list with descriptions
var cmdMarkerPatterns = map[string]*regexp.Regexp{
	"cmd":               regexp.MustCompile(`<!--\s*book:cmd\s+([a-zA-Z0-9_-]+(?:\s+[a-zA-Z0-9_-]+)*)\s*-->`),
	"cmd-group":         regexp.MustCompile(`<!--\s*book:cmd-group\s+([a-zA-Z0-9_-]+)\s*-->`),
	"cmd-all":           regexp.MustCompile(`<!--\s*book:cmd-all\s*-->`),
	"cmd-toc":           regexp.MustCompile(`<!--\s*book:cmd-toc\s*-->`),
	"categories-table":  regexp.MustCompile(`<!--\s*book:categories-table\s*-->`),
	"categories-list":   regexp.MustCompile(`<!--\s*book:categories-list\s*-->`),
	"category-section":  regexp.MustCompile(`<!--\s*book:category-section\s+([a-zA-Z0-9_-]+)\s*-->`),
	"category-commands": regexp.MustCompile(`<!--\s*book:category-commands\s+([a-zA-Z0-9_-]+)\s*-->`),
	"categories-index":  regexp.MustCompile(`<!--\s*book:categories-index\s*-->`),
}

// processCommandMarkers finds and replaces command help markers in staging markdown files.
func (p *Preprocessor) processCommandMarkers() error {
	p.log("    Processing command help markers...")

	// Get eac binary path
	cmdBinary := paths.CommandsBinaryPath(p.workspaceRoot)
	if _, err := os.Stat(cmdBinary); os.IsNotExist(err) {
		p.log("    Warning: eac binary not found at %s, skipping command markers", cmdBinary)
		return nil
	}

	replacements := 0
	filesModified := 0

	// Use file index for efficient iteration
	for _, path := range p.fileIndex.GetMarkdownFiles() {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := original
		fileReplacements := 0

		// Process each marker type
		for markerType, pattern := range cmdMarkerPatterns {
			matches := pattern.FindAllStringSubmatch(modified, -1)
			for _, match := range matches {
				var replacement string
				var err error

				switch markerType {
				case "cmd":
					// Single command help
					cmdName := match[1]
					replacement, err = p.formatSingleCommand(cmdBinary, cmdName)
				case "cmd-group":
					// Command group (all subcommands)
					groupName := match[1]
					replacement, err = p.formatCommandGroup(cmdBinary, groupName)
				case "cmd-all":
					// Full command reference
					replacement, err = p.formatAllCommands(cmdBinary)
				case "cmd-toc":
					// Table of contents
					replacement, err = p.formatCommandTOC(cmdBinary)
				case "categories-table":
					// Category quick reference table
					replacement, err = p.formatCategoriesTable(cmdBinary)
				case "categories-list":
					// Category list with descriptions
					replacement, err = p.formatCategoriesList(cmdBinary)
				case "category-section":
					// Single category section with description and link
					categoryName := match[1]
					replacement, err = p.formatCategorySection(cmdBinary, categoryName)
				case "category-commands":
					// Table of commands in a category
					categoryName := match[1]
					replacement, err = p.formatCategoryCommands(cmdBinary, categoryName)
				case "categories-index":
					// Full categories index page
					replacement, err = p.formatCategoriesIndex(cmdBinary)
				}

				if err != nil {
					// Fail build on missing commands - docs must reference valid commands
					if errors.Is(err, ErrCommandNotFound) {
						relPath, relErr := filepath.Rel(p.stagingDir, path)
						if relErr != nil {
							relPath = path // Fallback to absolute path
						}
						return fmt.Errorf("command marker in %s references non-existent command: %w", relPath, err)
					}
					p.warn("failed to process %s marker '%s': %v", markerType, match[0], err)
					continue
				}

				// Wrap replacement with generation markers
				wrapped := fmt.Sprintf("<!-- book:generated %s -->\n%s\n<!-- /book:generated -->", match[0][5:len(match[0])-4], replacement)
				modified = strings.Replace(modified, match[0], wrapped, 1)
				fileReplacements++
			}
		}

		if fileReplacements > 0 {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			replacements += fileReplacements
			filesModified++
		}
	}

	p.log("    Processed %d command markers in %d files", replacements, filesModified)
	return nil
}

// formatSingleCommand generates markdown documentation for a single command
// The document is expected to already have a title, so we don't include one.
func (p *Preprocessor) formatSingleCommand(cmdBinary, cmdName string) (string, error) {
	help, err := p.getCommandHelp(cmdBinary, cmdName)
	if err != nil {
		return "", err
	}

	// H2 level for sub-sections, no title (doc already has # Title)
	return p.formatCommandHelp(help, 2, false), nil
}

// formatCommandGroup generates markdown for all subcommands of a group.
func (p *Preprocessor) formatCommandGroup(cmdBinary, groupName string) (string, error) {
	commands, err := p.getValidCommands(cmdBinary)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s Commands\n\n", strings.Title(groupName)))

	// Find all commands in this group
	var groupCmds []CommandInfo
	for _, cmd := range commands {
		if strings.HasPrefix(cmd.Command, groupName+" ") || cmd.Command == groupName {
			groupCmds = append(groupCmds, cmd)
		}
	}

	if len(groupCmds) == 0 {
		return "", fmt.Errorf("no commands found in group '%s'", groupName)
	}

	// Sort by command name
	sort.Slice(groupCmds, func(i, j int) bool {
		return groupCmds[i].Command < groupCmds[j].Command
	})

	// Generate help for each command
	for i, cmd := range groupCmds {
		help, err := p.getCommandHelp(cmdBinary, cmd.Command)
		if err != nil {
			p.warn("failed to get help for '%s': %v", cmd.Command, err)
			continue
		}

		// H3 level with title (each command needs its own heading in a group)
		sb.WriteString(p.formatCommandHelp(help, 3, true))
		if i < len(groupCmds)-1 {
			sb.WriteString("\n---\n\n")
		}
	}

	return sb.String(), nil
}

// formatAllCommands generates markdown for all commands.
func (p *Preprocessor) formatAllCommands(cmdBinary string) (string, error) {
	commands, err := p.getValidCommands(cmdBinary)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("## Command Reference\n\n")

	// Group commands by their first word
	groups := make(map[string][]CommandInfo)
	for _, cmd := range commands {
		parts := strings.SplitN(cmd.Command, " ", 2)
		group := parts[0]
		groups[group] = append(groups[group], cmd)
	}

	// Sort group names
	var groupNames []string
	for name := range groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	// Generate help for each group
	for _, groupName := range groupNames {
		groupCmds := groups[groupName]
		sort.Slice(groupCmds, func(i, j int) bool {
			return groupCmds[i].Command < groupCmds[j].Command
		})

		sb.WriteString(fmt.Sprintf("### %s\n\n", strings.Title(groupName)))

		for _, cmd := range groupCmds {
			help, err := p.getCommandHelp(cmdBinary, cmd.Command)
			if err != nil {
				continue
			}
			// H4 level with title (each command needs its own heading)
			sb.WriteString(p.formatCommandHelp(help, 4, true))
			sb.WriteString("\n---\n\n")
		}
	}

	return sb.String(), nil
}

// formatCommandTOC generates a table of contents for all commands.
func (p *Preprocessor) formatCommandTOC(cmdBinary string) (string, error) {
	commands, err := p.getValidCommands(cmdBinary)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("## Commands\n\n")
	sb.WriteString("| Command | Description |\n")
	sb.WriteString("|---------|-------------|\n")

	for _, cmd := range commands {
		// Create anchor link from command name
		anchor := strings.ReplaceAll(cmd.Command, " ", "-")
		sb.WriteString(fmt.Sprintf("| [`%s`](#%s) | %s |\n", cmd.Command, anchor, cmd.Description))
	}

	return sb.String(), nil
}

// CategoryStats holds category information for documentation.
type CategoryStats struct {
	Name         string
	Description  string
	CommandCount int
}

// formatCategoriesTable generates a category quick reference table.
func (p *Preprocessor) formatCategoriesTable(cmdBinary string) (string, error) {
	categories, err := p.getCategoryStats(cmdBinary)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("| Category | Commands | Description |\n")
	sb.WriteString("|----------|----------|-------------|\n")

	for _, cat := range categories {
		sb.WriteString(fmt.Sprintf("| **%s** | %d | %s |\n", cat.Name, cat.CommandCount, cat.Description))
	}

	return sb.String(), nil
}

// formatCategoriesList generates a category list with descriptions.
func (p *Preprocessor) formatCategoriesList(cmdBinary string) (string, error) {
	categories, err := p.getCategoryStats(cmdBinary)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, cat := range categories {
		sb.WriteString(fmt.Sprintf("- **%s** (%d commands): %s\n", cat.Name, cat.CommandCount, cat.Description))
	}

	return sb.String(), nil
}

// formatCategorySection generates a single category section with description and link.
func (p *Preprocessor) formatCategorySection(cmdBinary, categoryName string) (string, error) {
	categories, err := p.getCategoryStats(cmdBinary)
	if err != nil {
		return "", err
	}

	// Find the category
	var cat *CategoryStats
	for i := range categories {
		if categories[i].Name == categoryName {
			cat = &categories[i]
			break
		}
	}

	if cat == nil {
		return "", fmt.Errorf("unknown category: %s", categoryName)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### %s (%d commands)\n\n", categoryName, cat.CommandCount))
	sb.WriteString(fmt.Sprintf("**Purpose**: %s\n\n", cat.Description))
	sb.WriteString(fmt.Sprintf("[Browse %s commands →](./%s.md)\n", categoryName, categoryName))

	return sb.String(), nil
}

// formatCategoryCommands generates a table of commands for a specific category.
func (p *Preprocessor) formatCategoryCommands(cmdBinary, categoryName string) (string, error) {
	commands, err := p.getValidCommands(cmdBinary)
	if err != nil {
		return "", err
	}

	// Load commands config for overrides
	cmdConfig, err := p.loadCommandsConfig()
	if err != nil {
		return "", err
	}

	// Find commands in this category
	var categoryCommands []CommandInfo
	for _, cmd := range commands {
		// Skip commands with explicit overrides
		if _, hasOverride := cmdConfig.Overrides[cmd.Command]; hasOverride {
			continue
		}

		parts := strings.SplitN(cmd.Command, " ", 2)
		if parts[0] == categoryName {
			categoryCommands = append(categoryCommands, cmd)
		}
	}

	if len(categoryCommands) == 0 {
		return "", fmt.Errorf("no commands found in category '%s'", categoryName)
	}

	// Sort by command name
	sort.Slice(categoryCommands, func(i, j int) bool {
		return categoryCommands[i].Command < categoryCommands[j].Command
	})

	var sb strings.Builder
	sb.WriteString("| Command | Description |\n")
	sb.WriteString("|---------|-------------|\n")

	for _, cmd := range categoryCommands {
		// Generate link to command doc
		parts := strings.SplitN(cmd.Command, " ", 2)
		var linkPath string
		if len(parts) == 1 {
			// Single word command like "build" -> ../build/build.md
			linkPath = fmt.Sprintf("../%s/%s.md", parts[0], parts[0])
		} else {
			// Multi-word command like "work create" -> ../work/create.md
			// Replace spaces with hyphens in the filename
			subCmd := strings.ReplaceAll(parts[1], " ", "-")
			linkPath = fmt.Sprintf("../%s/%s.md", parts[0], subCmd)
		}

		sb.WriteString(fmt.Sprintf("| [%s](%s) | %s |\n", cmd.Command, linkPath, escapeTableCell(cmd.Description)))
	}

	return sb.String(), nil
}

// formatCategoriesIndex generates the entire categories index page.
func (p *Preprocessor) formatCategoriesIndex(cmdBinary string) (string, error) {
	categories, err := p.getCategoryStats(cmdBinary)
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	// All Categories section
	sb.WriteString("## All Categories\n\n")

	for i, cat := range categories {
		sb.WriteString(fmt.Sprintf("### %s (%d commands)\n\n", cat.Name, cat.CommandCount))
		sb.WriteString(fmt.Sprintf("**Purpose**: %s\n\n", cat.Description))
		sb.WriteString(fmt.Sprintf("[Browse %s commands →](./%s.md)\n\n", cat.Name, cat.Name))

		if i < len(categories)-1 {
			sb.WriteString("---\n\n")
		}
	}

	// Quick Reference Table
	sb.WriteString("\n## Category Quick Reference\n\n")
	sb.WriteString("| Category | Commands | Description |\n")
	sb.WriteString("|----------|----------|-------------|\n")

	for _, cat := range categories {
		sb.WriteString(fmt.Sprintf("| **%s** | %d | %s |\n", cat.Name, cat.CommandCount, cat.Description))
	}

	return sb.String(), nil
}

// loadCommandsConfig loads the commands configuration from commands.yml.
func (p *Preprocessor) loadCommandsConfig() (*config.CommandsConfig, error) {
	configPath := filepath.Join(p.workspaceRoot, ".eac", config.CommandsFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return config.DefaultCommandsConfig(), nil
		}
		return nil, fmt.Errorf("failed to read commands config: %w", err)
	}

	var cfg config.CommandsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse commands config: %w", err)
	}

	return &cfg, nil
}

// getCategoryStats returns category statistics from valid commands.
func (p *Preprocessor) getCategoryStats(cmdBinary string) ([]CategoryStats, error) {
	commands, err := p.getValidCommands(cmdBinary)
	if err != nil {
		return nil, err
	}

	// Load commands config for categories and overrides
	cmdConfig, err := p.loadCommandsConfig()
	if err != nil {
		return nil, err
	}

	// Build set of known categories from config
	knownCategories := make(map[string]bool)
	for name := range cmdConfig.Categories {
		knownCategories[name] = true
	}

	// Determine uncategorized folder from config (no default - skip if not specified)
	uncategorizedFolder := cmdConfig.Defaults.UncategorizedFolder

	// Count commands per category
	cmdCounts := make(map[string]int)
	for _, cmd := range commands {
		// Skip commands with explicit overrides - they have their own paths
		if _, hasOverride := cmdConfig.Overrides[cmd.Command]; hasOverride {
			continue
		}

		parts := strings.SplitN(cmd.Command, " ", 2)
		category := parts[0]

		// Commands whose first word isn't a known category go to uncategorized
		if !knownCategories[category] {
			// Only add to uncategorized if we have a configured folder
			if uncategorizedFolder != "" {
				cmdCounts[uncategorizedFolder]++
			}
			continue
		}

		cmdCounts[category]++
	}

	// Build sorted category list
	var categoryNames []string
	for name := range cmdCounts {
		categoryNames = append(categoryNames, name)
	}
	sort.Strings(categoryNames)

	var categories []CategoryStats
	for _, name := range categoryNames {
		desc := ""
		if cat, ok := cmdConfig.Categories[name]; ok {
			desc = cat.Description
		}
		if desc == "" {
			desc = fmt.Sprintf("%s commands", strings.Title(name))
		}
		categories = append(categories, CategoryStats{
			Name:         name,
			Description:  desc,
			CommandCount: cmdCounts[name],
		})
	}

	return categories, nil
}

// getCommandHelp runs --help for a command and parses the output.
func (p *Preprocessor) getCommandHelp(cmdBinary, cmdName string) (*CommandHelp, error) {
	// Run command with --help
	args := append(strings.Fields(cmdName), "--help")
	cmd := exec.Command(cmdBinary, args...)
	cmd.Dir = p.workspaceRoot
	cmd.Env = append(os.Environ(), "NO_COLOR=1")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errOutput := stderr.String()
		errLower := strings.ToLower(errOutput)
		// Check if command doesn't exist
		if strings.Contains(errLower, "unknown command") ||
			strings.Contains(errLower, "command not found") ||
			strings.Contains(errLower, "not a valid command") ||
			strings.Contains(errLower, "no such command") {
			return nil, fmt.Errorf("%w: '%s'", ErrCommandNotFound, cmdName)
		}
		return nil, fmt.Errorf("failed to run '%s --help': %s", cmdName, errOutput)
	}

	return parseHelpOutput(cmdName, stdout.String())
}

// parseHelpOutput parses the --help output into a CommandHelp struct.
func parseHelpOutput(cmdName, output string) (*CommandHelp, error) {
	help := &CommandHelp{
		Name: cmdName,
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	section := "description"
	var currentLines []string
	var expectedOutputLines []string
	inExpectedOutput := false
	var currentFlag *FlagArg // Track current flag for multi-line descriptions

	for scanner.Scan() {
		line := scanner.Text()

		// Detect section headers
		switch {
		case strings.HasPrefix(line, "Usage:"):
			help.Description = formatDescription(currentLines)
			currentLines = nil
			help.Usage = strings.TrimPrefix(line, "Usage:")
			help.Usage = strings.TrimSpace(help.Usage)
			section = "usage"
			inExpectedOutput = false
			currentFlag = nil
			continue
		case strings.HasPrefix(line, "Arguments:"):
			section = "arguments"
			inExpectedOutput = false
			currentFlag = nil
			continue
		case strings.HasPrefix(line, "Flags:"):
			section = "flags"
			inExpectedOutput = false
			currentFlag = nil
			continue
		case strings.HasPrefix(line, "Example:") || strings.HasPrefix(line, "Examples:"):
			// Handle both singular and plural
			section = "examples"
			inExpectedOutput = false
			currentFlag = nil
			continue
		case strings.HasPrefix(line, "Expected Output:"):
			// Expected Output goes into Notes, not Description
			inExpectedOutput = true
			// Add blank line after heading so markdown list renders correctly
			expectedOutputLines = append(expectedOutputLines, "**Expected Output:**", "")
			currentFlag = nil
			continue
		}

		// Process line based on current section
		switch section {
		case "description":
			if inExpectedOutput {
				// Collect expected output lines for Notes
				expectedOutputLines = append(expectedOutputLines, line)
			} else if line != "" {
				currentLines = append(currentLines, line)
			} else if len(currentLines) > 0 {
				// Empty line - preserve paragraph break
				currentLines = append(currentLines, "")
			}
		case "arguments":
			if arg := parseFlagLine(line); arg != nil {
				help.Arguments = append(help.Arguments, *arg)
				currentFlag = &help.Arguments[len(help.Arguments)-1]
			} else if currentFlag != nil && strings.HasPrefix(line, "      ") {
				// Continuation line for argument description (6+ spaces)
				desc := strings.TrimSpace(line)
				if currentFlag.Description != "" {
					currentFlag.Description += " " + desc
				} else {
					currentFlag.Description = desc
				}
			} else if strings.TrimSpace(line) == "" {
				// Empty line might signal end of section or notes
			} else if !strings.HasPrefix(line, "  ") && line != "" {
				// This might be notes after arguments
				section = "notes"
				help.Notes = line
				currentFlag = nil
			}
		case "flags":
			if arg := parseFlagLine(line); arg != nil {
				help.Flags = append(help.Flags, *arg)
				currentFlag = &help.Flags[len(help.Flags)-1]
			} else if currentFlag != nil && strings.HasPrefix(line, "      ") {
				// Continuation line for flag description (6+ spaces)
				desc := strings.TrimSpace(line)
				if currentFlag.Description != "" {
					currentFlag.Description += " " + desc
				} else {
					currentFlag.Description = desc
				}
			} else if strings.TrimSpace(line) == "" {
				// Empty line
			} else if !strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "-") && line != "" {
				// This might be notes after flags (like the MkDocs section)
				section = "notes"
				help.Notes = line
				currentFlag = nil
			}
		case "notes":
			if help.Notes != "" {
				help.Notes += "\n"
			}
			help.Notes += line
		case "examples":
			if help.Examples != "" {
				help.Examples += "\n"
			}
			help.Examples += line
		}
	}

	// Handle case where description spans entire output (no Usage: line)
	if help.Description == "" && len(currentLines) > 0 {
		help.Description = formatDescription(currentLines)
	}

	// Add expected output to notes
	if len(expectedOutputLines) > 0 {
		expectedOutput := strings.Join(expectedOutputLines, "\n")
		if help.Notes != "" {
			help.Notes = expectedOutput + "\n\n" + help.Notes
		} else {
			help.Notes = expectedOutput
		}
	}

	return help, nil
}

// formatDescription joins description lines, preserving paragraph breaks.
func formatDescription(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	var paragraphs []string
	var currentParagraph []string

	for _, line := range lines {
		if line == "" {
			if len(currentParagraph) > 0 {
				paragraphs = append(paragraphs, strings.Join(currentParagraph, " "))
				currentParagraph = nil
			}
		} else {
			currentParagraph = append(currentParagraph, line)
		}
	}

	// Don't forget the last paragraph
	if len(currentParagraph) > 0 {
		paragraphs = append(paragraphs, strings.Join(currentParagraph, " "))
	}

	return strings.TrimSpace(strings.Join(paragraphs, "\n\n"))
}

// parseFlagLine parses a flag/argument line like "  --dry-run    Description here".
func parseFlagLine(line string) *FlagArg {
	if !strings.HasPrefix(line, "  ") {
		return nil
	}

	line = strings.TrimPrefix(line, "  ")
	if line == "" {
		return nil
	}

	// Find the gap between flag name and description (multiple spaces)
	// Pattern: "  --flag-name     Description text"
	parts := regexp.MustCompile(`\s{2,}`).Split(line, 2)
	if len(parts) < 1 {
		return nil
	}

	name := strings.TrimSpace(parts[0])
	desc := ""
	if len(parts) > 1 {
		desc = strings.TrimSpace(parts[1])
	}

	if name == "" {
		return nil
	}

	return &FlagArg{
		Name:        name,
		Description: desc,
	}
}

// formatCommandHelp formats a CommandHelp as markdown
// headingLevel controls the base heading level for the command name
// includeTitle controls whether to include the command name as a heading
// (set to false when the document already has a title).
func (p *Preprocessor) formatCommandHelp(help *CommandHelp, headingLevel int, includeTitle bool) string {
	var sb strings.Builder
	heading := strings.Repeat("#", headingLevel)
	subHeading := strings.Repeat("#", headingLevel+1)

	// Command name as heading (optional - skip when doc already has title)
	if includeTitle {
		sb.WriteString(fmt.Sprintf("%s %s\n\n", heading, help.Name))
	}

	// Description (escape Jinja2 syntax)
	if help.Description != "" {
		sb.WriteString(escapeJinja2(help.Description))
		sb.WriteString("\n\n")
	}

	// Usage
	if help.Usage != "" {
		sb.WriteString(fmt.Sprintf("**Usage:** `%s`\n\n", help.Usage))
	}

	// Arguments table
	if len(help.Arguments) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Arguments\n\n", subHeading))
		sb.WriteString("| Argument | Description |\n")
		sb.WriteString("|----------|-------------|\n")
		for _, arg := range help.Arguments {
			sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", arg.Name, escapeTableCell(arg.Description)))
		}
		sb.WriteString("\n")
	}

	// Flags table
	if len(help.Flags) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Flags\n\n", subHeading))
		sb.WriteString("| Flag | Description |\n")
		sb.WriteString("|------|-------------|\n")
		for _, flag := range help.Flags {
			sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", flag.Name, escapeTableCell(flag.Description)))
		}
		sb.WriteString("\n")
	}

	// Notes (if any, escape Jinja2 syntax)
	if help.Notes != "" {
		sb.WriteString(fmt.Sprintf("\n%s Notes\n\n", subHeading))
		sb.WriteString(escapeJinja2(help.Notes))
		sb.WriteString("\n\n")
	}

	// Examples (escape Jinja2 syntax)
	if help.Examples != "" {
		sb.WriteString(fmt.Sprintf("\n%s Examples\n\n", subHeading))
		sb.WriteString("```bash\n")
		// Clean up example formatting - remove common leading whitespace
		sb.WriteString(escapeJinja2(dedentExamples(help.Examples)))
		sb.WriteString("\n```\n")
	}

	return sb.String()
}

// getValidCommands retrieves the list of all valid commands.
func (p *Preprocessor) getValidCommands(cmdBinary string) ([]CommandInfo, error) {
	cmd := exec.Command(cmdBinary, "get", "valid-commands")
	cmd.Dir = p.workspaceRoot
	cmd.Env = append(os.Environ(), "NO_COLOR=1")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get valid-commands: %s", stderr.String())
	}

	var commands []CommandInfo
	if err := yaml.Unmarshal(stdout.Bytes(), &commands); err != nil {
		return nil, fmt.Errorf("failed to parse valid-commands output: %w", err)
	}

	return commands, nil
}

// escapeTableCell escapes pipe characters and newlines for markdown tables.
func escapeTableCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// escapeJinja2 escapes Jinja2 template syntax to prevent MkDocs macros from processing
// command help output as template expressions.
func escapeJinja2(s string) string {
	// Escape {{ and }} which Jinja2 interprets as expression delimiters
	// Use Jinja2's raw block alternative: wrap in {% raw %}...{% endraw %}
	// Or escape the braces: {{ becomes { {
	s = strings.ReplaceAll(s, "{{", "{ {")
	s = strings.ReplaceAll(s, "}}", "} }")
	// Also escape {% which is Jinja2 statement delimiter
	s = strings.ReplaceAll(s, "{%", "{ %")
	s = strings.ReplaceAll(s, "%}", "% }")
	return s
}

// dedentExamples removes common leading whitespace from example lines.
func dedentExamples(examples string) string {
	lines := strings.Split(strings.TrimSpace(examples), "\n")
	if len(lines) == 0 {
		return examples
	}

	// Find minimum leading whitespace (excluding empty lines)
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return strings.TrimSpace(examples)
	}

	// Remove common indent from all lines
	var result []string
	for _, line := range lines {
		if len(line) >= minIndent {
			result = append(result, line[minIndent:])
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
