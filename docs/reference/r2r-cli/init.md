# r2r init

Creates a minimal `.r2r/r2r-cli.yml` configuration file in the repository root. This is the first step in setting up R2R for extension management.

## Syntax

```bash
r2r init [flags]
```

## Description

The `init` command creates the `.r2r/r2r-cli.yml` file in your project. This file manages extension registration and configuration.

**What it does:**

1. Creates the `.r2r/` directory if needed
2. Generates a minimal `r2r-cli.yml` with empty extension list
3. Ready for extension installation via `r2r install`

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
r2r init
```

### Reset Configuration

Remove all existing configuration and start fresh:

```bash
r2r init --delete-configs
```

### Three-Step Setup

```bash
# Step 1: Initialize R2R configuration
r2r init

# Step 2: Install EAC extension
r2r install eac

# Step 3: Configure EAC
r2r eac init --ai-provider claude-api
```

## File Structure

After `r2r init`:

```text
your-project/
├── .r2r/
│   └── r2r-cli.yml      # Extension registry (empty)
└── (your project files)
```

After `r2r install eac`:

```text
your-project/
├── .r2r/
│   └── r2r-cli.yml      # Contains eac extension
└── (your project files)
```

After `r2r eac init`:

```text
your-project/
├── .r2r/
│   ├── r2r-cli.yml          # Extension registry
│   └── eac/                 # EAC configuration
│       ├── ai-provider.yml
│       └── repository.yml
└── (your project files)
```

## Configuration File

The generated configuration:

```yaml
# Empty configuration - extensions added via 'r2r install'
extensions: []
```

After installing EAC:

```yaml
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/ext-eac:latest'
    description: 'Everything-as-Code automation'
```

## See Also

- [R2R CLI Overview](index.md) - Command overview and architecture
- [install command](install.md) - Install extensions after init
- [Configuration Reference](configuration.md) - Detailed configuration options
- [Quick Start Tutorial](../../tutorials/getting-started/quick-start.md) - Complete setup guide
- [CLI vs Extensions](../r2r-eac/cli-vs-extensions.md) - Understanding the architecture
