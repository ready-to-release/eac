package content

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/render"
)

// FormatSingleCommand generates markdown documentation for a single command.
func FormatSingleCommand(ctx context.Context, workspaceRoot string, executor CommandExecutor, cmdName string) (string, error) {
	return GetCommandHelpMarkdown(ctx, workspaceRoot, executor, cmdName)
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
		md, err := GetCommandHelpMarkdown(ctx, workspaceRoot, executor, cmd.Command)
		if err != nil {
			warnf("failed to get help for '%s': %v", cmd.Command, err)
			continue
		}
		sb.WriteString(md)
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
			md, err := GetCommandHelpMarkdown(ctx, workspaceRoot, executor, cmd.Command)
			if err != nil {
				continue
			}
			sb.WriteString(md)
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

	rows := make([][]string, len(categories))
	for i, cat := range categories {
		rows[i] = []string{fmt.Sprintf("**%s**", cat.Name), fmt.Sprintf("%d", cat.CommandCount), cat.Description}
	}

	return FormatTable([]string{"Category", "Commands", "Description"}, rows), nil
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

		sb.WriteString(fmt.Sprintf("| [%s](%s) | %s |\n", cmd.Command, linkPath, render.EscapeTableCell(cmd.Description)))
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
