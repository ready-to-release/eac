# VS Code Extension Reference

Quick reference for commands, actions, servers, file locations, and packaging.

---

## Available Actions

Actions accessible via the robot button (🤖) in Source Control toolbar:

| Action | Description | Auto-generates Message | Requirements |
|--------|-------------|------------------------|--------------|
| **Git Commit** | Commits staged changes with AI-generated message | Yes - analyzes staged files | Staged changes |
| **Git Push** | Push commits to remote repository | N/A | Commits to push |
| **Git Pull** | Pull changes from remote repository | N/A | Remote configured |
| **Custom Action** | Run custom command with manual input | No - prompts for input | None |

---

## Available MCP Servers

Servers configured in `.mcp.json`:

| Server | Purpose | Location | Tools Provided |
|--------|---------|----------|----------------|
| **vscode** | VSCode actions (Git, etc.) | `src/mcp/vscode/` | `vscode-action` |
| **pwsh** | PowerShell commands | `src/mcp/pwsh/` | `pwsh-command` |
| **docs** | Documentation management | `src/mcp/docs/` | `docs-tool` |
| **github** | GitHub API integration | `src/mcp/github/` | `github-api` |

---

## Example Workflows

### Quick Commit with Auto-generated Message

```text
1. Make changes to files
2. Open Source Control (Ctrl+Shift+G)
3. Stage your changes (click + icon next to files)
4. Click robot icon (🤖)
5. Select "Git Commit"
6. Extension analyzes staged changes
7. AI generates semantic commit message
8. Commit is created automatically
9. Notification shows the generated message
```

### Push Changes to Remote

```text
1. Ensure you have commits ready to push
2. Click robot icon (🤖)
3. Select "Git Push"
4. Wait for operation to complete
5. View result in notification
```

### Pull Changes from Remote

```text
1. Click robot icon (🤖)
2. Select "Git Pull"
3. Wait for operation to complete
4. View result in notification
```

### Run Custom Action with Manual Input

```text
1. Click robot icon (🤖)
2. Select "Custom Action"
3. Enter command or message when prompted
4. View result in notification
```

---

## Viewing Results

### Notifications

- **Location**: Bottom-right corner of VSCode
- **Types**:
  - Information (blue) - Success messages
  - Warning (yellow) - Non-critical issues
  - Error (red) - Failures

### Debug Console

- **Location**: Original VSCode window → View → Debug Console (or `Ctrl+Shift+Y`)
- **Content**:
  - Extension log messages (`console.log()`)
  - MCP server stderr output
  - Error stack traces

### Output Panel

- **Location**: View → Output (or `Ctrl+Shift+U`)
- **Select**: Choose extension from dropdown
- **Content**: Structured logging output

---

## Common Commands

### Setup Commands

```bash
# First-time setup (installs dependencies, builds extension and servers)
./automation/sh/vscode/init.sh

# Restore dependencies (if node_modules or Go modules are corrupted)
./automation/sh/vscode/restore.sh

# Clean build artifacts (removes compiled files)
./automation/sh/vscode/clean.sh
```

### Development Commands

```bash
# Compile TypeScript once
cd .vscode/extensions/claude-mcp-vscode
npm run compile

# Auto-compile on save (recommended for development)
npm run watch

# Run tests
npm test

# Lint code
npm run lint
```

### Debugging Commands

```bash
# Open Source Control view
Ctrl+Shift+G (Windows/Linux)
Cmd+Shift+G (Mac)

# Launch Extension Development Host
F5 (in VSCode)

# Reload extension in Development Host
Ctrl+R (Windows/Linux)
Cmd+R (Mac)

# Open Debug Console
Ctrl+Shift+Y (Windows/Linux)
Cmd+Shift+Y (Mac)

# Open Command Palette
Ctrl+Shift+P (Windows/Linux)
Cmd+Shift+P (Mac)
```

### Git Commands (for manual testing)

```bash
# Check status
git status

# Stage files
git add <file>
git add .

# View staged changes
git diff --staged

# View commit history
git log --oneline -10
```

---

## File Locations

### Extension Files

```text
.vscode/extensions/claude-mcp-vscode/
├── src/
│   └── extension.ts          # Main extension code
├── package.json               # Extension manifest
│   ├── contributes           # Commands, menus, buttons
│   ├── activationEvents      # When extension loads
│   └── dependencies          # npm packages
├── tsconfig.json             # TypeScript configuration
├── out/                       # Compiled JavaScript (generated)
└── node_modules/             # Dependencies (generated)
```

### MCP Server Files

```text
src/mcp/
├── vscode/                    # VSCode action server
│   ├── main.go               # JSON-RPC handler
│   └── run.sh                # Server launcher
├── pwsh/                      # PowerShell commands
│   ├── main.go
│   └── run.sh
├── docs/                      # Documentation management
│   ├── main.go
│   └── run.sh
└── github/                    # GitHub API integration
    ├── main.go
    └── run.sh
```

### Configuration Files

```text
.mcp.json                      # MCP server configuration (project root)
.vscode/
├── launch.json               # Debug configurations
└── tasks.json                # Build tasks
```

---

## Packaging & Distribution

### Create VSIX Package

