# R2R CLI Command Reference

The R2R CLI manages containerized extensions. It provides framework commands for installing and running the EAC extension.

## Command Overview

| Command                       | Description                                   |
| ----------------------------- | --------------------------------------------- |
| [init](init.md)               | Initialize `.r2r/r2r-cli.yml` configuration   |
| [install](install.md)         | Install the EAC extension                     |
| [list](list.md)               | List available extensions                     |
| [validate](validate.md)       | Validate configuration syntax                 |
| [verify](verify.md)           | Verify system prerequisites                   |
| [cleanup](cleanup.md)         | Clean up old Docker images                    |
| [interactive](interactive.md) | Open interactive shell in extension container |
| [metadata](metadata.md)       | Retrieve extension metadata                   |
| [version](version.md)         | Display R2R CLI version                       |

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

See [Configuration Reference](configuration.md) for details.

## See Also

- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md)
- [Configuration Guide](configuration.md)
- [EAC Commands Reference](../../eac/commands/index.md)
- [CLI vs Extensions](../../eac/architecture/cli-integration.md)
