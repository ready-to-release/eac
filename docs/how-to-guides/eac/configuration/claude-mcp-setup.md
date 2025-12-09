# Claude Code MCP Server Setup

{{ page_breadcrumb() }}

This project uses MCP (Model Context Protocol) servers to give Claude access to:

- **Commands Server**: All `r2r` CLI commands (modules, tests, builds, specs, etc.) - Local implementation
- **GitHub Server**: GitHub operations (repos, issues, PRs, workflows) - Official GitHub MCP server

## Setup

### 1. Configure Your Shell Profile

Add environment variables and source the setup script in your shell profile:

#### PowerShell (Windows)

Edit `$PROFILE` and add:

```powershell
# Set GitHub token
$env:GITHUB_TOKEN = "ghp_your_token_here"

# Source setup script (update path to your project location)
. "C:\path\to\project\scripts\pwsh\claude\env-setup.ps1"
```

Reload: `. $PROFILE`

#### Bash/Zsh (Linux/macOS)

Edit `~/.bashrc` or `~/.zshrc` and add:

```bash
# Set GitHub token
export GITHUB_TOKEN="ghp_your_token_here"

# Source setup script (update path to your project location)
source "$HOME/path/to/project/scripts/sh/claude/env-setup.sh"
```

Reload: `source ~/.bashrc` or `source ~/.zshrc`

### 2. Generate GitHub Token

1. Go to [https://github.com/settings/tokens](https://github.com/settings/tokens)
2. Create a new token (classic) with scopes: `repo`, `workflow`
3. Copy the token and set it in your shell profile (step 1)

### 3. Restart Claude Code

**IMPORTANT**: MCP servers load at startup, not during active sessions.

1. Close any open Claude Code sessions
2. Close and reopen your terminal
3. Start Claude Code in the project directory

### 4. Verify Setup

Check the initialization message:

```text
MCP Server Status:
Commands Server (mcp__commands__*): ✅ CONNECTED
GitHub Server (mcp__github__*): ✅ CONNECTED
```

## How It Works

When you start Claude Code:

1. Claude reads `.claude/settings.json` in the project root
2. Launches MCP servers as child processes:
   - Commands: `go run ./go/eac/mcp/commands/main.go` (local implementation)
   - GitHub: Official GitHub MCP server (configured via npx or direct installation)
3. Servers inherit environment variables from your shell
4. Claude can call MCP tools like `mcp__commands__show-modules` and `mcp__github__*`

The setup scripts auto-detect the project root based on their location (two directories up from `scripts/pwsh/` or `scripts/sh/`).

## Troubleshooting

### MCP Servers Not Connected

**Check Go version:**

```bash
go version  # Should be ≥ 1.21
```

**Test servers manually:**

```bash
# Test commands server (local)
go run ./go/eac/mcp/commands/main.go < /dev/null

# GitHub server uses official implementation - check .claude/settings.json for configuration
# Verify GitHub token is set:
echo $GITHUB_TOKEN  # Bash/Zsh
$env:GITHUB_TOKEN   # PowerShell
```

**Verify environment:**

```bash
# PowerShell
$env:GITHUB_TOKEN

# Bash/Zsh
echo $GITHUB_TOKEN
```

**Check logs:**

- Session logs: `.claude/logs/session-YYYY-MM-DD.jsonl`

### GitHub Authentication Fails

Test your token:

```bash
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user
```

If it fails, generate a new token with proper scopes.

### Changes Not Taking Effect

**MCP configuration changes require a full restart:**

1. Close all Claude Code sessions
2. Close all terminal windows
3. Open new terminal
4. Verify environment variables are set: `echo $env:GITHUB_TOKEN` (PowerShell) or `echo $GITHUB_TOKEN` (Bash)
5. Launch Claude Code fresh

**Note**: Changes to `.mcp.json` or `.claude/settings.json` only take effect at startup.

## Configuration Files

- `.claude/settings.json` - MCP server configuration (managed by Claude Code)
- `scripts/pwsh/claude/env-setup.ps1` - PowerShell setup script
- `scripts/sh/claude/env-setup.sh` - Bash/Zsh setup script
- `.claude-settings-example.json` - Example settings reference (this directory)

## Resources

- [Claude Code MCP Docs](https://docs.claude.com/en/docs/claude-code/mcp)
- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [GitHub Tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)

{{ diataxis_footer() }}
