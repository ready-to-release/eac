# CLIE CLI Configuration

Complete reference for configuring CLIE via `.clie/clie.yml`.

## Overview

The CLIE CLI uses `.clie/clie.yml` to define the EAC extension, registry settings, and runtime behavior.

**Key concepts:**

- **Extensions** - Containerized tools (EAC)
- **Registry** - Container image fetch control
- **Defaults** - Fallback values for extensions
- **Environment** - Variables passed into containers

## Configuration File Location

```text
<repository-root>/.clie/clie.yml
```

The CLI discovers the repository root by walking up directories until it finds `.git`.

### Creating the Configuration

```bash
# Initialize configuration
clie init

# Result: Creates .clie/clie.yml
```

## Minimal Configuration

Simplest working configuration:

```yaml
extensions:
  - name: "eac"
    image: "ghcr.io/ready-to-release/eac-ext:latest"
```

Invokable as:

```bash
clie eac [command] [args...]
```

## Configuration Hierarchy

Values resolved in priority order (highest to lowest):

| Priority | Source                | Description                             |
| -------- | --------------------- | --------------------------------------- |
| 1        | Environment Variables | Runtime overrides via `CLIE_*` variables |
| 2        | User-Specific Config  | `.clie/clie.local.yml` (gitignored)   |
| 3        | Repository Config     | `.clie/clie.yml` (committed)          |
| 4        | Built-in Defaults     | Programmed into CLI                     |

### Configuration Layering

```text
.clie/
├── clie.yml           # Team configuration (committed)
└── clie.local.yml     # Personal overrides (gitignored)
```

**Team config** (`.clie/clie.yml`):

```yaml
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/eac-ext:latest'
```

**Personal overrides** (`.clie/clie.local.yml`):

```yaml
extensions:
  - name: 'eac'
    load_local: true
    image: 'eac-ext:dev'  # Use local development image
```

## Schema Reference

### Extensions Configuration

Define the EAC extension:

```yaml
extensions:
  - name: "eac"
    image: "ghcr.io/ready-to-release/eac-ext:latest"
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
image: "ghcr.io/ready-to-release/eac-ext:v1.2.3"

# Registry with organization, latest tag
image: "ghcr.io/ready-to-release/eac-ext:latest"

# Local development image
image: "eac-ext:dev"
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
    image: "ghcr.io/ready-to-release/eac-ext:latest"
```

### With Environment Variables

```yaml
extensions:
  - name: "eac"
    image: "ghcr.io/ready-to-release/eac-ext:latest"
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
    image: "eac-ext:dev"  # Local development image
    environment:
      - name: "EAC_DEBUG"
        value: "true"
      - name: "CLIE_LOCAL_DEV"
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
    image: "ghcr.io/ready-to-release/eac-ext:latest"
```

## User-Specific Overrides

### Local Development File

Create `.clie/clie.local.yml` (gitignored) for personal settings:

**Team config** (`.clie/clie.yml`):

```yaml
extensions:
  - name: "eac"
    image: "ghcr.io/ready-to-release/eac-ext:latest"
```

**Your overrides** (`.clie/clie.local.yml`):

```yaml
extensions:
  - name: "eac"
    pull_policy: "Never"
    image: "eac-ext:dev"
    environment:
      - name: "EAC_DEBUG"
        value: "true"
```

### Recommended .gitignore

```gitignore
# CLIE user-specific configuration
.clie/*.local.yml
.clie/*.personal.yml

# CLIE temporary files
.clie/.cache/
.clie/tmp/
```

## See Also

- [CLIE CLI Overview](index.md) - Command overview
- [init command](init.md) - Initialize configuration
- [install command](install.md) - Install extensions
- [validate command](validate.md) - Validate configuration
- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md) - Getting started guide
