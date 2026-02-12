# eac-mcp-server

MCP (Model Context Protocol) server that exposes EAC commands as tools for AI assistants. It reads JSON-RPC requests from stdin, dynamically discovers available commands via the EAC adapter, and returns structured text results.

## Key Types

- **`MCPRequest`** -- Inbound JSON-RPC 2.0 request envelope
- **`MCPResponse`** -- Outbound JSON-RPC 2.0 response envelope
- **`MCPError`** -- JSON-RPC error with code and message
- **`Tool`** -- MCP tool definition with name, description, input schema
- **`InputSchema`** -- JSON Schema describing tool input parameters
- **`Property`** -- Single property definition within an input schema
- **`CallToolParams`** -- Parsed arguments for a `tools/call` invocation
- **`ToolResult`** -- Structured content array returned from tool execution
- **`Content`** -- Single content entry (type + text) within a tool result
- **`CommandInfo`** -- Metadata for a single discovered EAC command
- **`CommandTree`** -- Full command hierarchy returned by `get commands`

## Key Functions

- `processRequests` -- Main loop reading JSON-RPC from scanner and dispatching
- `handleRequest` -- Routes requests by method to initialize, tools/list, or tools/call handlers
- `getCommandTools` -- Discovers commands and converts to MCP tool definitions
- `commandToTool` -- Converts a single CommandInfo to an MCP Tool (filters out get-commands)
- `callTool` -- Executes a named tool by translating back to EAC command invocation
- `execCommand` -- Runs a command via the EAC adapter and formats output
- `findRepoRoot` -- Locates repository root via `repository.GetRepositoryRoot`

## Patterns

- Stdio transport: Reads newline-delimited JSON-RPC from stdin, writes responses to stdout
- Dynamic discovery: Calls `get commands` through the EAC adapter to build the tool list
- Name mapping: Converts space-separated command names to hyphenated MCP tool names
- Stateless dispatch: Each `tools/call` creates a fresh adapter port and executes
- Self-filtering: The `get-commands` tool is excluded from the exposed tool list
- Protocol version: Implements MCP protocol version `2024-11-05`

## Internal Structure

| File    | Responsibility                                                                       |
| ------- | ------------------------------------------------------------------------------------ |
| main.go | Type definitions, JSON-RPC loop, request dispatch, command discovery, tool execution |

## Dependencies

- `go/adapters/eac` -- Adapter port for executing EAC commands
- `go/core/repository` -- Repository root discovery

## Role in System

The MCP server is the bridge between AI assistants (such as Claude Code) and the EAC command system. It translates MCP protocol requests into EAC adapter calls, allowing AI tools to query modules, dependencies, configurations, and other repository metadata without direct CLI invocation. The server is designed to run as a subprocess with stdio-based JSON-RPC communication.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- None identified
