// Package main implements an MCP (Model Context Protocol) server that exposes
// EAC commands as tools for AI assistants. It discovers available commands
// dynamically and executes them via the EAC adapter abstraction.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	eac "github.com/ready-to-release/eac/go/adapters/eac"
	"github.com/ready-to-release/eac/go/core/repository"
)

// cachedCommands holds the result of the first getCommands() call.
// The command set does not change during a server session, so we
// discover once and reuse on subsequent tools/list requests.
var (
	cachedCommands     CommandTree
	cachedCommandsOnce sync.Once
)

type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolResult struct {
	Content []Content `json:"content"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// CommandInfo from go/cli/eac/describe-commands.go.
type CommandInfo struct {
	Name        string   `json:"name"`
	Parts       []string `json:"parts"`
	Description string   `json:"description"`
	Parent      string   `json:"parent"`
	IsLeaf      bool     `json:"is_leaf"`
}

type CommandTree struct {
	Commands []CommandInfo       `json:"commands"`
	Tree     map[string][]string `json:"tree"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	processRequests(scanner, encoder)
}

// processRequests reads JSON-RPC requests from the scanner and dispatches them.
func processRequests(scanner *bufio.Scanner, encoder *json.Encoder) {
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		processLine(encoder, line)
	}
}

// processLine parses and handles a single JSON-RPC request line.
func processLine(encoder *json.Encoder, line string) {
	var req MCPRequest
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		sendError(encoder, nil, -32700, "Parse error")
		return
	}
	handleRequest(encoder, &req)
}

func handleRequest(encoder *json.Encoder, req *MCPRequest) {
	switch req.Method {
	case "initialize":
		handleInitialize(encoder, req)
	case "tools/list":
		handleToolsList(encoder, req)
	case "tools/call":
		handleToolCall(encoder, req)
	default:
		sendError(encoder, req.ID, -32601, "Method not found")
	}
}

// handleInitialize responds to the MCP initialize request.
func handleInitialize(encoder *json.Encoder, req *MCPRequest) {
	sendResponse(encoder, req.ID, initializeResponse())
}

// handleToolsList responds with the list of available tools.
func handleToolsList(encoder *json.Encoder, req *MCPRequest) {
	sendResponse(encoder, req.ID, map[string]interface{}{"tools": getCommandTools()})
}

// initializeResponse returns the MCP protocol initialization response.
func initializeResponse() map[string]interface{} {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"serverInfo":      map[string]string{"name": "eac-mcp-server", "version": "0.1.0"},
		"capabilities":    map[string]interface{}{"tools": map[string]bool{}},
	}
}

// handleToolCall processes a tools/call request.
func handleToolCall(encoder *json.Encoder, req *MCPRequest) {
	var params CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		sendError(encoder, req.ID, -32602, "Invalid params")
		return
	}
	sendResponse(encoder, req.ID, callTool(&params))
}

// getCommandTools discovers commands by calling "get commands".
func getCommandTools() []Tool {
	tree := getCommands()
	tools := make([]Tool, 0, len(tree.Commands))
	for i := range tree.Commands {
		if tool, ok := commandToTool(&tree.Commands[i]); ok {
			tools = append(tools, tool)
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
func getCommands() CommandTree {
	cachedCommandsOnce.Do(func() {
		cachedCommands = discoverCommands()
	})
	return cachedCommands
}

// discoverCommands calls the commands system to get command info.
func discoverCommands() CommandTree {
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		return emptyCommandTree()
	}
	port := eac.New(repoRoot)
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

// logError logs an error to stderr.
func logError(context string, err error) {
	fmt.Fprintf(os.Stderr, "Error %s: %v\n", context, err)
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
	port := eac.New(repoRoot)
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

// findRepoRoot walks up directory tree to find repository root.
func findRepoRoot() string {
	root, err := repository.GetRepositoryRoot("")
	if err != nil {
		return ""
	}
	return root
}

func textResult(text string) ToolResult {
	return ToolResult{
		Content: []Content{{
			Type: "text",
			Text: text,
		}},
	}
}

func sendResponse(encoder *json.Encoder, id, result interface{}) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	_ = encoder.Encode(resp) //nolint:errcheck // best-effort JSON-RPC response
}

func sendError(encoder *json.Encoder, id interface{}, code int, message string) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}
	_ = encoder.Encode(resp) //nolint:errcheck // best-effort JSON-RPC error
}
