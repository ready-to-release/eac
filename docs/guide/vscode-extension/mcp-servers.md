# MCP Servers Guide

The CLI project includes multiple Model Context Protocol (MCP) servers implemented in Go.

## Available Servers

All servers are configured in `.mcp.json`:

1. **pwsh** - Execute PowerShell commands
2. **docs** - Documentation management
3. **github** - GitHub API integration
4. **vscode** - VSCode action integration

## Running MCP Servers

### Via VSCode Extension

The VSCode extension automatically manages the server lifecycle. When you click the robot button, it spawns the appropriate MCP server process.

### Manual Testing

MCP servers are executed via their `run.sh` scripts:

```bash
# Example: Run the VSCode MCP server
bash src/mcp/vscode/run.sh
```

Or directly with Go:

```bash
cd src/mcp/vscode
go run .
```

## Testing MCP Servers

### Manual Test of VSCode MCP Server

You can test the MCP server directly before using it with the extension:

```bash
cd src/mcp/vscode
go run .
```

The server will wait for JSON-RPC input. Test it with:

**1. Initialize the server:**

```json
{"jsonrpc":"2.0","id":1,"method":"initialize"}
```

**Expected Response:**

```json
{
  "jsonrpc":"2.0",
  "id":1,
  "result":{
    "protocolVersion":"2024-11-05",
    "serverInfo":{"name":"mcp-server-vscode","version":"0.1.0"},
    "capabilities":{"tools":{}}
  }
}
```

**2. List available tools:**

```json
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
```

**Expected Response:**

```json
{
  "jsonrpc":"2.0",
  "id":2,
  "result":{
    "tools":[
      {
        "name":"vscode-action",
        "description":"Execute a VSCode action",
        "inputSchema":{
          "type":"object",
          "properties":{
            "action":{"type":"string","description":"Action to execute (e.g., 'git-commit', 'git-push', 'git-pull')"},
            "message":{"type":"string","description":"Optional message for the action"}
          },
          "required":["action"]
        }
      }
    ]
  }
}
```

**3. Call a tool:**

```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vscode-action","arguments":{"action":"git-commit","message":"test commit"}}}
```

**Expected Response:**

```json
{
  "jsonrpc":"2.0",
  "id":3,
  "result":{
    "content":[
      {"type":"text","text":"Executing action: git-commit with message: test commit"}
    ]
  }
}
```

Press `Ctrl+C` to exit the server.

## MCP Server Development

Each MCP server is a standalone Go module:

```bash
cd src/mcp/<server-name>
go run .              # Run the server
go build              # Build binary
go test               # Run tests (when added)
```

## Communication Flow

```
VSCode Extension
    ↓ (spawn process)
Go MCP Server (src/mcp/vscode/main.go)
    ↓ (JSON-RPC via stdin/stdout)
Tool: vscode-action
    ↓ (execute action)
Result
    ↓ (JSON response)
VSCode Notification
```

## Adding a New MCP Server

1. **Create server directory:**
   ```bash
   mkdir -p src/mcp/my-server
   cd src/mcp/my-server
   ```

2. **Initialize Go module:**
   ```bash
   go mod init my-server
   ```

3. **Implement MCP protocol:**
   - Handle JSON-RPC requests
   - Implement `initialize`, `tools/list`, `tools/call` methods
   - Return responses via stdout

4. **Add to `.mcp.json`:**
   ```json
   {
     "mcpServers": {
       "my-server": {
         "command": "bash",
         "args": ["src/mcp/my-server/run.sh"]
       }
     }
   }
   ```

5. **Create run script:**
   ```bash
   #!/bin/bash
   cd "$(dirname "$0")"
   go run .
   ```

## Troubleshooting

### Server Not Starting

1. Check Go installation: `go version`
2. Verify server path in `.mcp.json`
3. Test server manually: `cd src/mcp/<name> && go run .`

### JSON-RPC Errors

1. Validate JSON format
2. Check method name spelling
3. Ensure required parameters are provided
4. View server logs in VSCode Output panel

### Extension Can't Connect

1. Verify server runs without errors
2. Check file permissions on `run.sh` script
3. Ensure workspace path is correct
4. Review Debug Console for error messages

## Related Documentation

- [VSCode Extension Guide](index.md) - Using and developing the extension
- [Setup Guide](setup.md) - Installation and configuration
- [Usage Guide](USAGE.md) - Complete usage examples
