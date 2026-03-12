# Work merge

<!-- book:cmd work merge -->

Merges the current workspace branch back into the target branch using squash merge by default, creating a single well-documented commit with an AI-generated message.

The command validates the workspace is clean and up to date, switches to the target branch, performs the merge, and removes the workspace afterward (unless `--keep-worktree` is set).

## Usage

```bash
eac work merge [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--target` | `main` | Target branch to merge into |
| `--no-squash` | `false` | Use regular merge instead of squash merge |
| `--keep-worktree` | `false` | Keep workspace after merge |
| `--debug` | `false` | Enable debug mode |

## Examples

```bash
# Squash merge into main (default)
eac work merge

# Merge into a different branch
eac work merge --target=develop

# Regular merge (preserves individual commits)
eac work merge --no-squash

# Keep the workspace after merging
eac work merge --keep-worktree
```

## See Also

- [get squash-message](../get/squash-message.md)
- [work remove](./remove.md)
- [work Commands](../categories/work.md)
