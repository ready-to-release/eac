# Testing in External Repositories

{{ page_breadcrumb() }}

**Problem**: You want to test your r2r extension in a different repository to ensure it works correctly outside of the EAC development environment.

**Solution**: Build a local Docker image in the EAC repository, then configure external repositories to use it.

## Overview

### Why Test in External Repositories?

Testing in external repositories ensures your extension:

- Works with real-world repository structures
- Handles different configuration scenarios
- Functions correctly outside the development environment
- Integrates properly with existing workflows
- Validates Dockerfile and container configuration

### Development vs Testing Workflow

| Location | Workflow | Purpose |
|----------|----------|---------|
| **EAC Repository** | `importer.ps1` | Fast development and debugging |
| **External Repositories** | Docker (this guide) | Realistic integration testing |

**Workflow Summary:**

1. Develop and debug in EAC repository using `importer.ps1`
2. Build Docker image when ready to test
3. Set up external repositories to use the local Docker image
4. Test in realistic environments

## Prerequisites

Before testing in external repositories:

1. **Complete development in EAC repository** using `importer.ps1`:

   ```powershell
   cd C:\source\ready-to-release\eac
   .\scripts\pwsh\importer.ps1
   ```

2. **Build local Docker image** when ready for external testing:

   ```powershell
   cd C:\source\ready-to-release\eac
   .\scripts\pwsh\local-dev\build-local.ps1
   ```

3. **Ensure r2r binary is installed** (importer.ps1 handles this automatically)

4. **Have a target external repository** (with git initialized)

## Quick Start

Set up an external repository in one command:

```powershell
# Navigate to the external repository
cd C:\path\to\external-repo

# Run setup from EAC repository
C:\source\ready-to-release\eac\scripts\pwsh\local-dev\setup.ps1 -TargetRepo .
```

This configures the external repository to use your local Docker image.

## Rebuilding Docker Image After Code Changes

After making code changes in the EAC repository, rebuild and test:

```powershell
# 1. Navigate to EAC repository
cd C:\source\ready-to-release\eac

# 2. Rebuild Docker image
.\scripts\pwsh\local-dev\build-local.ps1

# 3. Test in external repo (no reconfiguration needed)
cd C:\path\to\external-repo
$env:R2R_REPO_ROOT = (Get-Location)
r2r eac <command>
```

**That's it!** External repositories automatically use the updated image.

## Step-by-Step Setup

### 1. Navigate to Target Repository

```powershell
cd C:\path\to\external-repo
```

**Requirements:**

- Directory must be a git repository
- If not, initialize git first:

  ```powershell
  git init
  ```

### 2. Run Setup Script

From the target repository, run the setup script from the EAC repository:

```powershell
C:\source\ready-to-release\eac\scripts\pwsh\local-dev\setup.ps1 -TargetRepo .
```

**Options:**

| Parameter | Description | Default |
|-----------|-------------|---------|
| `-TargetRepo` | Path to target repository | Current directory |
| `-SkipBuild` | Skip Docker build (use existing) | `false` |
| `-SkipInstall` | Skip r2r installation | `false` |
| `-SkipTest` | Skip validation tests | `false` |
| `-ImageTag` | Docker image tag to use | `ext-eac:dev` |
| `-Force` | Overwrite existing config | `false` |

### 3. Set Environment Variable

For each terminal session in the external repository:

```powershell
# Set to current directory
$env:R2R_REPO_ROOT = (Get-Location)

# Or set explicitly
$env:R2R_REPO_ROOT = "C:\path\to\external-repo"
```

**Important**: This must be set in every new terminal session.

### 4. Initialize EAC Configuration

Initialize the EAC configuration in the external repository:

```powershell
r2r eac init
```

Follow the prompts to configure:

- AI provider (Claude API, OpenAI, or Gemini)
- API tokens
- Git integration

### 5. Test Commands

Run EAC commands to verify the setup:

```powershell
# View available commands
r2r eac help

# Show module information
r2r eac show modules

# List module groups
r2r eac show module-groups
```

## Setup Verification

### Check Configuration Files

Verify the configuration files were created:

```powershell
# Check r2r local config
cat .r2r\r2r-cli.local.yml

# Check EAC config (after r2r eac init)
cat .r2r\eac\agent.yml
```

### Verify Docker Image

Confirm the local Docker image is available:

```powershell
docker images ext-eac:dev
```

### Test Command Execution

Run a simple command to ensure everything works:

```powershell
r2r eac show modules
```

## Testing Workflow

### 1. Test Basic Operations

