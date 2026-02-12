package main

import (
	"context"
	"fmt"
	"strings"

	eac "github.com/ready-to-release/eac/go/adapters/eac"
	"github.com/ready-to-release/eac/go/core/tool"
)

// callTool dispatches a tool call to the appropriate command.
func callTool(params *CallToolParams) ToolResult {
	commandName := strings.ReplaceAll(params.Name, "-", " ")
	args := extractArgs(params.Arguments)
	return textResult(execCommand(commandName, args))
}

// extractArgs extracts the "args" string from tool arguments.
func extractArgs(arguments map[string]interface{}) string {
	if argsVal, ok := arguments["args"].(string); ok {
		return argsVal
	}
	return ""
}

// execCommand executes a command via the commands system.
func execCommand(commandName, additionalArgs string) string {
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		return "Error: Could not find repository root"
	}
	cmdParts := buildCmdParts(commandName, additionalArgs)
	port := eac.New(repoRoot, tool.GlobalRegistry(), tool.GlobalExecutor())
	result, err := port.Execute(context.Background(), cmdParts, &eac.ExecConfig{
		WorkspaceRoot: repoRoot,
	})
	return formatCommandOutput(commandName, result, err)
}

// formatCommandOutput formats command output or error for the tool result.
func formatCommandOutput(commandName string, result *eac.Result, err error) string {
	if err != nil {
		return fmt.Sprintf("Error executing command '%s': %v", commandName, err)
	}
	if !result.Success() {
		combined := string(result.Stdout) + string(result.Stderr)
		return fmt.Sprintf("Error executing command '%s': exit code %d\n\nOutput:\n%s", commandName, result.ExitCode, combined)
	}
	return strings.TrimSpace(string(result.Stdout))
}

// buildCmdParts builds command arguments from command name and additional args.
func buildCmdParts(commandName, additionalArgs string) []string {
	cmdParts := strings.Fields(commandName)
	if additionalArgs != "" {
		cmdParts = append(cmdParts, strings.Fields(additionalArgs)...)
	}
	return cmdParts
}

// textResult wraps a string in a ToolResult with a single text content item.
func textResult(text string) ToolResult {
	return ToolResult{
		Content: []Content{{
			Type: "text",
			Text: text,
		}},
	}
}
