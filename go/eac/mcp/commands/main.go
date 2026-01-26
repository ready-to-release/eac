// Package main implements an MCP (Model Context Protocol) server that exposes
// EAC commands as tools for AI assistants. It discovers available commands
// dynamically and executes them via either direct binary or Docker-based r2r.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
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

// CommandInfo from go/eac/commands/describe-commands.go.
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
		"serverInfo":      map[string]string{"name": "mcp-server-commands", "version": "0.1.0"},
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

// getCommands calls the commands system to get command info.
func getCommands() CommandTree {
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		return emptyCommandTree()
	}
	output, err := buildCommand(repoRoot, []string{"get", "commands"}).Output()
	if err != nil {
		logError("getting commands", err)
		return emptyCommandTree()
	}
	return parseCommandTree(output)
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
	output, err := buildCommand(repoRoot, cmdParts).CombinedOutput()
	return formatCommandOutput(commandName, output, err)
}

// formatCommandOutput formats command output or error for the tool result.
func formatCommandOutput(commandName string, output []byte, err error) string {
	if err != nil {
		return fmt.Sprintf("Error executing command '%s': %v\n\nOutput:\n%s", commandName, err, string(output))
	}
	return strings.TrimSpace(string(output))
}

// buildCmdParts builds command arguments from command name and additional args.
func buildCmdParts(commandName, additionalArgs string) []string {
	cmdParts := strings.Fields(commandName)
	if additionalArgs != "" {
		cmdParts = append(cmdParts, strings.Fields(additionalArgs)...)
	}
	return cmdParts
}

// buildCommand creates an exec.Cmd for running EAC commands.
// Uses direct binary if EAC_USE_DIRECT_BINARY=true, otherwise uses r2r eac (Docker).
func buildCommand(repoRoot string, cmdParts []string) *exec.Cmd {
	var cmd *exec.Cmd
	if os.Getenv("EAC_USE_DIRECT_BINARY") == "true" {
		cmd = exec.Command(paths.CommandsBinaryPath(repoRoot), cmdParts...) //nolint:gosec // MCP server executes EAC commands by design
	} else {
		cmd = exec.Command("r2r", append([]string{"eac"}, cmdParts...)...) //nolint:gosec // MCP server executes EAC commands by design
	}
	cmd.Dir = repoRoot
	return cmd
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
