# create spec

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac create spec <module> <description>`
**Purpose**: Generate Gherkin specifications from natural language descriptions
**Category**: [create](../categories/create.md)

## Syntax

```bash
r2r eac create spec <description> [options]
```

## Options and Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--module` | `-m` | Target module for the specification | Auto-detected |
| `--output` | `-o` | Output file path | Auto-generated |
| `--prompt` | | Custom prompt file path (overrides team and system defaults) | System/team default |
| `--debug` | `-d` | Save intermediate outputs for debugging | `false` |

## Examples

```bash
# Generate spec from description (auto-detect module)
r2r eac create spec "User can login with valid credentials"

# Specify target module
r2r eac create spec "User can login" --module src-auth

# Specify output location
r2r eac create spec "API validates input" --module src-api --output specs/src-api/validation.feature

# Use custom prompt
r2r eac create spec "Feature description" --prompt /path/to/custom-prompt.md

# Debug generation process
r2r eac create spec "Feature description" --debug
```

## Custom Prompts

The spec generation supports **three-tier prompt system** for customization:

1. **Command Flag**: `--prompt /path/to/custom.md` (highest priority)
2. **Team Override**: `.r2r/eac/templates/ai/specs/specs.md` (team-wide customization)
3. **System Default**: `templates/ai/specs/specs.md` (fallback)

See [commit-message](./commit-message.md#custom-prompts) for detailed customization guide or:
```bash
cat .r2r/eac/templates/ai/README.md
```

## See Also

- [validate specs](../validate/specs.md)
- [test](../test/test.md)
- [create Commands](../categories/create.md)

{{ diataxis_footer() }}
