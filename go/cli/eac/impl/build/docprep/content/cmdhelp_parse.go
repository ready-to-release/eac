package content

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"gopkg.in/yaml.v3"
)

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

// parseState tracks mutable state during help output parsing.
type parseState struct {
	help                *CommandHelp
	section             string
	currentLines        []string
	expectedOutputLines []string
	inExpectedOutput    bool
	currentFlag         *FlagArg
}

// ParseHelpOutput parses the --help output into a CommandHelp struct.
func ParseHelpOutput(cmdName, output string) (*CommandHelp, error) {
	state := &parseState{
		help:    &CommandHelp{Name: cmdName},
		section: "description",
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if handleSectionHeader(state, line) {
			continue
		}
		parseSectionLine(state, line)
	}

	finalizeHelp(state)
	return state.help, nil
}

// handleSectionHeader checks if a line is a section header and transitions state.
// Returns true if the line was consumed as a header.
func handleSectionHeader(s *parseState, line string) bool {
	switch {
	case strings.HasPrefix(line, "Usage:"):
		s.help.Description = FormatDescription(s.currentLines)
		s.currentLines = nil
		s.help.Usage = strings.TrimSpace(strings.TrimPrefix(line, "Usage:"))
		s.section = "usage"
		s.inExpectedOutput = false
		s.currentFlag = nil
		return true
	case strings.HasPrefix(line, "Arguments:"):
		s.section = "arguments"
		s.inExpectedOutput = false
		s.currentFlag = nil
		return true
	case strings.HasPrefix(line, "Flags:"):
		s.section = "flags"
		s.inExpectedOutput = false
		s.currentFlag = nil
		return true
	case strings.HasPrefix(line, "Example:") || strings.HasPrefix(line, "Examples:"):
		s.section = "examples"
		s.inExpectedOutput = false
		s.currentFlag = nil
		return true
	case strings.HasPrefix(line, "Expected Output:"):
		s.inExpectedOutput = true
		s.expectedOutputLines = append(s.expectedOutputLines, "**Expected Output:**", "")
		s.currentFlag = nil
		return true
	}
	return false
}

// parseSectionLine dispatches a content line to the appropriate section parser.
func parseSectionLine(s *parseState, line string) {
	switch s.section {
	case "description":
		parseDescriptionLine(s, line)
	case "arguments":
		parseFlagOrArgLine(s, line, &s.help.Arguments)
	case "flags":
		parseFlagOrArgLine(s, line, &s.help.Flags)
	case "notes":
		if s.help.Notes != "" {
			s.help.Notes += "\n"
		}
		s.help.Notes += line
	case "examples":
		if s.help.Examples != "" {
			s.help.Examples += "\n"
		}
		s.help.Examples += line
	}
}

// parseDescriptionLine handles a line in the description section.
func parseDescriptionLine(s *parseState, line string) {
	if s.inExpectedOutput {
		s.expectedOutputLines = append(s.expectedOutputLines, line)
	} else if line != "" {
		s.currentLines = append(s.currentLines, line)
	} else if len(s.currentLines) > 0 {
		s.currentLines = append(s.currentLines, "")
	}
}

// parseFlagOrArgLine handles a line in the arguments or flags section.
// It appends parsed flags/args to the target slice, handles continuation lines,
// and transitions to the notes section when an unindented non-flag line is found.
func parseFlagOrArgLine(s *parseState, line string, target *[]FlagArg) {
	if arg := ParseFlagLine(line); arg != nil {
		*target = append(*target, *arg)
		s.currentFlag = &(*target)[len(*target)-1]
	} else if s.currentFlag != nil && strings.HasPrefix(line, "      ") {
		desc := strings.TrimSpace(line)
		if s.currentFlag.Description != "" {
			s.currentFlag.Description += " " + desc
		} else {
			s.currentFlag.Description = desc
		}
	} else if strings.TrimSpace(line) == "" {
		// empty line within section
	} else if !strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "-") && line != "" {
		s.section = "notes"
		s.help.Notes = line
		s.currentFlag = nil
	}
}

// finalizeHelp applies post-parse fixups (description fallback and expected output merging).
func finalizeHelp(s *parseState) {
	if s.help.Description == "" && len(s.currentLines) > 0 {
		s.help.Description = FormatDescription(s.currentLines)
	}

	if len(s.expectedOutputLines) > 0 {
		expectedOutput := strings.Join(s.expectedOutputLines, "\n")
		if s.help.Notes != "" {
			s.help.Notes = expectedOutput + "\n\n" + s.help.Notes
		} else {
			s.help.Notes = expectedOutput
		}
	}
}
