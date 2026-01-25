# R2R CLI Commands

Reference documentation for R2R CLI framework commands.

## Command Overview

| Command                           | Description                                   |
| --------------------------------- | --------------------------------------------- |
| [init](init.md)                   | Initialize `.r2r/r2r-cli.yml` configuration   |
| [install](install.md)             | Install extensions                            |
| [list](list.md)                   | List available extensions                     |
| [validate](validate.md)           | Validate configuration syntax                 |
| [verify](verify.md)               | Verify system prerequisites                   |
| [cleanup](cleanup.md)             | Clean up old Docker images                    |
| [interactive](interactive.md)     | Open interactive shell in extension container |
| [metadata](metadata.md)           | Retrieve extension metadata                   |
| [version](version.md)             | Display R2R CLI version                       |
| [configuration](../configuration/) | Configuration file reference                  |

## Quick Start

```bash
# Initialize R2R configuration
r2r init

# Install EAC extension
r2r install eac

# Use EAC commands
r2r eac build
r2r eac test
```

## See Also

- [R2R CLI Overview](../index.md) - R2R CLI reference index
- [R2R Architecture](../architecture/) - R2R CLI system architecture
- [EAC Commands](../../eac/commands/) - EAC extension commands
