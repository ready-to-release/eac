# templates install-ai

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r templates install ai`
**Purpose**: Install AI prompt templates without value replacements
**Category**: [templates](../categories/templates.md)

## Syntax

```bash
r2r templates install ai [--debug]
```

## Options

- `--debug, -d`: Save detailed logs to `out/logs/templates/install/`

## Examples

```bash
# Install AI prompt templates
r2r templates install ai

# Install with debug logging
r2r templates install ai --debug
```

## Details

Installs AI prompt template files from `templates/ai/` to `.r2r/eac/templates/ai/`.

**Source**: `templates/ai/` (fixed)
**Destination**: `.r2r/eac/templates/ai/` (fixed)

## Use Case

Install AI prompt templates (commit messages, design generation, risk assessments, specification generation) once to your project, then customize them as needed.

These templates are used by AI-powered commands like:

- `create commit-message` - Generate commit messages
- `create design` - Generate architecture diagrams
- `create risk-assess` - Generate risk assessments
- `create spec` - Generate Gherkin specifications

## See Also

- [templates install](./index.md) - Install command overview
- [templates install-docs](./install-docs.md) - Install documentation templates
- [create commit-message](../create/commit-message.md) - AI commit messages
- [create design](../create/design.md) - AI architecture design
- [create spec](../create/spec.md) - AI specification generation
- [templates Commands](../categories/templates.md)

{{ diataxis_footer() }}
