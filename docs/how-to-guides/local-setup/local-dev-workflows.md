# Local Development Workflows

{{ page_breadcrumb() }}

**Problem**: You want to develop and test EAC commands without pushing to remote registries during development.

**Solution**: Use the workflow that matches your goal:

- **EAC repository**: Use `importer.ps1` to load commands directly from source
- **External repository (quick test)**: Build the `eac` binary and run it directly — no Docker, no `clie`
- **External repository (production-like)**: Use Docker-based local development with locally built images

## Overview

### Three Development Workflows

#### 1. EAC Repository Development (importer.ps1)

When developing in the EAC repository itself, use `importer.ps1` to load commands directly:

```powershell
# In the EAC repository
.\importer.ps1

# Set commands path
$env:CLIE_COMMANDS_PATH = '.\go\cli\eac'

# Test your changes immediately
eac <command>
```

**Advantages:**

- Instant feedback on code changes
- No Docker build required
- Fastest iteration cycle
- Direct debugging

#### 2. External Repository Testing (Direct Binary)

When you just want to run `eac` commands against an external repository without `clie` or Docker,
build the binary once and run it directly:

```powershell
# In the EAC repository — build the binary once
go build -o out/eac.exe ./go/cli/eac

# Change into the target repository
cd C:\path\to\external-repo

# Run any eac command directly
C:\source\ready-to-release\eac\out\eac.exe init --scan
C:\source\ready-to-release\eac\out\eac.exe show modules
C:\source\ready-to-release\eac\out\eac.exe help
```

The binary auto-detects the repository root from the current directory (via git).

**Advantages:**

- No Docker required
- No `clie` setup required
- Works on any git repository immediately
- Fastest way to test `eac` commands on external repos

#### 3. External Repository Testing (Docker)

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

| Scenario                           | Use           | Reason                          |
| ---------------------------------- | ------------- | ------------------------------- |
| Writing new EAC commands           | importer.ps1  | Fastest iteration               |
| Debugging EAC code                 | importer.ps1  | Direct code execution           |
| Running `eac` on an external repo  | Direct binary | No Docker or clie needed        |
| Testing init scan on another repo  | Direct binary | Quickest path, just go build    |
| Validating container config        | Docker        | Production-like setup           |
| Testing multi-repo support via clie| Docker        | Integration testing             |

## EAC Repository Development

### Quick Start with importer.ps1

```powershell
# Navigate to EAC repository
cd C:\source\ready-to-release\eac

# Run importer to install/update clie binary
.\importer.ps1

# Set commands path
$env:CLIE_COMMANDS_PATH = '.\go\cli\eac'

# Test immediately
eac help
```

### Development Iteration

1. **Make code changes** in `go/cli/eac/`
2. **Test immediately**: `eac <command>`
3. **No Docker rebuild needed**

## External Repository Testing (Direct Binary)

This workflow runs `eac` directly on any external repository — no `clie`, no Docker.

### Prerequisites

- Go 1.21 or later
- EAC repository source code

### Quick Start

```powershell
# 1. Build the eac binary (run once from EAC repo root)
cd C:\source\ready-to-release\eac
go build -o out/eac.exe ./go/cli/eac

# 2. Change to your target repository
cd C:\path\to\your-repo

# 3. Run eac commands directly
C:\source\ready-to-release\eac\out\eac.exe init --scan
C:\source\ready-to-release\eac\out\eac.exe show modules
```

### Using a Personal AI Provider

If you want AI-enhanced generation (e.g. `init --scan` with `claude-cli`), copy your personal
AI provider config into the target repo before running:

```powershell
# Create .eac directory in the target repo
mkdir C:\path\to\your-repo\.eac

# Copy personal config from this workspace
Copy-Item C:\source\ready-to-release\eac\.eac\ai-provider.personal.yml `
          C:\path\to\your-repo\.eac\ai-provider.personal.yml

# Run init scan — eac picks up the personal config automatically
cd C:\path\to\your-repo
C:\source\ready-to-release\eac\out\eac.exe init --scan
```

The AI executor loads `.eac/ai-provider.personal.yml` automatically when it exists.
No `--ai-provider` flag is needed.

### Rebuilding After Code Changes

The binary must be rebuilt after modifying `go/cli/eac` source:

```powershell
cd C:\source\ready-to-release\eac
go build -o out/eac.exe ./go/cli/eac
```

### Tips

- The binary auto-detects the repo root via `git rev-parse --show-toplevel`
- Run the binary **from within** the target repository (not from the EAC repo)
- `out/eac.exe` is gitignored — safe to rebuild at any time

---

## External Repository Testing (Docker/clie)

This section covers testing your EAC extension in external repositories using Docker and `clie`.

### Components

The Docker-based workflow uses three scripts:

1. **Docker Image Builder** (`build-local.ps1`) - Builds local Docker images
2. **Local Config Generator** (`init-local.ps1`) - Creates clie configuration
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
2. Build the Docker image as `eac-ext:dev` in EAC repo
3. Install the clie binary (if not already installed)
4. Create `.eac/eac.local.yml` configuration in the external repo
5. Test the setup

## Step-by-Step Setup for External Repositories

This section shows how to manually set up an external repository step by step.

### 1. Build Local Docker Image

Build the EAC extension Docker image locally in the EAC repository:

```powershell
# Basic build (single platform, linux/amd64)
.\scripts\pwsh\local-dev\build-local.ps1

