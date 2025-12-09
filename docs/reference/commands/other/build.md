# build

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac build [module...] [--all]`
**Purpose**: Build one or more modules by moniker
**Category**: [other](../categories/other.md)

## Syntax

```bash
r2r eac build [module...] [--all] [options]
```

## Options

| Flag | Description |
|------|-------------|
| `--all` | Build all modules |
| `--clean` | Clean before building |
| `--parallel` | Build in parallel |
| `--verbose` | Verbose output |

## Examples

```bash
# Build single module
r2r eac build src-auth

# Build multiple modules
r2r eac build src-auth src-api

# Build all modules
r2r eac build --all

# Clean build
r2r eac build src-auth --clean

# Build changed modules
CHANGED=$(r2r eac get changed-modules | jq -r '.changed_modules[]')
r2r eac build $CHANGED
```

## Build Order

Builds respect dependency order automatically:

```bash
# Builds dependencies first
r2r eac build r2r-cli

# Execution order: eac-core → eac-commands → r2r-cli
```

## See Also

- [get execution-order](../get/execution-order.md) - Build order
- [get build-deps](../get/build-deps.md) - Build dependencies
- [show artifacts](../show/artifacts.md) - Verify artifacts
- [validate artifacts](../validate/artifacts.md)
- [other Commands](../categories/other.md)

{{ diataxis_footer() }}
