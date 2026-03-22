# Show Workspaces

<!-- book:cmd show workspaces -->

Lists all git worktrees (workspaces) in a formatted table showing their path, branch, and status.

## Usage

```bash
eac show workspaces [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--verbose`, `-v` | bool | Show additional information including commit SHA |
| `--debug`, `-d` | bool | Enable debug logging to `out/commands.log` |

## Output Sections

A table with columns:

| Column | Default | Verbose |
|--------|---------|---------|
| Path | Yes | Yes |
| Branch | Yes | Yes |
| SHA | No | Yes (first 7 chars) |
| Status | Yes | Yes |

Status is either `clean` (no uncommitted changes) or `dirty` (has uncommitted changes). Detached HEAD branches display as `(detached)`.

A summary line shows the total worktree count.

## Examples

```bash
# List all worktrees
eac show workspaces

# List with commit SHAs
eac show workspaces --verbose
```

## See Also

- [work create](../work/create.md) - Create workspace
- [work remove](../work/remove.md) - Remove workspace
- [work Commands](../../categories/work.md)
