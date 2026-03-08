# CLIE CLI Command Reference

The CLIE CLI manages containerized extensions.

It provides framework commands for installing and running the EAC extension.

## Command Overview

| Command                       | Description                                   |
| ----------------------------- | --------------------------------------------- |
| [init](init.md)               | Initialize `.clie/clie.yml` configuration     |
| [install](install.md)         | Install the EAC extension                     |
| [run](run.md)                 | Run an extension explicitly                   |
| [list](list.md)               | List available extensions                     |
| [validate](validate.md)       | Validate configuration syntax                 |
| [verify](verify.md)           | Verify system prerequisites                   |
| [cleanup](cleanup.md)         | Clean up old Docker images                    |
| [interactive](interactive.md) | Open interactive shell in extension container |
| [metadata](metadata.md)       | Retrieve extension metadata                   |
| [update self](update.md)      | Update the clie CLI binary to latest version  |
| [version](version.md)         | Display CLIE CLI version                      |

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

## Direct Extension Invocation

When CLIE starts, it reads `.clie/clie.yml` and registers each configured extension name
as a direct subcommand. This is why `clie eac build` works — `eac` becomes a registered
command that is an alias for `clie run eac`.

```bash
clie eac build       # alias: equivalent to clie run eac build
clie run eac build   # explicit form
```

Aliases are only registered when the configuration file loads successfully at startup.

!!! note "EAC can also run standalone"
    If you don't need containerized execution, you can install the `eac` CLI directly.
    See [Install EAC CLI](../../../how-to-guides/local-setup/install-eac.md).

## CLIE CLI vs EAC CLI

| Aspect       | CLIE CLI              | EAC Extension         |
| ------------ | --------------------- | --------------------- |
| **Purpose**  | Extension framework   | Automation tools      |
| **Commands** | init, install, list   | build, test, validate |
| **Runs**     | Host machine          | Docker container      |
| **Config**   | `.clie/clie.yml`      | `.eac/`               |

## Configuration File

CLIE CLI uses `.clie/clie.yml`:

```yaml
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/eac-ext:latest'
```

See [Configuration Reference](../configuration.md) for details.

## See Also

- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md)
- [Configuration Reference](../configuration.md)
- [EAC Commands Reference](../../eac/commands/index.md)
- [CLI vs Extensions](../../eac/architecture/cli-integration.md)
