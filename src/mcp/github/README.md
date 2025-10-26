# GitHub CLI MCP Server

A Model Context Protocol server for GitHub operations using the GitHub CLI.

## Features

- Repository management
- Issue and PR operations
- GitHub Actions workflow execution
- Repository search
- Code search
- Organization management

## Configuration

The server is configured in `.mcp.json` as:

```json
{
  "github": {
    "transport": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"]
    },
    "env": {
      "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_PERSONAL_ACCESS_TOKEN}"
    }
  }
}
```

## Environment Variables

- `GITHUB_TOKEN` - GitHub personal access token (required)
  - Create at: <https://github.com/settings/tokens>
  - Required scopes: `repo`, `workflow`, `read:org`

## Tools Provided

This uses the official MCP GitHub server which provides:

- `create-repository` - Create a new repository
- `get-repository` - Get repository details
- `create-issue` - Create an issue
- `create-pull-request` - Create a pull request
- `push-files` - Push files to repository
- `search-repositories` - Search GitHub repositories
- `search-code` - Search code across repositories
- `list-commits` - List commits in a repository

## Usage in Claude Code

```text
Create a new issue in this repository about adding tests
```

Or via slash command:

```text
/mcp__github__create-issue
```

## Setup Instructions

1. **Install GitHub CLI** (if not using npx):

   ```bash
   # Windows
   winget install --id GitHub.cli

   # macOS
   brew install gh

   # Linux
   sudo apt install gh
   ```

2. **Authenticate**:

   ```bash
   gh auth login
   ```

3. **Create Personal Access Token**:

   - Visit <https://github.com/settings/tokens>
   - Click "Generate new token (classic)"
   - Select scopes: `repo`, `workflow`, `read:org`
   - Copy token and set environment variable

4. **Configure in `.claude/settings.local.json`**:

   ```json
   {
     "env": {
       "GITHUB_TOKEN": "ghp_your_token_here"
     }
   }
   ```

## Implementation Status

✅ **GO IMPLEMENTATION** - Working MCP server written in Go using GitHub CLI.

Runs directly via `go run` - no build step required!

## Prerequisites

GitHub CLI must be installed and authenticated:

```bash
# Install gh CLI
winget install --id GitHub.cli

# Authenticate
gh auth login
```

## Development

```bash
# Test locally
cd .claude/mcp-servers/github
GITHUB_TOKEN=your_token go run main.go

# Test with initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | go run main.go

# Test tools list
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | go run main.go
```

## Documentation

- [GitHub CLI Documentation](https://cli.github.com/manual/)
- [MCP Protocol Specification](https://modelcontextprotocol.io/)
