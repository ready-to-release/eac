# R2R CLI Configuration

Complete reference for configuring R2R via `.r2r/r2r-cli.yml`.

## Overview

The R2R CLI uses `.r2r/r2r-cli.yml` to define the EAC extension, registry settings, and runtime behavior.

**Key concepts:**

- **Extensions** - Containerized tools (EAC)
- **Registry** - Container image fetch control
- **Defaults** - Fallback values for extensions
- **Environment** - Variables passed into containers

## Configuration File Location

```text
<repository-root>/.r2r/r2r-cli.yml
```

The CLI discovers the repository root by walking up directories until it finds `.git`.

### Creating the Configuration

```bash
# Initialize configuration
r2r init

# Result: Creates .r2r/r2r-cli.yml
```

## Minimal Configuration

Simplest working configuration:

```yaml
extensions:
  - name: "eac"
    image: "ghcr.io/ready-to-release/ext-eac:latest"
```

Invokable as:

```bash
r2r eac [command] [args...]
```

## Configuration Hierarchy

Values resolved in priority order (highest to lowest):

| Priority   | Source                | Description                             |
| ---------- | --------------------- | --------------------------------------- |
| 1          | Environment Variables | Runtime overrides via `R2R_*` variables |
| 2          | User-Specific Config  | `.r2r/r2r-cli.local.yml` (gitignored)   |
| 3          | Repository Config     | `.r2r/r2r-cli.yml` (committed)          |
| 4          | Built-in Defaults     | Programmed into CLI                     |

### Configuration Layering

```text
.r2r/
├── r2r-cli.yml           # Team configuration (committed)
└── r2r-cli.local.yml     # Personal overrides (gitignored)
```

**Team config** (`.r2r/r2r-cli.yml`):

```yaml
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/ext-eac:latest'
```

**Personal overrides** (`.r2r/r2r-cli.local.yml`):

```yaml
extensions:
  - name: 'eac'
    load_local: true
    image: 'ext-eac:dev'  # Use local development image
```

## Schema Reference

### Extensions Configuration

Define the EAC extension:

```yaml
extensions:
  - name: "eac"
    image: "ghcr.io/ready-to-release/ext-eac:latest"
    description: "Everything-as-Code automation"
    pull_policy: "IfNotPresent"
    timeout: 3600
    environment:
      - name: "EAC_DEBUG"
        value: "true"
```

#### Required Fields

| Field   | Type   | Description                             |
| ------- | ------ | --------------------------------------- |
| `name`  | string | Extension name (alphanumeric + hyphens) |
| `image` | string | Docker image reference                  |

#### Optional Fields

| Field          | Type    | Default       | Description                 |
| -------------- | ------- | ------------- | --------------------------- |
| `description`  | string  | -             | Human-readable description  |
| `pull_policy`  | enum    | From defaults | When to pull image          |
| `remove_after` | boolean | From defaults | Remove container after run  |
| `timeout`      | integer | From defaults | Execution timeout (seconds) |
| `memory_limit` | string  | From defaults | Memory limit                |
| `cpu_limit`    | string  | From defaults | CPU limit                   |
| `environment`  | array   | `[]`          | Extension-specific env vars |
| `volumes`      | array   | `[]`          | Additional volume mounts    |
| `working_dir`  | string  | `/workspace`  | Container working directory |

#### Image Reference Format

Valid formats:

```yaml
# Registry with organization and tag
image: "ghcr.io/ready-to-release/ext-eac:v1.2.3"

# Registry with organization, latest tag
image: "ghcr.io/ready-to-release/ext-eac:latest"

# Local development image
image: "ext-eac:dev"
```

### Environment Variables

Global and per-extension environment configuration:

```yaml
environment:
  global:
    - name: "TERM"
      value: "xterm-256color"
    - name: "CI"
      value: "true"
  secrets:
    - name: "GITHUB_TOKEN"
      env: "GH_TOKEN"
```

#### Global Variables

Passed to every extension:

```yaml
environment:
  global:
    - name: "LOG_LEVEL"
      value: "debug"
```

#### Secrets

Map container variables to host environment variables:

```yaml
environment:
  secrets:
    - name: "API_KEY"        # Variable name in container
      env: "HOST_API_KEY"    # Environment variable on host
```

Host environment:

```bash
export HOST_API_KEY="sk-..."
```

Container receives:

```bash
API_KEY="sk-..."
```

## Examples

### Single Extension (Minimal)

```yaml
extensions:
  - name: "eac"
    image: "ghcr.io/ready-to-release/ext-eac:latest"
```

### With Environment Variables

```yaml
extensions:
  - name: "eac"
    image: "ghcr.io/ready-to-release/ext-eac:latest"
    description: "Everything-as-Code automation"
    environment:
      - name: "EAC_DEBUG"
        value: "true"
```

### Development Configuration

Local development with debug enabled:

```yaml
defaults:
  pull_policy: "Never"  # Don't pull, use local images
  remove_after: false   # Keep containers for debugging

extensions:
  - name: "eac"
    image: "ext-eac:dev"  # Local development image
    environment:
      - name: "EAC_DEBUG"
        value: "true"
      - name: "R2R_LOCAL_DEV"
        value: "true"
```

### CI/CD Configuration

Optimized for continuous integration:

```yaml
defaults:
  pull_policy: "Always"    # Always get latest
  remove_after: true       # Clean up after each run
  timeout: 1800            # 30-minute timeout

environment:
  global:
    - name: "CI"
      value: "true"
  secrets:
    - name: "GITHUB_TOKEN"
      env: "GH_TOKEN"

extensions:
  - name: "eac"
    image: "ghcr.io/ready-to-release/ext-eac:latest"
```

## User-Specific Overrides

### Local Development File

Create `.r2r/r2r-cli.local.yml` (gitignored) for personal settings:

**Team config** (`.r2r/r2r-cli.yml`):

```yaml
extensions:
  - name: "eac"
    image: "ghcr.io/ready-to-release/ext-eac:latest"
```

**Your overrides** (`.r2r/r2r-cli.local.yml`):

```yaml
extensions:
  - name: "eac"
    pull_policy: "Never"
    image: "ext-eac:dev"
    environment:
      - name: "EAC_DEBUG"
        value: "true"
```

### Recommended .gitignore

```gitignore
# R2R user-specific configuration
.r2r/*.local.yml
.r2r/*.personal.yml

# R2R temporary files
.r2r/.cache/
.r2r/tmp/
```

## See Also

- [R2R CLI Overview](index.md) - Command overview
- [init command](init.md) - Initialize configuration
- [install command](install.md) - Install extensions
- [validate command](validate.md) - Validate configuration
- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md) - Getting started guide
