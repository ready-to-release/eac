# Work remove

<!-- book:cmd work remove -->

Removes a workspace (git worktree) and optionally deletes the associated local and remote branches.

By default removes the current workspace. Pass a branch name to remove a specific workspace. The command validates the workspace is clean before removal unless `--force` is used.

## Usage

```bash
eac work remove [branch-name] [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `branch-name` | Branch name of workspace to remove (default: current workspace) |

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--keep-branch` | | `false` | Keep local branch after removing workspace |
| `--delete-remote` | | `false` | Delete remote branch as well |
| `--force` | `-f` | `false` | Force remove even with uncommitted changes |
| `--debug` | `-d` | `false` | Enable debug logging |

## Examples

```bash
# Remove current workspace
eac work remove

# Remove a specific workspace by branch name
eac work remove feature/old-feature

# Keep the local branch
eac work remove --keep-branch

# Also delete the remote branch
eac work remove --delete-remote

# Force remove with uncommitted changes
eac work remove --force
```

## See Also

- [work create](./create.md)
- [show workspaces](../show/workspaces.md)
- [work Commands](../../categories/work.md)
