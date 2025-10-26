# VSCode Extension Setup

Complete setup guide for the VSCode extension and MCP servers.

## Prerequisites

- **Go** 1.21 or higher (for MCP servers)
- **Node.js** 18.x or higher (for VSCode extension)
- **npm** (comes with Node.js)
- **VSCode** 1.80.0 or higher (for extension development)
- **Git** (for version control)

## Setup Options

### Option 1: Automated Setup (Recommended)

Use the automation scripts for easy setup:

```bash
# First time setup - installs dependencies and verifies everything
./automation/sh/vscode/init.sh

# Restore dependencies (if you have issues)
./automation/sh/vscode/restore.sh

# Clean all build artifacts
./automation/sh/vscode/clean.sh
```

See [automation/sh/vscode/README.md](../../../automation/sh/vscode/README.md) for more details.

### Option 2: Manual Setup

If you prefer to set up manually:

#### 1. Install VSCode Extension

Navigate to the extension directory and install dependencies:

```bash
cd .vscode/extensions/claude-mcp-vscode
npm install
npm run compile
```

#### 2. Verify MCP Servers

The MCP servers are Go-based and run on-demand. Verify they work:

```bash
# Test the VSCode MCP server
cd src/mcp/vscode
go run .
# Press Ctrl+C to exit

# Test the PowerShell MCP server
cd src/mcp/pwsh
go run .
# Press Ctrl+C to exit
```

## Running the Extension

### Option A: Debug Mode (Recommended)

1. Open this workspace in VSCode
2. Press `F5` to launch the Extension Development Host
3. In the new window, open any Git repository
4. Open the Source Control view (`Ctrl+Shift+G` or `Cmd+Shift+G`)
5. Click the robot icon button in the toolbar

### Option B: Install Locally

```bash
cd .vscode/extensions/claude-mcp-vscode
npm run compile
code --install-extension .
```

## Development Workflow

### Watch Mode (Auto-Recompile)

For faster development, use watch mode:

```bash
cd .vscode/extensions/claude-mcp-vscode
npm run watch
```

Now TypeScript auto-recompiles on every save. Just reload the Extension Development Host (Ctrl+R) after changes.

### Build

```bash
npm run compile
```

### Package Extension

```bash
npm install -g @vscode/vsce
vsce package
```

## Troubleshooting

### Extension Not Showing in Toolbar

1. Verify you're in a Git repository
2. Check that `scmProvider == git` in your workspace
3. Restart the Extension Development Host

### MCP Server Errors

1. Verify Go is installed: `go version`
2. Check server logs in VSCode Output panel
3. Test server manually: `cd src/mcp/<name> && go run .`

### Build Errors

```bash
# Clean and rebuild extension
cd .vscode/extensions/claude-mcp-vscode
rm -rf node_modules out
npm install
npm run compile
```

## Configuration

### MCP Server Configuration

Edit `.mcp.json` to add or modify MCP servers:

```json
{
  "mcpServers": {
    "server-name": {
      "command": "bash",
      "args": ["src/mcp/server-name/run.sh"],
      "env": {
        "ENV_VAR": "value"
      }
    }
  }
}
```

### Extension Configuration

The extension reads configuration from:

- `package.json` - Extension manifest
- `.vscode/launch.json` - Debug configuration
- `.vscode/tasks.json` - Build tasks

## Next Steps

- [Quick Start Guide](QUICKSTART.md) - Get running in 5 minutes
- [Usage Guide](USAGE.md) - Detailed usage examples
- [Extension Development](index.md) - Customization and advanced topics
