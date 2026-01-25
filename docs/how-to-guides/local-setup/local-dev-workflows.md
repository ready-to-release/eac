# Local Development Workflows

{{ page_breadcrumb() }}

**Problem**: You want to develop and test r2r extensions without pushing to remote registries during development.

**Solution**: Use different workflows depending on where you're working:

- **EAC repository**: Use `importer.ps1` to load commands directly from source
- **External repositories**: Use Docker-based local development with locally built images

## Overview

### Two Development Workflows

#### 1. EAC Repository Development (importer.ps1)

When developing in the EAC repository itself, use `importer.ps1` to load commands directly:

```powershell
# In the EAC repository
.\scripts\pwsh\importer.ps1

# Set commands path
$env:R2R_COMMANDS_PATH = '.\go\eac\commands'

# Load commands
r2r load-commands

# Test your changes immediately
r2r eac <command>
```

**Advantages:**

- Instant feedback on code changes
- No Docker build required
- Fastest iteration cycle
- Direct debugging

#### 2. External Repository Testing (Docker)

When testing in external repositories (e.g., `ext-env-check`), use Docker-based local development:

```powershell
# Build local Docker image in EAC repo
cd C:\source\ready-to-release\eac
.\scripts\pwsh\local-dev\build-local.ps1

# Setup external repository
cd C:\path\to\external-repo
C:\source\ready-to-release\eac\scripts\pwsh\local-dev\setup.ps1 -TargetRepo .
```

**Advantages:**

- Tests in realistic deployment environment
- Validates Dockerfile and container configuration
- Tests across different repository structures
- Mirrors production usage

### When to Use Which Workflow

| Scenario                    | Use          | Reason                |
| --------------------------- | ------------ | --------------------- |
| Writing new EAC commands    | importer.ps1 | Fastest iteration     |
| Debugging EAC code          | importer.ps1 | Direct code execution |
| Testing in different repo   | Docker       | Realistic environment |
| Validating container config | Docker       | Production-like setup |
| Testing multi-repo support  | Docker       | Integration testing   |

## EAC Repository Development

### Quick Start with importer.ps1

```powershell
# Navigate to EAC repository
cd C:\source\ready-to-release\eac

# Run importer to install/update r2r binary
.\scripts\pwsh\importer.ps1

# Set commands path
$env:R2R_COMMANDS_PATH = '.\go\eac\commands'

# Load commands directly from source
r2r load-commands

# Test immediately
r2r eac help
```

### Development Iteration

1. **Make code changes** in `go/eac/commands/`
2. **Reload commands**: `r2r load-commands`
3. **Test immediately**: `r2r eac <command>`
4. **No Docker rebuild needed**

## External Repository Testing

This section covers testing your EAC extension in external repositories using Docker.

### Components

The Docker-based workflow uses three scripts:

1. **Docker Image Builder** (`build-local.ps1`) - Builds local Docker images
2. **Local Config Generator** (`init-local.ps1`) - Creates r2r configuration
3. **Setup Orchestrator** (`setup-local-dev.ps1`) - Coordinates complete setup

### Prerequisites

For Docker-based external repository testing:

- Docker Desktop installed and running
- Git installed
- PowerShell 5.1 or later (Windows)
- Access to the EAC repository source code
- An external repository to test in (with git initialized)

### Quick Start for External Repositories

Set up an external repository in one command:

```powershell
# Navigate to external repository
cd C:\path\to\external-repo

# Run setup from EAC repository
C:\source\ready-to-release\eac\scripts\pwsh\local-dev\setup.ps1 -TargetRepo .
```

This will:

1. Check prerequisites
2. Build the Docker image as `ext-eac:dev` in EAC repo
3. Install the r2r binary (if not already installed)
4. Create `.r2r/r2r-cli.local.yml` configuration in the external repo
5. Test the setup

## Step-by-Step Setup for External Repositories

This section shows how to manually set up an external repository step by step.

### 1. Build Local Docker Image

Build the EAC extension Docker image locally in the EAC repository:

```powershell
# Basic build (single platform, linux/amd64)
.\scripts\pwsh\local-dev\build-local.ps1

# With custom tag
.\scripts\pwsh\local-dev\build-local.ps1 -Tag "ext-eac:my-feature"

# Multi-platform build (amd64 and arm64)
.\scripts\pwsh\local-dev\build-local.ps1 -MultiPlatform
```

**Build Options:**

| Parameter        | Description                    | Default       |
| ---------------- | ------------------------------ | ------------- |
| `-Tag`           | Docker image tag               | `ext-eac:dev` |
| `-Platform`      | Target platform                | `linux/amd64` |
| `-MultiPlatform` | Build for both amd64 and arm64 | `false`       |
| `-Push`          | Push to registry after build   | `false`       |
| `-NoBuildKit`    | Disable Docker BuildKit        | `false`       |

