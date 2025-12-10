# work commit

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac work commit [--all] [--message <msg>]`
**Purpose**: Commit changes with AI-generated commit messages
**Category**: [work](../categories/work.md)

## Syntax

```bash
r2r eac work commit [--all] [--message <msg>]
```

## Options

| Flag | Description |
|------|-------------|
| `--all` | Stage all changes before committing |
| `--message` | Use manual message instead of AI-generated |

## Examples

```bash
# Commit staged changes (AI message)
r2r eac work commit

# Stage all and commit
r2r eac work commit --all

# Manual message
r2r eac work commit --all --message "fix: resolve auth bug"
```

## See Also

- [create commit-message](../create/commit-message.md) - Generate messages
- [work create](./create.md)
- [work Commands](../categories/work.md)

{{ diataxis_footer() }}
