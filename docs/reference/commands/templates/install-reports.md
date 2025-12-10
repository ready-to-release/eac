# templates install-reports

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac templates install-reports`
**Purpose**: Install report templates without value replacements
**Category**: [templates](../categories/templates.md)

## Syntax

```bash
r2r eac templates install-reports
```

## Examples

```bash
# Install report templates
r2r eac templates install-reports
```

## Use Case

Installs report template files (test reports, build summaries, security scan reports) into the project without applying variable substitutions. Templates are installed as-is for later customization.

## See Also

- [templates install](./install.md) - Install all templates
- [templates apply](./apply.md) - Apply with substitutions
- [show build-summary](../show/build-summary.md)
- [show test-summary](../show/test-summary.md)
- [templates Commands](../categories/templates.md)

{{ diataxis_footer() }}
