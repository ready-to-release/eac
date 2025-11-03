# Work with MCP Servers

Learn how to test, debug, and create new MCP servers for the VS Code extension.

---

## Overview

**What you'll learn:**

- Test MCP servers manually with JSON-RPC
- Debug server communication
- Create a new MCP server
- Configure servers in .mcp.json
- Use watch mode for development

**Time Required:** 20-30 minutes

---

## Available MCP Servers

Four servers are configured in `.mcp.json`:

| Server | Purpose | Location | Language |
|--------|---------|----------|----------|
| **vscode** | VSCode actions (Git, etc.) | `src/mcp/vscode/` | Go |
| **pwsh** | PowerShell commands | `src/mcp/pwsh/` | Go |
| **docs** | Documentation management | `src/mcp/docs/` | Go |
| **github** | GitHub API integration | `src/mcp/github/` | Go |

---

## Testing an MCP Server Manually

Manual testing helps you debug server logic without involving VSCode.

### Start the Server

```bash
cd src/mcp/vscode
go run .
```

The server will wait for JSON-RPC commands on stdin.

### Send JSON-RPC Commands

#### Initialize the Server

```json
{"jsonrpc":"2.0","id":1,"method":"initialize"}
```

**Expected response:**

```json
{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","serverInfo":{"name":"vscode-server","version":"1.0.0"}}}
```

#### List Available Tools

```json
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
```

**Expected response:**

```json
{"jsonrpc":"2.0","id":2,"result":{"tools":[{"name":"vscode-action","description":"Execute VSCode actions","inputSchema":{...}}]}}
```

#### Call a Tool

```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vscode-action","arguments":{"action":"git-commit","message":"test commit"}}}
```

**Expected response:**

```json
{"jsonrpc":"2.0","id":3,"result":{"content":[{"type":"text","text":"Committed successfully"}]}}
```

### Exit the Server

Press `Ctrl+C`

---

## Debugging Server Communication

### Add Logging to Server

**File:** `src/mcp/vscode/main.go`

```go
func executeAction(action string, message string) string {
    // Log to stderr (shown in VSCode Debug Console)
    fmt.Fprintf(os.Stderr, "Executing action: %s with message: %s\n", action, message)

    switch action {
    case "git-commit":
        result := handleGitCommit(message)
        fmt.Fprintf(os.Stderr, "Result: %s\n", result)
        return result
    // ...
    }
}
```

**Why stderr?**

- stdout is used for JSON-RPC messages
- stderr appears in VSCode Debug Console

### View Logs in VSCode

1. Open Debug Console in original VSCode window
2. Look for stderr output from MCP server
3. Check for errors, timing, data flow

### Common Debugging Patterns

**Log request received:**

```go
fmt.Fprintf(os.Stderr, "[DEBUG] Received request: %+v\n", request)
```

**Log response before sending:**

```go
fmt.Fprintf(os.Stderr, "[DEBUG] Sending response: %+v\n", response)
```

**Log errors:**

```go
if err != nil {
    fmt.Fprintf(os.Stderr, "[ERROR] %s: %v\n", operation, err)
    return fmt.Sprintf("Error: %v", err)
}
```

---

## Creating a New MCP Server

Let's create a custom server from scratch.

### Step 1: Create Server Directory

```bash
mkdir -p src/mcp/my-server
cd src/mcp/my-server
go mod init my-server
```

### Step 2: Implement JSON-RPC Handlers

**File:** `src/mcp/my-server/main.go`

```go
package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
)

type JSONRPCRequest struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      int             `json:"id"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      int         `json:"id"`
    Result  interface{} `json:"result,omitempty"`
    Error   interface{} `json:"error,omitempty"`
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)

    for scanner.Scan() {
        line := scanner.Text()

        var request JSONRPCRequest
        if err := json.Unmarshal([]byte(line), &request); err != nil {
            fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
            continue
        }

        var response JSONRPCResponse
        response.JSONRPC = "2.0"
        response.ID = request.ID

        switch request.Method {
        case "initialize":
            response.Result = handleInitialize()
        case "tools/list":
            response.Result = handleToolsList()
        case "tools/call":
            response.Result = handleToolsCall(request.Params)
        default:
            response.Error = map[string]interface{}{
                "code":    -32601,
                "message": "Method not found",
            }
        }

        output, _ := json.Marshal(response)
        fmt.Println(string(output))
    }
}

func handleInitialize() map[string]interface{} {
    return map[string]interface{}{
        "protocolVersion": "2024-11-05",
        "serverInfo": map[string]string{
            "name":    "my-server",
            "version": "1.0.0",
        },
    }
}

func handleToolsList() map[string]interface{} {
    return map[string]interface{}{
        "tools": []map[string]interface{}{
            {
                "name":        "my-tool",
                "description": "My custom tool",
                "inputSchema": map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "input": map[string]string{
                            "type":        "string",
                            "description": "Input parameter",
                        },
                    },
                    "required": []string{"input"},
                },
            },
        },
    }
}

