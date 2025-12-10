# work merge

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac work merge [--no-squash]`
**Purpose**: Merge workspace changes back to main (squash by default)
**Category**: [work](../categories/work.md)

## Syntax

```bash
r2r eac work merge [--no-squash]
```

## Options

| Flag | Description |
|------|-------------|
| `--no-squash` | Preserve commits instead of squashing |

## Examples

```bash
# Squash merge (default)
r2r eac work merge

# No squash (preserve commits)
r2r eac work merge --no-squash

# Full workflow
r2r eac work commit --all
r2r eac work pull
r2r eac test
r2r eac work merge
```

## See Also

- [create squash-message](../create/squash-message.md)
- [work remove](./remove.md)
- [work Commands](../categories/work.md)

{{ diataxis_footer() }}
