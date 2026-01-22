# work Commands

Workspace management using git worktrees for parallel development.

## Commands in this Category

| Command                    | Purpose                                       |
| -------------------------- | --------------------------------------------- |
| [work](./work.md)          | Base work command                             |
| [work commit](./commit.md) | Commit changes with AI-generated messages     |
| [work create](./create.md) | Create new workspace for parallel development |
| [work merge](./merge.md)   | Merge workspace changes back to main          |
| [work pull](./pull.md)     | Sync workspace with latest main               |
| [work remove](./remove.md) | Remove workspace and associated branches      |

## Quick Examples

```bash
# Create workspace
r2r eac work create feature-auth

# Commit with AI message
r2r eac work commit

# Merge back to main
r2r eac work merge
```

## See Also

- [Category Overview](../categories/work.md)
- [create commit-message](../create/commit-message.md)
