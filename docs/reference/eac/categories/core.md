# Core Commands

## Overview

Core commands are top-level commands that provide essential functionality and don't fit into a specific category.

These commands are invoked directly without a category prefix.

**Key Characteristics**:

- Top-level invocation (no category prefix)
- Essential utility commands
- Cross-cutting functionality

**When to use**: For general-purpose operations that apply across the entire repository or don't belong to a specific workflow category.

## All Core Commands

| Command | Description |
|---------|-------------|
| [build](../commands/build/build.md) | Build one or more modules |
| [lint](../commands/lint.md) | Run linters on modules |
| [scan](../commands/scan/index.md) | Run security scanners on modules |
| [test](../commands/test/index.md) | Run tests for modules |
| [extension-meta](../commands/extension-meta.md) | Output extension metadata for CLI integration |
| [init](../commands/init/index.md) | Initialize a new EAC repository |
| [help](../commands/help/index.md) | Display help information |

## Common Use Cases

### Linting Code

```bash
# Lint all modules
eac lint

# Lint specific module
eac lint eac-commands

# Lint with auto-fix
eac lint --fix
```

The lint command automatically selects appropriate linters based on module component types (Go, Markdown, etc.).

### Extension Metadata

```bash
# Output extension metadata for CLIE extension host
eac extension-meta
```

Provides metadata about the EAC extension for [CLIE](../../clie/index.md) integration. See [extension-meta](../commands/extension-meta.md) for details.

## See Also

- [build](../commands/build/build.md) - Build modules
- [test](../commands/test/index.md) - Test modules
- [validate](../commands/validate/index.md) - Validate repository structure
