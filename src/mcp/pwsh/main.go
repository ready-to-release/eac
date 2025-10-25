package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// MCP Server for PowerShell command execution

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
				"name":    "mcp-server-pwsh",
				"version": "0.1.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]bool{},
			},
		})

	case "tools/list":
		tools := []Tool{
			{
				Name:        "execute-pwsh",
				Description: "Execute a PowerShell command",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"command": {
							Type:        "string",
							Description: "PowerShell command to execute",
						},
					},
					Required: []string{"command"},
				},
			},
			{
				Name:        "get-pwsh-modules",
				Description: "List available PowerShell modules",
				InputSchema: InputSchema{
					Type:       "object",
					Properties: map[string]Property{},
				},
			},
		}
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

func callTool(params *CallToolParams) ToolResult {
	switch params.Name {
	case "execute-pwsh":
		command, ok := params.Arguments["command"].(string)
		if !ok {
			return ToolResult{
				Content: []Content{{
					Type: "text",
					Text: "Error: command must be a string",
				}},
			}
		}
		output := executePowerShell(command)
		return ToolResult{
			Content: []Content{{
				Type: "text",
				Text: output,
			}},
		}

	case "get-pwsh-modules":
		output := executePowerShell("Get-Module -ListAvailable | Select-Object Name, Version | ConvertTo-Json")
		return ToolResult{
			Content: []Content{{
				Type: "text",
				Text: output,
			}},
		}

	default:
		return ToolResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("Unknown tool: %s", params.Name),
			}},
		}
	}
}

func executePowerShell(command string) string {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-NonInteractive", "-Command", command)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Error: %v\nOutput: %s", err, string(output))
	}

	return strings.TrimSpace(string(output))
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
