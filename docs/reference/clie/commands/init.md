# clie init

Creates a minimal `.clie/clie-cli.yml` configuration file in the repository root.

This is the first step in setting up CLIE for extension management.

## Syntax

```bash
clie init [flags]
```

## Description

The `init` command creates the `.clie/clie-cli.yml` file in your project.

This file manages extension registration and configuration.

**What it does:**

1. Creates the `.clie/` directory if needed
2. Generates a minimal `clie-cli.yml` with empty extension list
3. Ready for extension installation via `clie install`

**What it creates:**

```yaml
extensions: []
```

## Flags

| Flag                | Description                                                       |
| ------------------- | ----------------------------------------------------------------- |
| `--delete-configs`  | Delete all configuration files including overrides                |
| `--use-pwd-as-root` | Use current directory as repository root (creates .git if needed) |
| `-h, --help`        | Display help information                                          |

## Examples

### Basic Initialization

```bash
cd /path/to/your/project
clie init
```

### Reset Configuration

Remove all existing configuration and start fresh:

```bash
clie init --delete-configs
```

### Three-Step Setup

```bash
# Step 1: Initialize CLIE configuration
clie init

# Step 2: Install EAC extension
clie install eac

# Step 3: Configure EAC
clie eac init --ai-provider claude-api
```

## File Structure

After `clie init`:

```text
your-project/
├── .clie/
│   └── clie-cli.yml      # Extension registry (empty)
└── (your project files)
```

After `clie install eac`:

```text
your-project/
├── .clie/
│   └── clie-cli.yml      # Contains eac extension
└── (your project files)
```

After `clie eac init`:

```text
your-project/
├── .clie/
│   ├── clie-cli.yml          # Extension registry
│   └── eac/                 # EAC configuration
│       ├── ai-provider.yml
│       └── repository.yml
└── (your project files)
```

## Configuration File

The generated configuration:

```yaml
# Empty configuration - extensions added via 'clie install'
extensions: []
```

After installing EAC:

```yaml
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/eac-ext:latest'
    description: 'Everything-as-Code automation'
```

## See Also

- [CLIE CLI Overview](index.md) - Command overview and architecture
- [install command](install.md) - Install extensions after init
- [Configuration Reference](configuration.md) - Detailed configuration options
- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md) - Complete setup guide
- [CLI vs Extensions](../../eac/architecture/cli-integration.md) - Understanding the architecture
