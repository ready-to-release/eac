package content

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/staging"
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

// CategoryStats holds category information for documentation.
type CategoryStats struct {
	Name         string
	Description  string
	CommandCount int
}

// flagSplitPattern splits flag lines on 2+ spaces.
var flagSplitPattern = regexp.MustCompile(`\s{2,}`)

// cmdMarkerPatterns matches command markers in markdown.
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

// ProcessCommandMarkers finds and replaces command help markers in staging markdown files.
func ProcessCommandMarkers(
	ctx context.Context,
	fileIndex *staging.FileIndex,
	stagingDir, workspaceRoot string,
	executor CommandExecutor,
	logf func(string, ...any),
	warnf func(string, ...any),
) error {
	logf("    Processing command help markers...")

	cmdBinary := paths.CommandsBinaryPath(workspaceRoot)
	if _, err := os.Stat(cmdBinary); os.IsNotExist(err) {
		logf("    Warning: eac binary not found at %s, skipping command markers", cmdBinary)
		return nil
	}

	replacements := 0
	filesModified := 0

	for _, path := range fileIndex.GetMarkdownFiles() {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := original
		fileReplacements := 0

		for markerType, pattern := range cmdMarkerPatterns {
			matches := pattern.FindAllStringSubmatch(modified, -1)
			for _, match := range matches {
				var replacement string
				var err error

				switch markerType {
				case "cmd":
					cmdName := match[1]
					replacement, err = FormatSingleCommand(ctx, workspaceRoot, executor, cmdName)
				case "cmd-group":
					groupName := match[1]
					replacement, err = FormatCommandGroup(ctx, workspaceRoot, executor, groupName, warnf)
				case "cmd-all":
					replacement, err = FormatAllCommands(ctx, workspaceRoot, executor)
				case "cmd-toc":
					replacement, err = FormatCommandTOC(ctx, workspaceRoot, executor)
				case "categories-table":
					replacement, err = FormatCategoriesTable(ctx, workspaceRoot, executor)
				case "categories-list":
					replacement, err = FormatCategoriesList(ctx, workspaceRoot, executor)
				case "category-section":
					categoryName := match[1]
					replacement, err = FormatCategorySection(ctx, workspaceRoot, executor, categoryName)
				case "category-commands":
					categoryName := match[1]
					replacement, err = FormatCategoryCommands(ctx, workspaceRoot, executor, categoryName)
				case "categories-index":
					replacement, err = FormatCategoriesIndex(ctx, workspaceRoot, executor)
				}

				if err != nil {
					if errors.Is(err, ErrCommandNotFound) {
						relPath, relErr := filepath.Rel(stagingDir, path)
						if relErr != nil {
							relPath = path
						}
						return fmt.Errorf("command marker in %s references non-existent command: %w", relPath, err)
					}
					warnf("failed to process %s marker '%s': %v", markerType, match[0], err)
					continue
				}

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

	logf("    Processed %d command markers in %d files", replacements, filesModified)
	return nil
}

// FormatSingleCommand generates markdown documentation for a single command.
func FormatSingleCommand(ctx context.Context, workspaceRoot string, executor CommandExecutor, cmdName string) (string, error) {
	help, err := GetCommandHelp(ctx, workspaceRoot, executor, cmdName)
	if err != nil {
		return "", err
	}
	return FormatCommandHelp(help, 2, false), nil
}

// FormatCommandGroup generates markdown for all subcommands of a group.
func FormatCommandGroup(ctx context.Context, workspaceRoot string, executor CommandExecutor, groupName string, warnf func(string, ...any)) (string, error) {
	commands, err := GetValidCommands(ctx, workspaceRoot, executor)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s Commands\n\n", strings.Title(groupName))) //nolint:staticcheck

	var groupCmds []CommandInfo
	for _, cmd := range commands {
		if strings.HasPrefix(cmd.Command, groupName+" ") || cmd.Command == groupName {
			groupCmds = append(groupCmds, cmd)
		}
	}

	if len(groupCmds) == 0 {
		return "", fmt.Errorf("no commands found in group '%s'", groupName)
	}

	sort.Slice(groupCmds, func(i, j int) bool {
		return groupCmds[i].Command < groupCmds[j].Command
	})

	for i, cmd := range groupCmds {
		help, err := GetCommandHelp(ctx, workspaceRoot, executor, cmd.Command)
		if err != nil {
			warnf("failed to get help for '%s': %v", cmd.Command, err)
			continue
		}

		sb.WriteString(FormatCommandHelp(help, 3, true))
		if i < len(groupCmds)-1 {
			sb.WriteString("\n---\n\n")
		}
	}

	return sb.String(), nil
}

// FormatAllCommands generates markdown for all commands.
func FormatAllCommands(ctx context.Context, workspaceRoot string, executor CommandExecutor) (string, error) {
	commands, err := GetValidCommands(ctx, workspaceRoot, executor)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("## Command Reference\n\n")

	groups := make(map[string][]CommandInfo)
	for _, cmd := range commands {
		parts := strings.SplitN(cmd.Command, " ", 2)
		group := parts[0]
		groups[group] = append(groups[group], cmd)
	}

	var groupNames []string
	for name := range groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	for _, groupName := range groupNames {
		groupCmds := groups[groupName]
		sort.Slice(groupCmds, func(i, j int) bool {
			return groupCmds[i].Command < groupCmds[j].Command
		})

		sb.WriteString(fmt.Sprintf("### %s\n\n", strings.Title(groupName))) //nolint:staticcheck

		for _, cmd := range groupCmds {
			help, err := GetCommandHelp(ctx, workspaceRoot, executor, cmd.Command)
			if err != nil {
				continue
			}
			sb.WriteString(FormatCommandHelp(help, 4, true))
			sb.WriteString("\n---\n\n")
		}
	}

	return sb.String(), nil
}

// FormatCommandTOC generates a table of contents for all commands.
func FormatCommandTOC(ctx context.Context, workspaceRoot string, executor CommandExecutor) (string, error) {
	commands, err := GetValidCommands(ctx, workspaceRoot, executor)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("## Commands\n\n")
	sb.WriteString("| Command | Description |\n")
	sb.WriteString("|---------|-------------|\n")

	for _, cmd := range commands {
		anchor := strings.ReplaceAll(cmd.Command, " ", "-")
		sb.WriteString(fmt.Sprintf("| [`%s`](#%s) | %s |\n", cmd.Command, anchor, cmd.Description))
	}

	return sb.String(), nil
}

// FormatCategoriesTable generates a category quick reference table.
func FormatCategoriesTable(ctx context.Context, workspaceRoot string, executor CommandExecutor) (string, error) {
	categories, err := GetCategoryStats(ctx, workspaceRoot, executor)
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

// FormatCategoriesList generates a category list with descriptions.
func FormatCategoriesList(ctx context.Context, workspaceRoot string, executor CommandExecutor) (string, error) {
	categories, err := GetCategoryStats(ctx, workspaceRoot, executor)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, cat := range categories {
		sb.WriteString(fmt.Sprintf("- **%s** (%d commands): %s\n", cat.Name, cat.CommandCount, cat.Description))
	}

	return sb.String(), nil
}

// FormatCategorySection generates a single category section with description and link.
func FormatCategorySection(ctx context.Context, workspaceRoot string, executor CommandExecutor, categoryName string) (string, error) {
	categories, err := GetCategoryStats(ctx, workspaceRoot, executor)
	if err != nil {
		return "", err
	}

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

// FormatCategoryCommands generates a table of commands for a specific category.
func FormatCategoryCommands(ctx context.Context, workspaceRoot string, executor CommandExecutor, categoryName string) (string, error) {
	commands, err := GetValidCommands(ctx, workspaceRoot, executor)
	if err != nil {
		return "", err
	}

	cmdConfig, err := LoadCommandsConfig(workspaceRoot)
	if err != nil {
		return "", err
	}

	var categoryCommands []CommandInfo
	for _, cmd := range commands {
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

	sort.Slice(categoryCommands, func(i, j int) bool {
		return categoryCommands[i].Command < categoryCommands[j].Command
	})

	var sb strings.Builder
	sb.WriteString("| Command | Description |\n")
	sb.WriteString("|---------|-------------|\n")

	for _, cmd := range categoryCommands {
		parts := strings.SplitN(cmd.Command, " ", 2)
		var linkPath string
		if len(parts) == 1 {
			linkPath = fmt.Sprintf("../%s/%s.md", parts[0], parts[0])
		} else {
			subCmd := strings.ReplaceAll(parts[1], " ", "-")
			linkPath = fmt.Sprintf("../%s/%s.md", parts[0], subCmd)
		}

		sb.WriteString(fmt.Sprintf("| [%s](%s) | %s |\n", cmd.Command, linkPath, EscapeTableCell(cmd.Description)))
	}

	return sb.String(), nil
}

// FormatCategoriesIndex generates the entire categories index page.
func FormatCategoriesIndex(ctx context.Context, workspaceRoot string, executor CommandExecutor) (string, error) {
	categories, err := GetCategoryStats(ctx, workspaceRoot, executor)
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	sb.WriteString("## All Categories\n\n")

	for i, cat := range categories {
		sb.WriteString(fmt.Sprintf("### %s (%d commands)\n\n", cat.Name, cat.CommandCount))
		sb.WriteString(fmt.Sprintf("**Purpose**: %s\n\n", cat.Description))
		sb.WriteString(fmt.Sprintf("[Browse %s commands →](./%s.md)\n\n", cat.Name, cat.Name))

		if i < len(categories)-1 {
			sb.WriteString("---\n\n")
		}
	}

	sb.WriteString("\n## Category Quick Reference\n\n")
	sb.WriteString("| Category | Commands | Description |\n")
	sb.WriteString("|----------|----------|-------------|\n")

	for _, cat := range categories {
		sb.WriteString(fmt.Sprintf("| **%s** | %d | %s |\n", cat.Name, cat.CommandCount, cat.Description))
	}

	return sb.String(), nil
}

// LoadCommandsConfig loads the commands configuration from commands.yml.
func LoadCommandsConfig(workspaceRoot string) (*config.CommandsConfig, error) {
	configPath := filepath.Join(workspaceRoot, ".eac", config.CommandsFileName)
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

// GetCategoryStats returns category statistics from valid commands.
func GetCategoryStats(ctx context.Context, workspaceRoot string, executor CommandExecutor) ([]CategoryStats, error) {
	commands, err := GetValidCommands(ctx, workspaceRoot, executor)
	if err != nil {
		return nil, err
	}

	cmdConfig, err := LoadCommandsConfig(workspaceRoot)
	if err != nil {
		return nil, err
	}

	knownCategories := make(map[string]bool)
	for name := range cmdConfig.Categories {
		knownCategories[name] = true
	}

	uncategorizedFolder := cmdConfig.Defaults.UncategorizedFolder

	cmdCounts := make(map[string]int)
	for _, cmd := range commands {
		if _, hasOverride := cmdConfig.Overrides[cmd.Command]; hasOverride {
			continue
		}

		parts := strings.SplitN(cmd.Command, " ", 2)
		category := parts[0]

		if !knownCategories[category] {
			if uncategorizedFolder != "" {
				cmdCounts[uncategorizedFolder]++
			}
			continue
		}

		cmdCounts[category]++
	}

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
			desc = fmt.Sprintf("%s commands", strings.Title(name)) //nolint:staticcheck
		}
		categories = append(categories, CategoryStats{
			Name:         name,
			Description:  desc,
			CommandCount: cmdCounts[name],
		})
	}

	return categories, nil
}

// GetCommandHelp runs --help for a command and parses the output.
func GetCommandHelp(ctx context.Context, workspaceRoot string, executor CommandExecutor, cmdName string) (*CommandHelp, error) {
	args := append(strings.Fields(cmdName), "--help")

	output, err := executor.Run(ctx, workspaceRoot, args)
	if err != nil {
		errLower := strings.ToLower(err.Error())
		if strings.Contains(errLower, "unknown command") ||
			strings.Contains(errLower, "command not found") ||
			strings.Contains(errLower, "not a valid command") ||
			strings.Contains(errLower, "no such command") {
			return nil, fmt.Errorf("%w: '%s'", ErrCommandNotFound, cmdName)
		}
		return nil, fmt.Errorf("failed to run '%s --help': %v", cmdName, err)
	}

	return ParseHelpOutput(cmdName, output)
}

// GetValidCommands retrieves the list of all valid commands.
func GetValidCommands(ctx context.Context, workspaceRoot string, executor CommandExecutor) ([]CommandInfo, error) {
	output, err := executor.Run(ctx, workspaceRoot, []string{"get", "valid-commands"})
	if err != nil {
		return nil, fmt.Errorf("failed to get valid-commands: %w", err)
	}

	var commands []CommandInfo
	if err := yaml.Unmarshal([]byte(output), &commands); err != nil {
		return nil, fmt.Errorf("failed to parse valid-commands output: %w", err)
	}

	return commands, nil
}

// ParseHelpOutput parses the --help output into a CommandHelp struct.
func ParseHelpOutput(cmdName, output string) (*CommandHelp, error) {
	help := &CommandHelp{
		Name: cmdName,
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	section := "description"
	var currentLines []string
	var expectedOutputLines []string
	inExpectedOutput := false
	var currentFlag *FlagArg

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "Usage:"):
			help.Description = FormatDescription(currentLines)
			currentLines = nil
			help.Usage = strings.TrimSpace(strings.TrimPrefix(line, "Usage:"))
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
			section = "examples"
			inExpectedOutput = false
			currentFlag = nil
			continue
		case strings.HasPrefix(line, "Expected Output:"):
			inExpectedOutput = true
			expectedOutputLines = append(expectedOutputLines, "**Expected Output:**", "")
			currentFlag = nil
			continue
		}

		switch section {
		case "description":
			if inExpectedOutput {
				expectedOutputLines = append(expectedOutputLines, line)
			} else if line != "" {
				currentLines = append(currentLines, line)
			} else if len(currentLines) > 0 {
				currentLines = append(currentLines, "")
			}
		case "arguments":
			if arg := ParseFlagLine(line); arg != nil {
				help.Arguments = append(help.Arguments, *arg)
				currentFlag = &help.Arguments[len(help.Arguments)-1]
			} else if currentFlag != nil && strings.HasPrefix(line, "      ") {
				desc := strings.TrimSpace(line)
				if currentFlag.Description != "" {
					currentFlag.Description += " " + desc
				} else {
					currentFlag.Description = desc
				}
			} else if strings.TrimSpace(line) == "" {
				// empty line
			} else if !strings.HasPrefix(line, "  ") && line != "" {
				section = "notes"
				help.Notes = line
				currentFlag = nil
			}
		case "flags":
			if arg := ParseFlagLine(line); arg != nil {
				help.Flags = append(help.Flags, *arg)
				currentFlag = &help.Flags[len(help.Flags)-1]
			} else if currentFlag != nil && strings.HasPrefix(line, "      ") {
				desc := strings.TrimSpace(line)
				if currentFlag.Description != "" {
					currentFlag.Description += " " + desc
				} else {
					currentFlag.Description = desc
				}
			} else if strings.TrimSpace(line) == "" {
				// empty line
			} else if !strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "-") && line != "" {
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

	if help.Description == "" && len(currentLines) > 0 {
		help.Description = FormatDescription(currentLines)
	}

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

// FormatDescription joins description lines, preserving paragraph breaks.
func FormatDescription(lines []string) string {
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

	if len(currentParagraph) > 0 {
		paragraphs = append(paragraphs, strings.Join(currentParagraph, " "))
	}

	return strings.TrimSpace(strings.Join(paragraphs, "\n\n"))
}

// ParseFlagLine parses a flag/argument line like "  --dry-run    Description here".
func ParseFlagLine(line string) *FlagArg {
	if !strings.HasPrefix(line, "  ") {
		return nil
	}

	line = strings.TrimPrefix(line, "  ")
	if line == "" {
		return nil
	}

	parts := flagSplitPattern.Split(line, 2)
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

// FormatCommandHelp formats a CommandHelp as markdown.
func FormatCommandHelp(help *CommandHelp, headingLevel int, includeTitle bool) string {
	var sb strings.Builder
	heading := strings.Repeat("#", headingLevel)
	subHeading := strings.Repeat("#", headingLevel+1)

	if includeTitle {
		sb.WriteString(fmt.Sprintf("%s %s\n\n", heading, help.Name))
	}

	if help.Description != "" {
		sb.WriteString(EscapeJinja2(help.Description))
		sb.WriteString("\n\n")
	}

	if help.Usage != "" {
		sb.WriteString(fmt.Sprintf("**Usage:** `%s`\n\n", help.Usage))
	}

	if len(help.Arguments) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Arguments\n\n", subHeading))
		sb.WriteString("| Argument | Description |\n")
		sb.WriteString("|----------|-------------|\n")
		for _, arg := range help.Arguments {
			sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", arg.Name, EscapeTableCell(arg.Description)))
		}
		sb.WriteString("\n")
	}

	if len(help.Flags) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Flags\n\n", subHeading))
		sb.WriteString("| Flag | Description |\n")
		sb.WriteString("|------|-------------|\n")
		for _, flag := range help.Flags {
			sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", flag.Name, EscapeTableCell(flag.Description)))
		}
		sb.WriteString("\n")
	}

	if help.Notes != "" {
		sb.WriteString(fmt.Sprintf("\n%s Notes\n\n", subHeading))
		sb.WriteString(EscapeJinja2(help.Notes))
		sb.WriteString("\n\n")
	}

	if help.Examples != "" {
		sb.WriteString(fmt.Sprintf("\n%s Examples\n\n", subHeading))
		sb.WriteString("```bash\n")
		sb.WriteString(EscapeJinja2(DedentExamples(help.Examples)))
		sb.WriteString("\n```\n")
	}

	return sb.String()
}

// EscapeTableCell escapes pipe characters and newlines for markdown tables.
func EscapeTableCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// EscapeJinja2 escapes Jinja2 template syntax to prevent MkDocs macros from processing
// command help output as template expressions.
func EscapeJinja2(s string) string {
	s = strings.ReplaceAll(s, "{{", "{ {")
	s = strings.ReplaceAll(s, "}}", "} }")
	s = strings.ReplaceAll(s, "{%", "{ %")
	s = strings.ReplaceAll(s, "%}", "% }")
	return s
}

// DedentExamples removes common leading whitespace from example lines.
func DedentExamples(examples string) string {
	lines := strings.Split(strings.TrimSpace(examples), "\n")
	if len(lines) == 0 {
		return examples
	}

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
