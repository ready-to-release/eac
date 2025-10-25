# CLI Project

An extensible CLI-in-a-box for managing software delivery flows with integrated MCP servers and VSCode extension.

## 📚 Documentation

### Quick Links

- **[Quick Start](QUICKSTART.md)** - Get up and running in 5 minutes
- **[Usage Guide](USAGE.md)** - Complete usage examples
- **[Full Documentation](docs/)** - Comprehensive MkDocs site

### Documentation Site

The full documentation is in the `docs/` directory and uses MkDocs:

#### Option 1: Using Docker (Recommended)

```bash
# Start documentation server
cd containers/mkdocs && ./serve.sh

# Open: http://localhost:8000
```

#### Option 2: Local Installation

```bash
# Install MkDocs
pip install mkdocs-material

# Serve locally
mkdocs serve

# Open: http://127.0.0.1:8000
```

**Key Documentation:**

- [VSCode Extension Guide](docs/guide/vscode-extension/index.md) - Using and developing the extension
- [Recording Demos](docs/guide/vscode-extension/recording-demos.md) - Create demo videos
- [Automation Scripts](automation/sh-vscode/README.md) - Build and maintenance scripts
- [Project Structure](docs/README-DOCS.md) - Documentation organization

## Project Overview

This project provides:

- **MCP Servers**: Go-based Model Context Protocol servers for various integrations
- **VSCode Extension**: Local extension with Git toolbar integration
- **Extensible Architecture**: Plugin-based system for custom workflows

## Project Structure

```text
.
├── src/mcp/                    # MCP servers (Go)
│   ├── pwsh/                   # PowerShell execution server
│   ├── docs/                   # Documentation server
│   ├── github/                 # GitHub integration server
│   └── vscode/                 # VSCode integration server
├── .vscode/extensions/         # VSCode extensions
│   └── claude-mcp-vscode/      # MCP integration extension
├── .mcp.json                   # MCP server configuration
└── CLAUDE.md                   # Claude Code instructions

```

## Prerequisites

- **Go** 1.21 or higher (for MCP servers)
- **Node.js** 18.x or higher (for VSCode extension)
- **npm** (comes with Node.js)
- **VSCode** 1.80.0 or higher (for extension development)
- **Git** (for version control)

## Quick Start

### Automated Setup (Recommended)

Use the automation scripts for easy setup:

```bash
# First time setup - installs dependencies and verifies everything
./automation/sh-vscode/init.sh

# Restore dependencies (if you have issues)
./automation/sh-vscode/restore.sh

# Clean all build artifacts
./automation/sh-vscode/clean.sh
```

See [automation/sh-vscode/README.md](automation/sh-vscode/README.md) for more details.

### Manual Setup

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

### Running the Extension

#### Option A: Debug Mode (Recommended)

1. Open this workspace in VSCode
2. Press `F5` to launch the Extension Development Host
3. In the new window, open any Git repository
4. Open the Source Control view (`Ctrl+Shift+G` or `Cmd+Shift+G`)
5. Click the robot icon button in the toolbar

#### Option B: Install Locally

```bash
cd .vscode/extensions/claude-mcp-vscode
npm run compile
code --install-extension .
```

---

## MCP Servers

### Available Servers

All servers are configured in `.mcp.json`:

1. **pwsh** - Execute PowerShell commands
2. **docs** - Documentation management
3. **github** - GitHub API integration
4. **vscode** - VSCode action integration

### Running MCP Servers

MCP servers are executed via their `run.sh` scripts:

```bash
# Example: Run the VSCode MCP server
bash src/mcp/vscode/run.sh
```

The VSCode extension automatically manages the server lifecycle.

## VSCode Extension Usage

### Features

- **Git Toolbar Button**: Robot icon in the Source Control view
- **Action Menu**: Quick pick to select actions (commit, push, pull, custom)
- **MCP Integration**: Automatically calls the local MCP server

### Using the Extension

1. Open Source Control view in VSCode
2. Click the robot icon button
3. Select an action from the menu:
   - **Git Commit** - Stage and commit changes
   - **Git Push** - Push to remote
   - **Git Pull** - Pull from remote
   - **Custom Action** - Run custom commands
4. Enter optional message when prompted
5. View the result notification

## Development

### VSCode Extension Development

**Watch mode** (auto-recompile on changes):

```bash
cd .vscode/extensions/claude-mcp-vscode
npm run watch
```

**Build**:

```bash
npm run compile
```

**Package extension**:

```bash
npm install -g @vscode/vsce
vsce package
```

### MCP Server Development

Each MCP server is a standalone Go module:

```bash
cd src/mcp/<server-name>
go run .              # Run the server
go build              # Build binary
go test               # Run tests (when added)
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

## Troubleshooting

### Extension not showing in toolbar

1. Verify you're in a Git repository
2. Check that `scmProvider == git` in your workspace
3. Restart the Extension Development Host

### MCP Server errors

1. Verify Go is installed: `go version`
2. Check server logs in VSCode Output panel
3. Test server manually: `cd src/mcp/<name> && go run .`

### Build errors

```bash
# Clean and rebuild extension
cd .vscode/extensions/claude-mcp-vscode
rm -rf node_modules out
npm install
npm run compile
```

## Next Steps

- Add more MCP server tools
- Implement actual Git operations in the VSCode MCP server
- Create additional VSCode extension commands
- Add unit tests for both extension and servers
- Set up CI/CD pipeline

## Contributing

When adding new features:

1. Update relevant documentation
2. Follow existing code patterns
3. Test in the Extension Development Host
4. Update CLAUDE.md if changing architecture
