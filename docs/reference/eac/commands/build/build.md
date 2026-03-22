# build

<!-- book:cmd build -->

## Language Support

The build command selects handlers based on module type:

- **Go** (`type: go`) - Compiles Go binaries with cross-platform support
- **TypeScript** (`type: typescript`) - Runs npm install and tsc compilation
- **Docker** (`type: container`) - Builds container images with buildx
- **MkDocs** (`type: docs`) - Generates documentation sites
- **Scripts** (`type: static` with scripts) - Custom build commands

## Build Order

Builds respect dependency order automatically:

```bash
# Builds dependencies first
eac build clie
# Execution order: eac-core → eac-commands → clie
```

## See Also

- [get build-deps](../get/build-deps.md) - Build dependencies
- [show artifacts](../show/artifacts.md) - Verify artifacts
- [build Commands](../../categories/build.md)