**Verify the build:**

```powershell
docker images ext-eac:dev
```

### 2. Create Local Configuration

Create a local r2r configuration that uses your Docker image:

```powershell
# Basic configuration
.\scripts\pwsh\cli\init-local.ps1

# With custom image tag
.\scripts\pwsh\cli\init-local.ps1 -ImageTag "ext-eac:my-feature"

# With workspace mounting for live development
.\scripts\pwsh\cli\init-local.ps1 -MountWorkspace

# Force overwrite existing config
.\scripts\pwsh\cli\init-local.ps1 -Force
```

**Configuration Options:**

| Parameter         | Description                           | Default       |
| ----------------- | ------------------------------------- | ------------- |
| `-ImageTag`       | Docker image tag to use               | `ext-eac:dev` |
| `-Force`          | Overwrite existing config             | `false`       |
| `-MountWorkspace` | Add volume mount for live development | `false`       |
| `-EnableDebug`    | Add debug environment variables       | `false`       |

This creates `.r2r/r2r-cli.local.yml` with:

```yaml
# Local development configuration (gitignored)
load_local: true

extensions:
  - name: 'eac'
    image: 'ext-eac:dev'
    load_local: true
    image_pull_policy: 'Never'
    auto_remove_children: false
```

### 3. Set Environment Variable

For each terminal session, set the repository root:

```powershell
$env:R2R_REPO_ROOT = (Get-Location)
```

Or for a specific path:

```powershell
$env:R2R_REPO_ROOT = "C:\path\to\your\repository"
```

### 4. Test Your Setup

```powershell
# Test basic functionality
r2r eac help

# Test command execution
r2r eac show modules
```

## Configuration Details

### Local Override Files

r2r uses a configuration hierarchy:

1. `.r2r/r2r-cli.yml` - Base configuration (committed to git)
2. `.r2r/r2r-cli.local.yml` - Local overrides (gitignored)

Local files override base configuration for development.

### Key Configuration Fields

For local development, these fields are critical:

```yaml
load_local: true                    # Skip registry pulls (global)

extensions:
  - name: 'eac'
    image: 'ext-eac:dev'           # Local Docker image tag
    load_local: true               # Skip registry pulls (per-extension)
    image_pull_policy: 'Never'     # Fail if image not found locally
    auto_remove_children: false    # Keep containers for debugging
```

### Debug Configuration

Enable debug mode for additional logging:

```powershell
.\scripts\pwsh\cli\init-local.ps1 -EnableDebug
```

This adds debug environment variables:

```yaml
extensions:
  - name: 'eac'
    # ... other config ...
    env:
      - name: 'EAC_DEBUG'
        value: 'true'
      - name: 'R2R_LOCAL_DEV'
        value: 'true'
```

### Workspace Mounting

Mount your workspace for live code changes (advanced):

```powershell
.\scripts\pwsh\cli\init-local.ps1 -MountWorkspace
```

This adds volume mounts:

```yaml
extensions:
  - name: 'eac'
    # ... other config ...
    volumes:
      - host: '.'
        container: '/workspace'
        readonly: false
      - host: '.git'
        container: '/.git'
        readonly: true
```

## Development Workflow

### Iteration Cycle

When developing and testing in external repositories:

1. **Modify code in EAC repository**:

   ```powershell
   cd C:\source\ready-to-release\eac
   # Edit files in go/eac/commands/
   ```

2. **Rebuild Docker image**:

   ```powershell
   .\scripts\pwsh\local-dev\build-local.ps1
   ```

3. **Test in external repository**:

   ```powershell
   cd C:\path\to\external-repo
   $env:R2R_REPO_ROOT = (Get-Location)
   r2r eac <command>
   ```

**Note**: For faster iteration when developing commands, use `importer.ps1` in the EAC repository instead of the Docker workflow.

### Comparison: importer.ps1 vs Docker

| Workflow                | Code Changes | Test Changes        | Iteration Time |
| ----------------------- | ------------ | ------------------- | -------------- |
| importer.ps1 (EAC repo) | Edit code    | `r2r load-commands` | Seconds        |
| Docker (external repo)  | Edit code    | Rebuild image       | Minutes        |

**Recommendation**: Develop and debug in EAC repo with `importer.ps1`, then validate in external repos with Docker.

## Advanced Usage

### Custom Image Tags

Use meaningful tags for feature branches:

```powershell
# Build with feature tag
.\scripts\pwsh\local-dev\build-local.ps1 -Tag "ext-eac:feature-xyz"

# Configure to use feature tag
.\scripts\pwsh\cli\init-local.ps1 -ImageTag "ext-eac:feature-xyz"
```