func handleToolsCall(params json.RawMessage) map[string]interface{} {
    var callParams struct {
        Name      string                 `json:"name"`
        Arguments map[string]interface{} `json:"arguments"`
    }

    json.Unmarshal(params, &callParams)

    // Implement your tool logic here
    result := fmt.Sprintf("Processed: %v", callParams.Arguments["input"])

    return map[string]interface{}{
        "content": []map[string]string{
            {
                "type": "text",
                "text": result,
            },
        },
    }
}
```

### Step 3: Create Run Script

**File:** `src/mcp/my-server/run.sh`

```bash
#!/bin/bash
cd "$(dirname "$0")"
go run .
```

Make it executable:

```bash
chmod +x run.sh
```

### Step 4: Add to .mcp.json

**File:** `.mcp.json` (project root)

```json
{
  "mcpServers": {
    "vscode": {
      "command": "bash",
      "args": ["src/mcp/vscode/run.sh"]
    },
    "pwsh": {
      "command": "bash",
      "args": ["src/mcp/pwsh/run.sh"]
    },
    "docs": {
      "command": "bash",
      "args": ["src/mcp/docs/run.sh"]
    },
    "github": {
      "command": "bash",
      "args": ["src/mcp/github/run.sh"]
    },
    "my-server": {
      "command": "bash",
      "args": ["src/mcp/my-server/run.sh"]
    }
  }
}
```

### Step 5: Test Your Server

```bash
cd src/mcp/my-server
go run .
```

Send test command:

```json
{"jsonrpc":"2.0","id":1,"method":"initialize"}
```

---

## Development Workflow

### Making Changes to Extension

**1. Edit the code:**

```typescript
// .vscode/extensions/claude-mcp-vscode/src/extension.ts
export function activate(context: vscode.ExtensionContext) {
    // Your changes here
}
```

**2. Reload:**

- Press `Ctrl+R` (or `Cmd+R`) in Extension Development Host
- Changes take effect immediately

**3. View logs:**

- Debug Console in original VSCode window
- Check `console.log()` output

### Watch Mode (Recommended)

Auto-recompile TypeScript on save:

```bash
cd .vscode/extensions/claude-mcp-vscode
npm run watch
```

**Benefits:**

- No manual compilation needed
- Instant feedback on syntax errors
- Faster iteration

---

## MCP Protocol Reference

### Required Methods

All MCP servers must implement:

**1. initialize**:

- Handshake between client and server
- Returns protocol version and server info

**2. tools/list**:

- Lists available tools
- Returns tool schemas

**3. tools/call**:

- Executes a tool
- Returns results or errors

### Message Format

**Request:**

```json
{
  "jsonrpc": "2.0",
  "id": <number>,
  "method": "<string>",
  "params": <object>
}
```

**Response:**

```json
{
  "jsonrpc": "2.0",
  "id": <number>,
  "result": <any> | "error": <object>
}
```

### Error Codes

| Code | Meaning |
|------|---------|
| -32700 | Parse error |
| -32600 | Invalid request |
| -32601 | Method not found |
| -32602 | Invalid params |
| -32603 | Internal error |

---

## Best Practices

**1. Always Log to stderr**:

- Use `fmt.Fprintf(os.Stderr, ...)` for debugging
- Keep stdout clean for JSON-RPC

**2. Handle Errors Gracefully**:

- Never crash on invalid input
- Return descriptive error messages

**3. Validate Input**:

- Check all parameters before processing
- Return clear validation errors

**4. Keep Servers Focused**:

- One server per domain (git, docs, github)
- Don't mix unrelated functionality

**5. Test Manually First**:

- Verify JSON-RPC protocol works
- Test edge cases and errors
- Then integrate with extension

---

## Troubleshooting

**Problem:** Server doesn't respond

**Solution:**

- Check if server is reading from stdin
- Verify JSON format (use a validator)
- Check for compilation errors

---

**Problem:** Can't see server output

**Solution:**

- Log to stderr, not stdout
- Open Debug Console in VSCode
- Check if server process is running

---

**Problem:** Extension can't find server

**Solution:**

- Verify path in .mcp.json
- Check run.sh has execute permissions
- Ensure run.sh changes to correct directory

---

## Related Documentation

- **Architecture**: [VS Code Extension Architecture](../../explanation/vscode-extension-architecture.md) - How MCP communication works
- **Add Actions**: [Add a New Action](add-action.md) - Using servers from extension
- **Troubleshooting**: [Troubleshoot Common Issues](troubleshoot.md) - More debugging tips
- **MCP Protocol**: [Model Context Protocol Spec](https://modelcontextprotocol.io/)
