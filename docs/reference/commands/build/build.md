# build

<!-- book:cmd build -->

## Language Support

The build command supports multiple languages via capability-based handlers:

- **Go** (`type: go`) - Compiles Go binaries with cross-platform support
- **TypeScript** (`type: typescript`) - Runs npm install and tsc compilation
- **Docker** (`type: container`) - Builds container images with buildx
- **MkDocs** (`type: docs`) - Generates documentation sites
- **Scripts** (`type: static` with scripts) - Custom build commands

The command automatically selects the appropriate build handler based on the module's declared capabilities. See [Module Types Reference](../../r2r-eac/module-types-reference.md) for configuration details.

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
- [build Commands](../categories/build.md)
