# create risk-profile

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac create risk-profile <assessment-file>`
**Purpose**: Create OSCAL profile from risk assessment using AI
**Category**: [create](../categories/create.md)

## Syntax

```bash
r2r eac create risk-profile <assessment-file> [options]
```

## Options and Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--prompt` | | Custom prompt file path (overrides team and system defaults) | System/team default |
| `--debug` | `-d` | Save intermediate outputs for debugging | `false` |

## Examples

```bash
# Generate security profile from risk assessment
r2r eac create risk-profile .r2r/risk/assessment.md

# Use custom prompt for profile generation
r2r eac create risk-profile risk-assessment.md --prompt /path/to/custom-prompt.md

# Debug generation process
r2r eac create risk-profile risk-assessment.md --debug

# Validate generated profile
r2r eac validate risk-profile
```

## Custom Prompts

The risk profile generation supports **three-tier prompt system** for customization:

1. **Command Flag**: `--prompt /path/to/custom.md` (highest priority)
2. **Team Override**: `.r2r/eac/templates/ai/risk/profile.md` (team-wide customization)
3. **System Default**: `templates/ai/risk/profile.md` (fallback)

See [commit-message](./commit-message.md#custom-prompts) for detailed customization guide or:
```bash
cat .r2r/eac/templates/ai/README.md
```

## See Also

- [create risk-assess](./risk-assess.md)
- [validate risk-profile](../validate/risk-profile.md)
- [scan](../categories/scan.md)
- [create Commands](../categories/create.md)

{{ diataxis_footer() }}
