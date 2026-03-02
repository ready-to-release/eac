package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	eac "github.com/ready-to-release/eac/go/adapters/eac"
	"github.com/ready-to-release/eac/go/core/tool"
)

// cachedCommands holds the result of the first getCommands() call.
// The command set does not change during a server session, so we
// discover once and reuse on subsequent tools/list requests.
var (
	cachedCommands     CommandTree
	cachedCommandsOnce sync.Once

	// toolNameToParts maps MCP tool names (e.g. "get-commit-message") to their
	// command parts (e.g. ["get", "commit-message"]). This avoids the lossy
	// roundtrip of converting all hyphens to spaces, which breaks hyphenated
	// subcommand names like "commit-message".
	toolNameToParts map[string][]string
)

// getCommandTools discovers commands by calling "get commands".
func getCommandTools() []Tool {
	tree := getCommands()
	tools := make([]Tool, 0, len(tree.Commands))
	for i := range tree.Commands {
		if t, ok := commandToTool(&tree.Commands[i]); ok {
			tools = append(tools, t)
		}
	}
	return tools
}

// commandToTool converts a CommandInfo to a Tool, returning false if skipped.
func commandToTool(cmd *CommandInfo) (Tool, bool) {
	toolName := strings.ReplaceAll(cmd.Name, " ", "-")
	if toolName == "get-commands" {
		return Tool{}, false
	}
	return Tool{
		Name:        toolName,
		Description: toolDescription(cmd),
		InputSchema: toolInputSchema(),
	}, true
}

// toolDescription returns a description for the tool.
func toolDescription(cmd *CommandInfo) string {
	if cmd.Description != "" {
		return cmd.Description
	}
	return fmt.Sprintf("Execute '%s' command", cmd.Name)
}

// toolInputSchema returns the standard input schema for command tools.
func toolInputSchema() InputSchema {
	return InputSchema{
		Type: "object",
		Properties: map[string]Property{
			"args": {Type: "string", Description: "Additional arguments (optional)"},
		},
	}
}

// getCommands returns the cached command tree, discovering commands on the
// first call only. The command set does not change during a server session,
// so sync.Once ensures we shell out exactly once (OO-094).
// Also populates the toolNameToParts lookup for correct command dispatch.
func getCommands() CommandTree {
	cachedCommandsOnce.Do(func() {
		cachedCommands = discoverCommands()
		toolNameToParts = buildToolPartsIndex(cachedCommands)
	})
	return cachedCommands
}

// buildToolPartsIndex creates a mapping from MCP tool names to command parts.
func buildToolPartsIndex(tree CommandTree) map[string][]string {
	index := make(map[string][]string, len(tree.Commands))
	for i := range tree.Commands {
		toolName := strings.ReplaceAll(tree.Commands[i].Name, " ", "-")
		index[toolName] = tree.Commands[i].Parts
	}
	return index
}

// discoverCommands calls the commands system to get command info.
func discoverCommands() CommandTree {
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		return emptyCommandTree()
	}
	port := eac.New(repoRoot, tool.GlobalRegistry(), tool.GlobalExecutor())
	result, err := port.Execute(context.Background(), []string{"get", "commands"}, &eac.ExecConfig{
		WorkspaceRoot: repoRoot,
	})
	if err != nil {
		logError("getting commands", err)
		return emptyCommandTree()
	}
	if !result.Success() {
		logError("getting commands", fmt.Errorf("exit code %d: %s", result.ExitCode, string(result.Stderr)))
		return emptyCommandTree()
	}
	return parseCommandTree(result.Stdout)
}

// emptyCommandTree returns an empty command tree.
func emptyCommandTree() CommandTree {
	return CommandTree{Commands: []CommandInfo{}}
}

// parseCommandTree parses JSON output into a CommandTree.
func parseCommandTree(output []byte) CommandTree {
	var tree CommandTree
	if err := json.Unmarshal(output, &tree); err != nil {
		logError("parsing command tree", err)
		return emptyCommandTree()
	}
	return tree
}
