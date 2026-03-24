package content

import (
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

// GetCommandHelpMarkdown runs --help --help-format=markdown and returns the markdown directly.
func GetCommandHelpMarkdown(ctx context.Context, workspaceRoot string, executor CommandExecutor, cmdName string) (string, error) {
	args := append(strings.Fields(cmdName), "--help", "--help-format=markdown")
	output, err := executor.Run(ctx, workspaceRoot, args)
	if err != nil {
		errLower := strings.ToLower(err.Error())
		if strings.Contains(errLower, "unknown command") ||
			strings.Contains(errLower, "command not found") ||
			strings.Contains(errLower, "not a valid command") ||
			strings.Contains(errLower, "no such command") {
			return "", fmt.Errorf("%w: '%s'", ErrCommandNotFound, cmdName)
		}
		return "", fmt.Errorf("failed to run '%s --help': %v", cmdName, err)
	}
	return strings.TrimSpace(output), nil
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
