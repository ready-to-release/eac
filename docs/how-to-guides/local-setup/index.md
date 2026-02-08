# Local Setup

Set up your local development environment for working with EAC and CLIE.

These guides cover installation, configuration, and platform-specific troubleshooting.

## In This Section

| Guide                                                     | Description                                     |
| --------------------------------------------------------- | ----------------------------------------------- |
| [Install Toolchain](./install-toolchain.md)               | Install CLIE CLI and EAC extension               |
| [Configure AI Provider](./configure-ai.md)                | Set up Anthropic Claude for AI features         |
| [Configure Claude Code](./configure-claude-code.md)       | Use Claude Code effectively in this repository  |
| [Local Dev Workflows](./local-dev-workflows.md)           | Development and testing workflows               |
| [Platform Troubleshooting](./platform-troubleshooting.md) | Windows Defender, permissions, and other issues |

## Quick Start

```bash
# 1. Install CLIE CLI
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash

# 2. Initialize configuration
clie init

# 3. Install EAC extension
clie install eac

# 4. Configure AI provider (optional)
clie eac init --ai claude-api
```

## Prerequisites

- **Docker** - Required for running extensions
- **Git** - Version control
- **Go 1.21+** - For local development (optional)

## Next Steps

After setup, explore:

- [EAC Commands](../eac/commands/index.md) - Available automation commands
- [Creating Extensions](../clie/creating-extensions.md) - Build your own CLIE extensions
