# MCP Commands Server

Model Context Protocol (MCP) server that exposes all EAC CLI commands as MCP tools.

## Overview

This MCP server provides programmatic access to all commands defined in `src/commands` via the MCP protocol. Commands are automatically discovered using the `describe commands` introspection capability.

## Features

- **Auto-discovery**: Automatically discovers all available commands from `src/commands`
- **Dynamic tool registration**: Each command becomes an MCP tool
- **Command execution**: Executes commands via `go run ./src/commands <command>`
- **Output capture**: Returns command stdout/stderr as tool results

## Architecture

```
┌─────────────────┐
│  Claude Desktop │
└────────┬────────┘
         │ JSON-RPC over stdio
         ▼
┌─────────────────────────────┐
│  MCP Commands Server        │
│  (src/mcp/commands/main.go) │
└────────┬────────────────────┘
         │ go run ./src/commands <cmd>
         ▼
┌─────────────────────────────┐
│  Commands Module            │
│  (src/commands/*.go)        │
└─────────────────────────────┘
```

## Usage

### Running the server

```bash
cd src/mcp/commands
go run .
```

The server communicates via stdin/stdout using JSON-RPC 2.0 format.

### MCP Protocol

**Initialize:**
```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
```

**List tools:**
```json
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
```

**Call tool:**
```json
{
  "jsonrpc":"2.0",
  "id":3,
  "method":"tools/call",
  "params":{
    "name":"show-modules",
    "arguments":{}
  }
}
```

### Available Tools

Tools are auto-discovered from `src/commands`. Each command is converted to kebab-case:

| Command Name | Tool Name | Description |
|--------------|-----------|-------------|
| `show modules` | `show-modules` | Show all module contracts |
| `show files` | `show-files` | Show repository files with module ownership |
| `show files changed` | `show-files-changed` | Show changed files |
| `show files staged` | `show-files-staged` | Show staged files |
| `test module` | `test-module` | Run tests for a module |
| `docs serve` | `docs-serve` | Start MkDocs server |
| `design serve` | `design-serve` | Start Structurizr server |
| ... | ... | ... |

To see all available tools, use the `tools/list` method.

## Configuration for Claude Desktop

Add to your Claude Desktop MCP configuration (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "eac-commands": {
      "command": "go",
      "args": ["run", ".", ""],
      "cwd": "C:/projects/eac/src/mcp/commands",
      "env": {}
    }
  }
}
```

## Tool Arguments

All tools accept an optional `args` parameter for additional command arguments:

```json
{
  "name": "test-module",
  "arguments": {
    "args": "src-commands"
  }
}
```

This executes: `go run ./src/commands test module src-commands`

## Implementation Details

### Command Discovery

The server discovers commands by calling:
```bash
go run ./src/commands describe commands
```

This returns JSON with all registered commands and their metadata.

### Command Execution

Commands are executed via:
```bash
go run ./src/commands <command-name> [args]
```

The server captures both stdout and stderr and returns them as the tool result.

### Repository Root Detection

The server automatically finds the repository root by walking up the directory tree looking for `src/commands`.

## Related Files

- **Command Implementations**: `src/commands/*.go`
- **Command Registry**: `src/commands/main.go`
- **Command Introspection**: `src/commands/describe-commands.go`

## See Also

- [MCP GitHub Server](../github/README.md) - GitHub CLI integration
- [MCP Docs Server](../docs/README.md) - Documentation access
- [MCP PowerShell Server](../pwsh/README.md) - PowerShell execution
- [Commands Module](../../commands/README.md) - CLI commands implementation
