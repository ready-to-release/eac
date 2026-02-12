package main

import "encoding/json"

// MCPRequest represents a JSON-RPC request from the MCP client.
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC response to the MCP client.
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents a JSON-RPC error.
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Tool represents an MCP tool definition.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema defines the JSON Schema for tool input parameters.
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

// Property defines a single property in the tool input schema.
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// CallToolParams holds the parameters for a tools/call request.
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Content []Content `json:"content"`
}

// Content represents a single content item in a tool result.
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

// CommandTree holds the full set of discovered commands and their tree structure.
type CommandTree struct {
	Commands []CommandInfo       `json:"commands"`
	Tree     map[string][]string `json:"tree"`
}

// sendResponse writes a successful JSON-RPC response.
func sendResponse(encoder *json.Encoder, id, result interface{}) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	_ = encoder.Encode(resp) //nolint:errcheck // best-effort JSON-RPC response
}

// sendError writes a JSON-RPC error response.
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
