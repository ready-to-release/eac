# CLIE CLI Command Reference

The CLIE CLI manages containerized extensions.

It provides framework commands for installing and running the EAC extension.

## Command Overview

| Command                       | Description                                   |
| ----------------------------- | --------------------------------------------- |
| [init](init.md)               | Initialize `.clie/clie-cli.yml` configuration   |
| [install](install.md)         | Install the EAC extension                     |
| [list](list.md)               | List available extensions                     |
| [validate](validate.md)       | Validate configuration syntax                 |
| [verify](verify.md)           | Verify system prerequisites                   |
| [cleanup](cleanup.md)         | Clean up old Docker images                    |
| [interactive](interactive.md) | Open interactive shell in extension container |
| [metadata](metadata.md)       | Retrieve extension metadata                   |
| [version](version.md)         | Display CLIE CLI version                       |

## Quick Start

```bash
# 1. Initialize configuration
clie init

# 2. Install EAC extension
clie install eac

# 3. Use EAC commands
clie eac build
clie eac test
```

## CLIE CLI vs EAC Extension

| Aspect       | CLIE CLI             | EAC Extension         |
| ------------ | ------------------- | --------------------- |
| **Purpose**  | Extension framework | Automation tools      |
| **Commands** | init, install, list | build, test, validate |
| **Runs**     | Host machine        | Docker container      |
| **Config**   | `.clie/clie-cli.yml`  | `.eac/`           |

## Configuration File

CLIE CLI uses `.clie/clie-cli.yml`:

```yaml
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/eac-ext:latest'
```

See [Configuration Reference](configuration.md) for details.

## See Also

- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md)
- [Configuration Guide](configuration.md)
- [EAC Commands Reference](../../eac/commands/index.md)
- [CLI vs Extensions](../../eac/architecture/cli-integration.md)
