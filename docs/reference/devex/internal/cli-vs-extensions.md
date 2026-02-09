# CLIE CLI vs Extensions: Understanding the Architecture

This guide explains the two-tier architecture of CLIE:

1. The core CLI framework
2. The containerized extensions.

test

## Overview

CLIE uses a two-tier command structure that separates framework management from automation tools:

```mermaid
graph TB
    subgraph tier1["Tier 1: CLIE CLI (Core)"]
        A[Extension Management Framework<br/>Commands: init, install, list, etc.]
    end

    subgraph tier2["Tier 2: Extensions (Tools)"]
        B[Containerized Automation Tools<br/>Example: EAC]
    end

    tier1 -->|Manages and Executes| tier2

    style tier1 fill:#e1f5ff,stroke:#0066cc,stroke-width:2px
    style tier2 fill:#fff4e6,stroke:#ff9900,stroke-width:2px
```

## Tier 1: CLIE CLI (Core Framework)

### Purpose

The CLIE CLI is the **framework** that manages containerized extensions. It handles:

- Extension installation and lifecycle
- Docker container orchestration
- Configuration management
- Registry interaction

### Commands

Framework commands that manage extensions:

| Command        | Purpose                     |
| -------------- | --------------------------- |
| `clie init`     | Initialize configuration    |
| `clie install`  | Install extensions          |
| `clie list`     | Browse available extensions |
| `clie validate` | Validate configuration      |
| `clie cleanup`  | Clean up old images         |
| `clie verify`   | Check system prerequisites  |

### Where It Runs

- **Host machine** (your computer/CI environment)
- **Direct binary execution** (not containerized)
- **Single installation** per system

### Configuration

- **File**: `.clie/clie.yml`
- **Purpose**: Extension registry and runtime settings
- **Scope**: Framework-level configuration

## Tier 2: Extensions (Automation Tools)

### Purpose

Extensions are **containerized tools** that provide automation capabilities. Each extension:

- Runs inside a Docker container
- Provides specific automation commands
- Has isolated dependencies
- Uses consistent environments

### Examples

| Extension          | Commands                    | Purpose                       |
| ------------------ | --------------------------- | ----------------------------- |
| **eac**            | build, test, validate, scan | Everything-as-Code automation |
| **pwsh**           | script execution            | PowerShell automation         |
| **python**         | script execution            | Python automation             |
| **docker-builder** | multi-platform builds       | Docker utilities              |

### Where They Run

- **Inside Docker containers**
- **Isolated from host**
- **Reproducible environments**

### Configuration

- **Directory**: `.clie/<extension>/`
- **Purpose**: Extension-specific settings
- **Scope**: Per-extension configuration

Example: `.eac/` contains EAC-specific config like `ai-provider.yml`.

## The Relationship

### How They Work Together

1. **CLIE CLI** manages extensions:

   ```bash
   clie init          # CLI creates config
   clie install eac   # CLI pulls Docker image
   clie list          # CLI queries registry
   ```

2. **Extensions** provide tools:

   ```bash
   eac build     # Extension runs in container
   eac test      # Extension runs in container
   clie pwsh script   # Extension runs in container
   ```

### Command Execution Flow

```text
User runs: eac build

   ↓
[CLIE CLI]
  ├─ Reads .clie/clie.yml
  ├─ Finds 'eac' extension config
  ├─ Pulls image if needed
  ├─ Mounts workspace volume
  └─ Starts Docker container
       ↓
   [EAC Extension Container]
     ├─ Runs 'eac build' inside container
     ├─ Has access to /workspace (your project)
     ├─ Reads .eac/ configuration
     └─ Executes build and returns results
          ↓
   [CLIE CLI]
     └─ Displays output to user
```

## Key Differences

| Aspect            | CLIE CLI (Tier 1)       | Extensions (Tier 2)       |
| ----------------- | ---------------------- | ------------------------- |
| **Purpose**       | Extension management   | Automation tools          |
| **Installation**  | System-wide binary     | Per-project Docker images |
| **Runs**          | On host machine        | Inside containers         |
| **Updates**       | `clie update`           | Automatic via registry    |
| **Configuration** | `.clie/clie.yml`     | `.clie/<extension>/`       |
| **Commands**      | Framework operations   | Tool-specific operations  |
| **Examples**      | init, install, cleanup | build, test, scan         |

## Configuration Layering

### Two-Level Configuration

