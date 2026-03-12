# Show Files Changed

<!-- book:cmd show files-changed -->

Displays modified (unstaged) files from `git diff HEAD` with their owning module(s). Useful for identifying which modules are affected by uncommitted changes.

## Usage

```bash
eac show files-changed
```

## Output Sections

A table with columns:

| Column | Description |
|--------|-------------|
| File | Relative file path |
| Modules | Comma-separated list of owning modules, or `NONE` if unowned |

Produces no output if there are no changed files.

## Examples

```bash
# Show changed files with module ownership
eac show files-changed
```

## See Also

- [show files-staged](./files-staged.md) - Staged files
- [get changed-modules](../get/changed-modules.md) - Modules affected (JSON)
- [work commit](../work/commit.md) - Commit changes
