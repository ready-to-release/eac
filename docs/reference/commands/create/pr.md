# create pr

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac create pr`
**Purpose**: Create pull request with AI-generated description
**Category**: [create](../categories/create.md)

## Syntax

```bash
r2r eac create pr [--title <title>] [--body <body>]
```

## Options

| Flag      | Description       |
| --------- | ----------------- |
| `--title` | Override PR title |
| `--body`  | Override PR body  |

## Examples

```bash
# Create PR with AI description
r2r eac create pr

# Manual title
r2r eac create pr --title "Add authentication feature"

# Check branch commits first
git log main..HEAD --oneline
r2r eac create pr
```

## See Also

- [create squash-message](./squash-message.md)
- [work merge](../work/merge.md)
- [create Commands](../categories/create.md)

{{ diataxis_footer() }}
