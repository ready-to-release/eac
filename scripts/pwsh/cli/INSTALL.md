# Windows Installation

## Quick Install (Latest Version)

```powershell
irm https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/cli/install.ps1 | iex
```

## Install with Options

Download the script first to pass arguments:

```powershell
# Download installer
Invoke-WebRequest -Uri https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/cli/install.ps1 -OutFile install.ps1

# Install specific version
.\install.ps1 -Version "r2r-cli/v1.0.0"

# Install system-wide (requires Administrator)
.\install.ps1 -System

# Install UPX-compressed binary (smaller, slightly slower startup)
.\install.ps1 -Upx

# Custom install directory
.\install.ps1 -InstallDir "C:\Tools\r2r"
```

## Default Install Location

- **User install**: `%LOCALAPPDATA%\r2r\r2r.exe`
- **System install**: `%ProgramFiles%\r2r\r2r.exe`

## Note

You may need to restart your terminal for PATH changes to take effect.
