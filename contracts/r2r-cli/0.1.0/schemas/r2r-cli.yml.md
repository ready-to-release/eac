# R2R CLI Configuration Reference

This document provides comprehensive documentation for the R2R CLI configuration file (`r2r-cli.yml`).

The configuration controls how containerized extensions are discovered, pulled, and executed.

## Table of Contents

- [Overview](#overview)
- [Configuration Location](#configuration-location)
- [Configuration Hierarchy](#configuration-hierarchy)
- [Quick Start](#quick-start)
- [Schema Reference](#schema-reference)
  - [Registry Configuration](#registry-configuration)
  - [Defaults Configuration](#defaults-configuration)
  - [Environment Configuration](#environment-configuration)
  - [Extensions Configuration](#extensions-configuration)
- [Environment Variables](#environment-variables)
- [CI/CD Considerations](#cicd-considerations)
- [User-Specific Overrides](#user-specific-overrides)
- [Complete Examples](#complete-examples)

---

## Overview

The R2R CLI uses a YAML configuration file to define containerized extensions that extend the CLI's capabilities.

Each extension runs in an isolated Docker container, providing consistent, reproducible tooling across development environments and CI/CD pipelines.

**Key Concepts:**

- **Extensions** are containerized tools invoked via the CLI
- **Registry** settings control how container images are fetched
- **Defaults** provide fallback values for all extensions
- **Environment** manages variables passed into containers

---

## Configuration Location

The configuration file must be located at:

```text
<repository-root>/.r2r/r2r-cli.yml
```

The CLI discovers the repository root by walking up from the current directory until it finds a `.git` folder.

---

## Configuration Hierarchy

Configuration values are resolved in the following priority order (highest to lowest):

| Priority | Source                | Description                                                 |
| -------- | --------------------- | ----------------------------------------------------------- |
| 1        | Environment Variables | Runtime overrides via `R2R_*` variables                     |
| 2        | User-Specific Config  | `.r2r/r2r-cli.local.yml`, `.r2r/r2r-cli.personal.yml`, etc. |
| 3        | Repository Config     | `.r2r/r2r-cli.yml`                                          |
| 4        | Built-in Defaults     | Sensible defaults programmed into the CLI                   |

---

## Quick Start

Minimal configuration with a single extension:

```yaml
extensions:
  - name: "example"
    image: "ghcr.io/my-org/my-extension:1.0.0"
```

This defines an extension named `example` that can be invoked as:

```bash
r2r example [args...]
```

---

## Schema Reference

### Registry Configuration

Controls how the CLI interacts with container registries.

```yaml
registry:
  default: "ghcr.io"
  timeout: 30
  retry_attempts: 3
  authentication:
    required: false
    username_env: "REGISTRY_USER"
    token_env: "REGISTRY_TOKEN"
```

#### Fields

| Field                         | Type    | Default     | Description                                                                            |
| ----------------------------- | ------- | ----------- | -------------------------------------------------------------------------------------- |
| `default`                     | string  | `"ghcr.io"` | Default registry hostname. Must match pattern `^[a-zA-Z0-9][a-zA-Z0-9.-]*[a-zA-Z0-9]$` |
| `timeout`                     | integer | `30`        | Registry operation timeout in seconds                                                  |
| `retry_attempts`              | integer | `3`         | Number of retry attempts for failed operations                                         |
| `authentication.required`     | boolean | `false`     | Whether authentication is mandatory                                                    |
| `authentication.username_env` | string  | -           | Environment variable containing the registry username                                  |
| `authentication.token_env`    | string  | -           | Environment variable containing the registry token/password                            |

#### Authentication Example

For private registries requiring authentication:

```yaml
registry:
  default: "ghcr.io"
  authentication:
    required: true
    username_env: "GH_USER"
    token_env: "GH_TOKEN"
```

The CLI reads credentials from the specified host environment variables at runtime.

---

### Defaults Configuration

Provides default values applied to all extensions unless overridden.

```yaml
defaults:
  registry: "ghcr.io/ready-to-release/r2r"
  pull_policy: "IfNotPresent"
  remove_after: true
  timeout: 0
  memory_limit: "512MB"
  cpu_limit: "1.0"
  environment:
    - name: "NO_COLOR"
      value: "1"
```

#### Fields

| Field          | Type    | Default          | Description                                            |
| -------------- | ------- | ---------------- | ------------------------------------------------------ |
| `registry`     | string  | -                | Default registry prefix prepended to image references  |
| `pull_policy`  | enum    | `"IfNotPresent"` | When to pull images: `Always`, `IfNotPresent`, `Never` |
| `remove_after` | boolean | `true`           | Remove container after execution completes             |
| `timeout`      | integer | `0`              | Execution timeout in seconds (0 = no limit)            |
| `memory_limit` | string  | -                | Memory limit (e.g., `512MB`, `2GB`, `1g`)              |
| `cpu_limit`    | string  | -                | CPU limit as decimal cores (e.g., `0.5`, `2.0`)        |
| `environment`  | array   | `[]`             | Default environment variables for all containers       |

#### Pull Policy Behavior

| Policy         | Behavior                                        |
| -------------- | ----------------------------------------------- |
| `Always`       | Always pull the image before running            |
| `IfNotPresent` | Pull only if the image doesn't exist locally    |
| `Never`        | Never pull; fail if image doesn't exist locally |

---

### Environment Configuration

Manages global environment variables and secrets passed to all extensions.

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
    - name: "NPM_TOKEN"
      env: "NPM_AUTH_TOKEN"
```

#### Global Environment Variables

Variables defined in `global` are passed to every extension container.

```yaml
environment:
  global:
    - name: "LOG_LEVEL"
      value: "debug"
```

#### Secrets

Secrets map a container environment variable name to a host environment variable, allowing secure credential injection without hardcoding values:

```yaml
environment:
  secrets:
    - name: "API_KEY"        # Variable name inside the container
      env: "HOST_API_KEY"    # Variable name on the host to read from
```

When the container runs, `API_KEY` inside the container will contain the value of `HOST_API_KEY` from the host.

#### Environment Variable Naming

All environment variable names must:

- Start with an uppercase letter
- Contain only uppercase letters, digits, and underscores
- Match pattern: `^[A-Z][A-Z0-9_]*$`

---

### Extensions Configuration

Extensions are the core of the configuration.

Each extension defines a containerized tool.

```yaml
extensions:
  - name: "eac"
    image: "ghcr.io/ready-to-release/ext-eac:sha-abc123"
    image_pull_policy: "IfNotPresent"

    env:
      - name: "POWERSHELL_TELEMETRY_OPTOUT"
        value: "1"

    volumes:
      - host: "/var/run/docker.sock"
        container: "/var/run/docker.sock"
        readonly: true

    ports:
      - host: 8080
        container: 80

    privileged: false
    network_mode: "bridge"
    memory_limit: "1GB"
    cpu_limit: "2.0"

    repo_url: "https://github.com/ready-to-release/eac"
    docs_url: "https://github.com/ready-to-release/eac#readme"
```

#### Required Fields

| Field   | Type   | Description                                                                                                          |
| ------- | ------ | -------------------------------------------------------------------------------------------------------------------- |
| `name`  | string | Unique extension name used as the CLI command. Pattern: `^[a-z]([a-z0-9]+(-[a-z0-9]+)*)?$` (lowercase, max 64 chars) |
| `image` | string | Container image reference with tag or digest                                                                         |

#### Optional Fields

| Field               | Type    | Default        | Description                                     |
| ------------------- | ------- | -------------- | ----------------------------------------------- |
| `description`       | string  | -              | Human-readable description shown in help        |
| `version`           | string  | -              | Extension version (semver format)               |
| `image_pull_policy` | enum    | `"AutoDetect"` | `Always`, `IfNotPresent`, `Never`, `AutoDetect` |
| `env`               | array   | `[]`           | Extension-specific environment variables        |
| `volumes`           | array   | `[]`           | Volume mounts from host to container            |
| `ports`             | array   | `[]`           | Port mappings from host to container            |
| `working_dir`       | string  | -              | Working directory inside container              |
| `entrypoint`        | array   | -              | Override container entrypoint                   |
| `command`           | array   | -              | Override container command                      |
| `privileged`        | boolean | `false`        | Run in privileged mode (use with caution)       |
| `network_mode`      | enum    | `"bridge"`     | `bridge`, `host`, `none`                        |
| `memory_limit`      | string  | -              | Memory limit (e.g., `512MB`, `1GB`)             |
| `cpu_limit`         | string  | -              | CPU limit (e.g., `0.5`, `2.0`)                  |
| `repo_url`          | string  | -              | URL to extension source repository              |
| `docs_url`          | string  | -              | URL to extension documentation                  |

#### Image Pull Policy: AutoDetect

When set to `AutoDetect` (the default), the CLI determines the pull policy based on the image tag:

| Tag Pattern                                      | Resolved Policy |
| ------------------------------------------------ | --------------- |
| `:latest`, `:main`, `:master`                    | `Always`        |
| No tag specified                                 | `Always`        |
| Specific version (e.g., `:1.0.0`, `:sha-abc123`) | `IfNotPresent`  |

#### Environment Variables in Extensions

Extension environment variables support three modes:

**Static Value:**

```yaml
env:
  - name: "DEBUG"
    value: "true"
```

**Passthrough from Host:**

```yaml
env:
  - name: "HOME"
    # No value = pass through from host
```

**Required Passthrough:**

```yaml
env:
  - name: "API_KEY"
    required: true  # Fail if not set on host
```

#### Volume Mounts

Mount host paths into the container:

```yaml
volumes:
  - host: "/path/on/host"
    container: "/path/in/container"
    readonly: false
```

The `readonly` field defaults to `false`. Set to `true` for read-only access.

#### Port Mappings

Expose container ports on the host:

```yaml
ports:
  - host: 8080
    container: 80
```

Port numbers must be between 1 and 65535.

---

## Environment Variables

The CLI automatically sets these environment variables in every container:

| Variable                 | Value           | Description                               |
| ------------------------ | --------------- | ----------------------------------------- |
| `R2R_CONTAINER_REPOROOT` | `/var/task`     | Repository root path inside the container |
| `R2R_HOST_REPOROOT`      | `<actual path>` | Repository root path on the host          |

### CI Environment Detection

When running in CI, the CLI automatically sets:

| Variable      | Value  |
| ------------- | ------ |
| `NO_COLOR`    | `1`    |
| `TERM`        | `dumb` |
| `FORCE_COLOR` | `0`    |

The CLI detects CI environments via standard CI environment variables (`CI`, `GITHUB_ACTIONS`, `GITLAB_CI`, etc.).

---

## CI/CD Considerations

### Image Pinning Requirement

**In CI environments, all extension images must be pinned to a specific version.**

Unpinned tags like `:latest`, `:main`, or implicit latest (no tag) will cause the CLI to fail with an error.

**Allowed in CI:**

```yaml
extensions:
  - name: "eac"
    image: "ghcr.io/org/eac:sha-abc123"   # Pinned to commit SHA
  - name: "node"
    image: "ghcr.io/org/node:1.2.3"       # Pinned to version
```

**Rejected in CI:**

```yaml
extensions:
  - name: "eac"
    image: "ghcr.io/org/eac:latest"       # ERROR: unpinned
  - name: "node"
    image: "ghcr.io/org/node"             # ERROR: implicit latest
```

### Why Pinning Matters

Unpinned images can cause:

- **Non-reproducible builds**: Different runs may use different image versions
- **Silent breakages**: Upstream changes can break your pipeline unexpectedly
- **Security risks**: Malicious updates could be pulled automatically

---

## User-Specific Overrides

Developers can create personal configuration overrides that are not committed to version control.

The CLI looks for these files in priority order:

| File                          | Purpose                                        |
| ----------------------------- | ---------------------------------------------- |
| `.r2r/r2r-cli.local.yml`      | Local development overrides (highest priority) |
| `.r2r/r2r-cli.personal.yml`   | Personal customizations                        |
| `.r2r/r2r-cli.dev.yml`        | Development environment settings               |
| `.r2r/r2r-cli.<username>.yml` | User-specific settings                         |

### Override Example

Base config (`.r2r/r2r-cli.yml`):

```yaml
extensions:
  - name: "eac"
    image: "ghcr.io/org/eac:sha-abc123"
```

Local override (`.r2r/r2r-cli.local.yml`):

```yaml
extensions:
  - name: "eac"
    image: "eac-local:dev"  # Use local development image
```

Add to `.gitignore`:

```text
.r2r/r2r-cli.local.yml
.r2r/r2r-cli.personal.yml
.r2r/r2r-cli.dev.yml
.r2r/r2r-cli.*.yml
!.r2r/r2r-cli.yml
```

---

## Complete Examples

### Minimal Configuration

```yaml
extensions:
  - name: "hello"
    image: "alpine:3.18"
```

### Development Team Configuration

```yaml
registry:
  default: "ghcr.io"
  timeout: 60
  retry_attempts: 5
  authentication:
    required: true
    username_env: "GH_USER"
    token_env: "GH_TOKEN"

defaults:
  registry: "ghcr.io/acme-corp/tools"
  pull_policy: "IfNotPresent"
  remove_after: true
  memory_limit: "1GB"
  cpu_limit: "1.0"

environment:
  global:
    - name: "TZ"
      value: "UTC"
  secrets:
    - name: "GITHUB_TOKEN"
      env: "GH_TOKEN"
    - name: "NPM_TOKEN"
      env: "NPM_AUTH_TOKEN"

extensions:
  - name: "build"
    description: "Build toolchain"
    version: "2.1.0"
    image: "ghcr.io/acme-corp/tools/build:sha-a1b2c3d"
    memory_limit: "4GB"
    cpu_limit: "4.0"
    env:
      - name: "NODE_ENV"
        value: "production"

  - name: "lint"
    description: "Code linting and formatting"
    version: "1.5.0"
    image: "ghcr.io/acme-corp/tools/lint:sha-e4f5g6h"
    env:
      - name: "ESLINT_USE_FLAT_CONFIG"
        value: "true"

  - name: "test"
    description: "Test runner"
    version: "3.0.0"
    image: "ghcr.io/acme-corp/tools/test:sha-i7j8k9l"
    volumes:
      - host: "/var/run/docker.sock"
        container: "/var/run/docker.sock"
    privileged: false
    network_mode: "host"
```

### Service Extension with Ports

```yaml
extensions:
  - name: "docs"
    description: "Documentation server"
    image: "ghcr.io/org/docs-server:1.0.0"
    ports:
      - host: 3000
        container: 80
    working_dir: "/docs"
    command: ["serve", "--port", "80"]
```

---

## Schema Validation

The configuration file is validated against the JSON Schema at:

```text
contracts/r2r-cli/0.1.0/r2r-cli.schema.json
```

Validation includes:

- Required field presence (`name`, `image` for extensions)
- Pattern matching for names, image references, and environment variables
- Enum validation for `pull_policy`, `network_mode`, etc.
- Type checking for all fields

---

## See Also

- [JSON Schema](./r2r-cli.schema.json) - Formal schema definition
- [Command EBNF](./command.ebnf) - CLI command grammar
