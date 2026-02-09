# Build Commands

## Overview

Build commands compile and package modules in the repository.

They handle dependency resolution, incremental builds, and cross-platform compilation.

**Key Characteristics**:

- Dependency-aware build ordering
- Incremental build support
- Multi-language support (Go, TypeScript, Docker)
- Parallel execution by default

**When to use**: When compiling modules, preparing releases, or verifying that code changes compile correctly.

## All Build Commands

<!-- book:category-commands build -->

## Common Workflows

### Building During Development

```bash
# Build all modules
eac build

# Build specific module
eac build eac-commands

# Build multiple modules
eac build eac-core clie
```

### Building with Dependencies

```bash
# Build module and its dependencies
eac build clie
# Automatically builds: eac-core → eac-commands → clie
```

### CI/CD Integration

```bash
# Build all modules in CI
eac build

# Build changed modules only
eac build $(eac get changed-modules | jq -r '.changed_modules[]')
```

## See Also

- [get build-deps](../get/build-deps.md) - View build dependencies
- [show build-summary](../show/build-summary.md) - Build execution summary
- [show build-times](../show/build-times.md) - Build performance analysis
- [test](./test.md) - Test after building
