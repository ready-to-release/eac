# Work commit

<!-- book:cmd work commit -->

Commits changes in the current workspace using AI to generate semantic commit messages that follow project conventions.

By default, commits only staged changes. Use `--all` to stage all changes before committing. If `--message` is provided, the AI generation step is skipped.

## Usage

```bash
eac work commit [flags]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--all` | `-a` | Stage all changes before committing |
| `--message` | `-m` | Custom commit message (skips AI generation) |
| `--debug` | `-d` | Enable debug mode |

## Examples

```bash
# Commit staged changes with AI-generated message
eac work commit

# Stage and commit all changes
eac work commit --all

# Commit with a custom message
eac work commit -m "fix: resolve authentication bug"
```

## See Also

- [get commit-message](../get/commit-message.md) - Generate messages
- [work create](./create.md)
- [work Commands](../../categories/work.md)
