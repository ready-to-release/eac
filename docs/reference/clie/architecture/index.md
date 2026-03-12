# CLIE CLI Architecture

Architecture documentation for the CLIE CLI framework.

## Overview

CLIE (CLI Extender) provides a container-based CLI framework for modular, extensible command-line tools.

The architecture focuses on:

- **Docker-based execution** - Isolated, reproducible environments
- **Extension system** - Pluggable command collections via containers
- **Git-aware workflows** - Automatic repository detection and mounting
- **Cross-platform support** - Windows, macOS, Linux compatibility

## Ecosystem Overview

![CLIE CLI Ecosystem Overview](../../../assets/clie/clie-overview.drawio.png)

## Core Design Principles

### 1. Container Isolation

Extensions execute in isolated Docker containers, providing:

- Reproducible builds across environments
- Dependency isolation (no global installs)
- Consistent behavior (Linux runtime for all platforms)
- Security boundaries between extensions

### 2. Repository-Centric

CLIE automatically detects git repositories and:

- Mounts repository root as `/workspace`
- Preserves working directory context
- Enables git-based workflows (change detection, hooks)
- Respects `.gitignore` for artifacts

### 3. Extension Modularity

Extensions are self-contained Docker images:

- Single binary entrypoint
- Declarative configuration (`.clie/clie.yml`)
- Independent versioning and releases
- Composable (extensions can invoke other extensions)

## Architecture Documentation

| Document                     | Description                                   |
| ---------------------------- | --------------------------------------------- |
| [CLIE CLI Module](module.md) | Detailed module architecture with C4 diagrams |

## Command Discovery

Extensions register commands dynamically:

1. CLIE detects extension name from first argument (`clie eac ...`)
2. Looks up extension in `.clie/clie.yml`
3. Passes remaining arguments to extension (`build ...`)
4. Extension handles subcommand routing internally

No central registry needed - extensions are self-documenting.

## Performance Optimizations

### Image Caching

- Docker layer cache for fast rebuilds
- Multi-stage builds for smaller images
- Base image reuse across extensions

### Volume Mounting

- Bind mounts (not copies) for instant access
- Read-only mounts for immutable data
- Cached mounts for dependencies

### Container Reuse (Future)

Planned: Long-running containers for faster execution

```yaml
extensions:
  - name: eac
    keep_alive: true # Don't remove after command
    idle_timeout: 5m # Remove after 5 min idle
```

## Security Model

### Isolation

- Extensions run as non-root user in container
- Limited host access (only mounted volumes)
- Network isolation (no internet by default)

### Trust Model

Extensions execute arbitrary code - users must trust extension authors.

**Mitigations**:

- Extensions are Docker images (auditable)
- Local build option (`load_local: true`)
- No automatic updates (explicit version pins)

## Error Handling

CLIE uses exit codes for error signaling:

| Exit Code | Meaning             |
| --------- | ------------------- |
| 0         | Success             |
| 1         | General error       |
| 2         | Configuration error |
| 3         | Docker error        |
| 125+      | Container exit code |

Errors propagate from extension to CLI to user.

## Comparison with Alternatives

| Feature             | CLIE | Make | Devbox | Task |
| ------------------- | ---- | ---- | ------ | ---- |
| Container isolation | ✅   | ❌   | ✅     | ❌   |
| Cross-platform      | ✅   | ⚠️   | ✅     | ✅   |
| Extension system    | ✅   | ❌   | ⚠️     | ❌   |
| Git integration     | ✅   | ❌   | ❌     | ❌   |
| Reproducible builds | ✅   | ⚠️   | ✅     | ⚠️   |

**CLIE sweet spot**: Teams needing reproducible, containerized workflows with modular command collections.

## Related Documentation

- [CLIE CLI Commands](../commands/index.md) - Command reference
- [EAC Architecture](../../eac/architecture/index.md) - EAC CLI architecture
- [How-to Guides: CLIE](../../../how-to-guides/clie/index.md) - Usage guides
- [Creating Extensions](../../../how-to-guides/clie/creating-extensions.md) - Extension development guide
