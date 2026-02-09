# Install EAC CLI

## What You'll Accomplish

Install the standalone EAC CLI binary for native Go performance — no Docker required.

## Which CLI Do I Need?

| CLI                  | Install method         | Requires Docker | Best for                                                             |
| -------------------- | ---------------------- | --------------- | -------------------------------------------------------------------- |
| **clie** + eac-ext   | `clie install eac`     | Yes             | Full workflow: extensions, Docker-based tools, team standardization  |
| **eac** (standalone) | Installer script below | No              | Direct CLI access, CI runners, lightweight environments, development |

> Most users should start with [CLIE CLI](./install-toolchain.md) which includes EAC via Docker extension.
> Install the standalone EAC CLI if you want native performance or can't use Docker.

## Prerequisites

- **Git** installed
- **Terminal access** (PowerShell on Windows, bash/zsh on macOS/Linux)

## Quick Install

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/eac/install.sh | bash
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/eac/install.ps1 | iex
```

## Install Options

### Linux / macOS

```bash
# Download installer to pass arguments
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/eac/install.sh -o install.sh
chmod +x install.sh

# Install specific version
./install.sh --version eac/v1.0.0

# Install system-wide (requires sudo)
./install.sh --system

# UPX-compressed binary (smaller download, slightly slower startup)
./install.sh --upx

# Custom directory
./install.sh --install-dir ~/bin
```

### Windows (PowerShell)

```powershell
# Download installer to pass arguments
Invoke-WebRequest -Uri https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/eac/install.ps1 -OutFile install.ps1

# Install specific version
.\install.ps1 -Version "eac/v1.0.0"

# Install system-wide (requires Administrator)
.\install.ps1 -System

# UPX-compressed binary
.\install.ps1 -Upx

# Custom install directory
.\install.ps1 -InstallDir "C:\Tools\eac"
```

## Manual Installation

1. **Download the binary** from the [releases page](https://github.com/ready-to-release/eac/releases?q=eac)
2. **Place** in a directory on your PATH
3. **Verify** installation:

```bash
eac help
```

## Default Install Locations

| Platform    | User install                 | System install               |
| ----------- | ---------------------------- | ---------------------------- |
| Linux/macOS | `~/.local/bin/eac`           | `/usr/local/bin/eac`         |
| Windows     | `%LOCALAPPDATA%\eac\eac.exe` | `%ProgramFiles%\eac\eac.exe` |

## Verify Installation

```bash
# Check CLI is accessible
eac help

# Check version
eac version
```

## Updating

Re-run the install script — it replaces the existing binary automatically.

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/eac/install.sh | bash
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/eac/install.ps1 | iex
```

## Troubleshooting

| Problem             | Solution                                                                                                                  |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| "Command not found" | Add install directory to PATH, restart terminal                                                                           |
| "Permission denied" | Use `--system` with sudo (Linux) or run as Administrator (Windows)                                                        |
| Download fails      | Check network connectivity, verify release exists on [GitHub](https://github.com/ready-to-release/eac/releases?q=eac) |

## Next Steps

- [Configure AI Provider](./configure-ai.md) - Enable AI-powered features
- [Local Dev Workflows](./local-dev-workflows.md) - Development iteration cycles

## See Also

- [Install CLIE CLI](./install-toolchain.md) - Full toolchain with Docker extensions
- [Platform Troubleshooting](./platform-troubleshooting.md)
