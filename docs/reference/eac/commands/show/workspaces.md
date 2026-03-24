# show workspaces

<!-- book:cmd show workspaces -->

## Output Sections

A table with columns:

| Column | Default | Verbose             |
| ------ | ------- | ------------------- |
| Path   | Yes     | Yes                 |
| Branch | Yes     | Yes                 |
| SHA    | No      | Yes (first 7 chars) |
| Status | Yes     | Yes                 |

Status is either `clean` (no uncommitted changes) or `dirty` (has uncommitted changes). Detached HEAD branches display as `(detached)`.

A summary line shows the total worktree count.

## See Also

- [work create](../work/create.md) - Create workspace
- [work remove](../work/remove.md) - Remove workspace
- [work Commands](../work/index.md)
