package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/core/repository"
)

// MCP Server for EAC Commands integration

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

// CommandInfo from src/commands/describe-commands.go
type CommandInfo struct {
	Name        string   `json:"name"`
	Parts       []string `json:"parts"`
	Description string   `json:"description"`
	Parent      string   `json:"parent"`
	IsLeaf      bool     `json:"is_leaf"`
}

type CommandTree struct {
	Commands []CommandInfo      `json:"commands"`
	Tree     map[string][]string `json:"tree"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req MCPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			sendError(encoder, nil, -32700, "Parse error")
			continue
		}

		handleRequest(encoder, &req)
	}
}

func handleRequest(encoder *json.Encoder, req *MCPRequest) {
	switch req.Method {
	case "initialize":
		sendResponse(encoder, req.ID, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    "mcp-server-commands",
				"version": "0.1.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]bool{},
			},
		})

	case "tools/list":
		tools := getCommandTools()
		sendResponse(encoder, req.ID, map[string]interface{}{
			"tools": tools,
		})

	case "tools/call":
		var params CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			sendError(encoder, req.ID, -32602, "Invalid params")
			return
		}

		result := callTool(&params)
		sendResponse(encoder, req.ID, result)

	default:
		sendError(encoder, req.ID, -32601, "Method not found")
	}
}

// getCommandTools discovers commands by calling "describe commands"
func getCommandTools() []Tool {
	tree := describeCommands()
	var tools []Tool

	for _, cmd := range tree.Commands {
		// Convert command name to kebab-case for tool name
		toolName := strings.ReplaceAll(cmd.Name, " ", "-")

		// Skip meta commands
		if toolName == "list-commands" || toolName == "describe-commands" {
			continue
		}

		description := cmd.Description
		if description == "" {
			description = fmt.Sprintf("Execute '%s' command", cmd.Name)
		}

		tools = append(tools, Tool{
			Name:        toolName,
			Description: description,
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"args": {
						Type:        "string",
						Description: "Additional arguments (optional)",
					},
				},
			},
		})
	}

	return tools
}

// describeCommands calls "go run ./src/commands describe commands" to get command info
func describeCommands() CommandTree {
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		return CommandTree{Commands: []CommandInfo{}}
	}

	cmdPath := filepath.Join(repoRoot, "src", "commands")
	cmd := exec.Command("go", "run", ".", "describe", "commands")
	cmd.Dir = cmdPath

	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error describing commands: %v\n", err)
		return CommandTree{Commands: []CommandInfo{}}
	}

	var tree CommandTree
	if err := json.Unmarshal(output, &tree); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing command tree: %v\n", err)
		return CommandTree{Commands: []CommandInfo{}}
	}

	return tree
}

func callTool(params *CallToolParams) ToolResult {
	// Convert tool name back to command name (kebab-case to space-separated)
	commandName := strings.ReplaceAll(params.Name, "-", " ")

	// Get additional args if provided
	args := ""
	if argsVal, ok := params.Arguments["args"].(string); ok {
		args = argsVal
	}

	output := execCommand(commandName, args)
	return textResult(output)
}

// execCommand executes a command via "go run ./src/commands <command> [args]"
func execCommand(commandName string, additionalArgs string) string {
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		return "Error: Could not find repository root"
	}

	cmdPath := filepath.Join(repoRoot, "src", "commands")

	// Build command arguments
	cmdParts := strings.Fields(commandName)
	if additionalArgs != "" {
		cmdParts = append(cmdParts, strings.Fields(additionalArgs)...)
	}

	// Prepend "go run ."
	cmdArgs := append([]string{"run", "."}, cmdParts...)

	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = cmdPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Error executing command '%s': %v\n\nOutput:\n%s", commandName, err, string(output))
	}

	return strings.TrimSpace(string(output))
}

// findRepoRoot walks up directory tree to find repository root
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

func sendResponse(encoder *json.Encoder, id interface{}, result interface{}) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	encoder.Encode(resp)
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
	encoder.Encode(resp)
}
