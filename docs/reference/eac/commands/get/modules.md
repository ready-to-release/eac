# get modules

<!-- book:cmd get modules -->

## Output Structure

Each module entry contains:

- `moniker` - Unique module identifier
- `type` - Module type (e.g., go, container, typescript, static)
- `root` - Root path relative to repository
- `depends_on` - List of dependency module monikers
- `versioning` - Versioning scheme (calver or semver)
- Additional metadata (books, files, workflows)

## See Also

- [show modules](../show/modules.md) - Formatted table
