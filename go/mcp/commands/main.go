// Package main implements an MCP (Model Context Protocol) server that exposes
// EAC commands as tools for AI assistants. It discovers available commands
// dynamically and executes them via the EAC adapter abstraction.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/core/repository"
)

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

// handleRequest dispatches a parsed request to the appropriate handler.
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

// handleToolCall processes a tools/call request.
func handleToolCall(encoder *json.Encoder, req *MCPRequest) {
	var params CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		sendError(encoder, req.ID, -32602, "Invalid params")
		return
	}
	sendResponse(encoder, req.ID, callTool(&params))
}

// initializeResponse returns the MCP protocol initialization response.
func initializeResponse() map[string]interface{} {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"serverInfo":      map[string]string{"name": "eac-mcp-server", "version": "0.1.0"},
		"capabilities":    map[string]interface{}{"tools": map[string]bool{}},
	}
}

// findRepoRoot walks up directory tree to find repository root.
func findRepoRoot() string {
	root, err := repository.GetRepositoryRoot("")
	if err != nil {
		return ""
	}
	return root
}

// logError logs an error to stderr.
// Note: stderr is intentionally used here because stdout is reserved for
// JSON-RPC protocol messages. This is the correct pattern for MCP servers.
func logError(context string, err error) {
	fmt.Fprintf(os.Stderr, "eac-mcp-server: error %s: %v\n", context, err)
}
