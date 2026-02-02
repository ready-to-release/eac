# Getting Started with EAC

First steps for adopting EAC in your repository.

## Prerequisites

Before starting:

- [ ] Docker Desktop installed and running
- [ ] Git repository initialized
- [ ] Terminal access (bash, zsh, or PowerShell)

## Step 1: Install R2R CLI

### Linux/macOS

```bash
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash
```

### Windows

```powershell
irm https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/cli/install.ps1 | iex
```

### Verify Installation

```bash
r2r version
```

## Step 2: Initialize R2R

```bash
r2r init
```

This creates `.r2r/r2r-cli.yml` with extension registry settings.

## Step 3: Install EAC Extension

```bash
r2r install eac
```

This pulls the EAC Docker image.

## Step 4: Initialize EAC

```bash
eac init
```

This creates the `.eac/` directory with default configuration.

## Step 5: Define Your First Module

Edit `.eac/repository.yml`:

```yaml
modules:
  - moniker: my-app
    name: My Application
    type: go-cli
    files:
      root: cmd/my-app
      source: ["**/*.go"]
```

## Step 6: Verify Setup

```bash
# Show modules
eac show-modules

# Validate configuration
eac validate

# Check module dependencies
eac show-dependencies
```

## Optional: Configure AI Provider

For AI-powered features (commit messages, PR descriptions):

```bash
# Set API key environment variable
export ANTHROPIC_API_KEY="sk-ant-..."

# Initialize AI provider
eac init --ai claude-api
```

## What You Now Have

After setup, your project includes:

```text
your-project/
├── .r2r/
│   ├── r2r-cli.yml          # R2R CLI configuration
│   └── eac/
│       └── repository.yml    # Module definitions
└── ... (your code)
```

## Common First Commands

```bash
# View all modules
eac show-modules

# Validate all contracts
eac validate

# Show available commands
eac help

# Build a module
eac build my-app

# Run tests
eac test my-app
```

## Troubleshooting

| Problem              | Solution                         |
| -------------------- | -------------------------------- |
| "Docker not running" | Start Docker Desktop             |
| "Image not found"    | Run `r2r install eac`            |
| "Module not found"   | Check `repository.yml` syntax    |
| "Invalid type"       | Use valid component type (go, typescript, etc.) |

## Next Steps

- [Configuration](./configuration.md) - Customize EAC settings
- [Project Structure](./project-structure.md) - Organize configuration files
- [Command Reference](../../eac/commands/index.md) - Explore all commands

## See Also

- [Install Toolchain](../../../how-to-guides/local-setup/install-toolchain.md) - Detailed installation guide
- [Configure AI Provider](../../../how-to-guides/local-setup/configure-ai.md) - AI setup guide
