# Project Architecture

## Overview

This monorepo provides a complete MCP server infrastructure for building developer tools and automation workflows.

## Project Structure

```text
.
├── src/mcp/                    # MCP servers (Go)
│   ├── pwsh/                   # PowerShell execution server
│   ├── docs/                   # Documentation management server
│   ├── github/                 # GitHub API integration server
│   └── vscode/                 # VSCode actions server
├── .vscode/extensions/         # VSCode extensions
│   └── claude-mcp-vscode/      # MCP integration extension
├── docs/                       # Documentation (MkDocs)
│   ├── guide/                  # User guides
│   └── reference/              # Technical reference
├── automation/                 # Build and maintenance scripts
│   └── sh-vscode/              # VSCode extension automation
├── containers/                 # Docker configurations
│   └── mkdocs/                 # Documentation container
├── .mcp.json                   # MCP server configuration
├── .markdownlint.yml          # Markdown linting rules
└── CLAUDE.md                   # Claude Code instructions
```

## Model Context Protocol

The MCP servers follow the official specification and communicate via JSON-RPC over stdin/stdout:

```text
Client Application
    ↓ (spawn process)
MCP Server (Go)
    ↓ (JSON-RPC stdin/stdout)
initialize → tools/list → tools/call
    ↓ (execute)
Tool Implementation
    ↓ (result)
JSON Response
```

## Technology Stack

- **Language:** Go 1.21+
- **Protocol:** Model Context Protocol (JSON-RPC 2.0)
- **Communication:** stdin/stdout
- **Documentation:** MkDocs Material
- **Containerization:** Docker (optional)

## MCP Servers

Four production-ready MCP servers implementing the Model Context Protocol:

| Server | Purpose | Tools Provided |
|--------|---------|----------------|
| **pwsh** | Execute PowerShell commands | `execute-pwsh`, `get-pwsh-modules` |
| **docs** | Documentation management | `search-docs`, `get-doc-page`, `build-docs`, `serve-docs` |
| **github** | GitHub API integration | `gh-repo-view`, `gh-issue-create`, `gh-pr-list`, `gh-run-list` |
| **vscode** | VSCode actions | `vscode-action` (commit, push, pull, custom) |

**Configuration:** `.mcp.json`

**See:** [MCP Servers Guide](mcp-servers.md) for detailed documentation.

## Development

### Prerequisites

- **Go** 1.21 or higher
- **Git** for version control
- **Docker** (optional, for documentation)

### Building MCP Servers

```bash
# Navigate to server directory
cd src/mcp/<server-name>

# Run in development
go run .

# Build binary
go build -o server

# Run tests
go test ./...
```

### Testing MCP Servers

Test servers using JSON-RPC commands:

```bash
cd src/mcp/pwsh
go run .

# In another terminal, send JSON-RPC:
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | nc localhost 3000
```

**See:** [MCP Servers Guide](mcp-servers.md) for detailed testing procedures.

## Configuration

### MCP Server Configuration

The `.mcp.json` file configures all available MCP servers:

```json
{
  "mcpServers": {
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
    "vscode": {
      "command": "bash",
      "args": ["src/mcp/vscode/run.sh"]
    }
  }
}
```

### Markdown Linting

Configure markdown linting rules in `.markdownlint.yml`. See [Markdownlint Summary](markdownlint-summary.md) for all available rules.

## Example Integrations

The project includes example integrations demonstrating MCP usage:

- **VSCode Extension** - Located in `.vscode/extensions/claude-mcp-vscode/`
  - See [VSCode Extension Guide](vscode-extension/index.md)
  - See [Quick Start](vscode-extension/QUICKSTART.md)

### Usage Patterns

```bash
# Start an MCP server
cd src/mcp/pwsh
./run.sh

# The server listens on stdin/stdout
# Send JSON-RPC commands to interact
```

## Automation

Build and maintenance scripts in `automation/`:

```bash
# VSCode extension automation
./automation/sh/vscode/init.sh      # Setup
./automation/sh/vscode/restore.sh   # Restore dependencies
./automation/sh/vscode/clean.sh     # Clean artifacts
```

**See:** [Automation Scripts README](../../automation/sh/vscode/README.md)

## Contributing

When contributing:

1. **Documentation:** Update relevant docs in `docs/`
2. **Code Style:** Follow existing Go patterns
3. **Testing:** Add tests for new MCP tools
4. **Architecture:** Update `CLAUDE.md` for architectural changes
5. **Commits:** Follow semantic commit conventions

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| MCP server won't start | Verify Go installation: `go version` |
| JSON-RPC errors | Validate JSON format and method names |
| Build failures | Run `go mod tidy` in server directory |
| Documentation not loading | Check Docker or MkDocs installation |

**Detailed troubleshooting:** See individual server documentation.

## Resources

- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- [Go Documentation](https://golang.org/doc/)
- [MkDocs Material](https://squidfunk.github.io/mkdocs-material/)
