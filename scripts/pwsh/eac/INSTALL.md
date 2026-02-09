# Windows Installation

## Quick Install (Latest Version)

```powershell
irm https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/eac/install.ps1 | iex
```

## Install with Options

Download the script first to pass arguments:

```powershell
# Download installer
Invoke-WebRequest -Uri https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/eac/install.ps1 -OutFile install.ps1

# Install specific version
.\install.ps1 -Version "eac/v1.0.0"

# Install system-wide (requires Administrator)
.\install.ps1 -System

# Install UPX-compressed binary (smaller, slightly slower startup)
.\install.ps1 -Upx

# Custom install directory
.\install.ps1 -InstallDir "C:\Tools\eac"
```

## Default Install Location

- **User install**: `%LOCALAPPDATA%\eac\eac.exe`
- **System install**: `%ProgramFiles%\eac\eac.exe`

## Note

You may need to restart your terminal for PATH changes to take effect.