```text
Project Root/
├── .clie/
│   ├── clie.yml              ← Tier 1: Framework config
│   │                                (Which extensions to use)
│   └── eac/                     ← Tier 2: Extension config
│       ├── ai-provider.yml          (How EAC behaves)
│       └── repository.yml
```

### Configuration Purposes

**Tier 1** (`.clie/clie.yml`):

```yaml
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/eac-ext:latest'
  - name: 'pwsh'
    image: 'ghcr.io/ready-to-release/ext-pwsh:latest'
```

→ Tells CLIE CLI which extensions are available

**Tier 2** (`.eac/ai-provider.yml`):

```yaml
ai:
  provider: claude-api
  model: claude-sonnet-4-5
```

→ Tells EAC extension which AI provider to use

## Why This Architecture?

### Benefits

1. **Isolation**: Extensions don't interfere with each other
2. **Reproducibility**: Same container = same environment everywhere
3. **Versioning**: Lock extensions to specific versions
4. **Portability**: Works on any system with Docker
5. **Security**: Containers provide sandboxing
6. **Flexibility**: Mix and match extensions as needed

### Tradeoffs

- **Requires Docker**: Must have Docker installed
- **Image storage**: Extensions consume disk space
- **Startup overhead**: Container startup adds milliseconds

## Practical Examples

### Example 1: New Project Setup

```bash
# Tier 1: Set up framework
clie init                    # Create extension config
clie install eac             # Install EAC extension
clie install pwsh            # Install PowerShell extension

# Tier 2: Configure extensions
eac init --ai-provider claude-api
clie pwsh configure

# Tier 2: Use extensions
eac build
eac test
clie pwsh run-script.ps1
```

### Example 2: Development vs CI

**Development** (`.clie/clie.local.yml`):

```yaml
extensions:
  - name: 'eac'
    image: 'eac-ext:dev'          # Local dev image
    pull_policy: 'Never'          # Don't pull
```

**CI** (`.clie/clie.yml`):

```yaml
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/eac-ext:latest'
    pull_policy: 'Always'         # Always get latest
```

### Example 3: Multiple Extension Versions

```yaml
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/eac-ext:v1.2.3'

  - name: 'eac-beta'
    image: 'ghcr.io/ready-to-release/eac-ext:v2.0.0-beta'
```

Use different versions side by side:

```bash
eac build          # Uses v1.2.3
eac-beta build     # Uses v2.0.0-beta
```

## Common Workflows

### Installing CLIE for the First Time

```bash
# Step 1: Install CLIE CLI binary (Tier 1)
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash

# Step 2: Initialize framework (Tier 1)
clie init

# Step 3: Install extensions (Tier 1)
clie install eac

# Step 4: Configure extensions (Tier 2)
eac init --ai-provider claude-api

# Step 5: Use extensions (Tier 2)
eac build
```

### Daily Development

```bash
# Framework commands (as needed)
clie cleanup            # Free disk space
clie verify             # Check system health

# Extension commands (frequently)
eac build
eac test
eac validate
clie pwsh run-tests.ps1
```

### Troubleshooting

```bash
# Framework diagnostics (Tier 1)
clie version            # Check CLI version
clie verify             # Verify prerequisites
clie validate           # Check configuration
clie list               # See available extensions

# Extension diagnostics (Tier 2)
clie interactive eac    # Open container shell
clie metadata eac       # Get extension info
eac --help         # Extension help
```

## Summary

### Quick Reference

**Use CLIE CLI commands when:**

- ✅ Setting up a new project
- ✅ Installing or managing extensions
- ✅ Maintaining Docker images
- ✅ Verifying system health
- ✅ Configuring extension registry

**Use Extension commands when:**

- ✅ Building, testing, or validating code
- ✅ Running automation workflows
- ✅ Executing project-specific tasks
- ✅ Working with project files
- ✅ Performing daily development tasks

### Mental Model

Think of it like package managers:

- **CLIE CLI** = `apt` / `brew` / `npm` (installs packages)
- **Extensions** = packages themselves (provide functionality)

```bash
# Package manager layer
npm install webpack        # Install package

# Package usage layer
webpack build             # Use package
```

Similarly:

```bash
# Framework layer
clie install eac           # Install extension

# Extension usage layer
eac build             # Use extension
```

## See Also

- [CLIE CLI Overview](../../clie/index.md) - Framework commands
- [EAC Commands Reference](../../eac/commands/index.md) - Extension commands
- [Configuration Guide](../../clie/configuration/index.md) - Detailed configuration
- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md) - Getting started
