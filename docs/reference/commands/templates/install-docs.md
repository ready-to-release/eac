# templates install-docs

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r templates install docs`
**Purpose**: Install documentation templates without value replacements
**Category**: [templates](../categories/templates.md)

## Syntax

```bash
r2r templates install docs [--destination <path>] [--debug]
```

## Options

- `--destination <path>`: Output directory (default: `docs/reference/`)
- `--debug, -d`: Save detailed logs to `out/logs/templates/install/`

## Examples

```bash
# Install documentation templates to default location
r2r templates install docs

# Install to custom location
r2r templates install docs --destination ./custom-docs

# Install with debug logging
r2r templates install docs --debug
```

## Details

Installs documentation template files from `templates/docs/` to `docs/reference/`.

**Source**: `templates/docs/` (fixed)
**Default Destination**: `docs/reference/`
**Custom Destination**: Supported via `--destination` flag

## Use Case

Install documentation templates (architecture docs, implementation plans, operations guides) once to your project, then customize them as needed.

## See Also

- [templates install](./index.md) - Install command overview
- [templates install-ai](./install-ai.md) - Install AI prompt templates
- [create design](../create/design.md) - Generate architecture diagrams
- [templates Commands](../categories/templates.md)

{{ diataxis_footer() }}
