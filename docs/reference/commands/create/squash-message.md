# create squash-message

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac create squash-message [branch]`
**Purpose**: Generate squash commit message from branch commits
**Category**: [create](../categories/create.md)

## Syntax

```bash
r2r eac create squash-message [branch] [options]
```

## Options and Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--prompt` | | Custom prompt file path (overrides team and system defaults) | System/team default |
| `--debug` | `-d` | Save intermediate outputs for debugging | `false` |

## Examples

```bash
# Generate from current branch
r2r eac create squash-message

# Generate from specific branch
r2r eac create squash-message feature/auth

# Use in merge workflow
git checkout main
git merge --squash feature/auth
r2r eac create squash-message feature/auth > .git/SQUASH_MSG
git commit

# Use custom prompt
r2r eac create squash-message --prompt /path/to/custom-prompt.md

# Debug generation process
r2r eac create squash-message --debug
```

## Custom Prompts

The squash message generation supports **three-tier prompt system** for customization:

1. **Command Flag**: `--prompt /path/to/custom.md` (highest priority)
2. **Team Override**: `.r2r/eac/templates/ai/commit/squash.md` (team-wide customization)
3. **System Default**: `templates/ai/commit/squash.md` (fallback)

See [commit-message](./commit-message.md#custom-prompts) for detailed customization guide or:
```bash
cat .r2r/eac/templates/ai/README.md
```

## See Also

- [work merge](../work/merge.md)
- [create pr](./pr.md)
- [create commit-message](./commit-message.md)
- [create Commands](../categories/create.md)

{{ diataxis_footer() }}
