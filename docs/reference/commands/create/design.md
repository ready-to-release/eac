# create design

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac create design <module>`
**Purpose**: Generate workspace.dsl for a module using AI
**Category**: [create](../categories/create.md)

## Syntax

```bash
r2r eac create design <module> [options]
```

## Options and Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--prompt` | | Custom prompt file path (overrides team and system defaults) | System/team default |
| `--skip-validation` | | Skip Structurizr CLI validation | `false` |
| `--debug` | `-d` | Save intermediate outputs for debugging | `false` |

## Examples

```bash
# Generate architecture diagram
r2r eac create design src-auth

# Use custom prompt
r2r eac create design src-auth --prompt /path/to/custom-prompt.md

# Skip validation for faster iteration
r2r eac create design src-auth --skip-validation

# Debug generation process
r2r eac create design src-auth --debug

# Validate generated design
r2r eac validate design src-auth

# View in browser
r2r eac serve design
```

## Custom Prompts

The design generation supports **three-tier prompt system** for customization:

1. **Command Flag**: `--prompt /path/to/custom.md` (highest priority)
2. **Team Override**: `.r2r/eac/templates/ai/design/design.md` (team-wide customization)
3. **System Default**: `templates/ai/design/design.md` (fallback)

See [commit-message](./commit-message.md#custom-prompts) for detailed customization guide or:
```bash
cat .r2r/eac/templates/ai/README.md
```

## See Also

- [update design](../update/design.md)
- [validate design](../validate/design.md)
- [serve design](../serve/design.md)
- [create Commands](../categories/create.md)

{{ diataxis_footer() }}
