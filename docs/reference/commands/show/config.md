# show config

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show config`
**Purpose**: Display all loaded configurations with defaults applied
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show config
```

## Examples

```bash
# Show full configuration
r2r eac show config

# Check AI provider
r2r eac show config | grep "Provider"

# Verify repository settings
r2r eac show config | grep -A 3 "Repository"
```

## See Also

- [get config](../get/config.md) - JSON output
- [init](../other/init.md) - Configure AI provider
- [Configuration Guide](../../../how-to-guides/eac/configuration/index.md)

{{ diataxis_footer() }}
