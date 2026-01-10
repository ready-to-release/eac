# eac-mcp-commands

The `eac-mcp-commands` module provides Model Context Protocol (MCP) server integration for AI assistants like Claude. It exposes EAC commands as MCP tools that can be discovered and invoked by AI agents.

## System Context

Shows how eac-mcp-commands bridges AI assistants with the EAC system.

<!-- structurizr:eac-mcp-commands:SystemContext -->

## Container Architecture

High-level view of the MCP server components.

<!-- structurizr:eac-mcp-commands:Containers -->

## Component Architecture

Internal structure of the MCP command system.

<!-- structurizr:eac-mcp-commands:Components -->

## Workflow Diagrams

### Tool Discovery

How AI assistants discover available EAC commands.

<!-- structurizr:eac-mcp-commands:ToolDiscovery -->

### Tool Call

The flow when an AI assistant invokes an EAC command.

<!-- structurizr:eac-mcp-commands:ToolCall -->

## Design File

- **Location**: `specs/eac-mcp-commands/.design/workspace.dsl`
- **Interactive**: `r2r eac serve-design --module eac-mcp-commands`

## MCP Integration

The MCP server enables AI assistants to:

- Discover available EAC commands dynamically
- Invoke commands with proper argument handling
- Receive structured responses for further processing
- Access repository state and build information

## Configuration

MCP server is configured in Claude Code settings:

```json
{
  "mcpServers": {
    "commands": {
      "command": "r2r",
      "args": ["eac", "mcp", "commands"]
    }
  }
}
```
