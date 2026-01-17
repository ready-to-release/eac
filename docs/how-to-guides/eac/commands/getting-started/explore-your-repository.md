# Explore Your Repository

## What You'll Accomplish

Discover what modules, files, and dependencies exist in your repository using show commands.

## Prerequisites

### Required Knowledge

**New to r2r?** Learn these concepts first:

- [Quick Start Guide](../../../../tutorials/getting-started/quick-start.md) - Understand r2r basics and repository structure

### Required Setup

- Working in an EAC-managed repository
- Repository has module contracts defined

## Steps

### 1. List All Modules

```bash
r2r eac show modules
```

**What happens**: Displays table of all modules with their types, paths, and file counts

### 2. View Module Dependencies

```bash
r2r eac show dependencies
```

**What happens**: Shows dependency graph with which modules depend on others

### 3. See File Ownership

```bash
r2r eac show files
```

**What happens**: Lists all files and which module owns each file

### 4. Check Module Types

```bash
r2r eac show moduletypes
```

**What happens**: Groups modules by type (go-cli, go-library, mkdocs-site, etc.)

## Example Scenario

You're new to the codebase and want to understand its structure:

```bash
# See what modules exist
r2r eac show modules

# Find dependencies for a specific module
r2r eac show dependencies | grep src-auth

# See what files belong to src-auth
r2r eac show files | grep src-auth

# Get machine-readable output for scripting
r2r eac get modules | jq '.modules[] | select(.moniker == "src-auth")'
```

## Common Issues

| Problem          | Solution                                               |
| ---------------- | ------------------------------------------------------ |
| Empty output     | Ensure you're in repository root with module contracts |
| Module not shown | Check module.yml file exists and is valid              |

## Next Steps

- [Create Feature Workspace](../development-workflow/create-feature-workspace.md) → Start working on feature
- [Build Single Module](../build-test-validate/build-single-module.md) → Build a module

## Related Commands

- [`show modules`](../../../../reference/commands/show/modules.md) - Display all modules
- [`show dependencies`](../../../../reference/commands/show/dependencies.md) - Show dependency graph
- [`show files`](../../../../reference/commands/show/files.md) - Show file ownership
- [`get modules`](../../../../reference/commands/get/modules.md) - Get modules as JSON
