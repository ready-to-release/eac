# r2r verify

Verify system prerequisites and environment setup for R2R and extensions.

## Syntax

```bash
r2r verify [flags]
```

## Description

The `verify` command checks that your system meets all requirements for running R2R and the EAC extension.

**What it checks:**

- Docker installation
- Docker daemon running
- Docker version
- Network connectivity to registry
- Permissions
- Disk space
- Configuration validity

## Examples

### Basic Verification

```bash
r2r verify
```

**Output when healthy:**

```text
✓ Docker is installed (version 24.0.5)
✓ Docker daemon is running
✓ Docker permissions are correct
✓ Network connectivity to registry (ghcr.io)
✓ Disk space sufficient (15 GB available)
✓ R2R configuration is valid

All prerequisites met. R2R CLI is ready to use.
```

**Output when issues found:**

```text
✓ Docker is installed (version 20.10.8)
✗ Docker daemon is not running
✗ Cannot reach registry (ghcr.io)
⚠ Low disk space (3 GB available, recommend 10 GB)
✓ R2R configuration is valid

Issues found. Please fix the problems above.
```

### Before Installation

```bash
# Verify prerequisites before installing extensions
r2r verify

# If successful, proceed with installation
r2r install eac
```

## Verification Checks

### Docker Installation

Checks that Docker binary is present and executable.

### Docker Daemon Running

Checks that Docker daemon is active and responding.

### Docker Version

Checks Docker version meets minimum requirements (20.10.0+).

### Docker Permissions

Checks current user can run Docker commands.

### Network Connectivity

Checks access to GitHub Container Registry (ghcr.io).

### Disk Space

Checks sufficient disk space for Docker images.

- **Recommended:** 10 GB
- **Minimum:** 5 GB

### Configuration Validity

Checks `.r2r/r2r-cli.yml` is valid.

## See Also

- [R2R CLI Overview](index.md) - Command overview
- [install command](install.md) - Install extensions after verification
- [validate command](validate.md) - Validate configuration
- [cleanup command](cleanup.md) - Free disk space
- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md) - Installation guide