### Multi-Platform Development

Build for multiple platforms:

```powershell
.\scripts\pwsh\local-dev\build-local.ps1 -MultiPlatform
```

**Note**: Multi-platform images cannot be loaded directly into local Docker. For local testing, use single-platform builds.

### Skip Setup Steps

Skip parts of the setup process:

```powershell
# Skip Docker build (use existing image)
.\scripts\pwsh\setup-local-dev.ps1 -SkipBuild

# Skip r2r installation (already installed)
.\scripts\pwsh\setup-local-dev.ps1 -SkipInstall

# Skip validation tests
.\scripts\pwsh\setup-local-dev.ps1 -SkipTest
```

## Troubleshooting

### Image Not Found

**Problem**: `r2r eac` command fails with "image not found"

**Solution**: Verify image exists locally:

```powershell
docker images ext-eac:dev
```

If missing, rebuild:

```powershell
.\scripts\pwsh\local-dev\build-local.ps1
```

### Still Pulling from Registry

**Problem**: Docker still tries to pull from ghcr.io

**Solution**: Check your local configuration:

```powershell
cat .r2r\r2r-cli.local.yml
```

Ensure it contains:

```yaml
load_local: true
extensions:
  - name: 'eac'
    load_local: true
    image_pull_policy: 'Never'
```

### Wrong Repository Root

**Problem**: Commands create files in wrong location

**Solution**: Set `R2R_REPO_ROOT` environment variable:

```powershell
$env:R2R_REPO_ROOT = (Get-Location)
```

Verify it's set correctly:

```powershell
echo $env:R2R_REPO_ROOT
```

### Configuration Not Applied

**Problem**: Local config changes not taking effect

**Solution**: Check file location and syntax:

```powershell
# Verify file exists
Test-Path .r2r\r2r-cli.local.yml

# Check for YAML syntax errors
cat .r2r\r2r-cli.local.yml
```

### Build Failures

**Problem**: Docker build fails

**Solutions**:

1. **Check Docker is running**:

   ```powershell
   docker version
   ```

2. **Check Dockerfile exists**:

   ```powershell
   Test-Path containers\ext-eac\Dockerfile
   ```

3. **Try legacy builder**:

   ```powershell
   .\scripts\pwsh\local-dev\build-local.ps1 -NoBuildKit
   ```

4. **Check repository root**:

   ```powershell
   git rev-parse --show-toplevel
   ```

## Best Practices

1. **Develop with importer.ps1 first**: Use `importer.ps1` in EAC repo for fast iteration
2. **Validate with Docker**: Test in external repos with Docker for realistic validation
3. **Use descriptive tags**: Tag images with feature names or versions
4. **Rebuild after changes**: Always rebuild the Docker image after code changes
5. **Keep configs gitignored**: Never commit `.r2r/*.local.yml` files
6. **Set environment variables**: Remember to set `R2R_REPO_ROOT` in each session
7. **Clean up old images**: Periodically remove unused Docker images

## Reference

### Script Locations

| Script         | Purpose                      | Path                                     |
| -------------- | ---------------------------- | ---------------------------------------- |
| Build Script   | Build local Docker images    | `scripts/pwsh/local-dev/build-local.ps1` |
| Init Script    | Create local config          | `scripts/pwsh/cli/init-local.ps1`        |
| Setup Script   | Complete setup orchestration | `scripts/pwsh/local-dev/setup.ps1`       |
| Install Script | Install r2r binary           | `scripts/pwsh/cli/install.ps1`           |

### Configuration Files

| File                     | Purpose            | Committed       |
| ------------------------ | ------------------ | --------------- |
| `.r2r/r2r-cli.yml`       | Base configuration | Yes             |
| `.r2r/r2r-cli.local.yml` | Local overrides    | No (gitignored) |

### Environment Variables

| Variable        | Purpose                    | Example                          |
| --------------- | -------------------------- | -------------------------------- |
| `R2R_REPO_ROOT` | Repository root path       | `C:\source\ready-to-release\eac` |
| `EAC_DEBUG`     | Enable debug logging       | `true`                           |
| `R2R_LOCAL_DEV` | Flag for local development | `true`                           |

## Related Guides

- [Creating Extensions](../r2r/creating-extensions.md) - Learn how to create r2r extensions
- [Testing in External Repositories](../r2r/testing-in-external-repos.md) - Test extensions in other projects

## See Also

- [R2R CLI init command](../../reference/r2r/commands/init.md) - Initialize R2R configuration
- [R2R CLI Configuration](../../reference/r2r/configuration/) - Configuration file reference
- [CLI vs Extensions Architecture](../../reference/devex/internal/cli-vs-extensions.md) - Understanding the two-tier system

{{ diataxis_footer() }}
