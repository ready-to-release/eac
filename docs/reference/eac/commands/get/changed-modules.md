# get changed-modules

<!-- book:cmd get changed-modules -->

## How It Works

Identifies modules affected by file changes:

- **File Ownership**: Maps changed files to owning modules via file patterns
- **Dependency Propagation**: Includes modules that depend on changed modules
- **Git Detection**: Uses git status to identify modified files
- **Incremental Builds**: Enables building only affected modules

Used to optimize CI/CD by skipping unaffected modules.

## See Also

- [get changed-modules-ci](./changed-modules-ci.md) - For CI pipelines
- [show files-changed](../show/files-changed.md)
- [build](../build/build.md)
