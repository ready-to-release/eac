# show files-changed

<!-- book:cmd show files-changed -->

## Output Sections

A table with columns:

| Column  | Description                                                  |
| ------- | ------------------------------------------------------------ |
| File    | Relative file path                                           |
| Modules | Comma-separated list of owning modules, or `NONE` if unowned |

Produces no output if there are no changed files.

## See Also

- [show files-staged](./files-staged.md) - Staged files
- [get changed-modules](../get/changed-modules.md) - Modules affected (JSON)
- [work commit](../work/commit.md) - Commit changes
