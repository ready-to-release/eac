# Linux/macOS Installation

## Quick Install (Latest Version)

```bash
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash
```

## Install with Options

Download the script first to pass arguments:

```bash
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh -o install.sh
chmod +x install.sh

# Install specific version
./install.sh --version clie-cli/v1.0.0

# Install system-wide (requires sudo)
./install.sh --system

# Install UPX-compressed binary (smaller, slightly slower startup)
./install.sh --upx

# Custom install directory
./install.sh --install-dir ~/bin
```

## Using wget

```bash
wget -qO- https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash
```

## Default Install Location

- **User install**: `~/.local/bin/clie`
- **System install**: `/usr/local/bin/clie`
