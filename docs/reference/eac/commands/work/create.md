# Work create

<!-- book:cmd work create -->

Creates a new git worktree in a sibling directory for parallel development with separate Claude Code sessions.

The workspace directory follows the naming pattern `<repo-name>-<branch-name>` and is created as a sibling of the current repository root.

## Usage

```bash
eac work create <branch-name> [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `branch-name` | Name of the new branch to create (required) |

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--from` | `main` | Base branch to create from |
| `--path` | | Custom path for workspace (default: `../<repo>-<branch>`) |
| `--debug` | `false` | Enable debug logging |

## Examples

```bash
# Create workspace for a feature branch
eac work create feature/authentication

# Create from a specific base branch
eac work create bugfix/issue-123 --from=develop

# Create at a custom location
eac work create feature/api --path=../custom-path
```

## See Also

- [work commit](./commit.md) - Commit with AI messages
- [work remove](./remove.md) - Remove workspace
- [show workspaces](../show/workspaces.md)
- [work Commands](../../categories/work.md)
