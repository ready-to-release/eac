# R2R CLI Reference

The R2R CLI manages containerized extensions. It provides framework commands for installing and running the EAC extension.

## Command Overview

| Command                                | Description                                   |
| -------------------------------------- | --------------------------------------------- |
| [init](commands/init.md)               | Initialize `.r2r/r2r-cli.yml` configuration   |
| [install](commands/install.md)         | Install the EAC extension                     |
| [list](commands/list.md)               | List available extensions                     |
| [validate](commands/validate.md)       | Validate configuration syntax                 |
| [verify](commands/verify.md)           | Verify system prerequisites                   |
| [cleanup](commands/cleanup.md)         | Clean up old Docker images                    |
| [interactive](commands/interactive.md) | Open interactive shell in extension container |
| [metadata](commands/metadata.md)       | Retrieve extension metadata                   |
| [version](commands/version.md)         | Display R2R CLI version                       |

## Quick Start

```bash
# 1. Initialize configuration
r2r init

# 2. Install EAC extension
r2r install eac

# 3. Use EAC commands
r2r eac build
r2r eac test
```

## R2R CLI vs EAC Extension

| Aspect       | R2R CLI             | EAC Extension         |
| ------------ | ------------------- | --------------------- |
| **Purpose**  | Extension framework | Automation tools      |
| **Commands** | init, install, list | build, test, validate |
| **Runs**     | Host machine        | Docker container      |
| **Config**   | `.r2r/r2r-cli.yml`  | `.r2r/eac/`           |

## Configuration File

R2R CLI uses `.r2r/r2r-cli.yml`:

```yaml
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/ext-eac:latest'
```

See [Configuration Reference](configuration/) for details.

## In This Section

| Reference                     | Description                 |
| ----------------------------- | --------------------------- |
| [Commands](commands/)         | R2R CLI command reference   |
| [Architecture](architecture/) | R2R CLI system architecture |

## See Also

- [Quick Start Tutorial](../../tutorials/getting-started/quick-start.md)
- [EAC Reference](../eac/) - EAC extension documentation
