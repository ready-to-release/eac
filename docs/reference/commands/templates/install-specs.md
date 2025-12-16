# templates install-specs

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r templates install specs`
**Purpose**: Install specification templates without value replacements
**Category**: [templates](../categories/templates.md)

## Syntax

```bash
r2r templates install specs [--debug]
```

## Options

- `--debug, -d`: Save detailed logs to `out/logs/templates/install/`

## Examples

```bash
# Install specification templates
r2r templates install specs

# Install with debug logging
r2r templates install specs --debug
```

## Details

Installs specification template files from `templates/specs/` to `specs/risk-controls/`.

**Source**: `templates/specs/` (fixed)
**Destination**: `specs/risk-controls/` (fixed)

## Use Case

Install specification templates (risk control specifications, compliance test scenarios) once to your project, then customize them as needed.

These templates provide starter Gherkin specifications for:

- Security control validation
- Compliance testing scenarios
- Risk control implementation tests
- Regulatory requirement verification

## See Also

- [templates install](./index.md) - Install command overview
- [templates install-docs](./install-docs.md) - Install documentation templates
- [create spec](../create/spec.md) - AI specification generation
- [validate specs](../validate/specs.md) - Validate Gherkin specifications
- [test](../test/test.md) - Run BDD tests
- [templates Commands](../categories/templates.md)

{{ diataxis_footer() }}
