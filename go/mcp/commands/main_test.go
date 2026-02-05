package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// handleTestRequest is a test helper that sends a request and returns the parsed response.
// Note: JSON unmarshals numeric IDs as float64, so response.ID will be float64.
func handleTestRequest(t *testing.T, req *MCPRequest) MCPResponse {
	t.Helper()
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	handleRequest(encoder, req)

	var resp MCPResponse
	err := json.Unmarshal(buf.Bytes(), &resp)
	require.NoError(t, err)
	return resp
}

func TestMCPServerInitialize(t *testing.T) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}

	resp := handleTestRequest(t, &req)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, float64(1), resp.ID)
	assert.Nil(t, resp.Error)
	require.NotNil(t, resp.Result)

	result, ok := resp.Result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "2024-11-05", result["protocolVersion"])

	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "mcp-server-commands", serverInfo["name"])
	assert.Equal(t, "0.1.0", serverInfo["version"])
}

func TestMCPServerToolsList(t *testing.T) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	resp := handleTestRequest(t, &req)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, float64(2), resp.ID)
	assert.Nil(t, resp.Error)
	require.NotNil(t, resp.Result)

	result, ok := resp.Result.(map[string]interface{})
	require.True(t, ok)
	_, hasTools := result["tools"]
	assert.True(t, hasTools, "Result should have 'tools' key")
}

func TestMCPServerToolsCall_ValidCommand(t *testing.T) {
	params := CallToolParams{
		Name:      "show-modules",
		Arguments: map[string]interface{}{},
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	resp := handleTestRequest(t, &req)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, float64(3), resp.ID)
	assert.NotNil(t, resp.Result)
}

func TestMCPServerToolsCall_WithArguments(t *testing.T) {
	params := CallToolParams{
		Name: "get-modules",
		Arguments: map[string]interface{}{
			"args": "--format json",
		},
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	resp := handleTestRequest(t, &req)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, float64(4), resp.ID)
	assert.NotNil(t, resp.Result)
}

func TestMCPServerToolsCall_InvalidParams(t *testing.T) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "tools/call",
		Params:  json.RawMessage(`{invalid json`),
	}

	resp := handleTestRequest(t, &req)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, float64(5), resp.ID)
	assert.Nil(t, resp.Result)
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32602, resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Invalid params")
}

func TestMCPServerMethodNotFound(t *testing.T) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      6,
		Method:  "unknown/method",
	}

	resp := handleTestRequest(t, &req)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, float64(6), resp.ID)
	assert.Nil(t, resp.Result)
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32601, resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Method not found")
}

func TestEnvironmentVariableMode(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{"unset defaults to Docker mode", "", false},
		{"true enables direct mode", "true", true},
		{"false uses Docker mode", "false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue == "" {
				os.Unsetenv(environments.EnvEACUseDirectBinary)
			} else {
				os.Setenv(environments.EnvEACUseDirectBinary, tt.envValue)
				defer os.Unsetenv(environments.EnvEACUseDirectBinary)
			}

			got := os.Getenv(environments.EnvEACUseDirectBinary) == "true"
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTextResult(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"non-empty text", "Hello, World!"},
		{"empty text", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := textResult(tt.input)

			require.Len(t, result.Content, 1)
			assert.Equal(t, "text", result.Content[0].Type)
			assert.Equal(t, tt.input, result.Content[0].Text)
		})
	}
}

func TestSendResponse(t *testing.T) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)

	sendResponse(encoder, 1, map[string]string{"status": "ok"})

	var resp MCPResponse
	err := json.Unmarshal(buf.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, float64(1), resp.ID)
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Result)
}

func TestSendError(t *testing.T) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)

	sendError(encoder, 1, -32600, "Invalid Request")

	var resp MCPResponse
	err := json.Unmarshal(buf.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, float64(1), resp.ID)
	assert.Nil(t, resp.Result)
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32600, resp.Error.Code)
	assert.Equal(t, "Invalid Request", resp.Error.Message)
}

func TestGetCommandTools_Structure(t *testing.T) {
	tools := getCommandTools()

	for _, tool := range tools {
		assert.NotEmpty(t, tool.Name, "Tool name should not be empty")
		assert.NotEmpty(t, tool.Description, "Tool description should not be empty")
		assert.Equal(t, "object", tool.InputSchema.Type)
		require.Contains(t, tool.InputSchema.Properties, "args")

		argsProp := tool.InputSchema.Properties["args"]
		assert.Equal(t, "string", argsProp.Type)
		assert.NotEmpty(t, argsProp.Description)
	}
}

func TestFindRepoRoot_NotInRepo(t *testing.T) {
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	root := findRepoRoot()
	assert.Equal(t, "", root, "Should return empty string when not in a repo")
}

func TestCommandTreeParsing(t *testing.T) {
	jsonData := `{
		"commands": [
			{
				"name": "show modules",
				"parts": ["show", "modules"],
				"description": "Display all modules",
				"parent": "",
				"is_leaf": true
			},
			{
				"name": "get dependencies",
				"parts": ["get", "dependencies"],
				"description": "Get module dependencies",
				"parent": "",
				"is_leaf": true
			}
		],
		"tree": {
			"show": ["modules"],
			"get": ["dependencies"]
		}
	}`

	var tree CommandTree
	err := json.Unmarshal([]byte(jsonData), &tree)
	require.NoError(t, err)

	require.Len(t, tree.Commands, 2)
	assert.Equal(t, "show modules", tree.Commands[0].Name)
	assert.Equal(t, "Display all modules", tree.Commands[0].Description)
	assert.True(t, tree.Commands[0].IsLeaf)
}
