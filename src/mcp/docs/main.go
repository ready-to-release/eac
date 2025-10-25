package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// MCP Server for MkDocs documentation access

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

type SearchResult struct {
	File    string   `json:"file"`
	Matches []string `json:"matches"`
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
				"name":    "mcp-server-docs",
				"version": "0.1.0",
			},
			"capabilities": map[string]interface{}{
				"tools":     map[string]bool{},
				"resources": map[string]bool{},
			},
		})

	case "tools/list":
		tools := []Tool{
			{
				Name:        "search-docs",
				Description: "Search documentation for a query",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"query": {
							Type:        "string",
							Description: "Search query",
						},
					},
					Required: []string{"query"},
				},
			},
			{
				Name:        "get-doc-page",
				Description: "Get content of a specific documentation page",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"path": {
							Type:        "string",
							Description: "Path to documentation page",
						},
					},
					Required: []string{"path"},
				},
			},
			{
				Name:        "list-docs",
				Description: "List all available documentation pages",
				InputSchema: InputSchema{
					Type:       "object",
					Properties: map[string]Property{},
				},
			},
			{
				Name:        "build-docs",
				Description: "Build MkDocs static site using Docker container",
				InputSchema: InputSchema{
					Type:       "object",
					Properties: map[string]Property{},
				},
			},
			{
				Name:        "serve-docs",
				Description: "Start MkDocs development server using Docker container",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"detached": {
							Type:        "boolean",
							Description: "Run in background (default: true)",
						},
					},
				},
			},
			{
				Name:        "stop-docs",
				Description: "Stop MkDocs development server",
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
	docsPath := os.Getenv("DOCS_PATH")
	if docsPath == "" {
		docsPath = "."
	}

	switch params.Name {
	case "search-docs":
		query, ok := params.Arguments["query"].(string)
		if !ok {
			return errorResult("query must be a string")
		}
		results := searchDocs(docsPath, strings.ToLower(query))
		jsonBytes, _ := json.MarshalIndent(results, "", "  ")
		return ToolResult{
			Content: []Content{{
				Type: "text",
				Text: string(jsonBytes),
			}},
		}

	case "get-doc-page":
		path, ok := params.Arguments["path"].(string)
		if !ok {
			return errorResult("path must be a string")
		}
		content, err := os.ReadFile(filepath.Join(docsPath, path))
		if err != nil {
			return errorResult(fmt.Sprintf("Error reading file: %v", err))
		}
		return ToolResult{
			Content: []Content{{
				Type: "text",
				Text: string(content),
			}},
		}

	case "list-docs":
		docs := listDocs(docsPath)
		jsonBytes, _ := json.MarshalIndent(docs, "", "  ")
		return ToolResult{
			Content: []Content{{
				Type: "text",
				Text: string(jsonBytes),
			}},
		}

	case "build-docs":
		output := buildDocs()
		return ToolResult{
			Content: []Content{{
				Type: "text",
				Text: output,
			}},
		}

	case "serve-docs":
		detached := true
		if val, ok := params.Arguments["detached"].(bool); ok {
			detached = val
		}
		output := serveDocs(detached)
		return ToolResult{
			Content: []Content{{
				Type: "text",
				Text: output,
			}},
		}

	case "stop-docs":
		output := stopDocs()
		return ToolResult{
			Content: []Content{{
				Type: "text",
				Text: output,
			}},
		}

	default:
		return errorResult(fmt.Sprintf("Unknown tool: %s", params.Name))
	}
}

func searchDocs(docsPath, query string) []SearchResult {
	var results []SearchResult

	filepath.Walk(docsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentLower := strings.ToLower(string(content))
		if strings.Contains(contentLower, query) {
			lines := strings.Split(string(content), "\n")
			var matches []string

			for i, line := range lines {
				if strings.Contains(strings.ToLower(line), query) {
					start := max(0, i-2)
					end := min(len(lines), i+3)
					context := strings.Join(lines[start:end], "\n")
					matches = append(matches, context)
					if len(matches) >= 3 {
						break
					}
				}
			}

			relPath, _ := filepath.Rel(docsPath, path)
			results = append(results, SearchResult{
				File:    relPath,
				Matches: matches,
			})
		}

		return nil
	})

	return results
}

func listDocs(docsPath string) []string {
	var docs []string

	filepath.Walk(docsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		relPath, _ := filepath.Rel(docsPath, path)
		docs = append(docs, relPath)
		return nil
	})

	return docs
}

func errorResult(message string) ToolResult {
	return ToolResult{
		Content: []Content{{
			Type: "text",
			Text: message,
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// buildDocs builds the static documentation site using Docker
func buildDocs() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get absolute path to containers/mkdocs directory
	mkdocsPath := filepath.Join("containers", "mkdocs")
	absPath, err := filepath.Abs(mkdocsPath)
	if err != nil {
		return fmt.Sprintf("Error getting absolute path: %v", err)
	}

	// Run docker-compose from the mkdocs directory
	cmd := exec.CommandContext(ctx, "docker-compose", "run", "--rm", "mkdocs", "mkdocs", "build", "--clean", "--strict")
	cmd.Dir = absPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Sprintf("Error building docs: %v\nOutput: %s", err, string(output))
	}

	return fmt.Sprintf("✓ Documentation built successfully!\nOutput: %s\n\nStatic site created in: site/", string(output))
}

// serveDocs starts the MkDocs development server using Docker
func serveDocs(detached bool) string {
	ctx := context.Background()

	// Get absolute path to containers/mkdocs directory
	mkdocsPath := filepath.Join("containers", "mkdocs")
	absPath, err := filepath.Abs(mkdocsPath)
	if err != nil {
		return fmt.Sprintf("Error getting absolute path: %v", err)
	}

	var cmd *exec.Cmd
	if detached {
		// Run in detached mode
		cmd = exec.CommandContext(ctx, "docker-compose", "up", "-d")
	} else {
		// Run in foreground (note: will block)
		cmd = exec.CommandContext(ctx, "docker-compose", "up")
	}
	cmd.Dir = absPath

	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Sprintf("Error starting docs server: %v\nOutput: %s", err, string(output))
	}

	if detached {
		return fmt.Sprintf("✓ Documentation server started!\n\nAccess at: http://localhost:8000\n\nTo stop: Use stop-docs tool or run:\n  cd containers/mkdocs && docker-compose down\n\nOutput: %s", string(output))
	}

	return fmt.Sprintf("Documentation server output:\n%s", string(output))
}

// stopDocs stops the MkDocs development server
func stopDocs() string {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get absolute path to containers/mkdocs directory
	mkdocsPath := filepath.Join("containers", "mkdocs")
	absPath, err := filepath.Abs(mkdocsPath)
	if err != nil {
		return fmt.Sprintf("Error getting absolute path: %v", err)
	}

	cmd := exec.CommandContext(ctx, "docker-compose", "down")
	cmd.Dir = absPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Sprintf("Error stopping docs server: %v\nOutput: %s", err, string(output))
	}

	return fmt.Sprintf("✓ Documentation server stopped.\n\nOutput: %s", string(output))
}
