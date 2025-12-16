# templates install-reports

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r templates install reports`
**Purpose**: Install report templates without value replacements
**Category**: [templates](../categories/templates.md)

## Syntax

```bash
r2r templates install reports [--debug]
```

## Options

- `--debug, -d`: Save detailed logs to `out/logs/templates/install/`

## Examples

```bash
# Install report templates
r2r templates install reports

# Install with debug logging
r2r templates install reports --debug
```

## Details

Installs report template files from `templates/reports/` to `.r2r/templates/reports/`.

**Source**: `templates/reports/` (fixed)
**Destination**: `.r2r/templates/reports/` (fixed)

## Use Case

Install report templates (test reports, build summaries, security scan reports) once to your project, then customize them as needed.

## See Also

- [templates install](./index.md) - Install command overview
- [templates install-docs](./install-docs.md) - Install documentation templates
- [show build-summary](../show/build-summary.md)
- [show test-summary](../show/test-summary.md)
- [templates Commands](../categories/templates.md)

{{ diataxis_footer() }}