```powershell
# Initialize repository analysis
r2r eac init

# Analyze modules
r2r eac analyze modules

# View results
r2r eac show modules
```

### 2. Test AI Integration

```powershell
# Create specification with AI
r2r eac create spec <module-name>

# Generate changelog
r2r eac create changelog <module-name>
```

### 3. Test Release Management

```powershell
# Show release information
r2r eac show releases

# Check module releases
r2r eac show module-releases
```

## Multiple External Repositories

### Using the Same Local Image

Configure multiple repositories to use the same local image:

```powershell
# Repository 1
cd C:\repos\project-a
C:\source\ready-to-release\eac\scripts\pwsh\local-dev\setup.ps1 -TargetRepo . -SkipBuild

# Repository 2
cd C:\repos\project-b
C:\source\ready-to-release\eac\scripts\pwsh\local-dev\setup.ps1 -TargetRepo . -SkipBuild

# Repository 3
cd C:\repos\project-c
C:\source\ready-to-release\eac\scripts\pwsh\local-dev\setup.ps1 -TargetRepo . -SkipBuild
```

**Note**: Use `-SkipBuild` to avoid rebuilding the Docker image for each repository.

### Managing Configurations

Each repository maintains its own configuration:

```
project-a/
  .r2r/
    r2r-cli.local.yml      # Points to ext-eac:dev
    eac/
      agent.yml             # Project A specific config

project-b/
  .r2r/
    r2r-cli.local.yml      # Points to ext-eac:dev
    eac/
      agent.yml             # Project B specific config
```

## Updating the Docker Image

### After Code Changes

When you make changes to the EAC extension:

1. **Develop with importer.ps1 first** for fast iteration:

   ```powershell
   cd C:\source\ready-to-release\eac

   # Make code changes, then reload
   r2r load-commands
   r2r eac <command>
   ```

2. **Rebuild the Docker image** when ready to test in external repos:

   ```powershell
   cd C:\source\ready-to-release\eac
   .\scripts\pwsh\local-dev\build-local.ps1
   ```

3. **Test in external repository** (no reconfiguration needed):

   ```powershell
   cd C:\path\to\external-repo
   $env:R2R_REPO_ROOT = (Get-Location)
   r2r eac <command>
   ```

The external repository automatically uses the updated image.

### Recommended Iteration Pattern

```
EAC Repo (importer.ps1)          External Repo (Docker)
┌──────────────────┐              ┌──────────────────┐
│ 1. Edit code     │              │                  │
│ 2. r2r load-cmds │              │                  │
│ 3. Test quickly  │              │                  │
│ 4. Iterate...    │              │                  │
│                  │   ──────>    │ 5. Build image   │
│                  │              │ 6. Test in repo  │
│                  │              │ 7. Validate      │
└──────────────────┘              └──────────────────┘
   Fast (seconds)                    Thorough (minutes)
```

### Testing Different Versions

Test different image versions by using tags:

```powershell
# Build image with version tag
cd C:\source\ready-to-release\eac
.\scripts\pwsh\local-dev\build-local.ps1 -Tag "ext-eac:test-v2"

# Configure external repo to use it
cd C:\path\to\external-repo
.\scripts\pwsh\cli\init-local.ps1 -ImageTag "ext-eac:test-v2" -Force
```

## Troubleshooting

### Repository Not Found

**Problem**: Setup script reports "Target repository not found"

**Solution**: Verify the path exists and is a git repository:

```powershell
# Check path exists
Test-Path C:\path\to\external-repo

# Check git repository
cd C:\path\to\external-repo
git rev-parse --show-toplevel

# Initialize if needed
git init
```

### Commands Create Files in Wrong Location

**Problem**: EAC commands create files in the wrong directory

**Solution**: Check and set `R2R_REPO_ROOT` environment variable:

```powershell
# Check current value
echo $env:R2R_REPO_ROOT

# Set to external repository
cd C:\path\to\external-repo
$env:R2R_REPO_ROOT = (Get-Location)
```

### Image Not Found

**Problem**: r2r reports "image not found" in external repository

**Solutions**:

1. **Verify image exists**:

   ```powershell
   docker images ext-eac:dev
   ```

2. **Check local config**:

   ```powershell
   cat .r2r\r2r-cli.local.yml
   ```

   Should contain:

   ```yaml
   load_local: true
   extensions:
     - name: 'eac'
       image: 'ext-eac:dev'
       image_pull_policy: 'Never'
   ```

3. **Rebuild if missing**:

   ```powershell
   cd C:\source\ready-to-release\eac
   .\scripts\pwsh\local-dev\build-local.ps1
   ```

