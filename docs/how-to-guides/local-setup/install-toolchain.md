# Install Toolchain

## What You'll Accomplish

Install the EAC CLI to enable automation commands in your repository.

Two installation methods are available:

- **Standalone EAC CLI (Recommended)** - Direct binary, no Docker required
- **CLIE + EAC Extension** - Containerized execution via Docker

## Prerequisites

- **Git** installed
- **Terminal access** (PowerShell on Windows, bash/zsh on macOS/Linux)

## Option 1: Standalone EAC CLI (Recommended)

Install the `eac` binary directly. No Docker required.

### Linux/macOS

```bash
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/eac/install.sh | bash
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/eac/install.ps1 | iex
```

### Verify

```bash
eac version
```

### Post-Installation

```bash
# Initialize EAC in your project
eac init --ai-provider claude-api

# Explore commands
eac help
```

## Option 2: CLIE Extension Host (Containerized)

Install the CLIE CLI framework to run EAC inside Docker containers. Use this for reproducible, isolated environments.

**Requires Docker.**

### Linux/macOS

```bash
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/cli/install.ps1 | iex
```

### Post-Installation

```bash
clie init
clie install eac
eac help
```

## Configuration Files Created

| File               | Purpose                                           |
| ------------------ | ------------------------------------------------- |
| `.eac/`            | EAC configuration (created by `eac init`)         |
| `.clie/clie.yml`   | Extension registry (only if using CLIE)           |

## Updating

### Update EAC CLI (Standalone)

```bash
# Re-run the install script
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/eac/install.sh | bash
```

### Update CLIE + EAC Extension

```bash
# Update CLIE CLI
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash

# Update EAC extension
clie cleanup --all
clie install eac
```

## Troubleshooting

| Problem             | Solution                                                    |
| ------------------- | ----------------------------------------------------------- |
| "Docker not found"  | Only needed for CLIE. Install Docker Desktop if using CLIE  |
| "Permission denied" | Run terminal as administrator (Windows) or use sudo (Linux) |
| "Image pull failed" | Check network connectivity and Docker Hub access            |
| "Command not found" | Add installation directory to PATH                          |

## Next Steps

- [Configure AI Provider](./configure-ai.md) - Enable AI-powered features
- [Local Dev Workflows](./local-dev-workflows.md) - Development iteration cycles

## See Also

- [Install EAC CLI](./install-eac.md) - Detailed standalone installation options
- [CLIE CLI Reference](../../reference/clie/commands/index.md) - Full CLIE command documentation
