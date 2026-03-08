# Local Setup

Set up your local development environment for working with EAC and CLIE.

These guides cover installation, configuration, and platform-specific troubleshooting.

## In This Section

| Guide                                                     | Description                                     |
| --------------------------------------------------------- | ----------------------------------------------- |
| [Install Toolchain](./install-toolchain.md)               | Install EAC CLI (standalone or via CLIE)          |
| [Configure AI Provider](./configure-ai.md)                | Set up Anthropic Claude for AI features         |
| [Configure Claude Code](./configure-claude-code.md)       | Use Claude Code effectively in this repository  |
| [Configure Azure MCP](./configure-azure-mcp.md)           | Enable Azure tools in Claude Code (Windows WSL) |
| [Local Dev Workflows](./local-dev-workflows.md)           | Development and testing workflows               |
| [Platform Troubleshooting](./platform-troubleshooting.md) | Windows Defender, permissions, and other issues |

## Quick Start

```bash
# 1. Install EAC CLI
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/eac/install.sh | bash

# 2. Initialize EAC in your project
eac init --ai-provider claude-api

# 3. Explore commands
eac help
```

## Prerequisites

- **Docker** - Required only if using CLIE container host
- **Git** - Version control
- **Go 1.21+** - For local development (optional)

## Next Steps

After setup, explore:

- [EAC Commands](../eac/commands/index.md) - Available automation commands
- [Creating CLIE Extensions](../clie/creating-extensions.md) - Build containerized extensions (optional)
