# Work pull

<!-- book:cmd work pull -->

Fetches the latest changes from the target branch (default: main) and rebases the current branch on top, keeping commit history linear.

If conflicts occur, the command provides clear instructions for resolution. Use `--autostash` to automatically stash and reapply uncommitted changes.

## Usage

```bash
eac work pull [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--target` | `main` | Target branch to rebase onto |
| `--autostash` | `false` | Automatically stash uncommitted changes before rebasing |
| `--no-fetch` | `false` | Skip fetching from remote (use local target branch) |
| `--debug` | `false` | Enable debug logging |

## Examples

```bash
# Rebase onto latest main
eac work pull

# Rebase onto a different branch
eac work pull --target=develop

# Auto-stash uncommitted changes
eac work pull --autostash
```

## See Also

- [work create](./create.md)
- [work merge](./merge.md)
- [work Commands](../categories/work.md)
