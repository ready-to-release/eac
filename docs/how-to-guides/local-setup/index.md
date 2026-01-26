# Local Setup

Set up your local development environment for working with R2R and EAC.

These guides cover installation, configuration, and platform-specific troubleshooting.

## In This Section

| Guide                                                     | Description                                     |
| --------------------------------------------------------- | ----------------------------------------------- |
| [Install Toolchain](./install-toolchain.md)               | Install R2R CLI and EAC extension               |
| [Configure AI Provider](./configure-ai.md)                | Set up Anthropic Claude for AI features         |
| [Configure Claude Code](./configure-claude-code.md)       | Use Claude Code effectively in this repository  |
| [Local Dev Workflows](./local-dev-workflows.md)           | Development and testing workflows               |
| [Platform Troubleshooting](./platform-troubleshooting.md) | Windows Defender, permissions, and other issues |

## Quick Start

```bash
# 1. Install R2R CLI
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash

# 2. Initialize configuration
r2r init

# 3. Install EAC extension
r2r install eac

# 4. Configure AI provider (optional)
r2r eac init --ai claude-api
```

## Prerequisites

- **Docker** - Required for running extensions
- **Git** - Version control
- **Go 1.21+** - For local development (optional)

## Next Steps

After setup, explore:

- [EAC Commands](../eac/commands/) - Available automation commands
- [Creating Extensions](../r2r/creating-extensions.md) - Build your own R2R extensions
