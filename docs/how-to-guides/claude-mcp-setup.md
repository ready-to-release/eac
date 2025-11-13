# Claude Code MCP Server Setup

This project includes two MCP servers that give Claude access to CLI commands and GitHub operations.

## Quick Setup

To make MCP servers auto-start when you launch Claude Code in this repo, you need:

### 1. Set your GitHub token

```bash
export GITHUB_TOKEN="your-github-token"
```

Add this to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to make it permanent.

### 2. Add configuration to your local Claude config

Edit `~/.claude.json` (Linux/Mac) or `C:\Users\<username>\.claude.json` (Windows).

Add this section for your repo path:

```json
"C:\\source\\ready-to-release\\eac": {
  "mcpServers": {
    "commands": {
      "type": "stdio",
      "command": "go",
      "args": ["run", "./src/mcp/commands/main.go"],
      "env": {}
    },
    "github": {
      "type": "stdio",
      "command": "go",
      "args": ["run", "./src/mcp/github/main.go"],
      "env": {
        "GITHUB_TOKEN": "your-token-here"
      }
    }
  },
  "approvedProjectMcpServers": ["commands", "github"],
  "hasTrustDialogAccepted": true
}
```

**Important:**

- Replace `C:\\source\\ready-to-release\\eac` with your actual repo path
- On Windows, use double backslashes (`\\`)
- On Linux/Mac, use forward slashes (`/home/user/ready-to-release/eac`)
- Replace `your-token-here` with your actual GitHub token

### 3. Restart Claude Code

Exit completely and start a fresh session in your repo directory:

```bash
cd /path/to/ready-to-release/eac
claude
```

Run `/mcp` to verify both `commands` and `github` servers are loaded.

---

## Prerequisites

- Claude Code installed
- Go runtime
- GitHub CLI (`gh`) authenticated

---

## Resources

- [Claude Code MCP Docs](https://docs.claude.com/en/docs/claude-code/mcp)
- [MCP Specification](https://spec.modelcontextprotocol.io/)
