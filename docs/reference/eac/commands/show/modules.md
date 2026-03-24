# show modules

<!-- book:cmd show modules -->

## Output Sections

A table with columns:

| Column     | Description                    |
| ---------- | ------------------------------ |
| Moniker    | Module identifier              |
| Group      | Module group                   |
| Components | Component types display string |

With `--with-artifacts`, three additional columns appear:

| Column    | Description                               |
| --------- | ----------------------------------------- |
| Artifacts | Total artifact count for current platform |
| Missing   | Number of artifacts not found on disk     |
| Overrides | Number of metadata overrides applied      |

## See Also

- [get modules](../get/modules.md) - JSON output
- [show component-kinds](./component-kinds.md) - Group by type
- [show dependencies](./dependencies.md) - Dependency graph
