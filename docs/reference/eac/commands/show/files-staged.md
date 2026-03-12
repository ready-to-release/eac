# Show Files Staged

<!-- book:cmd show files-staged -->

Displays staged files from `git diff --cached` with their owning module(s). Useful for verifying which modules are affected before committing.

## Usage

```bash
eac show files-staged
```

## Output Sections

A table with columns:

| Column | Description |
|--------|-------------|
| File | Relative file path |
| Modules | Comma-separated list of owning modules, or `NONE` if unowned |

## Examples

```bash
# Show staged files with module ownership
eac show files-staged
```

## See Also

- [show files-changed](./files-changed.md) - Unstaged changes
- [work commit](../work/commit.md) - Commit staged files
- [get commit-message](../get/commit-message.md)
