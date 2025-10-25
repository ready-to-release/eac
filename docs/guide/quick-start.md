# Quick Start Guide

Get started with the CLI project in minutes.

## Running MCP Servers

Run any MCP server directly:

```bash
# PowerShell execution server
cd src/mcp/pwsh && go run .

# Documentation server
cd src/mcp/docs && go run .

# GitHub integration server
cd src/mcp/github && go run .

# VSCode actions server
cd src/mcp/vscode && go run .
```

## Documentation Site

View the full documentation site locally:

```bash
# Using Docker (recommended)
cd containers/mkdocs && ./serve.sh

# Open: http://localhost:8000
```

### Alternative: Local Installation

```bash
# Install MkDocs
pip install mkdocs-material

# Serve locally
mkdocs serve

# Open: http://127.0.0.1:8000
```

### Build Static Site

```bash
mkdocs build
# Output: site/
```

## VSCode Extension

Get the VSCode extension running:

```bash
# Automated setup
./automation/sh-vscode/init.sh

# Launch extension (in VSCode, press F5)
```

**See:** [VSCode Extension Quick Start](vscode-extension/QUICKSTART.md) for detailed instructions.

## Next Steps

- [Project Architecture](architecture.md) - Understand the project structure
- [MCP Servers Guide](vscode-extension/mcp-servers.md) - Develop MCP servers
- [VSCode Extension Guide](vscode-extension/index.md) - Customize the extension