# With custom tag
.\scripts\pwsh\local-dev\build-local.ps1 -Tag "eac-ext:my-feature"

# Multi-platform build (amd64 and arm64)
.\scripts\pwsh\local-dev\build-local.ps1 -MultiPlatform
```

**Build Options:**

| Parameter        | Description                    | Default       |
| ---------------- | ------------------------------ | ------------- |
| `-Tag`           | Docker image tag               | `eac-ext:dev` |
| `-Platform`      | Target platform                | `linux/amd64` |
| `-MultiPlatform` | Build for both amd64 and arm64 | `false`       |
| `-Push`          | Push to registry after build   | `false`       |
| `-NoBuildKit`    | Disable Docker BuildKit        | `false`       |

**Verify the build:**

```powershell
docker images eac-ext:dev
```

### 2. Create Local Configuration

Create a local clie configuration that uses your Docker image:

```powershell
# Basic configuration
.\scripts\pwsh\cli\init-local.ps1

# With custom image tag
.\scripts\pwsh\cli\init-local.ps1 -ImageTag "eac-ext:my-feature"

# With workspace mounting for live development
.\scripts\pwsh\cli\init-local.ps1 -MountWorkspace

# Force overwrite existing config
.\scripts\pwsh\cli\init-local.ps1 -Force
```

**Configuration Options:**

| Parameter         | Description                           | Default       |
| ----------------- | ------------------------------------- | ------------- |
| `-ImageTag`       | Docker image tag to use               | `eac-ext:dev` |
| `-Force`          | Overwrite existing config             | `false`       |
| `-MountWorkspace` | Add volume mount for live development | `false`       |
| `-EnableDebug`    | Add debug environment variables       | `false`       |

This creates `.eac/eac.local.yml` with:

```yaml
# Local development configuration (gitignored)
load_local: true

extensions:
  - name: "eac"
    image: "eac-ext:dev"
    load_local: true
    image_pull_policy: "Never"
    auto_remove_children: false
```

### 3. Set Environment Variable

For each terminal session, set the repository root:

```powershell
$env:CLIE_REPO_ROOT = (Get-Location)
```

Or for a specific path:

```powershell
$env:CLIE_REPO_ROOT = "C:\path\to\your\repository"
```

### 4. Test Your Setup

```powershell
# Test basic functionality
eac help

# Test command execution
eac show modules
```

## Configuration Details

### Local Override Files

clie uses a configuration hierarchy:

1. `.eac/eac.yml` - Base configuration (committed to git)
2. `.eac/eac.local.yml` - Local overrides (gitignored)

Local files override base configuration for development.

### Key Configuration Fields

For local development, these fields are critical:

```yaml
load_local: true # Skip registry pulls (global)

extensions:
  - name: "eac"
    image: "eac-ext:dev" # Local Docker image tag
    load_local: true # Skip registry pulls (per-extension)
    image_pull_policy: "Never" # Fail if image not found locally
    auto_remove_children: false # Keep containers for debugging
```

### Debug Configuration

Enable debug mode for additional logging:

```powershell
.\scripts\pwsh\cli\init-local.ps1 -EnableDebug
```

This adds debug environment variables:

```yaml
extensions:
  - name: "eac"
    # ... other config ...
    env:
      - name: "EAC_DEBUG"
        value: "true"
      - name: "CLIE_LOCAL_DEV"
        value: "true"
```

### Workspace Mounting

Mount your workspace for live code changes (advanced):

```powershell
.\scripts\pwsh\cli\init-local.ps1 -MountWorkspace
```

This adds volume mounts:

```yaml
extensions:
  - name: "eac"
    # ... other config ...
    volumes:
      - host: "."
        container: "/workspace"
        readonly: false
      - host: ".git"
        container: "/.git"
        readonly: true
