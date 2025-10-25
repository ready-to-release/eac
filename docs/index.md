# CLI Project Documentation

Welcome to the CLI project documentation.

This is an extensible CLI-in-a-box for managing software delivery flows with integrated MCP servers and VSCode extension.

## Quick Navigation

### Getting Started

- [Quick Start Guide](../QUICKSTART.md) - Get up and running in 5 minutes
- [Complete Usage Guide](../USAGE.md) - Comprehensive usage examples
- [Installation](installation.md) - Detailed installation instructions

### Guides

- [VSCode Extension](guide/vscode-ext-claude-commit/index.md) - Using and developing the extension
- [Recording Demos](guide/vscode-ext-claude-commit/recording-demos.md) - Create demo videos
- [MCP Servers](reference/mcp-servers.md) - Working with MCP servers
- [Automation Scripts](reference/automation.md) - Build and maintenance scripts

### Reference

- [Project Structure](reference/structure.md) - Codebase organization
- [Configuration](reference/configuration.md) - Configuration files
- [API Reference](reference/api.md) - MCP server API

## Project Overview

This project provides three main components:

### 1. MCP Servers

Go-based Model Context Protocol servers for various integrations:

- **pwsh** - PowerShell command execution
- **docs** - Documentation management
- **github** - GitHub API integration
- **vscode** - VSCode action integration

### 2. VSCode Extension

Local extension that adds a button to the Git toolbar:

- Integrates with MCP servers
- Provides quick access to git operations
- Extensible action system

### 3. Automation Scripts

Shell scripts for common tasks:

- **init.sh** - First-time setup
- **restore.sh** - Restore dependencies
- **clean.sh** - Clean build artifacts

## Architecture

```text
.
├── src/mcp/                    # MCP servers (Go)
│   ├── pwsh/                   # PowerShell execution
│   ├── docs/                   # Documentation
│   ├── github/                 # GitHub integration
│   └── vscode/                 # VSCode integration
├── .vscode/extensions/         # VSCode extensions
│   └── claude-mcp-vscode/      # MCP integration extension
├── automation/sh-vscode/       # Automation scripts
├── docs/                       # Documentation (you are here)
└── .mcp.json                   # MCP server configuration
```

## Key Features

### MCP Integration

- JSON-RPC protocol for tool invocation
- Stdin/stdout communication
- Extensible tool system

### VSCode Extension

- Git toolbar integration
- MCP server lifecycle management
- Action quick-pick menu
- Real-time notifications

### Automation

- One-command initialization
- Dependency management
- Build artifact cleanup

## Prerequisites

- **Go** 1.21 or higher
- **Node.js** 18.x or higher
- **npm** (comes with Node.js)
- **VSCode** 1.80.0 or higher
- **Git** for version control

## Quick Start

```bash
# Initialize everything
./automation/sh-vscode/init.sh

# Open in VSCode
code .

# Press F5 to launch Extension Development Host
# Then: Ctrl+Shift+G → Click 🤖 robot button
```

## Documentation Structure

This documentation is organized into several sections:

### Guides - 2

Step-by-step instructions for common tasks:

- Using the VSCode extension
- Recording demo videos
- Developing MCP servers
- Automation workflows

### Tutorials

Hands-on learning materials:

- Building your first MCP tool
- Creating custom VSCode commands
- Extending the automation scripts

### Reference - 2

Technical specifications and API documentation:

- MCP server protocol
- Extension API
- Configuration options
- Project structure

## Getting Help

- Check the [Quick Start Guide](../QUICKSTART.md) for immediate help
- Review [Usage Examples](../USAGE.md) for common scenarios
- See [Troubleshooting](guide/vscode-ext-claude-commit/index.md#troubleshooting) for common issues
- Open an issue on GitHub for bugs or feature requests

## Contributing

When contributing to this project:

1. Follow the existing code patterns
2. Update relevant documentation
3. Test in Extension Development Host
4. Update CLAUDE.md if changing architecture

## Project Status

| Component | Status |
|-----------|--------|
| MCP Servers | ✅ 4 servers operational |
| VSCode Extension | ✅ Functional |
| Automation Scripts | ✅ Complete |
| Documentation | ✅ Comprehensive |
| Demo Videos | ⏳ Storyboards ready, awaiting recording |

## Resources

**External Links:**

- [Model Context Protocol](https://modelcontextprotocol.io/) - MCP specification
- [VSCode Extension API](https://code.visualstudio.com/api) - Extension development
- [Go Documentation](https://golang.org/doc/) - Go language

## Next Steps

1. **New Users**: Start with the [Quick Start Guide](../QUICKSTART.md)
2. **Developers**: Read the [VSCode Extension Guide](guide/vscode-ext-claude-commit/index.md)
3. **Contributors**: Check the [Project Structure](reference/structure.md)