### Configuration Not Applied

**Problem**: Changes to `.r2r/r2r-cli.local.yml` not taking effect

**Solutions**:

1. **Verify file location**:

   ```powershell
   # Should be in repository root
   Test-Path .r2r\r2r-cli.local.yml
   ```

2. **Check YAML syntax**:

   ```powershell
   cat .r2r\r2r-cli.local.yml
   ```

3. **Recreate configuration**:

   ```powershell
   C:\source\ready-to-release\eac\scripts\pwsh\cli\init-local.ps1 -Force
   ```

### Init Command Fails

**Problem**: `r2r eac init` fails in external repository

**Solutions**:

1. **Set repository root**:

   ```powershell
   $env:R2R_REPO_ROOT = (Get-Location)
   ```

2. **Verify Docker image**:

   ```powershell
   docker images ext-eac:dev
   ```

3. **Check local config exists**:

   ```powershell
   Test-Path .r2r\r2r-cli.local.yml
   ```

4. **Run init again**:

   ```powershell
   r2r eac init
   ```

### Different Behavior Than EAC Repo

**Problem**: Commands behave differently in external repo vs EAC repo

**Causes**:

1. **Different repository structures** - Expected behavior
2. **Different configurations** - Check `.r2r/eac/agent.yml`
3. **Environment variables** - Verify `R2R_REPO_ROOT` is set
4. **Docker image version** - Ensure using same image

**Solutions**:

```powershell
# Compare configurations
diff (cat C:\source\ready-to-release\eac\.r2r\r2r-cli.local.yml) `
     (cat C:\path\to\external-repo\.r2r\r2r-cli.local.yml)

# Verify environment
echo "EAC: $(cd C:\source\ready-to-release\eac; $env:R2R_REPO_ROOT)"
echo "External: $(cd C:\path\to\external-repo; $env:R2R_REPO_ROOT)"

# Check Docker image
docker images ext-eac:dev
```

## Best Practices

1. **Develop with importer.ps1 first**: Use `importer.ps1` in EAC repo for fast development
2. **Build Docker for validation**: Only build Docker when ready for external testing
3. **One setup per repository**: Run setup once per external repository
4. **Set environment variable**: Always set `R2R_REPO_ROOT` in new terminal sessions
5. **Keep configs local**: Don't commit `.r2r/*.local.yml` to git
6. **Test incrementally**: Test one command at a time to isolate issues
7. **Use meaningful names**: Name external test repositories clearly
8. **Document test cases**: Keep notes on what you're testing in each repository
9. **Clean up**: Remove test configurations when done

## Environment Variable Management

### PowerShell Profile

Add to your PowerShell profile for automatic setup:

```powershell
# Edit profile
notepad $PROFILE

# Add function
function Set-R2RRepo {
    $env:R2R_REPO_ROOT = (Get-Location).Path
    Write-Host "R2R_REPO_ROOT set to: $env:R2R_REPO_ROOT" -ForegroundColor Green
}

# Use in any repository
cd C:\path\to\external-repo
Set-R2RRepo
```

### Session Management

Create a setup script for each external repository:

```powershell
# C:\repos\project-a\setup-r2r.ps1
$env:R2R_REPO_ROOT = "C:\repos\project-a"
Write-Host "Ready to use r2r commands in project-a" -ForegroundColor Green
```

Usage:

```powershell
cd C:\repos\project-a
.\setup-r2r.ps1
r2r eac <command>
```

## Reference

### Setup Command

```powershell
C:\source\ready-to-release\eac\scripts\pwsh\setup-local-dev.ps1 `
    -TargetRepo <path> `
    [-SkipBuild] `
    [-SkipInstall] `
    [-SkipTest] `
    [-ImageTag <tag>] `
    [-Force]
```

### Required Files

After setup, external repository should have:

```
external-repo/
  .git/                           # Git repository
  .r2r/
    r2r-cli.local.yml            # R2R local config (gitignored)
    eac/
      agent.yml                   # EAC config (after r2r eac init)
  .gitignore                      # Updated with .r2r/*.local.yml
```

### Environment Variables

| Variable | Required | Purpose |
|----------|----------|---------|
| `R2R_REPO_ROOT` | Yes | Points to repository root |

Set in each session:

```powershell
$env:R2R_REPO_ROOT = (Get-Location)
```

## Related Guides

- [Local Development Workflows](./local-development.md) - Learn about importer.ps1 and Docker workflows
- [Creating Extensions](./creating-extensions.md) - Learn how r2r extensions work

{{ diataxis_footer() }}
