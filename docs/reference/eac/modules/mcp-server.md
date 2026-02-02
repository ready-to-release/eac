# mcp-server

The `mcp-server` module provides Model Context Protocol (MCP) server integration for AI assistants like Claude.

It exposes EAC commands as MCP tools that can be discovered and invoked by AI agents.

## System Context

Shows how mcp-server bridges AI assistants with the EAC system.

<!-- structurizr:mcp-server:SystemContext -->

## Container Architecture

High-level view of the MCP server components.

<!-- structurizr:mcp-server:Containers -->

## Component Architecture

Internal structure of the MCP command system.

<!-- structurizr:mcp-server:Components -->

## Workflow Diagrams

### Tool Discovery

How AI assistants discover available EAC commands.

<!-- structurizr:mcp-server:ToolDiscovery -->

### Tool Call

The flow when an AI assistant invokes an EAC command.

<!-- structurizr:mcp-server:ToolCall -->

## Design File

- **Location**: `specs/mcp-server/.design/workspace.dsl`
- **Interactive**: `r2r eac serve-design --module mcp-server`

## MCP Integration

The MCP server enables AI assistants to:

- Discover available EAC commands dynamically
- Invoke commands with proper argument handling
- Receive structured responses for further processing
- Access repository state and build information

## Configuration

### Production Configuration (Default)

MCP server is configured in Claude Code settings (`.mcp.json`):

```json
{
  "mcpServers": {
    "commands": {
      "command": "mcp-server"
    }
  }
}
```

Uses `r2r eac <command>` by default (Docker-based execution).

### Development Configuration (Local Override)

For local development with faster startup (`.mcp.json`):

```json
{
  "mcpServers": {
    "commands": {
      "command": "go",
      "args": ["run", "./go/mcp/commands/main.go"],
      "env": {
        "EAC_USE_DIRECT_BINARY": "true"
      }
    }
  }
}
```

Uses direct binary execution (faster, ~100ms startup vs ~2s Docker).

### Environment Variable

**`EAC_USE_DIRECT_BINARY`**:

- `true`: Forces direct binary execution (development)
- Not set or `false`: Uses `r2r eac` (production, default)

### Verification

```bash
# Test production mode
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | r2r eac mcp commands

# Test development mode
EAC_USE_DIRECT_BINARY=true go run ./go/mcp/commands/main.go < input.json
```
