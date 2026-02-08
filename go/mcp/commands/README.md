# eac-mcp-server

MCP (Model Context Protocol) server that exposes EAC commands as tools for AI assistants. It reads JSON-RPC requests from stdin, dynamically discovers available commands via the EAC adapter, and returns structured text results.

## Key Types

- **`MCPRequest`** -- Inbound JSON-RPC 2.0 request envelope
- **`MCPResponse`** -- Outbound JSON-RPC 2.0 response envelope
- **`MCPError`** -- JSON-RPC error with code and message
- **`Tool`** -- MCP tool definition with name, description, input schema
- **`CallToolParams`** -- Parsed arguments for a `tools/call` invocation
- **`ToolResult`** -- Structured content array returned from tool execution
- **`CommandInfo`** -- Metadata for a single discovered EAC command
- **`CommandTree`** -- Full command hierarchy returned by `get commands`

## Patterns

- Stdio transport: Reads newline-delimited JSON-RPC from stdin, writes responses to stdout
- Dynamic discovery: Calls `get commands` through the EAC adapter to build the tool list
- Name mapping: Converts space-separated command names to hyphenated MCP tool names
- Stateless dispatch: Each `tools/call` creates a fresh adapter port and executes
- Self-filtering: The `get-commands` tool is excluded from the exposed tool list
- Protocol version: Implements MCP protocol version `2024-11-05`

## Internal Structure

| File | Responsibility |
| --- | --- |
| main.go | JSON-RPC loop, request dispatch, command discovery, tool execution |
| main_test.go | Protocol conformance and round-trip tests |

## Supported Methods

| Method | Behavior |
| --- | --- |
| `initialize` | Returns server info and capabilities |
| `tools/list` | Discovers commands and returns as MCP tools |
| `tools/call` | Executes a named command with optional args |

## Dependencies

- `go/adapters/eac` -- Adapter port for executing EAC commands
- `go/core/repository` -- Repository root discovery

## Role in System

The MCP server is the bridge between AI assistants (such as Claude Code) and the EAC command system. It translates MCP protocol requests into EAC adapter calls, allowing AI tools to query modules, dependencies, configurations, and other repository metadata without direct CLI invocation. The server is designed to run as a subprocess with stdio-based JSON-RPC communication.

## Code Health

### Tech Debt
- All production code in a single main.go (321 lines) with 25+ functions; type definitions, request handling, command discovery, and execution are intermixed
- Protocol version `2024-11-05` is hardcoded inline in `initializeResponse`; updating requires editing function internals

### Pain Points
- `getCommands()` shells out to the EAC adapter on every `tools/list` call with no caching; repeated tool-list requests re-discover commands each time
- Error logging goes to stderr via `fmt.Fprintf` without structured logging (no `core/logging` integration)

### Optimization Opportunities
- Cache command discovery results after first `tools/list` call since the command set does not change during a server session (high feasibility, simple in-memory cache)
- Split main.go into protocol handling, command discovery, and tool execution files for maintainability (moderate feasibility, small enough to be optional)
