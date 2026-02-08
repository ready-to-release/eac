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
