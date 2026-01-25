# Architecture

The R2R CLI is the Ready-to-Release command-line interface for executing containerized development workflows. It manages Docker-based extensions that provide isolated, reproducible development environments.

## In This Section

| Document | Description |
|----------|-------------|
| [System Context](./system-context.md) | How R2R CLI interacts with external systems |
| [Subsystems](./subsystems.md) | Internal component architecture |
| [Workflows](./workflows.md) | Dynamic workflow diagrams |

## Key Features

| Feature | Description |
|---------|-------------|
| Containerized Execution | Run commands in isolated Docker containers |
| Multi-Config Support | Layer configs from repository, local, personal, and dev |
| Extension Pinning | SHA-based version pinning for reproducibility |
| TTY Support | Full terminal support including interactive sessions |
| Cross-Platform | Works on Linux, macOS, and Windows |

## Design File

- **Location**: `specs/r2r-cli/.design/workspace.dsl`
- **Interactive**: `r2r eac serve-design --module r2r-cli`

## See Also

- [R2R CLI Commands](../commands/) - Command reference
- [Configuration](../configuration/) - Configuration options