```bash
cd .vscode/extensions/claude-mcp-vscode

# Install packaging tool (once)
npm install -g @vscode/vsce

# Create package
vsce package

# Output: claude-mcp-vscode-0.1.0.vsix
```

### Install Locally

**From VSIX file:**

```bash
code --install-extension claude-mcp-vscode-0.1.0.vsix
```

**From directory:**

```bash
cd .vscode/extensions/claude-mcp-vscode
code --install-extension .
```

**Manual installation:**

1. Open VSCode
2. Extensions view (`Ctrl+Shift+X`)
3. Click `...` (more actions)
4. Select "Install from VSIX..."
5. Choose the .vsix file

### Publish to Marketplace

**Prerequisites:**

1. Create account at [Visual Studio Marketplace](https://marketplace.visualstudio.com/)
2. Create Personal Access Token (PAT) in Azure DevOps

**Publish:**

```bash
# Login (once)
vsce login <publisher-name>

# Publish extension
vsce publish

# Publish specific version
vsce publish 1.0.1

# Publish patch/minor/major
vsce publish patch
vsce publish minor
vsce publish major
```

### Update Published Extension

```bash
# Increment version in package.json
npm version patch  # 1.0.0 → 1.0.1
npm version minor  # 1.0.0 → 1.1.0
npm version major  # 1.0.0 → 2.0.0

# Publish update
vsce publish
```

---

## Keyboard Shortcuts

Default shortcuts (if configured):

| Action | Windows/Linux | Mac | Description |
|--------|---------------|-----|-------------|
| **Open Source Control** | `Ctrl+Shift+G` | `Cmd+Shift+G` | View staged/unstaged changes |
| **Command Palette** | `Ctrl+Shift+P` | `Cmd+Shift+P` | Access all commands |
| **Launch Extension** | `F5` | `F5` | Start Extension Development Host |
| **Reload Extension** | `Ctrl+R` | `Cmd+R` | Reload in Development Host |

Custom shortcuts can be added in `package.json` under `keybindings`. See [Add a Command](../how-to-guides/vscode-extension/add-command.md).

---

## Extension Configuration

### package.json Key Sections

**Metadata:**

```json
{
  "name": "claude-mcp-vscode",
  "displayName": "Claude MCP VSCode",
  "description": "VSCode extension with MCP integration",
  "version": "0.1.0",
  "publisher": "your-publisher-name",
  "engines": {
    "vscode": "^1.80.0"
  }
}
```

**Activation:**

```json
{
  "activationEvents": [
    "onStartupFinished"
  ]
}
```

**Commands:**

```json
{
  "contributes": {
    "commands": [
      {
        "command": "claude-mcp-vscode.action",
        "title": "MCP Action",
        "icon": "$(robot)"
      }
    ]
  }
}
```

**Menus:**

```json
{
  "contributes": {
    "menus": {
      "scm/title": [
        {
          "command": "claude-mcp-vscode.action",
          "group": "navigation",
          "when": "scmProvider == git"
        }
      ]
    }
  }
}
```

---

## MCP Configuration

### .mcp.json Format

```json
{
  "mcpServers": {
    "<server-name>": {
      "command": "<executable>",
      "args": ["<arg1>", "<arg2>"],
      "env": {
        "ENV_VAR": "value"
      }
    }
  }
}
```

**Example:**

```json
{
  "mcpServers": {
    "vscode": {
      "command": "bash",
      "args": ["src/mcp/vscode/run.sh"]
    },
    "custom": {
      "command": "python",
      "args": ["src/mcp/custom/server.py"],
      "env": {
        "DEBUG": "true"
      }
    }
  }
}
```

---

## JSON-RPC Message Examples

### Initialize Request

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {}
}
```

### Initialize Response

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "serverInfo": {
      "name": "vscode-server",
      "version": "1.0.0"
    }
  }
}
```

### Tools List Request

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list",
  "params": {}
}
```

### Tools Call Request

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "vscode-action",
    "arguments": {
      "action": "git-commit",
      "message": "feat: add new feature"
    }
  }
}
```

### Tools Call Response

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Successfully committed"
      }
    ]
  }
}
```

### Error Response

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": "Git command failed"
  }
}
```

---

## External Resources

### VSCode Extension Development

- [VSCode Extension API](https://code.visualstudio.com/api)
- [Extension Guidelines](https://code.visualstudio.com/api/references/extension-guidelines)
- [Publishing Extensions](https://code.visualstudio.com/api/working-with-extensions/publishing-extension)
- [Icons Reference](https://code.visualstudio.com/api/references/icons-in-labels)

### MCP Protocol

- [Model Context Protocol Spec](https://modelcontextprotocol.io/)
- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)

### Tools & Libraries

- [vsce](https://github.com/microsoft/vscode-vsce) - VSCode Extension Packager
- [TypeScript](https://www.typescriptlang.org/)
- [Go](https://golang.org/)

---

## Related Documentation

- **Get Started**: [Quick Start Tutorial](../tutorials/vscode-extension-quickstart.md)
- **Architecture**: [VS Code Extension Architecture](../explanation/vscode-extension-architecture.md)
- **How-to Guides**: [VS Code Extension How-to Guides](../how-to-guides/vscode-extension/)
- **Troubleshooting**: [Troubleshoot Common Issues](../how-to-guides/vscode-extension/troubleshoot.md)
