# PowerShell MCP Server

A Model Context Protocol server for executing PowerShell commands in native sessions.

## Features

- Execute PowerShell commands in isolated sessions
- Return structured JSON output
- Support for PowerShell modules and cmdlets
- Session state management

## Configuration

The server is configured in `.mcp.json` as:

```json
{
  "pwsh": {
    "transport": {
      "type": "stdio",
      "command": "pwsh",
      "args": ["-NoProfile", "-NonInteractive"]
    }
  }
}
```

## Tools Provided

- `execute-pwsh` - Execute a PowerShell command
- `get-pwsh-module` - List available PowerShell modules
- `invoke-pwsh-script` - Run a PowerShell script file

## Usage in Claude Code

```
Use the pwsh MCP server to run Get-Process
```

Or via slash command:
```
/mcp__pwsh__execute-pwsh Get-Process
```

## Implementation Status

✅ **GO IMPLEMENTATION** - Working MCP server written in Go.

Runs directly via `go run` - no build step required!

## Development

```bash
# Test locally
cd .claude/mcp-servers/pwsh
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | go run main.go

# Test a tool call
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | go run main.go
```
