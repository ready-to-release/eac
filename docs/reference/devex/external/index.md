# External: Using EAC in Your Project

Reference documentation for developers adopting EAC in their own repositories.

## Overview

EAC (Everything as Code) provides automation for:

- Building and testing modules
- Generating documentation
- Managing releases
- Security scanning
- AI-powered workflows

## In This Section

| Topic                                       | Description                          |
| ------------------------------------------- | ------------------------------------ |
| [Getting Started](./getting-started.md)     | First steps adopting EAC             |
| [Configuration](./configuration.md)         | Configure EAC for your project       |
| [Project Structure](./project-structure.md) | Recommended `.eac/` directory layout |

## Quick Start

```bash
# 1. Install EAC CLI
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/eac/install.sh | bash

# 2. Initialize EAC in your project
eac init

# 3. Start using commands
eac show modules
eac validate
```

## What EAC Provides

### Module Management

Define modules with dependencies, file ownership, and build configuration:

```yaml
# .eac/repository.yml
modules:
  - moniker: my-service
    type: go-cli
    depends_on: [my-library]
```

### Automation Commands

Many commands for common tasks:

```bash
eac build <module>      # Build modules
eac test <module>       # Run tests
eac validate            # Validate configuration
eac scan                # Security scanning
```

### AI-Powered Features

With an AI provider configured:

```bash
eac get commit-message      # Generate commit messages
eac create pr              # Create PRs with AI descriptions
eac create spec            # Generate Gherkin specifications
```

## Requirements

- **Git** - Version control
- **CLIE CLI** - Optional, for containerized execution

## Next Steps

1. [Getting Started](./getting-started.md) - Detailed setup guide
2. [Configuration](./configuration.md) - Customize EAC for your needs
3. [Project Structure](./project-structure.md) - Organize your `.clie/` directory

## See Also

- [Command Reference](../../eac/commands/index.md) - All EAC commands
- [Local Setup](../../../how-to-guides/local-setup/index.md) - Installation guides
