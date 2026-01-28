# Install Toolchain

## What You'll Accomplish

Install the R2R CLI and EAC extension to enable automation commands in your repository.

## Prerequisites

- **Docker Desktop** installed and running
- **Git** installed
- **Terminal access** (PowerShell on Windows, bash/zsh on macOS/Linux)

## Installation Options

### Option 1: Quick Install (Recommended)

#### Linux/macOS

```bash
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash
```

#### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/cli/install.ps1 | iex
```

### Option 2: Manual Installation

1. **Download the binary** from the [releases page](https://github.com/ready-to-release/eac/releases)
2. **Extract** to a directory in your PATH
3. **Verify** installation:

```bash
r2r version
```

## Post-Installation Setup

### 1. Initialize Configuration

```bash
r2r init
```

This creates `.r2r/r2r-cli.yml` with default extension registry settings.

### 2. Install EAC Extension

```bash
r2r install eac
```

This pulls the EAC Docker image and registers it with the CLI.

### 3. Verify Installation

```bash
# Check CLI version
r2r version

# Verify Docker connectivity
r2r verify

# Test EAC extension
r2r eac help
```

## Configuration Files Created

| File               | Purpose                                           |
| ------------------ | ------------------------------------------------- |
| `.r2r/r2r-cli.yml` | Extension registry and CLI settings               |
| `.r2r/eac/`        | EAC-specific configuration (created on first use) |

## Updating

### Update R2R CLI

```bash
# Re-run the install script
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash
```

### Update EAC Extension

```bash
r2r cleanup --all
r2r install eac
```

## Troubleshooting

| Problem             | Solution                                                    |
| ------------------- | ----------------------------------------------------------- |
| "Docker not found"  | Install Docker Desktop and ensure it's running              |
| "Permission denied" | Run terminal as administrator (Windows) or use sudo (Linux) |
| "Image pull failed" | Check network connectivity and Docker Hub access            |
| "Command not found" | Add installation directory to PATH                          |

## Next Steps

- [Configure AI Provider](./configure-ai.md) - Enable AI-powered features
- [Local Dev Workflows](./local-dev-workflows.md) - Development iteration cycles

## See Also

- [R2R CLI Reference](../../reference/r2r/commands/index.md) - Full CLI command documentation
- [CLI vs Extensions](../../reference/devex/internal/cli-vs-extensions.md) - Architecture overview
