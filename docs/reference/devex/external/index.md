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
| [Project Structure](./project-structure.md) | Recommended `.r2r/` directory layout |

## Quick Start

```bash
# 1. Install R2R CLI
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash

# 2. Initialize R2R
r2r init

# 3. Install EAC extension
r2r install eac

# 4. Initialize EAC in your project
eac init

# 5. Start using commands
eac show-modules
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
eac create commit-message   # Generate commit messages
eac create-pr              # Create PRs with AI descriptions
eac create-spec            # Generate Gherkin specifications
```

## Requirements

- **Docker** - EAC runs as a container
- **Git** - Version control
- **R2R CLI** - The framework that runs EAC

## Next Steps

1. [Getting Started](./getting-started.md) - Detailed setup guide
2. [Configuration](./configuration.md) - Customize EAC for your needs
3. [Project Structure](./project-structure.md) - Organize your `.r2r/` directory

## See Also

- [Command Reference](../../eac/commands/index.md) - All EAC commands
- [Local Setup](../../../how-to-guides/local-setup/index.md) - Installation guides
