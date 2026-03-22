# clie-eac-bundle

The `clie-eac-bundle` module provides the release bundle that packages CLIE CLI and EAC extension together for simplified distribution and installation.

## Purpose

The bundle module enables:

- **Unified Distribution**: Single downloadable package containing both CLIE and EAC
- **Simplified Installation**: One-step installation process for users
- **Version Alignment**: Ensures CLIE and EAC versions are compatible
- **Cross-Platform Support**: Bundles for Windows, macOS, and Linux
- **Automated Updates**: Self-updating bundle mechanism

## Bundle Contents

The release bundle includes:

### CLIE CLI Binary

Platform-specific CLIE executables:

- `clie-linux-amd64` - Linux executable
- `clie-windows-amd64.exe` - Windows executable
- `clie-darwin-amd64` - macOS executable (Intel)
- `clie-darwin-arm64` - macOS executable (Apple Silicon)

### EAC Extension Image

Pre-built EAC container image:

- Image: `ghcr.io/ready-to-release/eac-ext:latest`
- Tagged with release version (e.g., `eac-ext:2024.02.1`)

### Configuration

Default CLIE configuration:

- `.clie/clie.yml` - Extension configuration
- Registry settings
- Volume mount defaults

### Installation Scripts

Platform-specific installers:

- `install.sh` - Unix/Linux/macOS installer
- `install.ps1` - Windows PowerShell installer
- `install.bat` - Windows batch installer

## Bundle Structure

```text
clie-eac-bundle-{version}/
├── bin/
│   ├── clie-linux-amd64
│   ├── clie-windows-amd64.exe
│   ├── clie-darwin-amd64
│   └── clie-darwin-arm64
├── config/
│   └── clie.yml
├── images/
│   └── eac-ext-{version}.tar  # Docker image archive
├── scripts/
│   ├── install.sh
│   ├── install.ps1
│   └── install.bat
├── README.md
├── LICENSE
└── CHANGELOG.md
```

## Installation Process

### Unix/Linux/macOS

```bash
# Download bundle
curl -L -o clie-eac-bundle.tar.gz \
  https://github.com/ready-to-release/eac/releases/latest/download/clie-eac-bundle-linux-amd64.tar.gz

# Extract
tar -xzf clie-eac-bundle.tar.gz
cd clie-eac-bundle

# Install
sudo ./scripts/install.sh
```

The installer:

1. Copies `clie` binary to `/usr/local/bin/`
2. Loads EAC Docker image
3. Creates default `.clie/clie.yml` in user home
4. Verifies installation

### Windows

```powershell
# Download bundle
Invoke-WebRequest -Uri "https://github.com/ready-to-release/eac/releases/latest/download/clie-eac-bundle-windows-amd64.zip" -OutFile clie-eac-bundle.zip

# Extract
Expand-Archive clie-eac-bundle.zip

# Install (Run as Administrator)
cd clie-eac-bundle
.\scripts\install.ps1
```

The installer:

1. Copies `clie.exe` to `C:\Program Files\CLIE\`
2. Adds to system PATH
3. Loads EAC Docker image
4. Creates default configuration

## Release Process

### Building the Bundle

```bash
# Build CLIE binaries for all platforms
eac build clie --platforms linux,windows,darwin

# Build EAC extension image
eac build eac-ext

# Export image to tar
docker save eac-ext:latest -o eac-ext.tar

# Package bundle
eac release bundle --version 2024.02.1
```

### Bundle Artifacts

Release artifacts uploaded to GitHub:

- `clie-eac-bundle-linux-amd64.tar.gz`
- `clie-eac-bundle-windows-amd64.zip`
- `clie-eac-bundle-darwin-amd64.tar.gz`
- `clie-eac-bundle-darwin-arm64.tar.gz`

## Version Management

### Bundle Versioning

Bundle version matches EAC release version:

- Format: `YYYY.MM.PATCH` (CalVer)
- Example: `2024.02.1`

### Component Versions

Bundle includes:

- **CLIE**: Pinned version (e.g., `1.2.3`)
- **EAC**: Release version (e.g., `2024.02.1`)

Version compatibility matrix maintained in `bundle-versions.yml`:

```yaml
bundles:
  - version: "2024.02.1"
    clie: "1.2.3"
    eac: "2024.02.1"
    docker_image: "ghcr.io/ready-to-release/eac-ext:2024.02.1"
```

## Dependencies

- **Depends On**:
  - `clie`: CLIE CLI binary
  - `eac-ext`: EAC extension image
- **Used By**: End users, installation scripts

## Configuration

Default bundle configuration in `config/clie.yml`:

```yaml
extensions:
  - name: eac
    image: ghcr.io/ready-to-release/eac-ext:2024.02.1
    load_local: false
    volumes:
      - "${PWD}:/workspace"
    environment:
      - "EAC_ENV=production"
```

## Testing

### Bundle Testing

Test bundle installation on all platforms:

```bash
# Test Linux bundle
docker run -v $(pwd):/workspace ubuntu:latest /workspace/scripts/install.sh

# Test Windows bundle
# (Run in Windows VM or container)

# Test macOS bundle
# (Run on macOS machine)
```

### Verification

Verify bundle installation:

```bash
# Check CLIE is installed
which clie
clie --version

# Check EAC is available
clie eac --help

# Run test command
clie eac show modules
```

## See Also

- [EAC Extension](eac-ext.md) - EAC containerized extension
- [Release Management](../continuous-delivery/index.md) - Release workflows
- [Installation Guide](../../devex/external/getting-started.md) - Installation instructions
