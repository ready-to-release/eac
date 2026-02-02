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

<!-- book:category-commands core -->

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
# Output extension metadata for r2r CLI
eac extension-meta
```

Provides metadata about the EAC extension for CLI integration and tooling.

## See Also

- [build](./build.md) - Build modules
- [test](./test.md) - Test modules
- [update lint](../update/lint.md) - Update lint configurations
- [validate](./validate.md) - Validate repository structure
