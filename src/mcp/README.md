# MCP Servers

This directory contains Model Context Protocol (MCP) server implementations written in Go.

## Configured Servers

### 1. PowerShell (`pwsh`)
Execute PowerShell commands in native sessions.

**Tools:**
- `execute-pwsh` - Execute a PowerShell command
- `get-pwsh-modules` - List available PowerShell modules

**Location:** `.claude/mcp-servers/pwsh/`

### 2. MkDocs (`docs`)
Access and search repository documentation.

**Tools:**
- `search-docs` - Search documentation for a query
- `get-doc-page` - Get content of a specific documentation page
- `list-docs` - List all available documentation pages

**Location:** `.claude/mcp-servers/docs/`

### 3. GitHub (`github`)
GitHub operations via GitHub CLI.

**Tools:**
- `gh-repo-view` - View repository details
- `gh-issue-create` - Create a new issue
- `gh-pr-list` - List pull requests
- `gh-run-list` - List workflow runs

**Location:** `.claude/mcp-servers/github/`

## Setup

### Prerequisites

- **Go 1.21+** - `go version` to check
  - Windows: Download from https://go.dev/dl/
  - Linux/WSL: `sudo apt install golang-go`
- **PowerShell** - For pwsh server (Windows has it by default)
- **GitHub CLI** - For github server: `winget install --id GitHub.cli` or `sudo apt install gh`

### Cross-Platform Setup

The MCP servers work on both Windows and Linux using bash:
- **Linux/WSL**: Native bash
- **Windows**: Uses WSL bash (requires WSL to be installed)

Scripts are executable via `bash run.sh` wrapper.

### Configuration

**Important:** The `.mcp.json` file is at project root and works on all platforms.

1. **Configure servers** - Already done in `.mcp.json`

2. **Set environment variables** in `.claude/settings.local.json`:

```json
{
  "env": {
    "GITHUB_TOKEN": "ghp_your_token_here",
    "DOCS_PATH": "./docs"
  }
}
```

3. **Authenticate GitHub CLI** (if using github server):
```bash
gh auth login
```

### Usage in Claude Code

The servers run automatically via `go run` - no build step needed!

**Check server status:**
```
/mcp
```

**Use server tools:**
```
Use the pwsh server to run Get-Process
Search the docs for authentication
Create a GitHub issue about adding tests
```

**Or use slash commands:**
```
/mcp__pwsh__execute-pwsh Get-Process
/mcp__docs__search-docs authentication
/mcp__github__gh-repo-view owner/repo
```

### Testing Servers

Test each server directly:

```bash
# PowerShell server
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | go run .claude/mcp-servers/pwsh

# Docs server
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | go run .claude/mcp-servers/docs

# GitHub server
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | go run .claude/mcp-servers/github
```

## Adding New Servers

1. Create directory: `.claude/mcp-servers/server-name/`
2. Create `main.go` and `go.mod` files
3. Add configuration to `.mcp.json` using `go run`
4. Document in server's README.md

## MCP Resources

- [MCP Documentation](https://modelcontextprotocol.io/)
- [MCP Protocol Specification](https://spec.modelcontextprotocol.io/)
- [MCP Go Examples](https://github.com/mark3labs/mcp-go)
- [Official MCP Servers](https://github.com/modelcontextprotocol/servers)

## Transport Types

- **stdio** - Local process (used by all our servers)
- **http** - Remote HTTP server (recommended for production)
- **sse** - Server-Sent Events (deprecated)

## Development Tips

1. Test servers standalone before integrating
2. Use environment variables for secrets
3. Implement proper error handling
4. Add comprehensive logging
5. Document all tools and resources
