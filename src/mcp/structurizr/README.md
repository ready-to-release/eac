# Structurizr MCP Server

A Model Context Protocol server for creating C4 architecture diagrams using Structurizr DSL.

## Features

- Create architecture workspaces for modules
- Add containers to software systems
- Define relationships between elements
- Export workspace DSL files

All workspaces are saved to `docs/reference/design/<module>/workspace.dsl` for version control and viewing with Structurizr Lite.

## Configuration

The server is configured in `.mcp.json` as:

```json
{
  "structurizr": {
    "command": "go",
    "args": ["run", "./src/mcp/structurizr/main.go"],
    "env": {
      "STRUCTURIZR_WORKSPACE_ROOT": "docs/reference/design"
    }
  }
}
```

## Environment Variables

- `STRUCTURIZR_WORKSPACE_ROOT` - Root directory for workspace files (default: `docs/reference/design`)

## Tools Provided

### create_workspace

Create a new architecture workspace for a module.

**Parameters**:

- `module` (string, required): Module name (e.g., "cli", "vscode", "docs")
- `name` (string, required): Workspace name (e.g., "CLI Architecture")
- `description` (string, required): Workspace description

**Example**:

```json
{
  "module": "cli",
  "name": "R2R CLI Architecture",
  "description": "Ready to Release CLI architecture"
}
```

### add_container

Add a container to the architecture.

**Parameters**:

- `module` (string, required): Module name
- `name` (string, required): Container name
- `technology` (string, required): Technology/platform (e.g., "Go", "React")
- `description` (string, required): Container's purpose and responsibilities

### add_relationship

Define a relationship between elements.

**Parameters**:

- `module` (string, required): Module name
- `source` (string, required): Source element ID (snake_case)
- `destination` (string, required): Destination element ID (snake_case)
- `description` (string, required): Relationship description
- `technology` (string, optional): Technology/protocol used

### export_workspace

Export the workspace DSL content.

**Parameters**:

- `module` (string, required): Module name

## Usage in Claude Code

```text
Create architecture documentation for the CLI module
```

Claude will automatically:

1. Create the workspace at `docs/reference/design/cli/workspace.dsl`
2. Add containers and relationships based on the code
3. Generate accompanying documentation

## Viewing Architecture Diagrams

After creating architecture, view it with Structurizr Lite:

```bash
cd docs/reference/design/<module>
docker run -d --name structurizr-<module> -p 8081:8080 structurizr/lite
sleep 5
docker cp workspace.dsl structurizr-<module>:/usr/local/structurizr/workspace.dsl
docker restart structurizr-<module>
```

Then open <http://localhost:8081>

## Output Location

All workspace DSL files are saved to:

```text
docs/reference/design/<module>/workspace.dsl
```

This location is:

- ✅ Tracked in git
- ✅ Directly loadable by Structurizr Lite
- ✅ Follows project conventions

## Implementation Status

✅ **GO IMPLEMENTATION** - Working MCP server written in Go using Structurizr DSL.

Runs directly via `go run` - no build step required!

## Prerequisites

- Go 1.21 or higher
- Structurizr Lite (optional, for viewing diagrams - runs in Docker)

## Development

```bash
# Test locally (run from project root to resolve paths correctly)
export STRUCTURIZR_WORKSPACE_ROOT=docs/reference/design
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | go run ./src/mcp/structurizr/main.go

# Test tools list
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | go run ./src/mcp/structurizr/main.go

# Test create workspace
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"create_workspace","arguments":{"module":"test","name":"Test Workspace","description":"Test description"}}}' | go run ./src/mcp/structurizr/main.go
```

## Documentation

- [Structurizr DSL Documentation](https://github.com/structurizr/dsl)
- [C4 Model](https://c4model.com/)
- [MCP Protocol Specification](https://modelcontextprotocol.io/)