```

## Development Workflow

### Iteration Cycle

When developing and testing in external repositories:

1. **Modify code in EAC repository**:

   ```powershell
   cd C:\source\ready-to-release\eac
   # Edit files in go/cli/eac/
   ```

2. **Rebuild Docker image**:

   ```powershell
   .\scripts\pwsh\local-dev\build-local.ps1
   ```

3. **Test in external repository**:

   ```powershell
   cd C:\path\to\external-repo
   $env:CLIE_REPO_ROOT = (Get-Location)
   eac <command>
   ```

**Note**: For faster iteration when developing commands, use `importer.ps1` in the EAC repository instead of the Docker workflow.

### Comparison: All Workflows

| Workflow                       | Requires     | Test Changes          | Iteration Time |
| ------------------------------ | ------------ | --------------------- | -------------- |
| importer.ps1 (EAC repo)        | Go           | `eac <command>`       | Seconds        |
| Direct binary (external repo)  | Go           | `go build` + run      | Seconds        |
| Docker/clie (external repo)    | Docker+clie  | Rebuild image         | Minutes        |

**Recommendation**: Use `importer.ps1` when developing EAC itself, the direct binary when
running `eac` against an external repo, and Docker when you need to validate the full
`clie eac` integration.

## Advanced Usage

### Custom Image Tags

Use meaningful tags for feature branches:

```powershell
# Build with feature tag
.\scripts\pwsh\local-dev\build-local.ps1 -Tag "eac-ext:feature-xyz"

# Configure to use feature tag
.\scripts\pwsh\cli\init-local.ps1 -ImageTag "eac-ext:feature-xyz"
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

# Skip clie installation (already installed)
.\scripts\pwsh\setup-local-dev.ps1 -SkipInstall

# Skip validation tests
.\scripts\pwsh\setup-local-dev.ps1 -SkipTest
```

## Troubleshooting

### Image Not Found

**Problem**: `eac` command fails with "image not found"

**Solution**: Verify image exists locally:

```powershell
docker images eac-ext:dev
```

If missing, rebuild:

```powershell
.\scripts\pwsh\local-dev\build-local.ps1
```

### Still Pulling from Registry

**Problem**: Docker still tries to pull from ghcr.io

**Solution**: Check your local configuration:

```powershell
cat .eac\eac.local.yml
```

Ensure it contains:

```yaml
load_local: true
extensions:
  - name: "eac"
    load_local: true
    image_pull_policy: "Never"
```

### Wrong Repository Root

**Problem**: Commands create files in wrong location

**Solution**: Set `CLIE_REPO_ROOT` environment variable:

```powershell
$env:CLIE_REPO_ROOT = (Get-Location)
```

Verify it's set correctly:

```powershell
echo $env:CLIE_REPO_ROOT
```

### Configuration Not Applied

**Problem**: Local config changes not taking effect

**Solution**: Check file location and syntax:

```powershell
# Verify file exists
Test-Path .eac\eac.local.yml

# Check for YAML syntax errors
cat .eac\eac.local.yml
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
   Test-Path containers\eac-ext\Dockerfile
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
5. **Keep configs gitignored**: Never commit `.eac/*.local.yml` files
6. **Set environment variables**: Remember to set `CLIE_REPO_ROOT` in each session
7. **Clean up old images**: Periodically remove unused Docker images

## Reference

### Script Locations

| Script         | Purpose                      | Path                                     |
| -------------- | ---------------------------- | ---------------------------------------- |
| Build Script   | Build local Docker images    | `scripts/pwsh/local-dev/build-local.ps1` |
| Init Script    | Create local config          | `scripts/pwsh/cli/init-local.ps1`        |
| Setup Script   | Complete setup orchestration | `scripts/pwsh/local-dev/setup.ps1`       |
| Install Script | Install clie binary          | `scripts/pwsh/cli/install.ps1`           |

### Configuration Files

| File                   | Purpose            | Committed       |
| ---------------------- | ------------------ | --------------- |
| `.eac/eac.yml`       | Base configuration | Yes             |
| `.eac/eac.local.yml` | Local overrides    | No (gitignored) |

### Environment Variables

| Variable         | Purpose                    | Example                          |
| ---------------- | -------------------------- | -------------------------------- |
| `CLIE_REPO_ROOT` | Repository root path       | `C:\source\ready-to-release\eac` |
| `EAC_DEBUG`      | Enable debug logging       | `true`                           |
| `CLIE_LOCAL_DEV` | Flag for local development | `true`                           |

## Related Guides

- [Creating Extensions](../clie/creating-extensions.md) - Learn how to create clie extensions
- [Testing in External Repositories](../clie/testing-in-external-repos.md) - Test extensions in other projects

## See Also

- [CLIE CLI init command](../../reference/clie/commands/init.md) - Initialize CLIE configuration
- [CLIE CLI Configuration](../../reference/clie/configuration.md) - Configuration file reference
- [CLI vs Extensions Architecture](../../reference/eac/architecture/cli-integration.md) - Understanding the two-tier system

{{ diataxis_footer() }}
