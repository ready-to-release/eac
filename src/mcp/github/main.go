package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// MCP Server for GitHub CLI integration

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
				"name":    "mcp-server-github",
				"version": "0.1.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]bool{},
			},
		})

	case "tools/list":
		tools := []Tool{
			{
				Name:        "gh-repo-view",
				Description: "View repository details",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"repo": {
							Type:        "string",
							Description: "Repository in format owner/repo",
						},
					},
					Required: []string{"repo"},
				},
			},
			{
				Name:        "gh-issue-create",
				Description: "Create a new issue",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"title": {
							Type:        "string",
							Description: "Issue title",
						},
						"body": {
							Type:        "string",
							Description: "Issue body",
						},
					},
					Required: []string{"title"},
				},
			},
			{
				Name:        "gh-pr-list",
				Description: "List pull requests",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"state": {
							Type:        "string",
							Description: "PR state: open, closed, merged, all",
						},
					},
				},
			},
			{
				Name:        "gh-run-list",
				Description: "List workflow runs",
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
	case "gh-repo-view":
		repo, ok := params.Arguments["repo"].(string)
		if !ok {
			return errorResult("repo must be a string")
		}
		output := execGH("repo", "view", repo)
		return textResult(output)

	case "gh-issue-create":
		title, _ := params.Arguments["title"].(string)
		body, _ := params.Arguments["body"].(string)

		args := []string{"issue", "create", "--title", title}
		if body != "" {
			args = append(args, "--body", body)
		}
		output := execGH(args...)
		return textResult(output)

	case "gh-pr-list":
		state, ok := params.Arguments["state"].(string)
		if !ok {
			state = "open"
		}
		output := execGH("pr", "list", "--state", state, "--json", "number,title,author,createdAt")
		return textResult(output)

	case "gh-run-list":
		output := execGH("run", "list", "--json", "databaseId,name,status,conclusion,createdAt")
		return textResult(output)

	default:
		return errorResult(fmt.Sprintf("Unknown tool: %s", params.Name))
	}
}

func execGH(args ...string) string {
	ghPath := findGH()
	if ghPath == "" {
		return "Error: GitHub CLI (gh) not found. Please install it from https://cli.github.com/"
	}

	cmd := exec.Command(ghPath, args...)

	// Pass through GITHUB_TOKEN if set
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		cmd.Env = append(os.Environ(), "GH_TOKEN="+token)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Error: %v\nOutput: %s", err, string(output))
	}

	return strings.TrimSpace(string(output))
}

// findGH locates the gh CLI executable
func findGH() string {
	// Try finding gh in PATH first
	if path, err := exec.LookPath("gh"); err == nil {
		return path
	}

	// On Windows, check common installation locations
	if runtime.GOOS == "windows" {
		commonPaths := []string{
			filepath.Join(os.Getenv("ProgramFiles"), "GitHub CLI", "gh.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "GitHub CLI", "gh.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "GitHub CLI", "gh.exe"),
		}

		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return ""
}

func textResult(text string) ToolResult {
	return ToolResult{
		Content: []Content{{
			Type: "text",
			Text: text,
		}},
	}
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
